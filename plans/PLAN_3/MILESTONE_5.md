# Milestone 5: Refactor internal/cmd/*

## Goal

Improve CLI command structure and handler quality

## Context

- CLI commands use Cobra framework (OVERALL.md: lines 20)
- Commands need to be well-structured (OVERALL.md: lines 125-249)
- Primary user interaction layer
- Current state: 6 Go files in internal/cmd/

## Tasks

### Task 5.1: Analyze current cmd package structure and identify refactoring opportunities

Review all 6 files in internal/cmd/ package to understand current implementation:
- Review 6 existing command files
- Identify common patterns in command handlers
- Map out command hierarchy and flow
- Identify duplication across commands
- Note flag handling patterns
- Document findings for subsequent tasks

### Task 5.2: Refactor command definitions and initialization

Improve command structure consistency:
- Improve command structure consistency
- Optimize flag definitions and validation
- Enhance command help and documentation
- Ensure proper separation of concerns
- Apply consistent patterns across all commands

### Task 5.3: Refactor command handler logic

Simplify command handler implementations:
- Simplify command handler implementations
- Improve error handling and user feedback
- Enhance command execution flow
- Apply consistent error wrapping patterns
- Ensure proper error propagation

### Task 5.4: Optimize shared command functionality

Extract common patterns for reusability:
- Extract common patterns into reusable functions
- Improve flag handling and validation
- Enhance global flag integration (OVERALL.md: lines 116-121)
- Reduce code duplication across commands
- Ensure consistent behavior across commands

### Task 5.5: Apply Go style guide and best practices

Ensure all code follows STYLE.md guidelines:
- Run `go fmt ./...` on cmd package
- Fix naming convention violations across all 6 files
- Optimize imports and code organization
- Ensure proper formatting (STYLE.md: lines 15-24)
- Add godoc comments for all exported declarations
- Use `const` for constant values (STYLE.md: lines 152)

### Task 5.6: Update tests for refactored code

Refactor and enhance test coverage:
- Refactor existing tests for commands
- Ensure test coverage for all command variants
- Add tests for flag combinations (OVERALL.md: lines 116-121)
- Test various CLI scenarios (OVERALL.md: lines 207-249)
- Use table-driven tests for multiple scenarios
- Test both happy path and error cases

## Verification

- All tests pass: `go test ./internal/cmd/... -v -cover`
- Code formatting: `go fmt ./internal/cmd/...`
- Linting passes: `go vet ./internal/cmd/...`
- Test coverage maintained or improved
- No regression in functionality
- Code adheres to STYLE.md guidelines
