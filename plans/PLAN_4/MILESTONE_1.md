---
status: completed
---

# Milestone 1: Fix All Three Logging Issues

## Goal

Address all three logging inconsistencies to improve user experience and clarity:
1. First-run detection logs at INFO level instead of WARN
2. "First-Runs" section title reflects informational nature
3. Dry-run message clearly explains why directories are checked
4. All log paths consistently show provider names without hash suffixes

## Context

Test results from 2026-03-15 identified three UX improvements (PLAN_4.md:9-25):

**Issue #1**: Inconsistent severity level for first-run detection
- Labels as both `[WARN]` and "First-Run Errors" section
- Causes confusion about issue severity
- Affected code: `internal/sync/firstrun.go:64,69` and `internal/sync/result.go:115`

**Issue #2**: Misleading message about directory creation in dry-run mode
- Message suggests directories are being created even in dry-run
- Confusing for users expecting read-only operations
- Affected code: `internal/sync/directories.go:31`

**Issue #3**: Inconsistent path representation in logs
- rclone output shows provider hash suffixes (e.g., `gd{YRXYK}:path/`)
- syncerman logs don't include suffixes (e.g., `gd:path/`)
- Makes debugging harder for users
- Affected code: `internal/sync/execution.go:14,43,62,65`

Reference documents:
- `ISSUES.md:25-102` - All three issues with expected behavior
- `guides/OVERALL.md:311-337` - First-run handling details

## Tasks

### Task 1.1: Change first-run log severity from WARN to INFO

**File**: `internal/sync/firstrun.go:69`

**Changes required**:
- Change `logger.Warn` to `logger.Info`
- Update message to: "First-run detected, retrying with --resync (attempt %d/%d)"

**Rationale**: First-run detection is not an error condition - it's expected behavior for new sync paths. Using INFO level reduces confusion about severity.

---

### Task 1.2: Update error section header name

**File**: `internal/sync/result.go:115`

**Changes required**:
- Change "First-Run Errors" to "First-Runs"
- Update section to reflect informational nature (not errors)

**Rationale**: The section title should accurately reflect that these are informational first-run notifications, not error conditions.

---

### Task 1.3: Improve dry-run directory creation message

**File**: `internal/sync/directories.go:31`

**Changes required**:
- Change message to clarify: "Ensuring %d destination directories exist (required by rclone even in dry-run mode)..."

**Rationale**: The current message suggests directories are being created even in dry-run. The updated message clarifies that rclone requires directory verification even in dry-run mode, reducing user confusion.

---

### Task 1.4: Add helper function to strip provider hash suffix

**File**: `internal/sync/targets.go` (append new function)

**Changes required**:
- Create `StripProviderHash(path string) string` to remove `{...}` suffixes
- Parse rclone path format `provider{hash}:path` → `provider:path`

**Implementation details**:
- Use regex pattern: `^(\w+)\{[A-Za-z0-9]+\}:(.*)$`
- Return normalized path or original if no hash suffix found
- Handle edge cases (missing hash, invalid format)

**Rationale**: Provider hash suffixes are temporary rclone identifiers that clutter logs and make debugging harder for users.

---

### Task 1.5: Apply path normalization to sync execution logs

**File**: `internal/sync/execution.go:14,16,62,65`

**Changes required**:
- Wrap provider paths with `StripProviderHash()` before logging
- Ensure consistent path format in INFO level logs

**Implementation details**:
- Update log statements that display provider paths
- Apply normalization for source paths (line 14, 16)
- Apply normalization for destination paths (line 62, 65)

**Rationale**: Consistent path formatting improves readability and helps users identify sync operations.

---

### Task 1.6: Apply path normalization to debug rclone output

**File**: `internal/sync/execution.go:43`

**Changes required**:
- Parse rclone output and strip provider hash suffixes from debug log
- Ensure users see clean paths for debugging

**Implementation details**:
- Update debug output parsing to normalize paths
- Handle multiple path occurrences in output
- Maintain rclone output structure while cleaning paths

**Rationale**: Debug output should show clean paths to help users understand what's being synced.

---

### Task 1.7: Test all three fixes

**Test scenarios**:
1. Run dry-run sync and verify directory creation message clarity
2. Trigger first-run condition and verify INFO level logging
3. Trigger first-run condition and verify "First-Runs" section title
4. Verify rclone output shows normalized paths in logs
5. Run `go test ./...` to ensure no regressions
6. Run `go build` to verify compilation
7. Check code formatting with `go fmt ./...`

**Expected results**:
- First-run logs appear at INFO level, not WARN
- Section header shows "First-Runs" instead of "First-Run Errors"
- Dry-run message clarifies rclone's directory requirements
- All provider paths in logs show clean format without `{hash}` suffixes
- All tests pass with no regressions
- Code compiles successfully
- Code formatting passes checks

## Success Criteria

- First-run detection logs consistently at INFO level
- "First-Runs" section title reflects informational nature
- Dry-run message clearly explains why directories are checked
- All log paths consistently show provider names without hash suffixes
- All tests pass with no regressions
- Code passes formatting and linting checks

## Notes

- Task 1.4-1.6 require careful regex parsing of rclone output format
- Provider hash suffixes are in format `{ALPHANUMERIC}` after provider name (e.g., `gd{YRXYK}:path/`)
- Test first-run scenario with new sync paths to verify Task 1.1-1.2
- Use `--verbose` flag when testing Task 1.6 for debug output
- Ensure backward compatibility with existing functionality
