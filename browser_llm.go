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

// URL_llm 구조체: 주소 정보를 담는 바구니입니다.
type URL_llm struct {
	Scheme string // http 같은 프로토콜
	Host   string // 주소 (example.com)
	Port   int    // 포트 번호 (80, 443 등)
	Path   string // 경로 (/index.html)
}

// NewURL_llm: 주소 문자열을 분석해서 URL_llm 구조체를 만들어주는 함수입니다.
func NewURL_llm(urlStr string) (*URL_llm, error) {
	// data 스킴은 특별하게 처리 (data:text/html,... 형식으로 :// 없음)
	if strings.HasPrefix(urlStr, "data:") {
		// data: 이후 전체를 path로 저장
		return &URL_llm{
			Scheme: "data",
			Host:   "",
			Port:   0,
			Path:   urlStr[5:], // "data:" 이후 부분
		}, nil
	}

	// 1. "://"를 기준으로 프로토콜(Scheme)을 분리합니다.
	parts := strings.SplitN(urlStr, "://", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("주소 형식이 잘못되었습니다 (:// 없음)")
	}
	scheme := parts[0]

	// http, https, file 스킴 지원
	if scheme != "http" && scheme != "https" && scheme != "file" {
		return nil, fmt.Errorf("http, https 또는 file 프로토콜만 지원합니다")
	}

	// 2. 스킴에 따라 다르게 파싱
	rest := parts[1]
	var host, path string
	var port int

	if scheme == "file" {
		// file:// 스킴의 경우
		// file:///C:/path/to/file → rest = "/C:/path/to/file"
		// file:///home/user/file → rest = "/home/user/file"
		// file://./relative → rest = "./relative"
		// file://test.html → rest = "test.html"

		host = "" // file 스킴은 호스트 없음
		port = 0  // file 스킴은 포트 없음

		// 경로는 rest를 그대로 사용
		// - 절대 경로: /C:/path 또는 /home/user/file
		// - 상대 경로: test.html 또는 ./test.html
		path = rest
	} else {
		// http, https 스킴의 경우
		if strings.Contains(rest, "/") {
			hostPath := strings.SplitN(rest, "/", 2)
			host = hostPath[0]
			path = "/" + hostPath[1]
		} else {
			host = rest
			path = "/"
		}

		// 3. 포트 번호 파싱
		if strings.Contains(host, ":") {
			hostPort := strings.SplitN(host, ":", 2)
			host = hostPort[0]

			var err error
			port, err = strconv.Atoi(hostPort[1])
			if err != nil {
				return nil, fmt.Errorf("포트 번호가 올바르지 않습니다: %s", hostPort[1])
			}
		} else {
			// 포트가 명시되지 않은 경우 기본 포트 사용
			if scheme == "https" {
				port = 443
			} else {
				port = 80
			}
		}
	}

	// 4. 완성된 결과물을 돌려줍니다.
	return &URL_llm{
		Scheme: scheme,
		Host:   host,
		Port:   port,
		Path:   path,
	}, nil
}

// Request_llm: 실제로 서버에 연결해서 데이터를 가져오거나 파일을 읽는 메서드입니다.
func (u *URL_llm) Request_llm() (string, error) {
	// file:// 스킴의 경우 로컬 파일 읽기
	if u.Scheme == "file" {
		return u.requestFile()
	}

	// data:// 스킴의 경우 URL에 담긴 데이터 직접 파싱
	if u.Scheme == "data" {
		return u.requestData()
	}

	// http, https 스킴의 경우 네트워크 요청
	return u.requestHTTP()
}

// requestFile: 로컬 파일을 읽는 헬퍼 메서드
func (u *URL_llm) requestFile() (string, error) {
	filePath := u.Path

	// Windows 절대 경로 처리: /C:/path → C:/path
	// file:///C:/Users/... 형식을 C:/Users/... 로 변환
	if len(filePath) > 2 && filePath[0] == '/' && filePath[2] == ':' {
		filePath = filePath[1:] // 앞의 / 제거
	}

	// 파일 읽기
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("파일 읽기 실패: %v", err)
	}

	fmt.Printf("--- [파일] %s 읽기 완료 ---\n", filePath)
	return string(content), nil
}

// requestData: data 스킴의 데이터를 파싱하는 헬퍼 메서드
// data 스킴 형식: data:[<mediatype>][;base64],<data>
// 예: data:text/html,<h1>Hello</h1>
// 예: data:text/html;base64,PGgxPkhlbGxvPC9oMT4=
func (u *URL_llm) requestData() (string, error) {
	dataStr := u.Path

	// ","를 기준으로 메타데이터와 실제 데이터를 분리
	commaIdx := strings.Index(dataStr, ",")
	if commaIdx == -1 {
		return "", fmt.Errorf("data 스킴 형식이 잘못되었습니다 (쉼표 없음)")
	}

	metadata := dataStr[:commaIdx] // 예: "text/html" 또는 "text/html;base64"
	data := dataStr[commaIdx+1:]   // 예: "<h1>Hello</h1>" 또는 base64 인코딩된 문자열

	// base64 인코딩 확인
	if strings.Contains(metadata, ";base64") {
		// base64 디코딩
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return "", fmt.Errorf("base64 디코딩 실패: %v", err)
		}
		data = string(decoded)
		fmt.Printf("--- [data] base64 디코딩 완료 ---\n")
	} else {
		// URL 인코딩된 문자열 디코딩 (예: %20 → 공백)
		decoded, err := url.QueryUnescape(data)
		if err != nil {
			// 디코딩 실패 시 원본 그대로 사용
			decoded = data
		}
		data = decoded
		fmt.Printf("--- [data] URL 파싱 완료 ---\n")
	}

	return data, nil
}

