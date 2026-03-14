---
title: "Milestone 3: Rclone Integration Foundation"
status: "completed"
---

# Milestone 3: Rclone Integration Foundation

## Goal

Build rclone command execution and verification layer with support for listremotes, mkdir, and bisync commands.

## Context

This milestone implements the rclone integration layer that enables Syncerman to interact with rclone for:
- Checking available remotes (OVERALL.md: lines 186-204)
- Creating destination directories (OVERALL.md: lines 298-309)
- Executing bisync commands with specific flags (OVERALL.md: lines 256-267)
- Handling rclone output and errors

## Tasks

### 3.1: Define Rclone Types and Interfaces

Create the foundational types and interfaces for the rclone package:

- Define `Config` struct with rclone binary path
- Define `Executor` interface with methods: `Run(ctx context.Context, args ...string) (*Result, error)`
- Define `Result` struct with: ExitCode, Stdout, Stderr, Combined
- Define `Remote` struct representing a rclone remote (name, config)
- Add JSON/YAML tags for all struct fields
- Create doc.go file if not present

### 3.2: Implement Rclone Binary Detection

Add functionality to locate the rclone binary:

- Implement `FindRcloneBinary()` function that searches PATH
- Support custom binary path via environment variable
- Add error handling when rclone not found
- Return error with clear message suggesting installation
- Add unit tests for binary detection with mocked PATH

### 3.3: Implement Command Executor

Create the core command execution functionality:

- Implement `NewExecutor(binaryPath string) *Executor`
- Implement `Run()` method using `exec.CommandContext`
- Capture both stdout and stderr
- Handle context cancellation and timeout
- Implement error wrapping with RcloneError type
- Add logging support (verbose mode dumps full command)
- Add unit tests for successful execution
- Add unit tests for error scenarios (binary not found, timeout)

### 3.4: Implement ListRemotes Command

Create command to list all configured rclone remotes:

- Implement `ListRemotes(ctx context.Context) ([]string, error)`
- Execute `rclone listremotes` command
- Parse output (one remote per line, ends with colon)
- Strip colons from remote names
- Return empty slice if no remotes (not an error)
- Add unit tests with various output formats
- Add unit tests for error scenarios

### 3.5: Implement Mkdir Command

Create command to create directories on rclone remotes:

- Implement `Mkdir(ctx context.Context, remotePath string) error`
- Execute `rclone mkdir <remote:path>` command
- Handle directory exists gracefully
- Handle parent directory not found error
- Add proper error wrapping
- Add unit tests for successful creation
- Add unit tests for error scenarios (permission denied, invalid remote)

### 3.6: Implement Bisync Command Builder

Build rclone bisync commands with all required flags:

- Define `BisyncOptions` struct with all optional parameters
- Define `BisyncArgs` builder struct
- Implement `NewBisyncArgs()` constructor with default flags
- Add methods to add/modify flags: `WithResync()`, `WithDryRun()`, `WithArgs()`
- Build command with standard flags (OVERALL.md: lines 257-267):
  - `--create-empty-src-dirs`
  - `--compare size,modtime`
  - `--no-slow-hash`
  - `-MvP`
  - `--drive-skip-gdocs`
  - `--fix-case`
  - `--ignore-listing-checksum`
  - `--fast-list`
  - `--transfers=10 --resilient`
- Add user-specified args at the end with `${@}`
- Add unit tests for command building
- Test with resync, dry-run, and custom args

### 3.7: Implement First-Run Error Detection

Add error detection for first-run scenario:

- Define `FirstRunErrorPattern` as compiled regex
- Pattern matches: `cannot find prior Path1 or Path2 listings` and `here are the filenames`
- Implement `IsFirstRunError(stderr string) bool`
- Extract file paths from error message for logging
- Add unit tests for error pattern matching
- Test with actual error output from OVERALL.md: lines 321-326

### 3.8: Write Integration Tests

Create integration-style tests for the rclone package:

- Test end-to-end workflow: detect binary → list remotes → verify remote exists
- Test mkdir followed by bisync command building
- Test error detection with simulated first-run error
- Test executor with different exit codes
- Mock rclone execution using testify/mock or similar
- Ensure all tests are isolated and don't require actual rclone installation
- Target 85%+ code coverage

### 3.9: Update Package Documentation

Ensure proper documentation:

- Add godoc comments to all exported types and functions
- Add examples in godoc for ListRemotes, Mkdir, and BisyncArgs
- Document error behaviors and edge cases
- Document all supported bisync flags
- Add package-level usage example
