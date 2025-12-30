// Package net implements HTTP networking for the browser.
// This file contains HTTP response caching.
package net

import (
	"go-web-browser/logger"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CacheEntry는 캐시된 HTTP 응답을 나타냄
//
// 응답 본문, 헤더, 캐시 저장 시간,
// Cache-Control 헤더의 max-age 값을 저장함
type CacheEntry struct {
	Body      string            // 응답 본문
	Headers   map[string]string // 응답 헤더
	Timestamp int64             // 캐시 저장 시간 (Unix timestamp)
	MaxAge    int               // max-age 값 (초 단위, 0 = max-age 없음, -1 = no-store)
}

// Cache는 HTTP 응답 캐싱을 관리함
//
// URL 문자열을 키로 응답을 저장하고,
// Cache-Control 헤더(no-store, max-age)에 따라 캐시 정책을 적용함
//
// 캐시는 thread-safe하며 여러 goroutine에서 동시에 사용 가능함
type Cache struct {
	entries map[string]*CacheEntry // URL → CacheEntry
	mu      sync.Mutex             // entries map 보호
}

// NewCache는 새 Cache 인스턴스를 생성함
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
	}
}

// Get은 주어진 URL에 대한 캐시 엔트리를 가져옴
//
// 엔트리가 존재하고 유효하면 (entry, true) 반환,
// 존재하지 않거나 만료되었으면 (nil, false) 반환
//
// max-age에 따라 만료 여부를 확인함:
//   - max-age가 설정되고 만료되었으면 (nil, false) 반환
//   - max-age가 0이면 (max-age 없음) 항상 엔트리 반환
//   - max-age가 -1이면 (no-store) 이 경우는 발생하지 않아야 함
//
// Get은 동시 사용에 안전함
func (c *Cache) Get(url string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[url]
	if !ok {
		return nil, false
	}

	// 엔트리 만료 여부 확인 (max-age)
	if entry.MaxAge > 0 {
		elapsed := time.Now().Unix() - entry.Timestamp
		if elapsed > int64(entry.MaxAge) {
			// 만료됨 - 캐시에서 제거
			delete(c.entries, url)
			logger.Logger.Printf("캐시 만료 (max-age=%ds, elapsed=%ds): %s", entry.MaxAge, elapsed, url)
			return nil, false
		}
	}

	logger.Logger.Printf("캐시에서 응답 반환: %s", url)
	return entry, true
}

// Put은 응답을 캐시에 저장함
//
// Cache-Control 헤더를 파싱하여 캐시 가능 여부를 판단함:
//   - no-store: 캐시하지 않음
//   - max-age=N: 만료 시간과 함께 캐시
//   - Cache-Control 없음 또는 지원하지 않는 지시어: 캐시하지 않음 (보수적)
//
// # HTTP 규격에 따라 GET 요청의 200 응답만 캐시함
//
// Put은 동시 사용에 안전함
func (c *Cache) Put(url string, statusCode int, body string, headers map[string]string) {
	// GET 요청의 200 응답만 캐시
	if statusCode != 200 {
		return
	}

	// Cache-Control 헤더 파싱
	cacheControl := headers["cache-control"]
	noStore, maxAge := parseCacheControl(cacheControl)

	// no-store인 경우 캐시하지 않음
	if noStore {
		logger.Logger.Printf("캐시하지 않음 (Cache-Control: no-store): %s", url)
		return
	}

	// Cache-Control 헤더가 없으면 기본적으로 캐시
	// max-age가 있으면 사용
	// 지원하지 않는 지시어가 있으면 (maxAge == -2) 캐시하지 않음
	if maxAge == -2 {
		logger.Logger.Printf("캐시하지 않음 (지원하지 않는 Cache-Control): %s", url)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &CacheEntry{
		Body:      body,
		Headers:   headers,
		Timestamp: time.Now().Unix(),
		MaxAge:    maxAge, // max-age 없으면 0, max-age=N이면 N
	}

	c.entries[url] = entry

	if maxAge > 0 {
		logger.Logger.Printf("응답 캐시 저장 (max-age=%ds): %s", maxAge, url)
	} else {
		logger.Logger.Printf("응답 캐시 저장 (무제한): %s", url)
	}
}

// Clear는 캐시의 모든 엔트리를 제거함
//
// 테스트할 때 또는 강제로 새로 가져오고 싶을 때 유용함
//
// Clear는 동시 사용에 안전함
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	logger.Logger.Println("캐시 전체 삭제")
}

// parseCacheControl은 Cache-Control 헤더를 파싱하고 다음을 반환함:
//   - noStore: "no-store" 지시어가 있으면 true
//   - maxAge: max-age 값(초 단위), 또는:
//     0 - max-age 지시어 없음 (만료 없음)
//     -1 - no-store (noStore를 먼저 확인해야 함)
//     -2 - 지원하지 않는 지시어가 있음
//
// 예시:
//   - "no-store" → (true, -1)
//   - "max-age=60" → (false, 60)
//   - "max-age=0" → (false, 0)
//   - "" → (false, 0)
//   - "private" → (false, -2)
//   - "must-revalidate" → (false, -2)
func parseCacheControl(cacheControl string) (noStore bool, maxAge int) {
	if cacheControl == "" {
		// Cache-Control 헤더 없음 - 기본적으로 캐시
		return false, 0
	}

	// 쉼표로 분리하여 여러 지시어 처리
	directives := strings.Split(cacheControl, ",")

	foundMaxAge := false
	maxAgeValue := 0

	for _, directive := range directives {
		directive = strings.TrimSpace(directive)

		if directive == "no-store" {
			return true, -1
		}

		if strings.HasPrefix(directive, "max-age=") {
			// max-age 값 파싱
			ageStr := strings.TrimPrefix(directive, "max-age=")
			age, err := strconv.Atoi(ageStr)
			if err == nil && age >= 0 {
				foundMaxAge = true
				maxAgeValue = age
			}
		} else if directive != "" && directive != "max-age" {
			// 지원하지 않는 지시어 (private, public, must-revalidate, no-cache 등)
			// 보수적으로 캐시하지 않음
			return false, -2
		}
	}

	if foundMaxAge {
		return false, maxAgeValue
	}

	// max-age도 no-store도 없음 - 만료 없이 캐시
	return false, 0
}

// GlobalCache is the global Cache instance used by the HTTP fetcher
var GlobalCache = NewCache()
