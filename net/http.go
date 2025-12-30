// Package net implements HTTP networking for the browser.
// This file contains HTTP/HTTPS fetching logic with caching and Keep-Alive support.
package net

import (
	"crypto/tls"
	"fmt"
	"go-web-browser/logger"
	"go-web-browser/url"
	"net"
	"strings"
)

// HTTP protocol constants
const (
	HTTPVersion = "HTTP/1.1"
	UserAgent   = "GoWebBrowser/1.0"
)

// HTTP header names
const (
	HeaderHost       = "Host"
	HeaderConnection = "Connection"
	HeaderUserAgent  = "User-Agent"
)

// HTTP header values
const (
	ConnectionClose = "close"
)

// HTTPFetcher: http://, https:// 스킴을 처리하는 Fetcher 구현
type HTTPFetcher struct{}

// Fetch: HTTPFetcher의 Fetch 메서드 구현
func (h *HTTPFetcher) Fetch(u *url.URL) (string, error) {
	// 캐시에서 먼저 확인
	urlStr := u.String()
	if entry, found := GlobalCache.Get(urlStr); found {
		return entry.Body, nil
	}

	const maxRedirects = 10
	currentURL := u

	// 리다이렉트 루프: 최대 10번까지 리다이렉트를 따라감
	for i := 0; i < maxRedirects; i++ {
		statusCode, body, headers, err := h.doRequest(currentURL)
		if err != nil {
			return "", err
		}

		// 리다이렉트가 아니면 성공
		if statusCode < 300 || statusCode >= 400 {
			// 응답을 캐시에 저장한 후 반환
			GlobalCache.Put(urlStr, statusCode, body, headers)
			return body, nil
		}

		// 리다이렉트 처리 (300-399)
		location := headers["location"]
		if location == "" {
			return "", fmt.Errorf("리다이렉트 응답에 Location 헤더가 없습니다 (status %d)", statusCode)
		}

		logger.Logger.Printf("리다이렉트 %d: %d -> %s", i+1, statusCode, location)

		// Location을 절대 URL로 변환
		nextURL, err := resolveURL(currentURL, location)
		if err != nil {
			return "", fmt.Errorf("리다이렉트 URL 변환 실패 %q: %w", location, err)
		}

		currentURL = nextURL
	}

	return "", fmt.Errorf("최대 리다이렉트 횟수 초과 (최대 %d회)", maxRedirects)
}

// resolveURL resolves a potentially relative URL against a base URL.
//
// If location is an absolute URL (starts with http:// or https://), it is parsed directly.
// If location is a relative URL (starts with /), it uses the base URL's scheme and host.
//
// Examples:
//   - resolveURL("http://example.com/page", "https://other.com/new") -> "https://other.com/new"
//   - resolveURL("http://example.com/page", "/new") -> "http://example.com/new"
func resolveURL(base *url.URL, location string) (*url.URL, error) {
	// Absolute URL: parse directly
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		return url.NewURL(location)
	}

	// Relative URL: use base URL's scheme and host
	if strings.HasPrefix(location, "/") {
		// Construct absolute URL: scheme://host:port/path
		var absoluteURL string
		if base.Scheme == url.SchemeHTTPS && base.Port == url.DefaultHTTPSPort {
			// HTTPS default port: omit :443
			absoluteURL = fmt.Sprintf("https://%s%s", base.Host, location)
		} else if base.Scheme == url.SchemeHTTP && base.Port == url.DefaultHTTPPort {
			// HTTP default port: omit :80
			absoluteURL = fmt.Sprintf("http://%s%s", base.Host, location)
		} else {
			// Non-default port: include it
			absoluteURL = fmt.Sprintf("%s://%s:%d%s", base.Scheme, base.Host, base.Port, location)
		}
		return url.NewURL(absoluteURL)
	}

	return nil, fmt.Errorf("지원하지 않는 Location 형식: %q (절대 URL 또는 상대 경로가 아님)", location)
}

// doRequest performs a single HTTP request and returns status code, body, headers
func (h *HTTPFetcher) doRequest(u *url.URL) (int, string, map[string]string, error) {
	address := fmt.Sprintf("%s:%d", u.Host, u.Port)

	// 1. ConnectionPool에서 기존 연결 찾기
	conn, found := GlobalConnectionPool.Get(address)

	if !found {
		// 2. Create new connection if not in pool
		logger.Logger.Printf("Creating new connection to %s", address)
		var err error

		if u.Scheme == url.SchemeHTTPS {
			conn, err = tls.Dial("tcp", address, nil)
		} else {
			conn, err = net.Dial("tcp", address)
		}

		if err != nil {
			return 0, "", nil, err
		}
	}

	// HTTP 요청 메시지 만들기
	headers := map[string]string{
		HeaderHost: u.Host,
		// Connection: close 헤더 제거!
		// → HTTP/1.1의 기본 동작이 keep-alive이므로 생략
		HeaderUserAgent: UserAgent,
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
	_, err := conn.Write([]byte(request))
	if err != nil {
		conn.Close() // 전송 실패 시 연결 닫기
		return 0, "", nil, err
	}

	// Read and parse HTTP response
	logger.Logger.Printf("Request sent to %s:%d", u.Host, u.Port)

	statusCode, body, respHeaders, err := ParseResponse(conn)
	if err != nil {
		conn.Close() // Close on parse error
		return 0, "", nil, err
	}

	// 3. Return connection to pool for reuse
	GlobalConnectionPool.Put(address, conn)

	return statusCode, body, respHeaders, nil
}
