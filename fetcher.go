package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
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

const MaxConnectionPerHost = 6

var logger *log.Logger

func init() {
	if os.Getenv("PRODUCTION") != "" {
		logger = log.New(io.Discard, "", 0) // Silent by default
	} else {
		logger = log.New(os.Stderr, "[HTTP] ", log.Ltime)
	}
}

// ConnectionPool manages persistent HTTP connections for Keep-Alive.
//
// It maintains a pool of idle connections per server address, allowing
// connection reuse across multiple HTTP requests to the same host.
// This significantly reduces latency by avoiding repeated TCP handshakes.
//
// The pool is thread-safe and can be used concurrently from multiple goroutines.
type ConnectionPool struct {
	connections map[string][]net.Conn // "host:port" → []net.Conn (배열로 변경!)
	mu          sync.Mutex            // 동시성 제어 (thread-safe)
	maxPerHost  int                   // 서버당 최대 연결 수
}

// NewConnectionPool creates a new ConnectionPool with default settings.
//
// The pool will maintain up to MaxConnectionsPerHost idle connections
// per server address. Connections exceeding this limit are closed immediately.
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string][]net.Conn),
		maxPerHost:  MaxConnectionPerHost, // HTTP/1.1 권장사항: 서버당 최대 6개 연결
	}
}

// Get retrieves an idle connection from the pool for the given address.
//
// It returns (conn, true) if an idle connection is available, or (nil, false)
// if the pool is empty for this address. The retrieved connection is removed
// from the pool (check-out pattern) and should be returned with Put after use.
//
// Get is safe for concurrent use.
func (pool *ConnectionPool) Get(address string) (net.Conn, bool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conns := pool.connections[address]
	if len(conns) == 0 {
		// 사용 가능한 연결 없음
		return nil, false
	}

	// 마지막 연결 꺼내기 (stack처럼 LIFO)
	lastIdx := len(conns) - 1
	conn := conns[lastIdx]
	pool.connections[address] = conns[:lastIdx] // 제거

	logger.Printf("♻️  기존 연결 재사용: %s (남은 연결: %d개)\n", address, len(conns)-1)
	return conn, true
}

// Put returns a connection to the pool for future reuse.
//
// If the pool already contains maxPerHost connections for this address,
// the connection is closed immediately to prevent resource leaks.
// Otherwise, the connection is stored for reuse by future requests.
//
// Put is safe for concurrent use.
func (pool *ConnectionPool) Put(address string, conn net.Conn) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conns := pool.connections[address]

	if len(conns) < pool.maxPerHost {
		// 배열에 여유 있으면 저장
		pool.connections[address] = append(conns, conn)
		logger.Printf("💾 연결 저장: %s (총 %d/%d개)\n", address, len(conns)+1, pool.maxPerHost)
	} else {
		// Pool이 가득 차면 닫기 (누수 방지!)
		conn.Close()
		logger.Printf("🔌 Pool 가득 차서 연결 닫기: %s (%d/%d)\n", address, pool.maxPerHost, pool.maxPerHost)
	}
}

// Close closes all idle connections for the given address and removes them from the pool.
//
// This is useful when you want to force new connections on the next request,
// or when shutting down.
//
// Close is safe for concurrent use.
func (pool *ConnectionPool) Close(address string) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conns := pool.connections[address]
	for _, conn := range conns {
		conn.Close()
	}
	delete(pool.connections, address)
	logger.Printf("🔌 모든 연결 닫기: %s (%d개)\n", address, len(conns))
}

// 전역 ConnectionPool 인스턴스
var globalConnectionPool = NewConnectionPool()

// Fetcher 인터페이스: URL에서 콘텐츠를 가져오는 역할을 추상화
type Fetcher interface {
	Fetch(u *URL) (string, error)
}

// FileFetcher: file:// 스킴을 처리하는 Fetcher 구현
type FileFetcher struct{}

// DataFetcher: data:// 스킴을 처리하는 Fetcher 구현
type DataFetcher struct{}

// HTTPFetcher: http://, https:// 스킴을 처리하는 Fetcher 구현
type HTTPFetcher struct{}

