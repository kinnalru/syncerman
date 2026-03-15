# Milestone 2: Refactor internal/logger/*

## Goal

Enhance logging system code quality and functionality

## Context

- Logging system provides structured output (OVERALL.md: lines 24)
- Supports multiple log levels (verbose, quiet)
- Integral to debugging and user feedback
- Current state: 3 Go files in internal/logger/

## Tasks

### Task 2.1: Analyze current logger package structure and identify refactoring opportunities

Review all 3 files in internal/logger/ package to understand current implementation:
- Analyze existing implementation and patterns
- Identify areas for improvement
- Check interface design and implementation patterns
- Note any performance bottlenecks
- Document findings for subsequent tasks

### Task 2.2: Refactor logger interface and implementation

Improve interface design following STYLE.md guidelines (lines 78-81):
- Ensure proper separation of concerns
- Enhance extensibility for future log levels/formatters
- Apply interface design principles (accept interfaces, return structs)
- Use concrete types where appropriate
- Define interfaces where they are used
- Don't define interfaces before they are used

### Task 2.3: Optimize log formatting and output handling

Enhance logging operations performance and consistency:
- Improve log message formatting consistency
- Optimize performance of logging operations
- Ensure proper log level handling
- Use `defer` for cleanup where applicable
- Keep functions small and focused (STYLE.md: lines 95)
- Consider using `sync.Pool` for object pooling if beneficial

### Task 2.4: Refactor error handling in logger package

Apply consistent error handling patterns:
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Improve error context and messages
- Ensure proper error propagation
- Return errors as the last return value
- Handle errors at the appropriate level
- Error strings should not be capitalized or end with punctuation (STYLE.md: lines 69)

### Task 2.5: Apply Go style guide and best practices

Ensure all code follows STYLE.md guidelines:
- Run `go fmt ./...` on logger package
- Fix naming convention violations
- Optimize imports and reduce duplication
- Ensure proper formatting (STYLE.md: lines 15-24)
- Add godoc comments for all exported declarations
- Use `const` for constant values, avoid magic numbers (STYLE.md: lines 152)
- Exported fields: PascalCase, Unexported fields: camelCase (STYLE.md: lines 107-108)

### Task 2.6: Update tests for refactored code

Refactor and enhance test coverage:
- Refactor existing tests to match code changes
- Ensure test coverage for all logging levels
- Add tests for edge cases and error conditions
- Use table-driven tests for multiple scenarios
- Test both happy path and error cases
- Ensure tests are independent and deterministic
- Use `t.Run()` for subtests

## Verification

- All tests pass: `go test ./internal/logger/... -v -cover`
- Code formatting: `go fmt ./internal/logger/...`
- Linting passes: `go vet ./internal/logger/...`
- Test coverage maintained or improved
- No regression in functionality
- Code adheres to STYLE.md guidelines
