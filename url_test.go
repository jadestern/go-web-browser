package main

import "testing"

// ============================================
// NewURL 테스트
// ============================================

// TestNewURL_HTTP_DefaultPort HTTP URL 기본 포트 테스트
func TestNewURL_HTTP_DefaultPort(t *testing.T) {
	urlStr := "http://example.com/index.html"

	result, err := NewURL(urlStr)

	if err != nil {
		t.Fatalf("NewURL(%q) returned error: %v", urlStr, err)
	}

	if result.Scheme != SchemeHTTP {
		t.Errorf("Scheme = %q; want %q", result.Scheme, SchemeHTTP)
	}
	if result.Host != "example.com" {
		t.Errorf("Host = %q; want %q", result.Host, "example.com")
	}
	if result.Port != 80 {
		t.Errorf("Port = %d; want %d", result.Port, 80)
	}
	if result.Path != "/index.html" {
		t.Errorf("Path = %q; want %q", result.Path, "/index.html")
	}
}

// TestNewURL_HTTP_CustomPort HTTP URL 커스텀 포트 테스트
func TestNewURL_HTTP_CustomPort(t *testing.T) {
	urlStr := "http://example.com:8080/api"

	result, err := NewURL(urlStr)

	if err != nil {
		t.Fatalf("NewURL(%q) returned error: %v", urlStr, err)
	}

	if result.Scheme != SchemeHTTP {
		t.Errorf("Scheme = %q; want %q", result.Scheme, SchemeHTTP)
	}
	if result.Host != "example.com" {
		t.Errorf("Host = %q; want %q", result.Host, "example.com")
	}
	if result.Port != 8080 {
		t.Errorf("Port = %d; want %d", result.Port, 8080)
	}
	if result.Path != "/api" {
		t.Errorf("Path = %q; want %q", result.Path, "/api")
	}
}

// TestNewURL_HTTPS_DefaultPort HTTPS URL 기본 포트 테스트
func TestNewURL_HTTPS_DefaultPort(t *testing.T) {
	urlStr := "https://secure.example.com/login"

	result, err := NewURL(urlStr)

	if err != nil {
		t.Fatalf("NewURL(%q) returned error: %v", urlStr, err)
	}

	if result.Scheme != SchemeHTTPS {
		t.Errorf("Scheme = %q; want %q", result.Scheme, SchemeHTTPS)
	}
	if result.Host != "secure.example.com" {
		t.Errorf("Host = %q; want %q", result.Host, "secure.example.com")
	}
	if result.Port != 443 {
		t.Errorf("Port = %d; want %d", result.Port, 443)
	}
	if result.Path != "/login" {
		t.Errorf("Path = %q; want %q", result.Path, "/login")
	}
}

// TestNewURL_File_Windows Windows 파일 경로 테스트
func TestNewURL_File_Windows(t *testing.T) {
	urlStr := "file:///C:/Users/test/index.html"

	result, err := NewURL(urlStr)

	if err != nil {
		t.Fatalf("NewURL(%q) returned error: %v", urlStr, err)
	}

	if result.Scheme != SchemeFile {
		t.Errorf("Scheme = %q; want %q", result.Scheme, SchemeFile)
	}
	if result.Host != "" {
		t.Errorf("Host = %q; want empty string", result.Host)
	}
	if result.Port != 0 {
		t.Errorf("Port = %d; want %d", result.Port, 0)
	}
	if result.Path != "/C:/Users/test/index.html" {
		t.Errorf("Path = %q; want %q", result.Path, "/C:/Users/test/index.html")
	}
}