// ViewSourceFetcher: view-source:// 스킴을 처리하는 Fetcher 구현
type ViewSourceFetcher struct{}

// fetcherRegistry: scheme에 따른 Fetcher를 등록하는 레지스트리
var fetcherRegistry = map[Scheme]Fetcher{
	SchemeFile:       &FileFetcher{},
	SchemeData:       &DataFetcher{},
	SchemeHTTP:       &HTTPFetcher{},
	SchemeHTTPS:      &HTTPFetcher{},
	SchemeViewSource: &ViewSourceFetcher{},
}

// Request: URL에서 콘텐츠를 가져오는 메서드
func (u *URL) Request() (string, error) {
	fetcher, ok := fetcherRegistry[u.Scheme]
	if !ok {
		return "", fmt.Errorf("지원하지 않는 프로토콜: %s", u.Scheme)
	}
	return fetcher.Fetch(u)
}

// Fetch: FileFetcher의 Fetch 메서드 구현
func (f *FileFetcher) Fetch(u *URL) (string, error) {
	filePath := u.Path

	// Windows 절대 경로 처리: /C:/path → C:/path
	if len(filePath) > 2 && filePath[0] == '/' && filePath[2] == ':' {
		filePath = filePath[1:]
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("파일 읽기 실패: %v", err)
	}

	logger.Printf("--- 파일 %s 읽기 완료 ---\n", filePath)
	return string(content), nil
}

// Fetch: DataFetcher의 Fetch 메서드 구현
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
		logger.Printf("--- [data] base64 디코딩 완료 ---\n")
	} else {
		decoded, err := url.QueryUnescape(data)
		if err != nil {
			decoded = data
		}
		data = decoded
		logger.Println("--- [data] URL 파싱 완료 ---")
	}

	return data, nil
}

// Fetch: HTTPFetcher의 Fetch 메서드 구현
func (h *HTTPFetcher) Fetch(u *URL) (string, error) {
	const maxRedirects = 10
	currentURL := u

	for i := 0; i < maxRedirects; i++ {
		statusCode, body, headers, err := h.doRequest(currentURL)
		if err != nil {
			return "", err
		}

		if statusCode < 300 || statusCode >= 400 {
			return body, nil
		}

		location := headers["location"]
		if location == "" {
			return "", fmt.Errorf("리다이렉트 응답에 Location 헤더가 없습니다 (status %d)", statusCode)
		}

		logger.Printf("리다이렉트 %d: %d -> %s", i+1, statusCode, location)

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
func resolveURL(base *URL, location string) (*URL, error) {
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		return NewURL(location)
	}

	if strings.HasPrefix(location, "/") {
		var absoluteURL string
		if base.Scheme == SchemeHTTPS && base.Port == DefaultHTTPSPort {
			absoluteURL = fmt.Sprintf("https://%s%s", base.Host, location)
		} else if base.Scheme == SchemeHTTP && base.Port == DefaultHTTPPort {
			absoluteURL = fmt.Sprintf("http://%s%s", base.Host, location)
		} else {
			absoluteURL = fmt.Sprintf("%s://%s:%d%s", base.Scheme, base.Host, base.Port, location)
		}
		return NewURL(absoluteURL)
	}

	return nil, fmt.Errorf("지원하지 않는 Location 형식: %q (절대 URL 또는 상대 경로가 아님)", location)
}

// doRequest performs a single HTTP request and returns status code, body, headers
func (h *HTTPFetcher) doRequest(u *URL) (int, string, map[string]string, error) {
	address := fmt.Sprintf("%s:%d", u.Host, u.Port)

	// 1. ConnectionPool에서 기존 연결 찾기
	conn, found := globalConnectionPool.Get(address)

	if !found {
		// 2. Pool에 없으면 새로운 연결 생성
		logger.Printf("🆕 새 연결 생성: %s\n", address)
		var err error

		if u.Scheme == SchemeHTTPS {
			conn, err = tls.Dial("tcp", address, nil)
		} else {
			conn, err = net.Dial("tcp", address)
		}

		if err != nil {
			return 0, "", nil, err
		}
	}
	// (found == true인 경우는 Get()에서 "♻️ 기존 연결 재사용" 메시지 출력함)

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

	// 서버의 대답(응답) 읽기
	logger.Printf("--- [%s:%d] 연결 및 요청 완료 ---\n", u.Host, u.Port)

	statusCode, body, responseHeader, err := parseResponse(conn)
	if err != nil {
		conn.Close() // 응답 파싱 실패 시 연결 닫기
		return 0, "", nil, err
	}

	// 3. 성공하면 Pool에 연결 저장 (재사용을 위해)
	globalConnectionPool.Put(address, conn)

	return statusCode, body, responseHeader, nil
}

// readChunkedBody reads an HTTP response body with Transfer-Encoding: chunked.
//
// Chunked encoding format:
//
//	<hex-size>\r\n
//	<data>\r\n
//	<hex-size>\r\n
//	<data>\r\n
//	0\r\n
//	\r\n
//
// Example:
//
//	5\r\n
//	Hello\r\n
//	6\r\n
//	 World\r\n
//	0\r\n
//	\r\n
//
// → "Hello World"
//
// Returns:
//   - body bytes
//   - error if chunk parsing fails
func readChunkedBody(reader *bufio.Reader) ([]byte, error) {
	var body []byte

	for {
		// 1. Read chunk size line (hex number + \r\n)
		sizeLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk size: %w", err)
		}

		// 2. Parse hex size to decimal
		sizeLine = strings.TrimSpace(sizeLine)
		chunkSize, err := strconv.ParseInt(sizeLine, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chunk size %q: %w", sizeLine, err)
		}

		logger.Printf("Read chunk size: %d (0x%s)", chunkSize, sizeLine)

		// 3. If chunk size is 0, we're done
		if chunkSize == 0 {
			reader.ReadString('\n')
			break
		}

		// 4. Read chunk data (exactly chunkSize bytes)
		chunkData := make([]byte, chunkSize)
		_, err = io.ReadFull(reader, chunkData)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk data: %w", err)
		}

		// 5. Read trailing \r\n after chunk data
		_, err = reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk trailing CRLF: %w", err)
		}

		// 6. Append to body
		body = append(body, chunkData...)
	}
	return body, nil
}

