package main

import (
	"fmt"
	"sync"
)

func testConcurrentRequests() {
	fmt.Println("=== 동시 요청 10개 테스트 (Pool 최대 6개) ===\n")

	testURL := "http://httpbin.org/get"

	var wg sync.WaitGroup

	// 동시에 10개 요청
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()

			urlObj, err := NewURL(testURL)
			if err != nil {
				fmt.Printf("[요청 %d] URL 파싱 에러: %v\n", num, err)
				return
			}

			body, err := urlObj.Request()
			if err != nil {
				fmt.Printf("[요청 %d] 요청 실패: %v\n", num, err)
				return
			}

			fmt.Printf("[요청 %d] 완료! 응답 길이: %d 바이트\n", num, len(body))
		}(i)
	}

	wg.Wait()

	fmt.Println("\n=== 동시 요청 완료 ===")
	fmt.Println("\n💡 예상 결과:")
	fmt.Println("  - 🆕 새 연결 생성: 10번")
	fmt.Println("  - 💾 연결 저장: 6번 (Pool 최대)")
	fmt.Println("  - 🔌 Pool 가득 차서 닫기: 4번 (초과분)")
}

func testSequentialRequests() {
	fmt.Println("\n\n=== 순차 요청 3개 테스트 (재사용 확인) ===\n")

	testURL := "http://httpbin.org/get"

	for i := 1; i <= 3; i++ {
		fmt.Printf("\n[요청 %d]\n", i)
		urlObj, _ := NewURL(testURL)
		urlObj.Request()
	}

	fmt.Println("\n=== 순차 요청 완료 ===")
	fmt.Println("\n💡 예상 결과:")
	fmt.Println("  - 요청 1: 🆕 새 연결, 💾 저장")
	fmt.Println("  - 요청 2: ♻️  재사용, 💾 저장")
	fmt.Println("  - 요청 3: ♻️  재사용, 💾 저장")
}

func main() {
	testConcurrentRequests()
	testSequentialRequests()
}
