package main

import "testing"

// TestParseHTML_BasicTag 기본 태그 제거 테스트
func TestParseHTML_BasicTag(t *testing.T) {
	input := "<h1>Hello</h1>"
	expected := "Hello"

	result := parseHTML(input)

	if result != expected {
		t.Errorf("parseHTML(%q) = %q; want %q", input, result, expected)
	}
}

// TestParseHTML_HTMLEntities HTML 엔티티 변환 테스트
func TestParseHTML_HTMLEntities(t *testing.T) {
	input := "&lt;div&gt;"
	expected := "<div>"

	result := parseHTML(input)

	if result != expected {
		t.Errorf("parseHTML(%q) = %q; want %q", input, result, expected)
	}
}

// TestParseHTML_MixedContent 태그와 엔티티 혼합 테스트
func TestParseHTML_MixedContent(t *testing.T) {
	input := "<p>&lt;code&gt;&amp;&lt;/code&gt;</p>"
	expected := "<code>&</code>"

	result := parseHTML(input)

	if result != expected {
		t.Errorf("parseHTML(%q) = %q; want %q", input, result, expected)
	}
}

// TestParseHTML_NoTags 태그 없는 텍스트 테스트
func TestParseHTML_NoTags(t *testing.T) {
	input := "Hello world!"
	expected := "Hello world!"

	result := parseHTML(input)

	if result != expected {
		t.Errorf("parseHTML(%q) = %q; want %q", input, result, expected)
	}
}

// TestParseHTML_MultipleTags 여러 태그 테스트
func TestParseHTML_MultipleTags(t *testing.T) {
	input := "<h1>Title</h1><p>Paragraph</p>"
	expected := "TitleParagraph"

	result := parseHTML(input)

	if result != expected {
		t.Errorf("parseHTML(%q) = %q; want %q", input, result, expected)
	}
}
