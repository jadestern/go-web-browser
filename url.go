package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
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
	// 1. "://"를 기준으로 프로토콜(Scheme)을 분리합니다.
	// SplitN(문자열, 구분자, 개수) -> 최대 2개로 나눕니다.
	parts := strings.SplitN(urlStr, "://", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("주소 형식이 잘못되었습니다 (:// 없음)")
	}
	scheme := parts[0]
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("http 또는 https 프로토콜만 지원합니다")
	}

	// 2. 호스트와 경로를 분리합니다.
	rest := parts[1]
	var host, path string

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

	var port int
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
	var conn net.Conn
	var err error

	// Port 필드를 사용하여 주소 구성
	address := fmt.Sprintf("%s:%d", u.Host, u.Port)

	if u.Scheme == "https" {
		conn, err = tls.Dial("tcp", address, nil)
	} else {
		conn, err = net.Dial("tcp", address)
	}
	if err != nil {
		return "", err
	}
	// 함수가 끝나기 직전에 반드시 연결을 닫으라고 예약해둡니다.
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Printf("연결 종료 에러: %v\n", closeErr)
		}
	}()

	// 2. HTTP 요청 메시지 만들기 (서버에 보낼 편지)
	// \r\n은 HTTP 규격에서 사용하는 줄바꿈입니다.
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
	// Write는 바이트([]byte) 형태로 보내야 하므로 타입 변환을 해줍니다.
	_, err = conn.Write([]byte(request))
	if err != nil {
		return "", err
	}

	// 4. 서버의 대답(응답) 읽기
	fmt.Printf("--- [%s] 연결 및 요청 완료 ---\n", u.Host)

	reader := bufio.NewReader(conn)

	// 2. Status Line 읽기 (첫 줄 읽어서 넘기기)
	_, err = reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// 3. Headers 건너뛰기
	// 빈 줄(\r\n)이 나올 때까지 계속 읽습니다.
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		// 헤더와 본문 사이의 빈 줄을 체크합니다.
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	// 4. 나머지 모든 데이터를 Body로 읽기
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
	if len(os.Args) < 2 {
		fmt.Println("사용법: ./url <URL>")
		fmt.Println("예시: ./url http://example.com")
		return
	}

	load(os.Args[1])
}