// TestNewURL_Data data URL 테스트
func TestNewURL_Data(t *testing.T) {
	urlStr := "data:text/html,<h1>Hello</h1>"

	result, err := NewURL(urlStr)

	if err != nil {
		t.Fatalf("NewURL(%q) returned error: %v", urlStr, err)
	}

	if result.Scheme != SchemeData {
		t.Errorf("Scheme = %q; want %q", result.Scheme, SchemeData)
	}
	if result.Host != "" {
		t.Errorf("Host = %q; want empty string", result.Host)
	}
	if result.Port != 0 {
		t.Errorf("Port = %d; want %d", result.Port, 0)
	}
	if result.Path != "text/html,<h1>Hello</h1>" {
		t.Errorf("Path = %q; want %q", result.Path, "text/html,<h1>Hello</h1>")
	}
}

// TestNewURL_NoPath 경로 없는 URL 테스트
func TestNewURL_NoPath(t *testing.T) {
	urlStr := "http://example.com"

	result, err := NewURL(urlStr)

	if err != nil {
		t.Fatalf("NewURL(%q) returned error: %v", urlStr, err)
	}

	if result.Path != "/" {
		t.Errorf("Path = %q; want %q", result.Path, "/")
	}
}

// TestNewURL_InvalidScheme 잘못된 스킴 테스트
func TestNewURL_InvalidScheme(t *testing.T) {
	urlStr := "ftp://example.com/file"

	_, err := NewURL(urlStr)

	if err == nil {
		t.Errorf("NewURL(%q) should return error for unsupported scheme", urlStr)
	}
}

// TestNewURL_MissingScheme 스킴 없는 URL 테스트
func TestNewURL_MissingScheme(t *testing.T) {
	urlStr := "example.com/path"

	_, err := NewURL(urlStr)

	if err == nil {
		t.Errorf("NewURL(%q) should return error for missing scheme", urlStr)
	}
}

// ============================================
// parsePort 테스트
// ============================================

// TestParsePort_HTTP_DefaultPort HTTP 기본 포트 테스트
func TestParsePort_HTTP_DefaultPort(t *testing.T) {
	scheme := SchemeHTTP
	host := "example.com"

	cleanHost, port, err := parsePort(scheme, host)

	if err != nil {
		t.Fatalf("parsePort(%q, %q) returned error: %v", scheme, host, err)
	}

	if cleanHost != "example.com" {
		t.Errorf("cleanHost = %q; want %q", cleanHost, "example.com")
	}
	if port != 80 {
		t.Errorf("port = %d; want %d", port, 80)
	}
}

// TestParsePort_HTTP_CustomPort HTTP 커스텀 포트 테스트
func TestParsePort_HTTP_CustomPort(t *testing.T) {
	scheme := SchemeHTTP
	host := "example.com:8080"

	cleanHost, port, err := parsePort(scheme, host)

	if err != nil {
		t.Fatalf("parsePort(%q, %q) returned error: %v", scheme, host, err)
	}

	if cleanHost != "example.com" {
		t.Errorf("cleanHost = %q; want %q", cleanHost, "example.com")
	}
	if port != 8080 {
		t.Errorf("port = %d; want %d", port, 8080)
	}
}

// TestParsePort_HTTPS_DefaultPort HTTPS 기본 포트 테스트
func TestParsePort_HTTPS_DefaultPort(t *testing.T) {
	scheme := SchemeHTTPS
	host := "secure.example.com"

	cleanHost, port, err := parsePort(scheme, host)

	if err != nil {
		t.Fatalf("parsePort(%q, %q) returned error: %v", scheme, host, err)
	}

	if cleanHost != "secure.example.com" {
		t.Errorf("cleanHost = %q; want %q", cleanHost, "secure.example.com")
	}
	if port != 443 {
		t.Errorf("port = %d; want %d", port, 443)
	}
}

// TestParsePort_HTTPS_CustomPort HTTPS 커스텀 포트 테스트
func TestParsePort_HTTPS_CustomPort(t *testing.T) {
	scheme := SchemeHTTPS
	host := "secure.example.com:8443"

	cleanHost, port, err := parsePort(scheme, host)

	if err != nil {
		t.Fatalf("parsePort(%q, %q) returned error: %v", scheme, host, err)
	}

	if cleanHost != "secure.example.com" {
		t.Errorf("cleanHost = %q; want %q", cleanHost, "secure.example.com")
	}
	if port != 8443 {
		t.Errorf("port = %d; want %d", port, 8443)
	}
}

