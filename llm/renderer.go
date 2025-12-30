package main

import (
	"fmt"
	"go-web-browser/llm/url"
)

// Renderer 인터페이스: 콘텐츠 렌더링을 추상화
type Renderer interface {
	Render(content string)
}

// HTMLRenderer: HTML을 파싱해서 텍스트만 렌더링
type HTMLRenderer struct{}

func (h *HTMLRenderer) Render(content string) {
	show(content)
}

// SourceRenderer: 원본 소스를 그대로 렌더링
type SourceRenderer struct{}

func (s *SourceRenderer) Render(content string) {
	fmt.Print(content)
}

// rendererRegistry: scheme에 따른 Renderer를 등록하는 레지스트리
var rendererRegistry = map[url.Scheme]Renderer{
	url.SchemeViewSource: &SourceRenderer{},
}

// getRenderer: scheme에 맞는 Renderer 반환, 기본은 HTMLRenderer
func getRenderer(scheme url.Scheme) Renderer {
	if renderer, ok := rendererRegistry[scheme]; ok {
		return renderer
	}
	return &HTMLRenderer{} // 기본 렌더러
}
