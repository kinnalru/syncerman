# Milestone 3: Refactor internal/rclone/*

## Goal

Improve rclone integration layer code quality and maintainability

## Context

- Rclone package executes complex commands (OVERALL.md: lines 250-310)
- Handles output parsing and error detection
- Critical for sync operations
- Current state: 14 Go files in internal/rclone/

## Tasks

### Task 3.1: Analyze current rclone package structure and identify refactoring opportunities

Review all 14 files in internal/rclone/ package to understand current implementation:
- Review existing files and their relationships
- Identify complex patterns that need simplification
- Map out current architecture
- Identify duplication across 14 files
- Note performance bottlenecks
- Document findings for subsequent tasks

### Task 3.2: Refactor command building and execution logic

Simplify and optimize command construction:
- Extract command construction into cleaner functions
- Improve argument handling and validation
- Optimize command execution patterns
- Keep functions small and focused (STYLE.md: lines 95)
- Use multiple return values, error is last (STYLE.md: lines 96)
- Limit parameters to 3-4, consider struct for more (STYLE.md: lines 97)

### Task 3.3: Refactor output parsing and error detection

Enhance parsing logic for rclone output:
- Improve parsing logic for rclone output (OVERALL.md: lines 311-337)
- Enhance first-run error detection with REGEXP (OVERALL.md: lines 321-326)
- Simplify error pattern matching
- Extract parsing logic into reusable functions
- Ensure proper error context in all parsed errors

### Task 3.4: Improve rclone verification and directory creation

Refactor verification and directory handling:
- Refactor remote verification logic
- Optimize directory creation handling
- Improve error handling for rclone operations
- Apply consistent error wrapping patterns
- Ensure all errors have proper context

### Task 3.5: Apply Go style guide and best practices

Ensure all code follows STYLE.md guidelines:
- Run `go fmt ./...` on rclone package
- Fix naming convention violations across all 14 files
- Optimize imports and reduce code duplication
- Ensure proper formatting (STYLE.md: lines 15-24)
- Use `const` for constant values, avoid magic numbers (STYLE.md: lines 152)
- Add godoc comments for all exported declarations

### Task 3.6: Enhance error handling and context

Apply consistent error handling patterns:
- Wrap all errors with context using `fmt.Errorf("context: %w", err)`
- Improve error messages with context
- Ensure proper error propagation
- Return errors as the last return value
- Handle errors at the appropriate level
- Use `errors.Is()` and `errors.As()` for error comparisons where needed

### Task 3.7: Update tests for refactored code

Refactor and enhance test coverage for 14 files:
- Refactor existing tests to match code changes
- Ensure test coverage for command building and execution
- Add tests for error patterns and edge cases
- Use table-driven tests for multiple scenarios
- Test both happy path and error cases
- Ensure tests are independent and deterministic

## Verification

- All tests pass: `go test ./internal/rclone/... -v -cover`
- Code formatting: `go fmt ./internal/rclone/...`
- Linting passes: `go vet ./internal/rclone/...`
- Test coverage maintained or improved
- No regression in functionality
- Code adheres to STYLE.md guidelines
