// Package main implements a web browser from scratch.
// This file contains HTTP fetching logic with Keep-Alive connection pooling.
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

// MaxConnectionsPerHost is the maximum number of idle Keep-Alive connections
// per host, as recommended by HTTP/1.1 (RFC 2616).
const MaxConnectionsPerHost = 6

// Logger for HTTP fetching operations.
// Set to nil to disable logging, or configure with log.SetOutput/SetFlags.
var logger *log.Logger

func init() {
	// Enable logging only if DEBUG environment variable is set
	if os.Getenv("DEBUG") != "" {
		logger = log.New(os.Stderr, "[HTTP] ", log.Ltime)
	} else {
		logger = log.New(io.Discard, "", 0) // Silent by default
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
	connections map[string][]net.Conn // "host:port" → []net.Conn
	mu          sync.Mutex             // protects connections map
	maxPerHost  int                    // maximum idle connections per host
}

// NewConnectionPool creates a new ConnectionPool with default settings.
//
// The pool will maintain up to MaxConnectionsPerHost idle connections
// per server address. Connections exceeding this limit are closed immediately.
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string][]net.Conn),
		maxPerHost:  MaxConnectionsPerHost,
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
		return nil, false
	}

	// Pop last connection (LIFO - most recently used)
	lastIdx := len(conns) - 1
	conn := conns[lastIdx]
	pool.connections[address] = conns[:lastIdx]

	logger.Printf("Reusing connection to %s (remaining: %d)", address, len(conns)-1)
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
		pool.connections[address] = append(conns, conn)
		logger.Printf("Stored connection to %s (total: %d/%d)", address, len(conns)+1, pool.maxPerHost)
	} else {
		conn.Close()
		logger.Printf("Pool full, closed connection to %s (%d/%d)", address, pool.maxPerHost, pool.maxPerHost)
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
	logger.Printf("Closed all connections to %s (%d connections)", address, len(conns))
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
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	logger.Printf("Read file: %s", filePath)
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
			return "", fmt.Errorf("base64 decode failed: %v", err)
		}
		data = string(decoded)
		logger.Println("Decoded base64 data URL")
	} else {
		decoded, err := url.QueryUnescape(data)
		if err != nil {
			decoded = data
		}
		data = decoded
		logger.Println("Decoded URL-encoded data URL")
	}

	return data, nil
}

// Fetch: HTTPFetcher의 Fetch 메서드 구현
func (h *HTTPFetcher) Fetch(u *URL) (string, error) {
	address := fmt.Sprintf("%s:%d", u.Host, u.Port)

	// 1. ConnectionPool에서 기존 연결 찾기
	conn, found := globalConnectionPool.Get(address)

	if !found {
		// 2. Create new connection if not in pool
		logger.Printf("Creating new connection to %s", address)
		var err error

		if u.Scheme == SchemeHTTPS {
			conn, err = tls.Dial("tcp", address, nil)
		} else {
			conn, err = net.Dial("tcp", address)
		}

		if err != nil {
			return "", err
		}
	}

	// HTTP 요청 메시지 만들기
	headers := map[string]string{
		HeaderHost:      u.Host,
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
		return "", err
	}

	// Read and parse HTTP response
	logger.Printf("Request sent to %s:%d", u.Host, u.Port)

	body, _, err := parseResponse(conn) // Ignore headers for now
	if err != nil {
		conn.Close() // Close on parse error
		return "", err
	}

	// 3. Return connection to pool for reuse
	globalConnectionPool.Put(address, conn)

	return body, nil
}

// readChunkedBody reads an HTTP response body with Transfer-Encoding: chunked.
//
// Chunked encoding format:
//   <hex-size>\r\n
//   <data>\r\n
//   <hex-size>\r\n
//   <data>\r\n
//   0\r\n
//   \r\n
//
// Example:
//   5\r\n
//   Hello\r\n
//   6\r\n
//    World\r\n
//   0\r\n
//   \r\n
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
			// Read trailing \r\n
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
			return nil, fmt.Errorf("failed to read header: %w", err)
		}

		// Empty line signals end of headers
		if line == "\r\n" || line == "\n" {
			break
		}

		// Parse "Key: Value" format
		line = strings.TrimSpace(line)
		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 {
			key := strings.TrimSpace(line[:colonIdx])
			value := strings.TrimSpace(line[colonIdx+1:])
			headers[key] = value
		}
	}

	// Log Connection header for Keep-Alive debugging
	if connHeader, ok := headers["Connection"]; ok {
		logger.Printf("Server Connection header: %s", connHeader)
	}

	// DEBUG: Print all headers
	logger.Println("=== All Response Headers ===")
	for key, value := range headers {
		logger.Printf("%s: %s", key, value)
	}
	logger.Println("==============================")

	return headers, nil
}

// readBody reads HTTP response body based on headers.
//
// It uses different strategies depending on the headers:
//   1. If Transfer-Encoding: chunked → read chunked body
//   2. If Content-Length present → read exact bytes
//   3. Otherwise → read until EOF
//
// Strategies 1 and 2 allow connection reuse (Keep-Alive).
// Strategy 3 closes the connection.
//
// Returns:
//   - body bytes
//   - error: if body reading fails
func readBody(reader *bufio.Reader, headers map[string]string) ([]byte, error) {
	// Priority 1: Transfer-Encoding: chunked
	if transferEncoding, ok := headers["Transfer-Encoding"]; ok && transferEncoding == "chunked" {
		bodyBytes, err := readChunkedBody(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunked body: %w", err)
		}
		logger.Println("Read chunked body, connection reusable")
		return bodyBytes, nil
	}

	// Priority 2: Content-Length
	if contentLengthStr, ok := headers["Content-Length"]; ok {
		contentLength, parseErr := strconv.Atoi(contentLengthStr)
		if parseErr != nil || contentLength < 0 {
			return nil, fmt.Errorf("invalid Content-Length: %v", parseErr)
		}

		bodyBytes := make([]byte, contentLength)
		_, err := io.ReadFull(reader, bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read body (Content-Length: %d): %w", contentLength, err)
		}

		logger.Printf("Read %d bytes (Content-Length), connection reusable", contentLength)
		return bodyBytes, nil
	}

	// Priority 3: No explicit length → read until EOF
	logger.Println("No Content-Length or Transfer-Encoding header, reading until EOF")
	bodyBytes, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	return bodyBytes, nil
}

// parseResponse parses an HTTP response and returns the body and headers.
//
// It reads the status line, parses headers, and reads the body.
// This function orchestrates the parsing process by delegating to:
//   - readHeaders() for header parsing
//   - readBody() for body reading with appropriate strategy
//
// Returns:
//   - body: response body as string
//   - headers: map of header names to values
//   - error: any error encountered during parsing
func parseResponse(r io.Reader) (body string, headers map[string]string, err error) {
	reader := bufio.NewReader(r)

	// 1. Read status line (e.g., "HTTP/1.1 200 OK")
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return "", nil, fmt.Errorf("failed to read status line: %w", err)
	}
	_ = statusLine // TODO: parse and return status code

	// 2. Parse headers
	headers, err = readHeaders(reader)
	if err != nil {
		return "", nil, err
	}

	// 3. Read body
	bodyBytes, err := readBody(reader, headers)
	if err != nil {
		return "", headers, err
	}

	return string(bodyBytes), headers, nil
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
		return "", fmt.Errorf("view-source: inner URL request failed: %v", err)
	}

	logger.Println("view-source: returning raw source")
	return content, nil
}
