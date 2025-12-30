package net_test

import (
	"go-web-browser/net"
	"go-web-browser/url"
	stdnet "net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ============================================
// FileFetcher 테스트
// ============================================

// TestFileFetcher_SimpleHTML testdata의 simple.html 읽기
func TestFileFetcher_SimpleHTML(t *testing.T) {
	urlStr := "file://testdata/simple.html"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// simple.html에는 <h1>Hello, World!</h1>가 있을 것으로 예상
	if content == "" {
		t.Error("content should not be empty")
	}

	// HTML 태그가 포함되어 있는지 확인
	if !containsAny(content, "<", ">") {
		t.Errorf("content should contain HTML tags, got: %q", content)
	}
}

// TestFileFetcher_EmptyHTML testdata의 empty.html 읽기
func TestFileFetcher_EmptyHTML(t *testing.T) {
	urlStr := "file://testdata/empty.html"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// empty.html은 빈 파일
	if content != "" {
		t.Errorf("content should be empty, got: %q", content)
	}
}

// TestFileFetcher_EntitiesHTML testdata의 entities.html 읽기
func TestFileFetcher_EntitiesHTML(t *testing.T) {
	urlStr := "file://testdata/entities.html"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// entities.html에는 HTML 엔티티가 포함되어 있음
	if content == "" {
		t.Error("content should not be empty")
	}
}

// TestFileFetcher_FileNotFound 존재하지 않는 파일 에러 처리
func TestFileFetcher_FileNotFound(t *testing.T) {
	urlStr := "file://testdata/nonexistent.html"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = net.Request(u)
	if err == nil {
		t.Error("Request() should return error for nonexistent file")
	}
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// ============================================
// DataFetcher 테스트
// ============================================

// TestDataFetcher_PlainText 일반 텍스트 data URL
func TestDataFetcher_PlainText(t *testing.T) {
	urlStr := "data:text/html,<h1>Hello World</h1>"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "<h1>Hello World</h1>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestDataFetcher_Base64 base64 인코딩된 data URL
func TestDataFetcher_Base64(t *testing.T) {
	// "<h1>Hello</h1>"를 base64 인코딩: PGgxPkhlbGxvPC9oMT4=
	urlStr := "data:text/html;base64,PGgxPkhlbGxvPC9oMT4="

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "<h1>Hello</h1>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestDataFetcher_URLEncoded URL 인코딩된 data URL
func TestDataFetcher_URLEncoded(t *testing.T) {
	// 공백이 %20으로 인코딩됨
	urlStr := "data:text/html,Hello%20World"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "Hello World"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestDataFetcher_ComplexHTML 복잡한 HTML data URL
func TestDataFetcher_ComplexHTML(t *testing.T) {
	urlStr := "data:text/html,<html><body><p>Test</p></body></html>"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "<html><body><p>Test</p></body></html>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestDataFetcher_MissingComma 쉼표 없는 잘못된 data URL
func TestDataFetcher_MissingComma(t *testing.T) {
	urlStr := "data:text/html"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = net.Request(u)
	if err == nil {
		t.Error("Request() should return error for data URL without comma")
	}
}

// TestDataFetcher_InvalidBase64 잘못된 base64 인코딩
func TestDataFetcher_InvalidBase64(t *testing.T) {
	urlStr := "data:text/html;base64,invalid!!!"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = net.Request(u)
	if err == nil {
		t.Error("Request() should return error for invalid base64")
	}
}

// ============================================
// HTTPFetcher 테스트
// ============================================

// TestHTTPFetcher_Success 성공적인 HTTP 요청
func TestHTTPFetcher_Success(t *testing.T) {
	// Mock HTTP 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 요청 검증
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("User-Agent") != net.UserAgent {
			t.Errorf("Expected User-Agent %q, got %q", net.UserAgent, r.Header.Get("User-Agent"))
		}

		// 응답 전송
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Page</h1></body></html>"))
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", server.URL, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "<html><body><h1>Test Page</h1></body></html>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestHTTPFetcher_WithPath 경로가 있는 HTTP 요청
func TestHTTPFetcher_WithPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 경로 검증
		if r.URL.Path != "/test/page.html" {
			t.Errorf("Expected path /test/page.html, got %s", r.URL.Path)
		}

		w.Write([]byte("<p>Path test</p>"))
	}))
	defer server.Close()

	// server.URL에 경로 추가
	urlStr := server.URL + "/test/page.html"
	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	if content != "<p>Path test</p>" {
		t.Errorf("unexpected content: %q", content)
	}
}

// TestHTTPFetcher_EmptyResponse 빈 응답
func TestHTTPFetcher_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// 빈 응답
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", server.URL, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	if content != "" {
		t.Errorf("Expected empty content, got %q", content)
	}
}

// TestHTTPFetcher_HTTPS HTTPS URL 파싱 검증
// 참고: 실제 HTTPS 연결 테스트는 인증서 검증 문제로 mock 서버로 어려움
// 실제 프로덕션 환경에서는 정상 작동함
func TestHTTPFetcher_HTTPS(t *testing.T) {
	urlStr := "https://example.com/index.html"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	// HTTPS URL이 올바르게 파싱되는지 확인
	if u.Scheme != "https" {
		t.Errorf("Scheme = %q; want %q", u.Scheme, "https")
	}
	if u.Port != 443 {
		t.Errorf("Port = %d; want %d", u.Port, 443)
	}

	// 실제 네트워크 요청은 테스트 환경에 따라 실패할 수 있으므로 스킵
	t.Skip("Skipping actual HTTPS request test - would require valid certificate or mock setup")
}

// TestHTTPFetcher_InvalidHost 존재하지 않는 호스트
func TestHTTPFetcher_InvalidHost(t *testing.T) {
	urlStr := "http://invalid-host-that-does-not-exist-12345.com/"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = net.Request(u)
	if err == nil {
		t.Error("Request() should return error for invalid host")
	}
}

// ============================================
// ViewSourceFetcher 테스트
// ============================================

// TestViewSourceFetcher_DataURL view-source:data URL 테스트
func TestViewSourceFetcher_DataURL(t *testing.T) {
	urlStr := "view-source:data:text/html,<h1>Hello</h1>"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// view-source는 원본 HTML을 그대로 반환해야 함 (태그 포함)
	expected := "<h1>Hello</h1>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestViewSourceFetcher_HTTP view-source:http URL 테스트
func TestViewSourceFetcher_HTTP(t *testing.T) {
	// Mock HTTP 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><h1>Test</h1></body></html>"))
	}))
	defer server.Close()

	urlStr := "view-source:" + server.URL

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// view-source는 원본 HTML을 그대로 반환 (태그 포함)
	expected := "<html><body><h1>Test</h1></body></html>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestViewSourceFetcher_File view-source:file URL 테스트