// readHeaders reads HTTP response headers from reader.
//
// It reads lines until it encounters an empty line (\r\n or \n),
// which signals the end of headers. Each header is parsed as "Key: Value"
// and stored in a map.
//
// Returns:
//   - headers: map of header names to values
//   - error: if header reading fails
func readHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("헤더 읽기 실패: %w", err)
		}

		// 빈 줄이 나오면 헤더 끝
		if line == "\r\n" || line == "\n" {
			break
		}

		// 헤더 파싱: "Content-Length: 1234\r\n" → key: "Content-Length", value: "1234"
		line = strings.TrimSpace(line) // 앞뒤 공백 제거
		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 {
			key := strings.TrimSpace(line[:colonIdx])     // "Content-Length"
			value := strings.TrimSpace(line[colonIdx+1:]) // "1234"
			// Normalize header names to lowercase (HTTP headers are case-insensitive)
			headers[strings.ToLower(key)] = value
		}
	}

	// 디버깅: 서버가 keep-alive로 응답했는지 확인
	if connHeader, ok := headers["connection"]; ok {
		logger.Printf("🔌 서버 응답 Connection 헤더: %s\n", connHeader)
	} else {
		fmt.Println("🔌 Connection 헤더 없음 (HTTP/1.1 기본 = keep-alive)")
	}

	logger.Println("=== All Response Headers ===")
	for key, value := range headers {
		logger.Printf("%s: %s", key, value)
	}
	logger.Println("=========================")

	return headers, nil
}

