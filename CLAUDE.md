# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an educational web browser implementation in Go, built from scratch to learn how browsers work internally. The project follows a structured learning path from basic HTTP networking through HTML/CSS parsing to rendering.

**See learning_progress.md for detailed concepts learned and their locations in code.**

## Learning Workflow

This is a hands-on learning project. Claude assists with the learning process following these rules:

### 1. Working with LLM Files

**CRITICAL: Use ONE main LLM file (`browser_llm.go`) for all work. DO NOT create multiple separate LLM files.**

When exploring or adding code:

- **File naming**: Work ONLY in `browser_llm.go` (the main LLM working file)
  - DO NOT create new files like `data_llm.go`, `show_llm.go`, etc.
  - All new features are added to the existing `browser_llm.go`

- **Naming convention**: Add `_llm` postfix to ALL type and function names to avoid conflicts
  - Example: `URL` → `URL_llm`, `NewURL` → `NewURL_llm`, `show` → `show_llm`
  - This allows the LLM file to be built and tested independently in the same folder
  - When integrating into actual files, simply remove the `_llm` postfix

- **Claude's workflow**:
  1. Read `browser_llm.go` to understand current implementation
  2. Modify `browser_llm.go` to add/change features (using `_llm` postfix)
  3. Build: `go build browser_llm.go`
  4. Test: `go run browser_llm.go <test-url>` or `.\browser_llm.exe <test-url>`
  5. If successful, provide integration instructions using the **Before/After format** (see below)

- **Student's role**: Manually review and apply changes to the actual implementation file (`browser.go`)

- **Important**:
  - Claude must NEVER directly modify `browser.go`. This ensures hands-on learning.
  - Claude must NEVER create separate `*_llm.go` files. Only work in `browser_llm.go`.

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
# Build the actual project file
go build url.go

# Run the program (Windows)
.\url.exe

# Run the program (Linux/Mac)
./url

# Run directly without building
go run url.go

# Build and test LLM working files (independent build)
go build show_llm.go
.\show_llm.exe

# Run tests (when test files exist)
go test -v
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
