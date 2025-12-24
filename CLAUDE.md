# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an educational web browser implementation in Go, built from scratch to learn how browsers work internally. The project follows a structured learning path from basic HTTP networking through HTML/CSS parsing to rendering.

**See learning_progress.md for detailed concepts learned and their locations in code.**

## Learning Workflow

This is a hands-on learning project. Claude assists with the learning process following these rules:

### 1. Working with LLM Files

When exploring or adding code:

- **File naming**: Create a `*_llm.go` file (e.g., `show_llm.go`) as a working copy
- **Naming convention**: Add `_llm` postfix to ALL type and function names to avoid conflicts
  - Example: `URL` → `URL_llm`, `NewURL` → `NewURL_llm`, `show` → `show_llm`
  - This allows the LLM file to be built and tested independently in the same folder
  - When integrating into actual files, simply remove the `_llm` postfix
- **Claude's role**:
  - Add, modify, and test code ONLY in `*_llm.go` files
  - Use `_llm` postfix for all identifiers (types, functions, methods)
  - Build and verify the code works correctly
  - Provide integration instructions with `_llm` removed
- **Student's role**: Manually review and apply changes to the actual implementation files
- **Important**: Claude must NEVER directly modify the actual implementation files (e.g., `url.go`). This ensures hands-on learning.

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
