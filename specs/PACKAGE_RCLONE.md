# Rclone Package Specification

## Package Overview

Provides integration with rclone CLI for bidirectional synchronization. Acts as an abstraction layer over rclone command complexity through direct binary execution with context support.

**Core Responsibilities:**
- Locate and configure rclone binary (system PATH or custom environment variable)
- Execute rclone commands with cancellation and timeout handling
- Build and manage bisync command arguments (builder pattern)
- Detect first-run errors for automatic resync handling
- Manage remote storage discovery and validation
- Create directories idempotently with parent directory support

## Dependencies

**Error Handling:** RcloneError wrapper for command failures and context cancellation
**Logging:** Logger interface for command execution (ConsoleLogger default, optional)

## Core Behaviors

### Binary Discovery Priority
1. Environment variable SYNCERMAN_RCLONE_PATH (if set and path exists) - highest priority
2. System PATH search using exec.LookPath("rclone")
3. All paths resolved to absolute form
4. Validation failure returns error (custom path missing or rclone not found)

### Command Execution Flow
Execute with context, capture all output streams, return structured result with exit code.

**Error Detection Rules:**
- Context cancellation during start → RcloneError with context error, Result = nil
- Context cancellation during execution → RcloneError with context error, Result with partial output
- Non-zero exit code → Result with exit code + RcloneError
- Success → Result with exit code 0 + nil error
- ExitCode -1 (signal termination) or ProcessState nil → defaults to 1
- Nil logger → graceful degradation (skip logging)
- Binary not executable → wrapped error returned

### Bisync Arguments Builder
Constructs rclone bisync commands with standard and optional flags.

**Standard Flags:**
```
--create-empty-src-dirs, --compare=size,modtime, --no-slow-hash
-Mv, --drive-skip-gdocs, --fix-case
--ignore-listing-checksum, --fast-list, --transfers=10
--resilient
```

**Builder Methods:**
- Set source and destination
- Enable --resync flag for first run
- Enable --dry-run flag (no changes)
- Append custom arbitrary arguments
- Build final argument list in consistent order

### First-Run Error Detection

**Detection Logic (both patterns required):**
- Pattern 1: "cannot find prior Path1 or Path2 listings" (case-insensitive)
- Pattern 2: "here are(?:\s+the)?\s+filenames" (case-insensitive)

**Corner Cases:**
- Only Pattern 1 or only Pattern 2 → Not a first-run error (IsFirstRunError = false)
- Empty stderr → Not a first-run error, returns empty paths
- Path1 or Path2 alone → ExtractFirstRunErrorPaths returns [Path] (not error)
- Paths reversed (Path2 mentioned without Path1) → Does NOT match
- Case variations matched due to case-insensitive flag

### Idempotent Directory Operations

**Success Conditions:**
- Directory created successfully
- Directory already exists (detected via error messages: "already exists", "file exists", "path already exists")

**Error Cases:**
- Empty remote path → Error before execution
- Parent directory doesn't exist → Enhanced error message
- Permission denied → Raw error from rclone
- Network failure → Raw error from rclone

**Method Differences:**
- CreatePath: Creates directory and its path
- Not all rclone remotes support parent flag equivalent (S3=yes, some clouds=no)

### Remote Name Normalization

**Output Processing:**
- Execute rclone listremotes command
- Parse each line, trim whitespace
- Remove trailing colon from each remote name
- Return array of clean remote names

**Validation:**
- Case-sensitive comparison
- Input must be plain name (no colons)
- Returns false for non-existent remotes

**Corner Cases:**
- No remotes configured → Returns empty slice (not error)
- Non-zero exit code from listremotes → Returns empty slice (not error)
- Executor.Run failure → Returns nil slice, propagates error
- Empty lines or whitespace-only lines → Ignored
- Trailing colons → Always trimmed

## Edge Cases

### Binary Discovery
- Custom path in SYNCERMAN_RCLONE_PATH missing → Error
- Rclone not found in PATH → Error
- Relative path with separators → Resolved to absolute
- Simple filename without separators → Uses PATH lookup

### Executor
- Context cancellation → RcloneError with partial/complete output depending on timing
- ProcessState nil → ExitCode defaults to 1 (signal termination)
- ExitCode -1 → Defaults to 1
- Logger nil → No logging (safe operation)

### Bisync Builder
- Nil options → Creates empty internal options
- Empty source/destination → Builds anyway, will fail on execution
- Multiple WithResync calls → Sets flag true (idempotent)
- Multiple WithArgs calls → Accumulates all arguments (append)

### First-Run Detection
- Both patterns must match for IsFirstRunError = true
- Case-insensitive matching
- Path extraction works even if only one path exists in message
- Both "Path1 OR Path2" and "Path2 or Path1" variations checked

### Directory Operations
- Empty path → Pre-execution validation error
- "Already exists" variations → Treated as success (idempotent)

### Remote Management
- Empty output → Empty slice returned
- Parse errors on individual lines → Line skipped, continues processing
- Trailing colons → Always removed
- Mixed whitespace → Trimmed, empty results skipped

## Invariants

### Execution Invariants
1. BinaryPath is always non-empty after successful initialization
2. Executor.Run always returns Result (non-nil), even on context cancellation
3. All errors wrapped with context information
4. Exit codes 0-255 preserved, -1/nil defaulted to 1
5. Context passed unchanged to command execution

### String Processing Invariants
1. Remote names never include trailing colons in output
2. Flag ordering is consistent: command, standard flags, optional flags, src, dst, extras
3. Path trimming always removes trailing colon if present
4. First-run patterns are case-insensitive

### State Invariants
1. Builder Build() does not modify state (rebuildable)
2. BisyncOptions always created internally (nil-safe)
3. Logger optional (nil-safe operation)
4. Combined output always equals Stdout + Stderr

## Public API

### Binary Discovery
- FindRcloneBinary() → (path: string, error)
- ConfigFromEnv() → (config: Config, error)

### Executor Creation
- NewConfig() → Config
- NewExecutor(config) → Executor
- NewExecutorWithLogger(config, logger) → Executor

### Bisync Operations
- NewBisyncArgs(src, dst, options) → BisyncArgs builder
- BisyncArgs.WithResync() → BisyncArgs
- BisyncArgs.WithDryRun() → BisyncArgs
- BisyncArgs.WithArgs(args...) → BisyncArgs
- BisyncArgs.Build() → []string

### Directory Operations
- CreatePath(ctx, executor, remotePath) → error

### Remote Management
- ListRemotes(ctx, executor) → ([]string, error)

### First-Run Detection
- IsFirstRunError(Combined) → bool
- ExtractFirstRunErrorPaths(Combined) → []string
- ParseFirstRunError(Combined) → FirstRunError

### Executor Interface
- Run(ctx, args...) → (Result, error)

### Result Structure
- ExitCode: int
- Stdout: string
- Stderr: string
- Combined: string (Stdout + Stderr)
- Success(): bool → ExitCode == 0
- Error(): error → nil on success, wrapper otherwise

---

**Document Version:** 2.0 (Simplified)  
**Purpose:** Language-agnostic specification for implementation porting  
**Reference Implementation:** internal/rclone/