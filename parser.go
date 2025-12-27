package main

import (
	"fmt"
	"html"
	"strings"
)

// parseHTML: HTML 태그를 제거하고 텍스트만 추출하는 순수 함수
func parseHTML(body string) string {
	// 태그를 제거하고 텍스트만 추출
	inTag := false
	var textBuilder strings.Builder

	for _, c := range body {
		if c == '<' {
			inTag = true
		} else if c == '>' {
			inTag = false
		} else if !inTag {
			// 태그 안이 아닐 때만 텍스트 수집
			textBuilder.WriteRune(c)
		}
	}

	text := html.UnescapeString(textBuilder.String())

	return text
}

// show: HTML을 파싱하고 결과를 출력하는 함수
func show(body string) {
	fmt.Print(parseHTML(body))
}
