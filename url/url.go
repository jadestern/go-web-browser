package url

import (
	"fmt"
	"strconv"
	"strings"
)

// Scheme 타입: URL 스킴을 타입 안전하게 표현
type Scheme string

// 프로토콜 스킴 상수
const (
	SchemeHTTP       Scheme = "http"
	SchemeHTTPS      Scheme = "https"
	SchemeFile       Scheme = "file"
	SchemeData       Scheme = "data"
	SchemeViewSource Scheme = "view-source"
)

// 기본 포트 번호
const (
	DefaultHTTPPort  = 80
	DefaultHTTPSPort = 443
)

// URL 구분자
const (
	SchemeDelimiter = "://"
	PathDelimiter   = "/"
	PortDelimiter   = ":"
)

// URL 구조체: 주소 정보를 담는 바구니입니다.
type URL struct {
	Scheme Scheme // http 같은 프로토콜 (타입 안전)
	Host   string // 주소 (example.com)
	Port   int
	Path   string // 경로 (/index.html)
}

// String: URL 객체를 문자열로 변환합니다. (fmt.Stringer 인터페이스 구현)
func (u *URL) String() string {
	if u.Scheme == SchemeData {
		return fmt.Sprintf("data:%s", u.Path)
	}
	if u.Scheme == SchemeViewSource {
		return fmt.Sprintf("view-source:%s", u.Path)
	}
	if u.Scheme == SchemeFile {
		return fmt.Sprintf("file://%s", u.Path)
	}

	// HTTP/HTTPS
	if (u.Scheme == SchemeHTTP && u.Port == DefaultHTTPPort) ||
		(u.Scheme == SchemeHTTPS && u.Port == DefaultHTTPSPort) {
		return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	}

	return fmt.Sprintf("%s://%s:%d%s", u.Scheme, u.Host, u.Port, u.Path)
}

// NewURL NewURL: 주소 문자열을 분석해서 URL 구조체를 만들어주는 함수입니다.
func NewURL(urlStr string) (*URL, error) {
	// view-source 스킴 특별 처리: view-source:http://example.org/
	if strings.HasPrefix(urlStr, string(SchemeViewSource)+PortDelimiter) {
		return &URL{
			Scheme: SchemeViewSource,
			Host:   "",
			Port:   0,
			Path:   urlStr[12:], // "view-source:" 길이는 12
		}, nil
	}

	// data 스킴 특별 처리: data:text/html,<html>
	if strings.HasPrefix(urlStr, string(SchemeData)+PortDelimiter) {
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
	scheme := Scheme(parts[0])

	if scheme != SchemeHTTP && scheme != SchemeHTTPS && scheme != SchemeFile {
		return nil, fmt.Errorf("지원하지 않는 프로토콜입니다: %s", scheme)
	}

	rest := parts[1]

	// 2. host와 path 분리
	host, path := parseHostPath(scheme, rest)

	// 3. 포트 파싱
	var port int
	var err error
	host, port, err = parsePort(scheme, host)
	if err != nil {
		return nil, fmt.Errorf("포트 파싱 실패: %w", err)
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
func parsePort(scheme Scheme, host string) (cleanHost string, port int, err error) {
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
			return "", 0, fmt.Errorf("포트 번호가 올바르지 않습니다 (%s): %w", parts[1], err)
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
func parseHostPath(scheme Scheme, rest string) (host, path string) {
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
