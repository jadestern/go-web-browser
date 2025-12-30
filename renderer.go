package main

import (
	"fmt"
	"go-web-browser/url"
)

type Renderer interface {
	Render(content string)
}

type HTMLRenderer struct{}

func (h *HTMLRenderer) Render(content string) {
	show(content)
}

type SourceRenderer struct{}

func (s *SourceRenderer) Render(content string) {
	fmt.Print(content)
}

var rendererRegistry = map[url.Scheme]Renderer{
	url.SchemeViewSource: &SourceRenderer{},
}

func getRenderer(scheme url.Scheme) Renderer {
	if renderer, ok := rendererRegistry[scheme]; ok {
		return renderer
	}
	return &HTMLRenderer{}
}
