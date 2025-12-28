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

When providing integration instructions, use the **Before/After format** with detailed explanations:

**Structure for each change:**
1. **Feature description**: What functionality is being added/changed
2. **Why it's needed**: Explain the purpose and reasoning
3. **Before code**: Show the original code from `browser.go`
4. **After code**: Show what the code should look like with changes (remove `_llm` postfix)

**Example format:**

```markdown
### Change 1: [Feature Name]

**What:** Brief description of the feature
**Why:** Explanation of why this change is necessary

**Before:**
```go
// Original code from browser.go
func OriginalFunction() {
    // existing code
}
```

**After:**
```go
// Modified code (without _llm postfix)
func OriginalFunction() {
    // new code added
    // existing code
}
```
```

**Key principles:**
- Don't just say "add this at line X" - explain the **purpose and context**
- Show **complete code blocks** for Before/After, not just snippets
- Explain **why** each change is necessary for understanding
- Include **comments** in the After code to guide the student
- For new methods/functions, show where they should be placed relative to existing code

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