// requestHTTP: HTTP/HTTPS 요청을 수행하는 헬퍼 메서드
func (u *URL_llm) requestHTTP() (string, error) {
	// 1. 서버에 연결하기
	var conn net.Conn
	var err error

	// Port 필드를 사용하여 주소 구성
	address := fmt.Sprintf("%s:%d", u.Host, u.Port)

	if u.Scheme == "https" {
		// HTTPS: TLS 암호화 연결
		conn, err = tls.Dial("tcp", address, nil)
	} else {
		// HTTP: 일반 TCP 연결
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return "", err
	}

	// defer에서 Close() 에러 처리
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// 연결 종료 에러는 일반적으로 무시해도 되지만
			// 디버깅을 위해 출력할 수 있음
			// fmt.Printf("연결 종료 에러: %v\n", closeErr)
		}
	}()

	// 2. HTTP 요청 메시지 만들기
	// HTTP/1.1 사용 및 헤더를 맵으로 관리하여 확장 가능하게 구성
	headers := map[string]string{
		"Host":       u.Host,
		"Connection": "close",
		"User-Agent": "GoWebBrowser/1.0",
	}

	// Request Line 구성: GET /path HTTP/1.1
	requestLine := fmt.Sprintf("GET %s HTTP/1.1\r\n", u.Path)

	// 헤더들을 문자열로 조합
	var headerLines strings.Builder
	headerLines.WriteString(requestLine)
	for key, value := range headers {
		headerLines.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	// 헤더와 본문 사이의 빈 줄
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

	// Status Line 읽기 (첫 줄 읽어서 넘기기)
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

// show_llm: HTML 태그를 제거하고 텍스트만 출력하는 함수
// 파이썬의 show 함수를 Go로 변환한 버전입니다.
//
// 파이썬 원본:
// def show(body):
//
//	in_tag = False
//	for c in body:
//	    if c == "<":
//	        in_tag = True
//	    elif c == ">":
//	        in_tag = False
//	    elif not in_tag:
//	        print(c, end="")
func show_llm(body string) {
	// 태그 안에 있는지 추적하는 플래그
	inTag := false

	// range를 사용해서 문자열의 각 문자(rune)를 순회
	// _ 는 인덱스 (사용하지 않으므로 무시)
	// c 는 rune 타입 (Go의 유니코드 문자 타입, int32의 별칭)
	for _, c := range body {
		if c == '<' {
			// '<' 를 만나면 태그 시작
			inTag = true
		} else if c == '>' {
			// '>' 를 만나면 태그 종료
			inTag = false
		} else if !inTag {
			// 태그 안이 아닐 때만 출력
			// rune을 string으로 변환 필요
			fmt.Print(string(c))
		}
	}
}

// load_llm: URL 객체를 받아서 요청하고 화면에 표시하는 통합 함수
// 파이썬의 load 함수를 Go로 변환한 버전입니다.
//
// 파이썬 원본:
// def load(url):
//
//	body = url.request()
//	show(body)
func load_llm(urlObj *URL_llm) {
	// 1. URL 객체의 Request_llm 메서드 호출해서 HTML 가져오기
	body, err := urlObj.Request_llm()
	if err != nil {
		fmt.Println("요청 에러:", err)
		return
	}

	// 2. show_llm 함수로 HTML 태그 제거하고 텍스트만 출력
	show_llm(body)
}

func main() {
	var urlStr string

	// 인자가 없으면 테스트 모드로 실행
	if len(os.Args) < 2 {
		fmt.Println("=== data 스킴 테스트 모드 ===\n")

		// 테스트할 data URL 목록
		testURLs := []string{
			"data:text/html,Hello world!",
			"data:text/html,<h1>Hello</h1>",
			"data:text/html,<h1>안녕하세요</h1><p>data 스킴 테스트</p>",
			"data:text/html;base64,PGgxPkhlbGxvPC9oMT4=", // <h1>Hello</h1>의 base64
		}

		for i, testURL := range testURLs {
			fmt.Printf("테스트 %d: %s\n", i+1, testURL)

			urlObj, err := NewURL_llm(testURL)
			if err != nil {
				fmt.Println("분석 에러:", err)
				fmt.Println()
				continue
			}

			body, err := urlObj.Request_llm()
			if err != nil {
				fmt.Println("요청 에러:", err)
				fmt.Println()
				continue
			}

			fmt.Print("결과: ")
			show_llm(body)
			fmt.Println("\n")
		}
		return
	} else {
		// 커맨드 라인 인자를 URL로 사용
		urlStr = os.Args[1]
	}

	// URL 파싱 및 로드
	urlObj, err := NewURL_llm(urlStr)
	if err != nil {
		fmt.Println("분석 에러:", err)
		return
	}

	load_llm(urlObj)
}
