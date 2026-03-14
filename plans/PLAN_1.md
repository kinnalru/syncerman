# PLAN_1: Syncerman Application Development Plan

## Overview

Building a CLI application for synchronizing targets using rclone, with YAML-based configuration and flexible sync operations.

---

## Milestones

### Milestone 1: Project Foundation and Core Structure
**Goal**: Establish project structure, CLI framework, and base utilities

**Context**:
- Existing stub in `main.go` needs expansion
- Follow Go style guidelines from `guides/STYLE.md`
- Set up internal package structure for maintainability

**Status**: Completed

---

### Milestone 2: Configuration System
**Goal**: Implement YAML configuration loading and validation

**Context**:
- Configuration format specified with provider/path/destination structure
- Optional fields: `args` (array), `resync` (bool)
- Need robust validation and error messages

**Status**: Completed

---

### Milestone 3: Rclone Integration Foundation
**Goal**: Build rclone command execution and verification layer

**Context**:
- Must execute rclone bisync with specific flags
- Need to check rclone remotes via `rclone listremotes`
- Must create destination directories with `rclone mkdir`

**Status**: Pending

---

### Milestone 4: Sync Execution Engine
**Goal**: Build core synchronization logic with error handling

**Context**:
- Sequential sync processing required
- Must detect and handle "cannot find prior listings" error
- First-run handling with --resync flag
- Dry-run mode support

**Status**: Pending

---

### Milestone 5: CLI Commands Implementation
**Goal**: Implement all CLI command variants

**Context**:
- Need commands: sync all, sync specific target, check config, check remotes
- Support dry-run flag across all sync operations
- Provide clear user feedback

**Status**: Pending

---

### Milestone 6: Testing and Quality Assurance
**Goal**: Ensure code quality with comprehensive testing

**Context**:
- Follow style guide requirements
- Run lint, fmt, vet before commits
- Test critical paths including error scenarios

**Status**: Pending

---

## Verification Strategy

Each milestone will be verified:
- All tests pass (go test ./... -v -cover)
- Code formatting (go fmt, goimports)
- Linting passes (golangci-lint run, go vet)
- Binary builds successfully (make build)
- Manual verification of CLI functionality

---

## Dependencies

- yaml v3 (gopkg.in/yaml.v3)
- cobra or similar CLI framework
- Testing: testify
- Mock framework for rclone commands (testify/mock)
