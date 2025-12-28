package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

// HTTP 관련 상수
const (
	HTTPVersion = "HTTP/1.1"
	UserAgent   = "GoWebBrowser/1.0"
)

// HTTP 헤더 이름
const (
	HeaderHost       = "Host"
	HeaderConnection = "Connection"
	HeaderUserAgent  = "User-Agent"
)

// ConnectionClose HTTP 헤더 값
const (
	ConnectionClose = "close"
)

// Fetcher 인터페이스: URL에서 콘텐츠를 가져오는 역할을 추상화
type Fetcher interface {
	Fetch(u *URL) (string, error)
}

// FileFetcher: file:// 스킴을 처리하는 Fetcher 구현
type FileFetcher struct{}

// DataFetcher: data:// 스킴을 처리하는 Fetcher 구현
type DataFetcher struct{}

// HTTPFetcher: http://, https:// 스킴을 처리하는 Fetcher 구현
type HTTPFetcher struct{}

// ViewSourceFetcher: view-source:// 스킴을 처리하는 Fetcher 구현
type ViewSourceFetcher struct{}

// fetcherRegistry: scheme에 따른 Fetcher를 등록하는 레지스트리
var fetcherRegistry = map[Scheme]Fetcher{
	SchemeFile:       &FileFetcher{},
	SchemeData:       &DataFetcher{},
	SchemeHTTP:       &HTTPFetcher{},
	SchemeHTTPS:      &HTTPFetcher{},
	SchemeViewSource: &ViewSourceFetcher{},
}

// Request: URL에서 콘텐츠를 가져오는 메서드
func (u *URL) Request() (string, error) {
	fetcher, ok := fetcherRegistry[u.Scheme]
	if !ok {
		return "", fmt.Errorf("지원하지 않는 프로토콜: %s", u.Scheme)
	}
	return fetcher.Fetch(u)
}

// Fetch: FileFetcher의 Fetch 메서드 구현
func (f *FileFetcher) Fetch(u *URL) (string, error) {
	filePath := u.Path

	// Windows 절대 경로 처리: /C:/path → C:/path
	if len(filePath) > 2 && filePath[0] == '/' && filePath[2] == ':' {
		filePath = filePath[1:]
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("파일 읽기 실패: %v", err)
	}

	fmt.Printf("--- 파일 %s 읽기 완료 ---\n", filePath)
	return string(content), nil
}

// Fetch: DataFetcher의 Fetch 메서드 구현
func (d *DataFetcher) Fetch(u *URL) (string, error) {
	dataStr := u.Path

	commaIdx := strings.Index(dataStr, ",")
	if commaIdx == -1 {
		return "", fmt.Errorf("data 스킴 형식이 잘못되었습니다 (쉼표 없음)")
	}

	metadata := dataStr[:commaIdx]
	data := dataStr[commaIdx+1:]

	if strings.Contains(metadata, ";base64") {
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return "", fmt.Errorf("base64 디코딩 실패: %v", err)
		}
		data = string(decoded)
		fmt.Printf("--- [data] base64 디코딩 완료 ---\n")
	} else {
		decoded, err := url.QueryUnescape(data)
		if err != nil {
			decoded = data
		}
		data = decoded
		fmt.Println("--- [data] URL 파싱 완료 ---")
	}

	return data, nil
}

// Fetch: HTTPFetcher의 Fetch 메서드 구현
func (h *HTTPFetcher) Fetch(u *URL) (string, error) {
	var conn net.Conn
	var err error

	address := fmt.Sprintf("%s:%d", u.Host, u.Port)

	if u.Scheme == SchemeHTTPS {
		conn, err = tls.Dial("tcp", address, nil)
	} else {
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return "", err
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Printf("연결 종료 에러: %v\n", closeErr)
		}
	}()

	// HTTP 요청 메시지 만들기
	headers := map[string]string{
		HeaderHost:       u.Host,
		HeaderConnection: ConnectionClose,
		HeaderUserAgent:  UserAgent,
	}

	requestLine := fmt.Sprintf("GET %s %s\r\n", u.Path, HTTPVersion)

	var headerLines strings.Builder
	headerLines.WriteString(requestLine)
	for key, value := range headers {
		headerLines.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	headerLines.WriteString("\r\n")

	request := headerLines.String()

	// 서버에 메시지 보내기
	_, err = conn.Write([]byte(request))
	if err != nil {
		return "", err
	}

	// 서버의 대답(응답) 읽기
	fmt.Printf("--- [%s:%d] 연결 및 요청 완료 ---\n", u.Host, u.Port)

	return parseResponse(conn)
}

// parseResponse: 서버의 응답을 읽어 상태 라인, 헤더를 처리하고 바디를 반환합니다.
func parseResponse(r io.Reader) (string, error) {
	reader := bufio.NewReader(r)

	// 1. Status Line 읽기 (예: HTTP/1.1 200 OK)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("상태 라인 읽기 실패: %w", err)
	}
	_ = statusLine // 현재는 상태 코드를 검사하지 않지만, 나중에 확장을 위해 저장

	// 2. Headers 건너뛰기
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("헤더 읽기 실패: %w", err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	// 3. Body 읽기
	bodyBytes, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("바디 읽기 실패: %w", err)
	}

	return string(bodyBytes), nil
}

// Fetch: ViewSourceFetcher의 Fetch 메서드 구현
func (v *ViewSourceFetcher) Fetch(u *URL) (string, error) {
	// Path에는 내부 URL 전체가 들어있음 (예: "http://example.org/")
	innerURLStr := u.Path

	if innerURLStr == "" {
		return "", fmt.Errorf("view-source: 내부 URL이 없습니다")
	}

	// 내부 URL 파싱
	innerURL, err := NewURL(innerURLStr)
	if err != nil {
		return "", fmt.Errorf("view-source: 내부 URL 파싱 실패: %v", err)
	}

	// 내부 URL로 콘텐츠 가져오기 (원본 그대로 반환)
	content, err := innerURL.Request()
	if err != nil {
		return "", fmt.Errorf("view-source: 내부 URL 요청 실패: %v", err)
	}

	fmt.Println("--- [view-source] 원본 소스 반환 ---")
	return content, nil
}
