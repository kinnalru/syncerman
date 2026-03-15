# PLAN_4: Logging and UX Improvements

## Objective

Improve user experience by fixing three low-severity logging inconsistencies found during testing.

## Context

Test results from 2026-03-15 (ISSUES.md) identified three UX improvements:

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

## Milestones

### Milestone 1: Fix All Three Logging Issues

Address all three logging inconsistencies to improve user experience and clarity.

#### Tasks

**Task 1.1:** Change first-run log severity from WARN to INFO
- File: `internal/sync/firstrun.go:69`
- Change `logger.Warn` to `logger.Info`
- Update message to: "First-run detected, retrying with --resync (attempt %d/%d)"

**Task 1.2:** Update error section header name
- File: `internal/sync/result.go:115`
- Change "First-Run Errors" to "First-Runs"
- Update section to reflect informational nature (not errors)

**Task 1.3:** Improve dry-run directory creation message
- File: `internal/sync/directories.go:31`
- Change message to clarify: "Ensuring %d destination directories exist (required by rclone even in dry-run mode)..."

**Task 1.4:** Add helper function to strip provider hash suffix
- File: `internal/sync/targets.go` (append new function)
- Create `StripProviderHash(path string) string` to remove `{...}` suffixes
- Parse rclone path format `provider{hash}:path` → `provider:path`

**Task 1.5:** Apply path normalization to sync execution logs
- File: `internal/sync/execution.go:14,16,62,65`
- Wrap provider paths with `StripProviderHash()` before logging
- Ensure consistent path format in INFO level logs

**Task 1.6:** Apply path normalization to debug rclone output
- File: `internal/sync/execution.go:43`
- Parse rclone output and strip provider hash suffixes from debug log
- Ensure users see clean paths for debugging

**Task 1.7:** Test all three fixes
- Run dry-run sync and verify directory creation message
- Trigger first-run condition and verify INFO level logging
- Verify rclone output shows normalized paths in logs
- Run `go test ./...` to ensure no regressions

## Success Criteria

- First-run detection logs consistently at INFO level
- "First-Runs" section title reflects informational nature
- Dry-run message clearly explains why directories are checked
- All log paths consistently show provider names without hash suffixes
- All tests pass with no regressions
- Code passes formatting and linting checks

## Notes

- Task 1.4-1.6 require careful regex parsing of rclone output format
- Provider hash suffixes are in format `{ALPHANUMERIC}` after provider name
- Test first-run scenario with new sync paths to verify Task 1.1-1.2
- Use `--verbose` flag when testing Task 1.6 for debug output
