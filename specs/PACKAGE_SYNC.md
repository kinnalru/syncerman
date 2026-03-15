# Sync Package Specification

## Package Overview

The `sync` package provides the core synchronization engine, orchestrating rclone bisync operations with automatic error recovery, validation, and result aggregation.

**Responsibilities:**
- Expand configuration to executable sync targets preserving YAML order
- Validate providers and destinations before sync execution
- Execute rclone bisync commands with first-run error handling
- Prepare directories by creating source and destination paths
- Aggregate results and generate comprehensive reports
- Manage dry-run mode at engine and operation level

**Architecture Position:**
```
config (Config) → sync (Engine) → rclone (Executor)
                       ↓
                  logger.Logger
```

## Key Components & Responsibilities

| File | Responsibilities |
|------|------------------|
| `targets.go` | Flattens nested config into executable targets. Validates providers/destinations. Normalizes paths including 'local' provider. Ensures `RemoteExists` invariant. |
| `directories.go` | Idempotent directory creation for all sources and destinations. Handles existing directories gracefully. |
| `execution.go` | Executes single sync operations. Parses command output. Detects and retries first-run errors. Executes all targets in config order. |
| `result.go` | Aggregates sync results into comprehensive report. Handles nil results safely. |

**Order Guarantee:**
```
YAML Order → ExpandTargets() → RunAll()
provider1:/path1 → target1
provider1:/path2 → target2
provider2:/path3 → target3
```

**Context Cancellation:**
- Returns all completed results
- Graceful shutdown at inter-target boundaries

## Dependencies

**`internal/config`** - Configuration structure with ordered collections
**`internal/rclone`** - Rclone command execution
**`internal/logger`** - Logger interface for domain operations

## Key Behaviors & Invariants

### Sequential Execution with Order Preservation

Critical for linear synchronization chains. YAML configuration order is preserved throughout entire sync pipeline to ensure correct file propagation through multiple destinations.

```
Configuration:
  providers:
    local:
      '/data':
        - to: 'gd:syncerman/'    # Target 1
    gd:
      'syncerman/':
        - to: 'yd:syncerman/'    # Target 2
    yd:
      'syncerman/':
        - to: '/backup/yd_backup/'  # Target 3

Execution Order (RunAll):
  1. local:/data → gd:syncerman/    # Initial data from local
  2. gd:syncerman/ → yd:syncerman/   # Files propagated from gd to yd
  3. yd:syncerman/ → /backup/yd_backup/  # Final backup
```

### First-Run Auto-Retry Strategy

**Error Detection:**
```
Pattern 1: (?i)cannot find prior Path1 or Path2 listings
Pattern 2: (?i)here are(?:\s+the)?\s+filenames
Both must match -> IsFirstRunError() returns true
```

**Recovery Logic:**
1. Initial run without --resync flag fails with first-run error
2. Sync detects pattern
3. Re-execute with `args.WithResync()` (adds --resync flag)
4. If still fails after maxRetries (default: 1), return error

### Dry-Run Mode

Two-level control: engine-level via `engine.SetDryRun(true)` or operation-level via `SyncOptions{DryRun: true}`. Both combined still uses dry-run due to engine flag.

### Edge Cases

- Empty configuration (no providers)
- Empty source paths
- No destinations configured
- Empty destination "to" field
- First-run error detection and retry
- Max retry exhaustion
- Context cancellation mid-execution
- Directory already exists
- Provider not found in rclone config
- Dry-run mode (engine-level and operation-level)
- Configuration order preservation
- Stop-on-error with partial results
- Nil result handling in report collection

## Code Patterns

### No Logging in Domain Code

Functions do not perform I/O. Parameters are inputs, return values are outputs. Logging performed via injected `Logger` interface.

### Error Wrapping with Context

Pattern: `"context: %w"` - always include context, wrap underlying error with `%w`.

### Context-First Parameters

Pass context from main() through all layers. Check `ctx.Done()` in loops. Return `ctx.Err()` on cancellation.

### Fluent Builder Pattern

```go
args := rclone.NewBisyncArgs("gdrive:docs", "s3:backup", options)
    .WithResync()
    .WithDryRun()
    .WithArgs("--fast-list", "--max-age 30d")
```

### Dependency Injection

Interface-based testing enables easy mocking without external dependencies.

## Exit Codes

| Code | Meaning | Sync Package Behavior |
|-------|----------|---------------------|
| 0 | Success | All sync targets succeeded |
| 1 | General error | Any sync target failed |

**Application-Level Exit Codes:** See `guides/OVERALL.md:452-461`

## Implementation Notes

### Thread Safety

The `Engine` struct is **not thread-safe** and should not be used concurrently.

### Regex Performance

First-run detection patterns are compiled at package init time to avoid recompilation overhead.

---

**Document Version:** 1.0  
**Last Updated:** 2026-03-15  
**Reference Implementation:** `internal/sync/`  
**Related Documents:** `guides/OVERALL.md`, `guides/STYLE.md`
