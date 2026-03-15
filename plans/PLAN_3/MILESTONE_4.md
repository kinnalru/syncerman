# Milestone 4: Refactor internal/sync/*

## Goal

Enhance sync engine code quality and orchestration

## Context

- Sync engine handles sequential processing (OVERALL.md: lines 376-383)
- Manages error handling and retry logic
- Core orchestration layer
- Current state: 11 Go files in internal/sync/

## Tasks

### Task 4.1: Analyze current sync package structure and identify refactoring opportunities

Review all 11 files in internal/sync/ package to understand current implementation:
- Review existing files and their orchestration patterns
- Identify complex logic that needs simplification
- Map out sync flow and dependencies
- Identify duplication across 11 files
- Note error handling patterns
- Document findings for subsequent tasks

### Task 4.2: Refactor sync orchestration logic

Simplify sequential processing flow:
- Improve sequential processing flow (OVERALL.md: lines 376-383)
- Improve error handling and propagation
- Enhance retry and recovery mechanisms
- Keep functions small and focused (STYLE.md: lines 95)
- Use `defer` for cleanup where applicable (STYLE.md: lines 100)

### Task 4.3: Refactor target sync execution

Improve sync command building and execution:
- Improve sync command building and execution
- Optimize first-run error handling (OVERALL.md: lines 311-337)
- Enhance dry-run mode support
- Extract common patterns into reusable functions
- Ensure proper error context in all operations

### Task 4.4: Optimize sync state management

Enhance tracking and reporting:
- Improve tracking of sync operations
- Enhance progress reporting
- Simplify status management
- Apply consistent patterns across all 11 files
- Ensure proper separation of concerns

### Task 4.5: Apply Go style guide and best practices

Ensure all code follows STYLE.md guidelines:
- Run `go fmt ./...` on sync package
- Fix naming convention violations across all 11 files
- Optimize imports and code organization
- Ensure proper formatting (STYLE.md: lines 15-24)
- Add godoc comments for all exported declarations
- Use `const` for constant values (STYLE.md: lines 152)

### Task 4.6: Update tests for refactored code

Refactor and enhance test coverage for 11 files:
- Refactor existing tests to match code changes
- Ensure test coverage for sync scenarios
- Add tests for error cases and edge conditions
- Use table-driven tests for multiple scenarios
- Test both happy path and error cases
- Ensure tests are independent and deterministic

## Verification

- All tests pass: `go test ./internal/sync/... -v -cover`
- Code formatting: `go fmt ./internal/sync/...`
- Linting passes: `go vet ./internal/sync/...`
- Test coverage maintained or improved
- No regression in functionality
- Code adheres to STYLE.md guidelines
