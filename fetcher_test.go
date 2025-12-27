package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============================================
// FileFetcher 테스트
// ============================================

// TestFileFetcher_SimpleHTML testdata의 simple.html 읽기
func TestFileFetcher_SimpleHTML(t *testing.T) {
	urlStr := "file://testdata/simple.html"

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = url.Request()
	if err == nil {
		t.Error("Request() should return error for data URL without comma")
	}
}

// TestDataFetcher_InvalidBase64 잘못된 base64 인코딩
func TestDataFetcher_InvalidBase64(t *testing.T) {
	urlStr := "data:text/html;base64,invalid!!!"

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = url.Request()
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
		if r.Header.Get("User-Agent") != UserAgent {
			t.Errorf("Expected User-Agent %q, got %q", UserAgent, r.Header.Get("User-Agent"))
		}

		// 응답 전송
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Page</h1></body></html>"))
	}))
	defer server.Close()

	url, err := NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", server.URL, err)
	}

	content, err := url.Request()
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
	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(server.URL)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", server.URL, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	// HTTPS URL이 올바르게 파싱되는지 확인
	if url.Scheme != "https" {
		t.Errorf("Scheme = %q; want %q", url.Scheme, "https")
	}
	if url.Port != 443 {
		t.Errorf("Port = %d; want %d", url.Port, 443)
	}

	// 실제 네트워크 요청은 테스트 환경에 따라 실패할 수 있으므로 스킵
	t.Skip("Skipping actual HTTPS request test - would require valid certificate or mock setup")
}

// TestHTTPFetcher_InvalidHost 존재하지 않는 호스트
func TestHTTPFetcher_InvalidHost(t *testing.T) {
	urlStr := "http://invalid-host-that-does-not-exist-12345.com/"

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	content, err := url.Request()
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

	url, err := NewURL(urlStr)
	if err != nil {
		t.Fatalf("NewURL(%q) failed: %v", urlStr, err)
	}

	_, err = url.Request()
	if err == nil {
		t.Error("Request() should return error for view-source with no inner URL")
	}
}
