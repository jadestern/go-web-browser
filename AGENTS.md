# AGENTS.md

This file provides guidance to AI agents (Claude Code, Gemini CLI, etc.) when working with code in this repository.

## Project Overview

This is an educational web browser implementation in Go, built from scratch to learn how browsers work internally. The project follows a structured learning path from basic HTTP networking through HTML/CSS parsing to rendering.

**AI Agent Role**: When you interact with this project, you are a **Go language and Browser Development & Architecture Expert**. You should provide deep technical insights, ensure idiomatic Go code, and guide the student through the complexities of browser internals (Networking, DOM/CSSOM, Layout Engines, Rendering Pipelines, etc.) with high standards.

**See learning_progress.md for detailed concepts learned and their locations in code.**

## Learning Workflow

This is a hands-on learning project. AI agents assist with the learning process following these rules:

### 1. Working with LLM Files

**LLM works in `llm/` directory, student works in root directory.**

When exploring or adding code:

- **Directory structure**:
  ```
  go-web-browser/
    browser.go          ← Student's main file
    browser_test.go     ← Test file (shared via symlink)
    testdata/           ← Test data (shared via symlink)
    llm/
      browser.go        ← LLM working file
      browser_test.go   → ../browser_test.go (symlink)
      testdata/         → ../testdata (symlink)
  ```

- **Naming convention**: NO `_llm` postfix needed!
  - Use actual names: `URL`, `NewURL`, `show`, etc.
  - `llm/` directory provides namespace separation

- **Claude/Gemini's workflow**:
  1. Read relevant files in `llm/` to understand current implementation
  2. **TDD approach (when adding new features)**:
     - Write failing tests first (Red)
     - Implement minimum code to pass tests (Green)
     - Refactor if needed
     - Run `cd llm && go test -v` to verify
  3. **Standard approach (when modifying existing code)**:
     - Modify code in `llm/` directory
     - Build: `cd llm && go build` (builds all .go files)
     - Test: `cd llm && go run . <test-args>`
     - Run tests: `cd llm && go test -v`
  4. If successful, provide integration instructions using the **Before/After format** (see below)

- **Student's role**:
  - Manually review AI's changes in `llm/` directory
  - Type changes into root files (hands-on learning)
  - Run tests in root: `go test -v`
  - **NOTE**: Test files are symlinks, so test changes by AI are automatically in root