// TestParsePort_File file 스킴 테스트 (포트 없음)
func TestParsePort_File(t *testing.T) {
	scheme := SchemeFile
	host := ""

	cleanHost, port, err := parsePort(scheme, host)

	if err != nil {
		t.Fatalf("parsePort(%q, %q) returned error: %v", scheme, host, err)
	}

	if cleanHost != "" {
		t.Errorf("cleanHost = %q; want empty string", cleanHost)
	}
	if port != 0 {
		t.Errorf("port = %d; want %d", port, 0)
	}
}

// TestParsePort_InvalidPort 잘못된 포트 번호 테스트
func TestParsePort_InvalidPort(t *testing.T) {
	scheme := SchemeHTTP
	host := "example.com:abc"

	_, _, err := parsePort(scheme, host)

	if err == nil {
		t.Errorf("parsePort(%q, %q) should return error for invalid port", scheme, host)
	}
}

// ============================================
// parseHostPath 테스트
// ============================================

// TestParseHostPath_HTTP_WithPath HTTP URL에서 host/path 분리
func TestParseHostPath_HTTP_WithPath(t *testing.T) {
	scheme := SchemeHTTP
	rest := "example.com/index.html"

	host, path := parseHostPath(scheme, rest)

	if host != "example.com" {
		t.Errorf("host = %q; want %q", host, "example.com")
	}
	if path != "/index.html" {
		t.Errorf("path = %q; want %q", path, "/index.html")
	}
}

// TestParseHostPath_HTTP_NoPath 경로 없는 HTTP URL
func TestParseHostPath_HTTP_NoPath(t *testing.T) {
	scheme := SchemeHTTP
	rest := "example.com"

	host, path := parseHostPath(scheme, rest)

	if host != "example.com" {
		t.Errorf("host = %q; want %q", host, "example.com")
	}
	if path != "/" {
		t.Errorf("path = %q; want %q", path, "/")
	}
}

// TestParseHostPath_HTTP_WithPort 포트가 포함된 HTTP URL
func TestParseHostPath_HTTP_WithPort(t *testing.T) {
	scheme := SchemeHTTP
	rest := "example.com:8080/api/users"

	host, path := parseHostPath(scheme, rest)

	if host != "example.com:8080" {
		t.Errorf("host = %q; want %q", host, "example.com:8080")
	}
	if path != "/api/users" {
		t.Errorf("path = %q; want %q", path, "/api/users")
	}
}

// TestParseHostPath_HTTPS_WithPath HTTPS URL에서 host/path 분리
func TestParseHostPath_HTTPS_WithPath(t *testing.T) {
	scheme := SchemeHTTPS
	rest := "secure.example.com/login"

	host, path := parseHostPath(scheme, rest)

	if host != "secure.example.com" {
		t.Errorf("host = %q; want %q", host, "secure.example.com")
	}
	if path != "/login" {
		t.Errorf("path = %q; want %q", path, "/login")
	}
}

// TestParseHostPath_File file 스킴 처리
func TestParseHostPath_File(t *testing.T) {
	scheme := SchemeFile
	rest := "/C:/Users/test/index.html"

	host, path := parseHostPath(scheme, rest)

	if host != "" {
		t.Errorf("host = %q; want empty string", host)
	}
	if path != "/C:/Users/test/index.html" {
		t.Errorf("path = %q; want %q", path, "/C:/Users/test/index.html")
	}
}

// TestParseHostPath_File_Relative file 스킴 상대 경로
func TestParseHostPath_File_Relative(t *testing.T) {
	scheme := SchemeFile
	rest := "test.html"

	host, path := parseHostPath(scheme, rest)

	if host != "" {
		t.Errorf("host = %q; want empty string", host)
	}
	if path != "test.html" {
		t.Errorf("path = %q; want %q", path, "test.html")
	}
}
