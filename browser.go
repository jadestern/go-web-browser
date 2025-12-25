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
	"strconv"
	"strings"
)

// URL 구조체: 주소 정보를 담는 바구니입니다.
type URL struct {
	Scheme string // http 같은 프로토콜
	Host   string // 주소 (example.com)
	Port   int
	Path   string // 경로 (/index.html)
}

// NewURL NewURL: 주소 문자열을 분석해서 URL 구조체를 만들어주는 함수입니다.
func NewURL(urlStr string) (*URL, error) {
	if strings.HasPrefix(urlStr, "data:") {
		return &URL{
			Scheme: "data",
			Host:   "",
			Port:   0,
			Path:   urlStr[5:],
		}, nil
	}
	// 1. "://"를 기준으로 프로토콜(Scheme)을 분리합니다.
	// SplitN(문자열, 구분자, 개수) -> 최대 2개로 나눕니다.
	parts := strings.SplitN(urlStr, "://", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("주소 형식이 잘못되었습니다 (:// 없음)")
	}
	scheme := parts[0]

	if scheme != "http" && scheme != "https" && scheme != "file" {
		return nil, fmt.Errorf("http, https, file 또는 data 프로토콜만 지원합니다")
	}

	rest := parts[1]
	var host, path string
	var port int

	if scheme == "file" {
		host = ""
		port = 0

		path = rest
	} else {
		// 슬래시(/)가 포함되어 있는지 확인합니다.
		if strings.Contains(rest, "/") {
			// naver.com/search 같은 경우 "/" 기준으로 나눕니다.
			hostPath := strings.SplitN(rest, "/", 2)
			host = hostPath[0]
			path = "/" + hostPath[1]
		} else {
			// naver.com 처럼 슬래시가 없는 경우 전체가 호스트이고 경로는 "/"입니다.
			host = rest
			path = "/"
		}
	}

	if strings.Contains(host, ":") {
		hostPort := strings.SplitN(host, ":", 2)
		host = hostPort[0]

		var err error
		port, err = strconv.Atoi(hostPort[1])
		if err != nil {
			return nil, fmt.Errorf("포트 번호가 올바르지 않습니다: %s", hostPort[1])
		}
	} else {
		if scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}

	// 3. 완성된 결과물을 돌려줍니다.
	return &URL{
		Scheme: scheme,
		Host:   host,
		Port:   port,
		Path:   path,
	}, nil
}

// Request Request: 실제로 서버에 연결해서 데이터를 가져오는 메서드입니다.
func (u *URL) Request() (string, error) {
	if u.Scheme == "file" {
		return u.requestFile()
	}

	if u.Scheme == "data" {
		return u.requestData()
	}

	return u.requestHTTP()
}

func (u *URL) requestData() (string, error) {
	dataStr := u.Path

	commaIdx := strings.Index(dataStr, ",")
	if commaIdx == -1 {
		return "", fmt.Errorf("data 스킴 형식이 잘못되었습니다 (쉼표 없음")
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

func (u *URL) requestFile() (string, error) {
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

func (u *URL) requestHTTP() (string, error) {
	var conn net.Conn
	var err error

	address := fmt.Sprintf("%s:%d", u.Host, u.Port)

	if u.Scheme == "https" {
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
		"Host":       u.Host,
		"Connection": "close",
		"User-Agent": "GoWebBrowser/1.0",
	}

	requestLine := fmt.Sprintf("GET %s HTTP/1.0\r\n", u.Path)

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

func show(body string) {
	inTag := false

	for _, c := range body {
		if c == '<' {
			inTag = true
		} else if c == '>' {
			inTag = false
		} else if !inTag {
			fmt.Print(string(c))
		}
	}
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