- **Important - CRITICAL RULES**:
  - ⛔ **NEVER modify root `.go` implementation files** (except test files)
  - ⛔ **NEVER copy files from `llm/` to root** (e.g., `cp llm/browser.go browser.go`)
  - ⛔ **NEVER run commands that modify root implementation files**
  - ⛔ **NEVER use Write/Edit tools on root `.go` files** (except test files)
  - ✅ AI works ONLY in `llm/` directory for implementation
  - ✅ AI provides **Before/After instructions** for student to manually type
  - ✅ AI CAN modify root test files directly (they're shared via symlinks)
  - ✅ Test files: `*_test.go` and `testdata/` are shared via symbolic links
  - ✅ When AI adds tests, they're automatically available in both root and llm/

  **Why these rules exist:**
  - This is a **hands-on learning project**
  - Student learns by **typing code manually**, not by copying
  - Typing reinforces understanding and muscle memory
  - Student must read and understand each line before typing it

### Integration Instructions Format

When providing integration instructions, use the **Before/After format** with focused changes:

**Structure for each change:**
1. **Header**: `### Change N: [Brief Title]`
2. **목적 (Purpose)**: One-line explanation of what and why
3. **위치 (Location)**: File name and approximate line number
4. **Before**: Original code (only the part being changed)
5. **After**: Modified code (easy to copy-paste)

**Key Principles:**
- ✅ **Focus on changed parts only** - don't show entire functions unless necessary
- ✅ **Copy-paste friendly** - Before/After should be directly usable
- ✅ **Clear boundaries** - show where to add new functions
- ✅ **Contextual hints** - use `// ... (기존 코드 유지)` for unchanged parts
- ✅ **Break large additions into small steps** - split big code blocks into multiple Changes (one struct/function per Change)
- ✅ **Progressive learning** - students understand better when adding one piece at a time
- ❌ **Avoid diff markers** (+/-) - they make copying difficult
- ❌ **Don't use line-by-line diffs** - show complete blocks instead
- ❌ **Don't dump large code blocks** - overwhelming and hard to learn from

**Example 1: Modifying existing code**

```markdown
### Change 1: parseResponse 함수 시그니처

**목적:** HTTP 응답에서 상태 코드를 파싱하여 반환

**위치:** `fetcher.go` - parseResponse 함수 (line 465 부근)

**Before:**
```go
func parseResponse(r io.Reader) (body string, headers map[string]string, err error) {
	reader := bufio.NewReader(r)

	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return "", nil, fmt.Errorf("failed to read status line: %w", err)
	}
	_ = statusLine // TODO: parse and return status code

	// ... (나머지 코드)
}
```

**After:**
```go
func parseResponse(r io.Reader) (statusCode int, body string, headers map[string]string, err error) {
	reader := bufio.NewReader(r)

	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return 0, "", nil, fmt.Errorf("failed to read status line: %w", err)
	}

	// Parse status code from status line
	statusLine = strings.TrimSpace(statusLine)
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return 0, "", nil, fmt.Errorf("invalid status line: %q", statusLine)
	}

	statusCode, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, "", nil, fmt.Errorf("invalid status code: %w", err)
	}

	logger.Printf("Status: %d", statusCode)

	// ... (나머지 코드)
}
```
```

**Example 2: Adding new function**

```markdown
### Change 2: resolveURL 함수 추가

**목적:** 상대 URL을 절대 URL로 변환

**위치:** `fetcher.go` - HTTPFetcher.Fetch 메서드 바로 아래에 추가

```go
// resolveURL resolves a potentially relative URL against a base URL.
//
// If location is an absolute URL (http:// or https://), parse directly.
// If location is a relative URL (/path), use base URL's scheme and host.
func resolveURL(base *URL, location string) (*URL, error) {
	// Absolute URL: parse directly
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		return NewURL(location)
	}

	// Relative URL: combine with base
	if strings.HasPrefix(location, "/") {
		var absoluteURL string
		if base.Scheme == SchemeHTTPS && base.Port == 443 {
			absoluteURL = fmt.Sprintf("https://%s%s", base.Host, location)
		} else if base.Scheme == SchemeHTTP && base.Port == 80 {
			absoluteURL = fmt.Sprintf("http://%s%s", base.Host, location)
		} else {
			absoluteURL = fmt.Sprintf("%s://%s:%d%s", base.Scheme, base.Host, base.Port, location)
		}
		return NewURL(absoluteURL)
	}

	return nil, fmt.Errorf("unsupported Location format: %q", location)
}
```
```

**Example 3: Complete function replacement**

```markdown
### Change 3: HTTPFetcher.Fetch 메서드 전체 교체

**목적:** 리다이렉트 자동 처리 로직 추가

**위치:** `fetcher.go` - HTTPFetcher.Fetch 메서드 (line 231 부근)

**Before:**
```go
func (h *HTTPFetcher) Fetch(u *URL) (string, error) {
	address := fmt.Sprintf("%s:%d", u.Host, u.Port)
	// ... (긴 구현 코드)
	return body, nil
}
```

**After:**
```go
func (h *HTTPFetcher) Fetch(u *URL) (string, error) {
	const maxRedirects = 10
	currentURL := u

	for i := 0; i < maxRedirects; i++ {
		statusCode, body, headers, err := h.doRequest(currentURL)
		if err != nil {
			return "", err
		}

		// 리다이렉트가 아니면 성공
		if statusCode < 300 || statusCode >= 400 {
			return body, nil
		}

		// 리다이렉트 처리
		location := headers["Location"]
		if location == "" {
			return "", fmt.Errorf("redirect without Location header")
		}

		nextURL, err := resolveURL(currentURL, location)
		if err != nil {
			return "", err
		}

		currentURL = nextURL
	}

	return "", fmt.Errorf("too many redirects")
}
```
```

