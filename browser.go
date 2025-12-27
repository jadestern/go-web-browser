package main

import (
	"fmt"
	"os"
	"strings"
)

// load: URL 문자열을 받아서 요청하고 화면에 표시하는 통합 함수
func load(urlStr string) {
	urlObj, err := NewURL(urlStr)
	if err != nil {
		fmt.Println("분석 에러: ", err)
		return
	}

	body, err := urlObj.Request()
	if err != nil {
		fmt.Println("요청 에러: ", err)
		return
	}

	renderer := getRenderer(urlObj.Scheme)
	renderer.Render(body)
}

func main() {
	var urlStr string

	if len(os.Args) < 2 {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("현재 디렉토리를 가져올 수 없습니다: ", err)
		}

		urlStr = fmt.Sprintf("file:///%s/index.html", strings.ReplaceAll(cwd, "\\", "/"))
		fmt.Printf("기본 파일 열기: %s\n", urlStr)
	} else {
		urlStr = os.Args[1]
	}

	load(urlStr)
}