func TestViewSourceFetcher_File(t *testing.T) {
	urlStr := "view-source:file://testdata/simple.html"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// view-source는 파일의 원본 HTML을 그대로 반환
	if content == "" {
		t.Error("content should not be empty")
	}

	// HTML 태그가 포함되어 있어야 함
	if !containsAny(content, "<", ">") {
		t.Errorf("content should contain HTML tags, got: %q", content)
	}
}

// TestViewSourceFetcher_InvalidFormat view-source 잘못된 형식
func TestViewSourceFetcher_InvalidFormat(t *testing.T) {
	urlStr := "view-source:"

	u, err := url.NewURL(urlStr)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = net.Request(u)
	if err == nil {
		t.Error("Request() should return error for view-source with no inner URL")
	}
}

// ============================================
// ConnectionPool 테스트
// ============================================

// mockAddr: 테스트용 가짜 stdnet.Addr
type mockAddr struct{}

func (m *mockAddr) Network() string { return "tcp" }
func (m *mockAddr) String() string  { return "mock:0" }

// mockConn: 테스트용 가짜 net.Conn
type mockConn struct {
	closed bool
	id     int // 연결 구분용
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { m.closed = true; return nil }
func (m *mockConn) LocalAddr() stdnet.Addr             { return &mockAddr{} }
func (m *mockConn) RemoteAddr() stdnet.Addr            { return &mockAddr{} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// TestConnectionPool_GetPut 기본 Get/Put 동작
func TestConnectionPool_GetPut(t *testing.T) {
	pool := net.NewConnectionPool()
	address := "example.com:80"

	// 1. 빈 Pool에서 Get → 없어야 함
	conn, found := pool.Get(address)
	if found {
		t.Error("Get() should return false for empty pool")
	}
	if conn != nil {
		t.Error("Get() should return nil for empty pool")
	}

	// 2. Put으로 연결 저장
	mockConn1 := &mockConn{id: 1}
	pool.Put(address, mockConn1)

	// 3. Get으로 가져오기
	conn, found = pool.Get(address)
	if !found {
		t.Error("Get() should return true after Put()")
	}
	if conn != mockConn1 {
		t.Error("Get() should return the same connection that was Put()")
	}

	// 4. 다시 Get → 없어야 함 (이미 꺼냈으므로)
	conn, found = pool.Get(address)
	if found {
		t.Error("Get() should return false after already getting the connection")
	}
}

// TestConnectionPool_MaxPerHost Pool이 6개로 제한되는지 테스트
func TestConnectionPool_MaxPerHost(t *testing.T) {
	pool := net.NewConnectionPool()
	address := "example.com:80"

	// 1. 10개 연결 Put
	conns := make([]*mockConn, 10)
	for i := 0; i < 10; i++ {
		conns[i] = &mockConn{id: i}
		pool.Put(address, conns[i])
	}

	// 2. Pool에서 모두 Get (최대 6개만 있어야 함)
	retrieved := 0
	for {
		_, found := pool.Get(address)
		if !found {
			break
		}
		retrieved++
	}

	if retrieved != 6 {
		t.Errorf("Pool should contain max 6 connections, got %d", retrieved)
	}

	// 3. 초과분(7, 8, 9, 10번째)은 Close 되었어야 함
	for i := 6; i < 10; i++ {
		if !conns[i].closed {
			t.Errorf("Connection %d should be closed (exceeded maxPerHost)", i)
		}
	}

	// 4. Pool에 저장된 것들(0~5번째)은 Close 안 되었어야 함
	for i := 0; i < 6; i++ {
		if conns[i].closed {
			t.Errorf("Connection %d should not be closed (within maxPerHost)", i)
		}
	}
}

// TestConnectionPool_MultiplHosts 여러 호스트 동시 관리
func TestConnectionPool_MultipleHosts(t *testing.T) {
	pool := net.NewConnectionPool()

	address1 := "example.com:80"
	address2 := "google.com:80"

	// 각 호스트에 연결 저장
	conn1 := &mockConn{id: 1}
	conn2 := &mockConn{id: 2}
	pool.Put(address1, conn1)
	pool.Put(address2, conn2)

	// 각 호스트에서 Get
	retrieved1, found1 := pool.Get(address1)
	retrieved2, found2 := pool.Get(address2)

	if !found1 || !found2 {
		t.Error("Get() should return connections for both hosts")
	}

	if retrieved1 != conn1 || retrieved2 != conn2 {
		t.Error("Get() should return correct connection for each host")
	}
}

// TestConnectionPool_Close 특정 호스트의 모든 연결 닫기
func TestConnectionPool_Close(t *testing.T) {
	pool := net.NewConnectionPool()
	address := "example.com:80"

	// 3개 연결 저장
	conns := make([]*mockConn, 3)
	for i := 0; i < 3; i++ {
		conns[i] = &mockConn{id: i}
		pool.Put(address, conns[i])
	}

	// Close 호출
	pool.Close(address)

	// 모두 닫혔어야 함
	for i := 0; i < 3; i++ {
		if !conns[i].closed {
			t.Errorf("Connection %d should be closed after pool.Close()", i)
		}
	}

	// Pool에서 Get → 없어야 함
	_, found := pool.Get(address)
	if found {
		t.Error("Get() should return false after pool.Close()")
	}
}

// TestHTTPFetcher_ChunkedEncoding: Transfer-Encoding: chunked 응답 처리
func TestHTTPFetcher_ChunkedEncoding(t *testing.T) {
	// Mock HTTP server that sends chunked response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send chunked response manually
		// Don't use w.Write() - it auto-adds Content-Length
		// Write raw HTTP response instead
		conn, buf, _ := w.(http.Hijacker).Hijack()
		defer conn.Close()

		// Status line
		buf.WriteString("HTTP/1.1 200 OK\r\n")
		// Headers
		buf.WriteString("Transfer-Encoding: chunked\r\n")
		buf.WriteString("Connection: keep-alive\r\n")
		buf.WriteString("\r\n")
		// Chunked body: "Hello World"
		buf.WriteString("5\r\n")      // chunk size: 5 bytes
		buf.WriteString("Hello\r\n")  // chunk data
		buf.WriteString("6\r\n")      // chunk size: 6 bytes
		buf.WriteString(" World\r\n") // chunk data
		buf.WriteString("0\r\n")      // last chunk (size 0)
		buf.WriteString("\r\n")       // trailing CRLF
		buf.Flush()
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("url.NewURL(%q) failed: %v", server.URL, err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "Hello World"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestHTTPFetcher_ChunkedEncodingMultipleChunks: 여러 chunk 테스트
func TestHTTPFetcher_ChunkedEncodingMultipleChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, buf, _ := w.(http.Hijacker).Hijack()
		defer conn.Close()

		buf.WriteString("HTTP/1.1 200 OK\r\n")
		buf.WriteString("Transfer-Encoding: chunked\r\n")
		buf.WriteString("\r\n")
		// Many small chunks
		buf.WriteString("1\r\nA\r\n")
		buf.WriteString("1\r\nB\r\n")
		buf.WriteString("1\r\nC\r\n")
		buf.WriteString("1\r\nD\r\n")
		buf.WriteString("0\r\n\r\n")
		buf.Flush()
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "ABCD"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestHTTPFetcher_ChunkedEncodingLarge: 큰 chunk 테스트
func TestHTTPFetcher_ChunkedEncodingLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, buf, _ := w.(http.Hijacker).Hijack()
		defer conn.Close()

		buf.WriteString("HTTP/1.1 200 OK\r\n")
		buf.WriteString("Transfer-Encoding: chunked\r\n")
		buf.WriteString("\r\n")

		// Large chunk: 1000 'X' characters
		largeData := strings.Repeat("X", 1000)
		// 1000 in hex = 0x3E8
		buf.WriteString("3e8\r\n")
		buf.WriteString(largeData + "\r\n")
		buf.WriteString("0\r\n\r\n")
		buf.Flush()
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := strings.Repeat("X", 1000)
	if content != expected {
		t.Errorf("content length = %d; want %d", len(content), len(expected))
	}
}

// ============================================
// Redirect 테스트
// ============================================

// TestHTTPFetcher_Redirect302: 기본 302 리다이렉트
func TestHTTPFetcher_Redirect302(t *testing.T) {
	// 최종 페이지
	finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Final Page</h1>"))
	}))
	defer finalServer.Close()

	// 리다이렉트 서버
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", finalServer.URL)
		w.WriteHeader(http.StatusFound) // 302
	}))
	defer redirectServer.Close()

	u, err := url.NewURL(redirectServer.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "<h1>Final Page</h1>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestHTTPFetcher_RedirectRelativeURL: 상대 URL 리다이렉트 (/)
func TestHTTPFetcher_RedirectRelativeURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			// 상대 경로로 리다이렉트
			w.Header().Set("Location", "/final")
			w.WriteHeader(http.StatusFound)
		} else if r.URL.Path == "/final" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<h1>Final Page</h1>"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL + "/redirect")
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "<h1>Final Page</h1>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestHTTPFetcher_RedirectChain: 연쇄 리다이렉트 (A → B → C)
func TestHTTPFetcher_RedirectChain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/step1":
			w.Header().Set("Location", "/step2")
			w.WriteHeader(http.StatusFound)
		case "/step2":
			w.Header().Set("Location", "/step3")
			w.WriteHeader(http.StatusFound)
		case "/step3":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<h1>Step 3 Final</h1>"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL + "/step1")
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "<h1>Step 3 Final</h1>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// TestHTTPFetcher_RedirectTooMany: 최대 리다이렉트 초과
func TestHTTPFetcher_RedirectTooMany(t *testing.T) {
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		// 무한 리다이렉트 (자기 자신으로)
		w.Header().Set("Location", r.URL.Path)
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL + "/loop")
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	_, err = net.Request(u)
	if err == nil {
		t.Error("Request() should return error for too many redirects")
	}

	// 에러 메시지에 "리다이렉트" 또는 "redirect" 포함되어야 함
	errMsg := strings.ToLower(err.Error())
	if !strings.Contains(errMsg, "redirect") && !strings.Contains(errMsg, "리다이렉트") {
		t.Errorf("Error should mention 'redirect' or '리다이렉트', got: %v", err)
	}
}