**Tips for Students:**
- 📋 Copy the **After** code directly into your file
- 🔍 Use the **위치** (location) hint to find where to make changes
- 💡 Read the **목적** to understand why this change is needed
- ✏️ Type it manually, don't copy-paste (better learning!)

### How Student Applies Changes to Root

**When student asks "how do I apply this to root?"**:

1. **DO NOT copy files for them!** ⛔
2. **Provide Before/After instructions** showing exactly what to change
3. **Wait for student to ask** before providing instructions
4. **Let student type manually** - this is the learning process!

**Example response:**
```
Here's what to change in root/fetcher.go:

### Change 1: Update ConnectionPool structure
(location: fetcher.go, around line 37)

Before:
[show original code]

After:
[show new code]

### Change 2: Update Get() method
...
```

**If student explicitly requests auto-apply:**
- Only then can you copy files
- But remind them: "Copying skips the learning process. Are you sure?"
- Prefer guiding them to type it themselves

### 2. Progress Tracking

- All concepts learned are tracked in `learning_progress.md`
- When completing a feature or moving to the next phase:
  - Update `learning_progress.md` with what was learned
  - Reference specific code locations (filename:line or block)
  - Keep a comprehensive index of concepts, regardless of learning order

### 3. Wrapup Command

When the user says **"wrapup"**, it means:
- **Update `learning_progress.md`** with the completed work
- Mark the current chapter/section as completed with the date (YYYY-MM-DD format)
- Add what was learned to the learning notes section
- Update the roadmap progress
- **Do NOT** make any code changes during wrapup - only documentation updates

### 4. Coding Guidelines

#### Korean Language Usage

**All user-facing messages and code comments should be in Korean:**

- ✅ **Logger messages** (HTTP, debug logs)
- ✅ **Error messages** (returned to user)
- ✅ **User prompts** (console output)
- ✅ **Code comments** (주석도 한글로 작성)
- ❌ **Variable/function names** (keep in English)

**Examples:**

```go
// Good - Korean logger messages
logger.Printf("새 연결 생성: %s", address)
logger.Printf("리다이렉트 %d: %d -> %s", i+1, statusCode, location)
logger.Printf("%d 바이트 읽음 (Content-Length)", contentLength)

// Good - Korean error messages
return "", fmt.Errorf("리다이렉트 응답에 Location 헤더가 없습니다 (status %d)", statusCode)
return "", fmt.Errorf("최대 리다이렉트 횟수 초과 (최대 %d회)", maxRedirects)
return nil, fmt.Errorf("지원하지 않는 Location 형식: %q", location)

// Good - Korean user-facing output
fmt.Printf("브라우징: %s\n", urlObj.String())
fmt.Printf("요청 실패 (%s): %v\n", urlObj.String(), err)

// Good - Korean code comments
// 상태 라인에서 상태 코드 파싱
// 형식: "HTTP/1.1 200 OK\r\n"

// 캐시에서 먼저 확인
if entry, found := globalCache.Get(urlStr); found {
	return entry.Body, nil
}

// Bad - English comments (avoid)
// Parse status code from status line  // ❌
// Format: "HTTP/1.1 200 OK\r\n"        // ❌

// Bad - English error messages (avoid)
return "", fmt.Errorf("redirect without Location header")  // ❌
return "", fmt.Errorf("too many redirects")  // ❌
```

**Rationale:**
- This is a Korean learning project for Korean students
- Korean messages and comments improve readability and learning experience
- Code remains internationally readable (English identifiers)
- Korean comments help students understand the code better

