package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// 프로토콜 스킴 상수
const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
	SchemeFile  = "file"
	SchemeData  = "data"
)

// 기본 포트 번호
const (
	DefaultHTTPPort  = 80
	DefaultHTTPSPort = 443
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

// URL 구분자
const (
	SchemeDelimiter = "://"
	PathDelimiter   = "/"
	PortDelimiter   = ":"
)

type Fetcher interface {
	Fetch(u *URL) (string, error)
}

// URL 구조체: 주소 정보를 담는 바구니입니다.
type URL struct {
	Scheme string // http 같은 프로토콜
	Host   string // 주소 (example.com)
	Port   int
	Path   string // 경로 (/index.html)
}

// NewURL NewURL: 주소 문자열을 분석해서 URL 구조체를 만들어주는 함수입니다.
func NewURL(urlStr string) (*URL, error) {
	if strings.HasPrefix(urlStr, SchemeData+PortDelimiter) {
		return &URL{
			Scheme: SchemeData,
			Host:   "",
			Port:   0,
			Path:   urlStr[5:],
		}, nil
	}
	// 1. "://"를 기준으로 프로토콜(Scheme)을 분리합니다.
	// SplitN(문자열, 구분자, 개수) -> 최대 2개로 나눕니다.
	parts := strings.SplitN(urlStr, SchemeDelimiter, 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("주소 형식이 잘못되었습니다 (:// 없음)")
	}
	scheme := parts[0]

	if scheme != SchemeHTTP && scheme != SchemeHTTPS && scheme != SchemeFile {
		return nil, fmt.Errorf("http, https, file 또는 data 프로토콜만 지원합니다")
	}

	rest := parts[1]

	// 2. host와 path 분리
	host, path := parseHostPath(scheme, rest)

	// 3. 포트 파싱
	var port int
	var err error
	host, port, err = parsePort(scheme, host)
	if err != nil {
		return nil, err
	}

	// 4. 완성된 결과물을 돌려줍니다.
	return &URL{
		Scheme: scheme,
		Host:   host,
		Port:   port,
		Path:   path,
	}, nil
}

// parsePort: scheme과 host를 받아서 포트 번호를 파싱하고 클린한 호스트를 반환합니다.
// file 스킴의 경우 포트 파싱을 하지 않고 0을 반환합니다.
// http/https 스킴의 경우:
//   - host에 포트가 명시되어 있으면 파싱해서 반환
//   - 포트가 없으면 scheme에 따라 기본 포트 반환 (http: 80, https: 443)
//
// 반환값:
//   - cleanHost: 포트 번호가 제거된 호스트 이름
//   - port: 파싱된 포트 번호 또는 기본 포트
//   - err: 포트 파싱 실패 시 에러
func parsePort(scheme, host string) (cleanHost string, port int, err error) {
	// file 스킴은 포트가 없음
	if scheme == SchemeFile {
		return host, 0, nil
	}

	// host에 포트가 명시되어 있는지 확인
	if strings.Contains(host, PortDelimiter) {
		// host:port 형식 파싱
		parts := strings.SplitN(host, PortDelimiter, 2)
		cleanHost = parts[0]

		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("포트 번호가 올바르지 않습니다: %s", parts[1])
		}

		return cleanHost, port, nil
	}

	// 포트가 명시되지 않은 경우: scheme에 따라 기본 포트 사용
	if scheme == SchemeHTTPS {
		return host, DefaultHTTPSPort, nil
	}

	return host, DefaultHTTPPort, nil
}

// parseHostPath: scheme과 rest(스킴 이후의 문자열)를 받아서 host와 path를 분리합니다.
// file 스킴의 경우: rest 전체를 path로 사용하고 host는 빈 문자열
// http/https 스킴의 경우: "/" 기준으로 host와 path를 분리
//
// 반환값:
//   - host: 호스트 이름 (file 스킴의 경우 빈 문자열)
//   - path: 경로 (http/https는 "/" 시작, file은 rest 전체)
func parseHostPath(scheme, rest string) (host, path string) {
	// file 스킴: rest 전체가 경로
	if scheme == SchemeFile {
		return "", rest
	}

	// http/https 스킴: "/" 기준으로 host와 path 분리
	if strings.Contains(rest, PathDelimiter) {
		// "example.com/index.html" → host="example.com", path="/index.html"
		parts := strings.SplitN(rest, PathDelimiter, 2)
		return parts[0], PathDelimiter + parts[1]
	}

	// 경로가 없는 경우: "example.com" → host="example.com", path="/"
	return rest, PathDelimiter
}

type FileFetcher struct{}
type DataFetcher struct{}
type HTTPFetcher struct{}

var fetcherRegistry = map[string]Fetcher{
	SchemeFile:  &FileFetcher{},
	SchemeData:  &DataFetcher{},
	SchemeHTTP:  &HTTPFetcher{},
	SchemeHTTPS: &HTTPFetcher{},
}

// Request Request: 실제로 서버에 연결해서 데이터를 가져오는 메서드입니다.
func (u *URL) Request() (string, error) {
	fetcher, ok := fetcherRegistry[u.Scheme]
	if !ok {
		return "", fmt.Errorf("지원하지 않는 프로토콜: %s", u.Scheme)
	}
	return fetcher.Fetch(u)
}

func (f *FileFetcher) Fetch(u *URL) (string, error) {
	filePath := u.Path

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

	// 2. HTTP 요청 메시지 만들기
	// (기존 HTTP 요청 코드 그대로 유지)
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

	// 3. 서버에 메시지 보내기
	_, err = conn.Write([]byte(request))
	if err != nil {
		return "", err
	}

	// 4. 서버의 대답(응답) 읽기
	fmt.Printf("--- [%s:%d] 연결 및 요청 완료 ---\n", u.Host, u.Port)

	reader := bufio.NewReader(conn)

	// Status Line 읽기
	_, err = reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Headers 건너뛰기
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	// Body 읽기
	bodyBytes, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(bodyBytes), nil
}

func parseHTML(body string) string {
	// 태그를 제거하고 텍스트만 추출
	inTag := false
	var textBuilder strings.Builder

	for _, c := range body {
		if c == '<' {
			inTag = true
		} else if c == '>' {
			inTag = false
		} else if !inTag {
			// 태그 안이 아닐 때만 텍스트 수집
			textBuilder.WriteRune(c)
		}
	}

	text := html.UnescapeString(textBuilder.String())

	return text
}

func show(body string) {
	fmt.Print(parseHTML(body))
}

func load(urlStr string) {
	urlObj, err := NewURL(urlStr)
	if err != nil {
		fmt.Println("분석 에러: ", err)
		return
	}

	body, err := urlObj.Request()
	if err != nil {
		fmt.Println("요청 에러: ", err)
		return
	}

	show(body)
}

func main() {
	var urlStr string

	if len(os.Args) < 2 {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("현재 디렉토리를 가져올 수 없습니다: ", err)
		}

		urlStr = fmt.Sprintf("file:///%s/index.html", strings.ReplaceAll(cwd, "\\", "/"))
		fmt.Printf("기본 파일 열기: %s\n", urlStr)
	} else {
		urlStr = os.Args[1]
	}

	load(urlStr)
}