// TestHTTPFetcher_RedirectNoLocation: Location 헤더 없는 리다이렉트
func TestHTTPFetcher_RedirectNoLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Location 헤더 없이 302 응답
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	_, err = net.Request(u)
	if err == nil {
		t.Error("Request() should return error for redirect without Location header")
	}

	// 에러 메시지에 "Location" 포함되어야 함
	if !strings.Contains(err.Error(), "Location") {
		t.Errorf("Error should mention 'Location', got: %v", err)
	}
}

// TestHTTPFetcher_Redirect301: 영구 리다이렉트 (301)
func TestHTTPFetcher_Redirect301(t *testing.T) {
	finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Permanent Location</h1>"))
	}))
	defer finalServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", finalServer.URL)
		w.WriteHeader(http.StatusMovedPermanently) // 301
	}))
	defer redirectServer.Close()

	u, err := url.NewURL(redirectServer.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	content, err := net.Request(u)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	expected := "<h1>Permanent Location</h1>"
	if content != expected {
		t.Errorf("content = %q; want %q", content, expected)
	}
}

// ============================================
// Caching 테스트
// ============================================

// TestHTTPFetcher_CacheBasic: 기본 캐싱 동작 - 동일한 URL 두 번 요청
func TestHTTPFetcher_CacheBasic(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Cached Page</h1>"))
	}))
	defer server.Close()

	// Clear cache before test
	net.GlobalCache.Clear()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	// First request - should hit server
	content1, err := net.Request(u)
	if err != nil {
		t.Fatalf("First Request() failed: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("First request should hit server once, got %d requests", requestCount)
	}

	// Second request - should hit cache
	content2, err := net.Request(u)
	if err != nil {
		t.Fatalf("Second Request() failed: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("Second request should not hit server (cached), got %d requests", requestCount)
	}

	// Both should return same content
	if content1 != content2 {
		t.Errorf("Cached content differs: %q vs %q", content1, content2)
	}

	expected := "<h1>Cached Page</h1>"
	if content2 != expected {
		t.Errorf("content = %q; want %q", content2, expected)
	}
}

// TestHTTPFetcher_CacheNoStore: Cache-Control: no-store는 캐시하지 않음
func TestHTTPFetcher_CacheNoStore(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>No Cache</h1>"))
	}))
	defer server.Close()

	// Clear cache before test
	net.GlobalCache.Clear()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	// First request
	_, err = net.Request(u)
	if err != nil {
		t.Fatalf("First Request() failed: %v", err)
	}

	// Second request - should still hit server (no-store)
	_, err = net.Request(u)
	if err != nil {
		t.Fatalf("Second Request() failed: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("no-store should not cache, expected 2 server requests, got %d", requestCount)
	}
}

