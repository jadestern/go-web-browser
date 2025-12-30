// Package net implements HTTP networking for the browser.
// This file contains HTTP protocol parsing helpers.
package net

import (
	"bufio"
	"fmt"
	"go-web-browser/llm/logger"
	"io"
	"strconv"
	"strings"
)

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

		logger.Logger.Printf("Read chunk size: %d (0x%s)", chunkSize, sizeLine)

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
			// Normalize header names to lowercase (HTTP headers are case-insensitive)
			headers[strings.ToLower(key)] = value
		}
	}

	// Log Connection header for Keep-Alive debugging
	if connHeader, ok := headers["connection"]; ok {
		logger.Logger.Printf("Server Connection header: %s", connHeader)
	}

	// DEBUG: Print all headers
	logger.Logger.Println("=== All Response Headers ===")
	for key, value := range headers {
		logger.Logger.Printf("%s: %s", key, value)
	}
	logger.Logger.Println("==============================")

	return headers, nil
}

// readBody reads HTTP response body based on headers.
//
// It uses different strategies depending on the headers:
//  1. If Transfer-Encoding: chunked → read chunked body
//  2. If Content-Length present → read exact bytes
//  3. Otherwise → read until EOF
//
// Strategies 1 and 2 allow connection reuse (Keep-Alive).
// Strategy 3 closes the connection.
//
// Returns:
//   - body bytes
//   - error: if body reading fails
func readBody(reader *bufio.Reader, headers map[string]string) ([]byte, error) {
	// Priority 1: Transfer-Encoding: chunked
	if transferEncoding, ok := headers["transfer-encoding"]; ok && transferEncoding == "chunked" {
		bodyBytes, err := readChunkedBody(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunked body: %w", err)
		}
		logger.Logger.Println("Read chunked body, connection reusable")
		return bodyBytes, nil
	}

	// Priority 2: Content-Length
	if contentLengthStr, ok := headers["content-length"]; ok {
		contentLength, parseErr := strconv.Atoi(contentLengthStr)
		if parseErr != nil || contentLength < 0 {
			return nil, fmt.Errorf("invalid Content-Length: %v", parseErr)
		}

		bodyBytes := make([]byte, contentLength)
		_, err := io.ReadFull(reader, bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read body (Content-Length: %d): %w", contentLength, err)
		}

		logger.Logger.Printf("Read %d bytes (Content-Length), connection reusable", contentLength)
		return bodyBytes, nil
	}

	// Priority 3: No explicit length → read until EOF
	logger.Logger.Println("No Content-Length or Transfer-Encoding header, reading until EOF")
	bodyBytes, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	return bodyBytes, nil
}

// ParseResponse parses an HTTP response and returns the status code, body and headers.
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
func ParseResponse(r io.Reader) (statusCode int, body string, headers map[string]string, err error) {
	reader := bufio.NewReader(r)

	// 1. Read status line (e.g., "HTTP/1.1 200 OK")
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return 0, "", nil, fmt.Errorf("failed to read status line: %w", err)
	}

	// Parse status code from status line
	// Format: "HTTP/1.1 200 OK\r\n"
	statusLine = strings.TrimSpace(statusLine)
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return 0, "", nil, fmt.Errorf("invalid status line: %q", statusLine)
	}

	statusCode, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, "", nil, fmt.Errorf("invalid status code in status line %q: %w", statusLine, err)
	}

	logger.Logger.Printf("Status: %d %s", statusCode, statusLine)

	// 2. Parse headers
	headers, err = readHeaders(reader)
	if err != nil {
		return statusCode, "", nil, err
	}

	// 3. Read body
	bodyBytes, err := readBody(reader, headers)
	if err != nil {
		return statusCode, "", headers, err
	}

	return statusCode, string(bodyBytes), headers, nil
}