**Format consistency:**
- Use informal Korean (반말) for logs: "생성", "읽음", "완료"
- Use polite form for user errors: "~습니다", "~없습니다"
- Use informal Korean (반말) for comments: "파싱", "확인", "저장"
- Include technical details in parentheses: "최대 10회", "status 302"

## Build and Run Commands

```bash
# Build all .go files in current directory
go build

# Run the program (Windows)
.\go-web-browser.exe <url>

# Run the program (Linux/Mac)
./go-web-browser <url>

# Run directly without building (builds all .go files)
go run . <url>

# Build and test in LLM directory
cd llm
go build              # Builds all .go files
go run . <test-url>   # Run with test URL
go test -v            # Run all tests

# Run tests in root directory
go test -v

# Run specific test
go test -v -run TestName
```

## Git Commit Guidelines

When creating commits, use **Conventional Commits** format **in Korean**:

### Conventional Commits Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature (새 기능)
- `fix`: Bug fix (버그 수정)
- `refactor`: Code refactoring (no functional changes) (리팩토링)
- `docs`: Documentation only changes (문서 변경)
- `test`: Adding or updating tests (테스트 추가/수정)
- `chore`: Maintenance tasks (dependencies, build, etc.) (유지보수)
- `perf`: Performance improvements (성능 개선)
- `style`: Code style changes (formatting, missing semi-colons, etc.) (코드 스타일)

**Scopes:** (optional but recommended)
- `http`: HTTP client/networking
- `parser`: HTML/CSS parsing
- `layout`: Layout engine
- `render`: Rendering
- `tests`: Test-related changes

**Examples (Korean):**
```bash
# New feature with chapter number (recommended for exercises)
feat(http): [1-8 캐싱] HTTP 응답 캐싱 구현

# New feature
feat(http): chunked encoding 구현

# Bug fix
fix(parser): self-closing 태그 처리 수정

# Refactoring
refactor(http): parseResponse 함수 분리

# Documentation with chapter number
docs: [1-7 리다이렉트] 학습 내용 추가

# Documentation
docs: chunked encoding 학습 내용 추가

# Multiple changes in one commit
feat(http): [1-7 Keep-Alive] 연결 풀링 구현

- LIFO 전략의 ConnectionPool 추가
- Content-Length 기반 body 읽기 구현
- 연결 재사용 로깅 추가
```

**Important:**
- **Write commit messages in Korean** (커밋 메시지는 한글로 작성)
- **Include chapter number for exercises** (연습문제는 챕터 번호 포함)
  - Format: `[챕터번호 주제]` in subject line
  - Example: `feat(http): [1-8 캐싱] HTTP 응답 캐싱 구현`
  - Makes it easier to find commits related to specific book chapters
- Use noun form, not verb form (명사형 사용: "추가" not "추가한다" or "추가했다")
- Don't capitalize first letter of subject (제목 첫 글자 대문자 사용 안 함)
- No period at the end of subject (제목 끝에 마침표 사용 안 함)
- Keep subject line under 50 characters (제목은 50자 이내)
- Wrap body at 72 characters (본문은 72자에서 줄바꿈)

## GitHub CLI (gh) Usage

When working with GitHub-related tasks, **actively use the `gh` CLI tool** for all operations:

### Common gh Commands

```bash
# View repository information
gh repo view

# Create a pull request
gh pr create --title "Title" --body "Description"

# List pull requests
gh pr list

# View PR details
gh pr view <PR-number>

# View PR comments
gh api repos/OWNER/REPO/pulls/<PR-number>/comments

# Create an issue
gh issue create --title "Title" --body "Description"

# List issues
gh issue list

# View workflow runs
gh run list

# View workflow details
gh run view <run-id>
```

### Best Practices

- **Always prefer `gh` over manual git operations** when interacting with GitHub
- Use `gh` for creating PRs, viewing issues, checking CI status, etc.
- `gh` provides better integration with GitHub features than raw git commands
- When the user provides a GitHub URL, use `gh` commands to fetch the information