// TestHTTPFetcher_CacheMaxAge: Cache-Control: max-age=60은 캐시함
func TestHTTPFetcher_CacheMaxAge(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Max Age Cache</h1>"))
	}))
	defer server.Close()

	// Clear cache before test
	net.GlobalCache.Clear()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	// First request
	content1, err := net.Request(u)
	if err != nil {
		t.Fatalf("First Request() failed: %v", err)
	}

	// Second request - should hit cache (within max-age)
	content2, err := net.Request(u)
	if err != nil {
		t.Fatalf("Second Request() failed: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("max-age should cache, expected 1 server request, got %d", requestCount)
	}

	if content1 != content2 {
		t.Errorf("Cached content differs: %q vs %q", content1, content2)
	}
}

// TestHTTPFetcher_CacheExpired: max-age 초과 시 캐시 무효화
func TestHTTPFetcher_CacheExpired(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// 1초 max-age
		w.Header().Set("Cache-Control", "max-age=1")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Expiring Cache</h1>"))
	}))
	defer server.Close()

	// Clear cache before test
	net.GlobalCache.Clear()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	// First request
	_, err = net.Request(u)
	if err != nil {
		t.Fatalf("First Request() failed: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(2 * time.Second)

	// Second request - should hit server (cache expired)
	_, err = net.Request(u)
	if err != nil {
		t.Fatalf("Second Request() failed: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("Expired cache should re-fetch, expected 2 server requests, got %d", requestCount)
	}
}

// TestHTTPFetcher_CacheOtherDirectives: 다른 Cache-Control 값은 캐시하지 않음
func TestHTTPFetcher_CacheOtherDirectives(t *testing.T) {
	testCases := []struct {
		name         string
		cacheControl string
	}{
		{"must-revalidate", "must-revalidate"},
		{"private", "private"},
		{"public", "public"},
		{"no-cache", "no-cache"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				w.Header().Set("Cache-Control", tc.cacheControl)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<h1>Test</h1>"))
			}))
			defer server.Close()

			// Clear cache before test
			net.GlobalCache.Clear()

			u, err := url.NewURL(server.URL)
			if err != nil {
				t.Fatalf("NewURL failed: %v", err)
			}

			// First request
			_, err = net.Request(u)
			if err != nil {
				t.Fatalf("First Request() failed: %v", err)
			}

			// Second request - should still hit server (unknown directive)
			_, err = net.Request(u)
			if err != nil {
				t.Fatalf("Second Request() failed: %v", err)
			}

			if requestCount != 2 {
				t.Errorf("%s should not cache, expected 2 server requests, got %d", tc.name, requestCount)
			}
		})
	}
}

