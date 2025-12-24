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

// URL_llm 구조체: 주소 정보를 담는 바구니입니다.
type URL_llm struct {
	Scheme string // http 같은 프로토콜
	Host   string // 주소 (example.com)
	Port   int    // 포트 번호 (80, 443 등)
	Path   string // 경로 (/index.html)
}

// NewURL_llm: 주소 문자열을 분석해서 URL_llm 구조체를 만들어주는 함수입니다.
func NewURL_llm(urlStr string) (*URL_llm, error) {
	// 1. "://"를 기준으로 프로토콜(Scheme)을 분리합니다.
	parts := strings.SplitN(urlStr, "://", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("주소 형식이 잘못되었습니다 (:// 없음)")
	}
	scheme := parts[0]

	// http와 https 모두 지원
	// 파이썬의 ssl 모듈처럼 암호화된 연결 지원
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("http 또는 https 프로토콜만 지원합니다")
	}

	// 2. 호스트와 경로를 분리합니다.
	rest := parts[1]
	var host, path string

	if strings.Contains(rest, "/") {
		hostPath := strings.SplitN(rest, "/", 2)
		host = hostPath[0]
		path = "/" + hostPath[1]
	} else {
		host = rest
		path = "/"
	}

	// 3. 포트 번호 파싱 (Python 코드와 동일한 로직)
	// Python 원본:
	// if ":" in self.host:
	//     self.host, port = self.host.split(":", 1)
	//     self.port = int(port)
	var port int
	if strings.Contains(host, ":") {
		// host에서 포트 분리: "example.com:8080" -> ["example.com", "8080"]
		hostPort := strings.SplitN(host, ":", 2)
		host = hostPort[0]

		// 포트 문자열을 정수로 변환
		// Python의 int(port)와 동일
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

	// 4. 완성된 결과물을 돌려줍니다.
	return &URL_llm{
		Scheme: scheme,
		Host:   host,
		Port:   port,
		Path:   path,
	}, nil
}

// Request_llm: 실제로 서버에 연결해서 데이터를 가져오는 메서드입니다.
func (u *URL_llm) Request_llm() (string, error) {
	// 1. 서버에 연결하기
	// scheme에 따라 다른 연결 방식 사용
	// Port 필드를 사용하여 주소 구성
	// 파이썬: ctx.wrap_socket(s, server_hostname=host)
	// Go: tls.Dial() 또는 net.Dial()

	var conn net.Conn
	var err error

	// Port 필드를 사용하여 주소 구성
	address := fmt.Sprintf("%s:%d", u.Host, u.Port)

	if u.Scheme == "https" {
		// HTTPS: TLS 암호화 연결
		// tls.Dial은 자동으로 TLS 핸드셰이크 수행
		// nil = 기본 설정 사용 (안전한 기본값)
		conn, err = tls.Dial("tcp", address, nil)
	} else {
		// HTTP: 일반 TCP 연결
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return "", err
	}

	// defer에서 Close() 에러 처리
	// Go의 모범 사례: defer 함수에서 에러를 명시적으로 처리
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// 연결 종료 에러는 일반적으로 무시해도 되지만
			// 디버깅을 위해 출력할 수 있음
			// fmt.Printf("연결 종료 에러: %v\n", closeErr)
		}
	}()

	// 2. HTTP 요청 메시지 만들기
	request := fmt.Sprintf(
		"GET %s HTTP/1.0\r\n"+
			"Host: %s\r\n"+
			"\r\n",
		u.Path, u.Host,
	)

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
	// 커맨드 라인 인자 처리
	// 파이썬의 if __name__ == "__main__": 와 동일한 역할
	// Go에서는 main 함수가 항상 진입점이므로 조건문 불필요

	// os.Args는 커맨드 라인 인자를 담은 문자열 슬라이스
	// os.Args[0] = 프로그램 이름
	// os.Args[1] = 첫 번째 인자
	// os.Args[2] = 두 번째 인자...

	// 인자 개수 확인 (최소 2개 필요: 프로그램명 + URL)
	if len(os.Args) < 2 {
		fmt.Println("사용법: ./show_llm <URL>")
		fmt.Println("예시: ./show_llm http://example.com")
		return
	}

	// 첫 번째 커맨드 라인 인자를 URL로 사용
	// 파이썬: sys.argv[1]
	// Go: os.Args[1]
	urlStr := os.Args[1]

	// load_llm 함수로 한 번에 처리
	// 파이썬: load(URL(sys.argv[1]))
	// Go: load_llm을 사용하려면 URL 객체 필요
	urlObj, err := NewURL_llm(urlStr)
	if err != nil {
		fmt.Println("분석 에러:", err)
		return
	}

	load_llm(urlObj)
}