// readBody reads HTTP response body based on headers.
//
// It uses different strategies depending on the headers:
//  1. If Transfer-Encoding: chunked → read chunked body
//  2. If Content-Length present → read exact bytes
//  3. Otherwise, → read until EOF
//
// Strategies 1 and 2 allow connection reuse (Keep-Alive).
// Strategy 3 closes the connection.
//
// Returns:
//   - body bytes
//   - error: if body reading fails
func readBody(reader *bufio.Reader, headers map[string]string) ([]byte, error) {
	if transferEncoding, ok := headers["transfer-Encoding"]; ok && transferEncoding == "chunked" {
		bodyBytes, err := readChunkedBody(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunked body: %w", err)
		}
		logger.Println("Read chunked body, connection reusable")
		return bodyBytes, nil
	} else if contentLengthStr, ok := headers["content-Length"]; ok {
		// Content-Length가 있으면: 정확히 그만큼만 읽기
		logger.Printf("📏 Content-Length 헤더 발견: %s 바이트\n", contentLengthStr)

		// string → int 변환 (예: "1234" → 1234)
		contentLength, parseErr := strconv.Atoi(contentLengthStr)
		if parseErr != nil || contentLength < 0 {
			return nil, fmt.Errorf("Content-Length 파싱 실패: %v", parseErr)
		}

		// 정확히 contentLength 바이트만 읽기
		bodyBytes := make([]byte, contentLength) // N바이트 버퍼 생성
		_, err := io.ReadFull(reader, bodyBytes) // 정확히 N바이트 읽기
		if err != nil {
			return nil, fmt.Errorf("바디 읽기 실패 (Content-Length: %d): %w", contentLength, err)
		}

		logger.Printf("✅ %d 바이트 정확히 읽음 (소켓 유지 가능!)\n", contentLength)
		return bodyBytes, nil
	}

	// Content-Length가 없으면: 기존 방식 (io.ReadAll)
	logger.Println("⚠️  Content-Length 없음, 연결 끝까지 읽기")
	bodyBytes, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("바디 읽기 실패: %w", err)
	}

	return bodyBytes, nil
}

// parseResponse parses an HTTP response and returns the status code, body and headers.
//
// It reads the status line, parses headers, and reads the body.
// This function orchestrates the parsing process by delegating to:
//   - readHeaders() for header parsing
//   - readBody() for body reading with appropriate strategy
//
// Returns:
//   - statusCode: HTTP status code (e.g., 200, 302, 404)
//   - body: response body as string
//   - headers: map of header names to values
//   - error: any error encountered during parsing
func parseResponse(r io.Reader) (statusCode int, body string, headers map[string]string, err error) {
	reader := bufio.NewReader(r)

	// 1. Status Line 읽기 (예: HTTP/1.1 200 OK)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return 0, "", nil, fmt.Errorf("상태 라인 읽기 실패: %w", err)
	}
	_ = statusLine // 현재는 상태 코드를 검사하지 않지만, 나중에 확장을 위해 저장

	statusLine = strings.TrimSpace(statusLine)
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return 0, "", nil, fmt.Errorf("invalid status line: %q", statusLine)
	}

	statusCode, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, "", nil, fmt.Errorf("invalid status code in status line %q: %w", statusLine, err)
	}

	logger.Printf("Status: %d %s", statusCode, statusLine)

	headers, err = readHeaders(reader)
	if err != nil {
		return statusCode, "", nil, err
	}

	// 3. Body 읽기: Content-Length에 따라 다르게 처리
	bodyBytes, err := readBody(reader, headers)
	if err != nil {
		return statusCode, "", headers, err
	}

	return statusCode, string(bodyBytes), headers, nil
}

// Fetch: ViewSourceFetcher의 Fetch 메서드 구현
func (v *ViewSourceFetcher) Fetch(u *URL) (string, error) {
	// Path에는 내부 URL 전체가 들어있음 (예: "http://example.org/")
	innerURLStr := u.Path

	if innerURLStr == "" {
		return "", fmt.Errorf("view-source: 내부 URL이 없습니다")
	}

	// 내부 URL 파싱
	innerURL, err := NewURL(innerURLStr)
	if err != nil {
		return "", fmt.Errorf("view-source: 내부 URL 파싱 실패: %v", err)
	}

	// 내부 URL로 콘텐츠 가져오기 (원본 그대로 반환)
	content, err := innerURL.Request()
	if err != nil {
		return "", fmt.Errorf("view-source: 내부 URL 요청 실패: %v", err)
	}

	logger.Println("--- [view-source] 원본 소스 반환 ---")
	return content, nil
}