// TestHTTPFetcher_CacheNoCacheControl: Cache-Control 헤더 없으면 기본 캐시
func TestHTTPFetcher_CacheNoCacheControl(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// No Cache-Control header
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Default Cache</h1>"))
	}))
	defer server.Close()

	// Clear cache before test
	net.GlobalCache.Clear()

	u, err := url.NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	// First request
	_, err = net.Request(u)
	if err != nil {
		t.Fatalf("First Request() failed: %v", err)
	}

	// Second request - should hit cache (default cacheable)
	_, err = net.Request(u)
	if err != nil {
		t.Fatalf("Second Request() failed: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("No Cache-Control should cache by default, expected 1 server request, got %d", requestCount)
	}
}

// TestHTTPFetcher_CacheDifferentURLs: 다른 URL은 별도로 캐시
func TestHTTPFetcher_CacheDifferentURLs(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>" + r.URL.Path + "</h1>"))
	}))
	defer server.Close()

	// Clear cache before test
	net.GlobalCache.Clear()

	u1, err := url.NewURL(server.URL + "/page1")
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	u2, err := url.NewURL(server.URL + "/page2")
	if err != nil {
		t.Fatalf("NewURL failed: %v", err)
	}

	// Request url1
	content1, err := net.Request(u1)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// Request url2 - different URL, should hit server
	content2, err := net.Request(u2)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("Different URLs should not share cache, expected 2 requests, got %d", requestCount)
	}

	if content1 == content2 {
		t.Error("Different URLs should have different content")
	}

	// Request url1 again - should hit cache
	_, err = net.Request(u1)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("Second request to url1 should hit cache, expected 2 total requests, got %d", requestCount)
	}
}
