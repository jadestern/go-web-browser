package main

import (
	"fmt"
	"go-web-browser/net"
	"go-web-browser/url"
	"os"
	"strings"
)

// load: URL 문자열을 받아서 요청하고 화면에 표시하는 통합 함수
func load(urlStr string) {
	urlObj, err := url.NewURL(urlStr)
	if err != nil {
		fmt.Printf("URL 분석 에러 (%s): %v\n", urlStr, err)
		return
	}

	fmt.Printf("브라우징: %s\n", urlObj.String())

	body, err := net.Request(urlObj)
	if err != nil {
		fmt.Printf("요청 실패 (%s): %v\n", urlObj.String(), err)
		return
	}

	renderer := getRenderer(urlObj.Scheme)
	renderer.Render(body)
}

func main() {
	fmt.Println("=== Go Web Browser ===")
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
