// Package net implements HTTP networking for the browser.
// This file contains Fetcher interface and implementations for different protocols.
package net

import (
	"encoding/base64"
	"fmt"
	"go-web-browser/llm/logger"
	"go-web-browser/llm/url"
	stdurl "net/url"
	"os"
	"strings"
)

// Fetcher 인터페이스: URL에서 콘텐츠를 가져오는 역할을 추상화
type Fetcher interface {
	Fetch(u *url.URL) (string, error)
}

// FileFetcher: file:// 스킴을 처리하는 Fetcher 구현
type FileFetcher struct{}

// DataFetcher: data:// 스킴을 처리하는 Fetcher 구현
type DataFetcher struct{}

// ViewSourceFetcher: view-source:// 스킴을 처리하는 Fetcher 구현
type ViewSourceFetcher struct{}

// FetcherRegistry: scheme에 따른 Fetcher를 등록하는 레지스트리
var FetcherRegistry = map[url.Scheme]Fetcher{
	url.SchemeFile:       &FileFetcher{},
	url.SchemeData:       &DataFetcher{},
	url.SchemeHTTP:       &HTTPFetcher{},
	url.SchemeHTTPS:      &HTTPFetcher{},
	url.SchemeViewSource: &ViewSourceFetcher{},
}

// Request: URL에서 콘텐츠를 가져오는 함수
func Request(u *url.URL) (string, error) {
	fetcher, ok := FetcherRegistry[u.Scheme]
	if !ok {
		return "", fmt.Errorf("지원하지 않는 프로토콜: %s", u.Scheme)
	}
	return fetcher.Fetch(u)
}

// Fetch: FileFetcher의 Fetch 메서드 구현
func (f *FileFetcher) Fetch(u *url.URL) (string, error) {
	filePath := u.Path

	// Windows 절대 경로 처리: /C:/path → C:/path
	if len(filePath) > 2 && filePath[0] == '/' && filePath[2] == ':' {
		filePath = filePath[1:]
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	logger.Logger.Printf("Read file: %s", filePath)
	return string(content), nil
}

// Fetch: DataFetcher의 Fetch 메서드 구현
func (d *DataFetcher) Fetch(u *url.URL) (string, error) {
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
		logger.Logger.Println("Decoded base64 data URL")
	} else {
		decoded, err := stdurl.QueryUnescape(data)
		if err != nil {
			decoded = data
		}
		data = decoded
		logger.Logger.Println("Decoded URL-encoded data URL")
	}

	return data, nil
}

// Fetch: ViewSourceFetcher의 Fetch 메서드 구현
func (v *ViewSourceFetcher) Fetch(u *url.URL) (string, error) {
	// Path에는 내부 URL 전체가 들어있음 (예: "http://example.org/")
	innerURLStr := u.Path

	if innerURLStr == "" {
		return "", fmt.Errorf("view-source: 내부 URL이 없습니다")
	}

	// 내부 URL 파싱
	innerURL, err := url.NewURL(innerURLStr)
	if err != nil {
		return "", fmt.Errorf("view-source: 내부 URL 파싱 실패: %v", err)
	}

	// 내부 URL로 콘텐츠 가져오기 (원본 그대로 반환)
	// Request() 메서드를 사용해야 하는데, 이것은 url.URL에 있음
	// 하지만 이건 순환 의존성 문제가 있음...
	// 해결책: Request()를 별도로 처리하거나, ViewSourceFetcher가 직접 FetcherRegistry 사용
	fetcher, ok := FetcherRegistry[innerURL.Scheme]
	if !ok {
		return "", fmt.Errorf("지원하지 않는 프로토콜: %s", innerURL.Scheme)
	}

	content, err := fetcher.Fetch(innerURL)
	if err != nil {
		return "", fmt.Errorf("view-source: inner URL request failed: %v", err)
	}

	logger.Logger.Println("view-source: returning raw source")
	return content, nil
}
