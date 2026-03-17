---
title: "Milestone 7: Refactor internal/cmd"
status: "⌛ In Progress"
---

# Milestone 7: Refactor internal/cmd

## Goal

Refactor internal/cmd package to improve code quality, eliminate duplication, consolidate flag handling, and ensure all code follows Go style guidelines. Review CLI command structure, improve error handling consistency, and verify all tests are comprehensive.

## Context

The internal/cmd package provides CLI command definitions and execution logic using Cobra framework. Current state:

- 8 files: root.go (167 lines), sync.go (143 lines), check.go (101 lines), config.go (28 lines), version.go (26 lines), doc.go (56 lines), root_test.go (201 lines), integration_test.go (397 lines)
- Commands: root, sync, check, version
- Uses Cobra framework for CLI
- CommandConfig struct manages configuration and logging
- ExitCodeError struct for error handling with exit codes
- Logger already consolidated - uses internal/logger.ConsoleLogger
- Tests pass (all tests in package)
- Duplicate executor/engine creation in check.go and sync.go
- Repeated logger retrieval and context creation patterns

Reference: guides/STYLE.md, guides/PLANING.md, guides/OVERALL.md, PLAN_7.md lines 170-184, AGENTS.md

## Tasks

### 1. Consolidate Executor and Engine Creation

Extract duplicate executor/engine creation logic into reusable helper functions. Currently both check.go and sync.go create executor and engine instances with same pattern. Consolidate this to eliminate duplication.

Actions:
- Create helper function to create executor with rclone.NewExecutor(rclone.NewConfig())
- Create helper function to create engine with sync.NewEngine(cfg, executor, logger)
- Update check.go to use helper functions
- Update sync.go to use helper functions
- Ensure code remains functionally equivalent

### 2. Remove Unused Code and Verify All Functions are Used

Analyze all functions and types in cmd package to verify they are actively used across the codebase. Remove any unused functions, types, or variables.

Actions:
- Verify CommandConfig struct and all methods are used
- Check exit code constants are referenced
- Verify utility functions (GetLogger, GetConfig, GetConfigFile, IsDryRun, IsVerbose, IsQuiet) are used
- Search entire codebase for usage of all exported identifiers
- Remove any unused code or uncomment if all are necessary
- Update imports after any removals

### 3. Review and Consolidate Flag Handling Logic

Review all flag definitions and handling across command modules to ensure proper consolidation and consistency.

Actions:
- Verify all global flags are properly defined in root.go
- Confirm no duplicate flag definitions exist
- Check flag handling logic for --verbose, --quiet, --dry-run conflicts
- Ensure flag setters follow consistent patterns
- Verify flag documentation is accurate and up to date

### 4. Review and Improve Command Organization

Analyze command structure and identify opportunities to improve organization, readability, and maintainability.

Actions:
- Review command separation between root.go, sync.go, check.go, version.go
- Consider if any commands should be further separated
- Verify command initialization order and dependencies
- Check for opportunities to extract common command patterns
- Ensure command naming follows Cobra conventions

### 5. Simplify Code and Apply DRY Principle

Identify and eliminate code duplication across cmd package. Extract repeated patterns into helper functions.

Actions:
- Extract repeated logger retrieval pattern into inline use
- Consolidate context creation pattern if duplicated
- Extract repeated error handling patterns if any
- Simplify complex logic where present
- Reduce unnecessary intermediate variables
- Ensure each function has single responsibility

### 6. Verify All Tests Exist and Improve Coverage

Ensure every production file in cmd package has corresponding comprehensive test coverage.

Actions:
- Verify root.go has test coverage (root_test.go) - present
- Verify sync.go has test coverage (integration_test.go covers sync command) - present
- Verify check.go has test coverage (integration_test.go covers check command) - present
- Verify config.go has test coverage (tested via integration tests) - present
- Verify version.go has test coverage (integration_test.go covers version command) - present
- Check if any files lack adequate test coverage
- Add missing tests if needed

### 7. Update Command and README Documentation

Review and update all command documentation to ensure accuracy with current implementation. Update README.md if needed.

Actions:
- Review command help text in root.go Long field
- Update doc.go package documentation to reflect current state
- Verify all examples in documentation work correctly
- Check consistency between doc.go, root.go help, and README.md
- Update README.md CLI section if needed
- Ensure documentation follows Go documentation conventions

### 8. Apply Code Style Guidelines

Ensure all code in the cmd package follows the project Go style guidelines.

Actions:
- Ensure all exported types and functions have godoc comments
- Verify naming conventions follow guidelines
- Check error handling follows project conventions
- Run go fmt, go vet, goimports, golangci-lint
- Fix all issues found by linting tools
- Ensure code formatting is consistent
