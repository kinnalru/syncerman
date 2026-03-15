# COMPLEX_BATTLE_TESING.md

## Overview

This document provides a comprehensive battle testing plan for the Syncerman CLI application in production environment with real rclone providers (`gd` and `yd`).

**Test Environment:**
- Real rclone providers: `gd` (Google Drive) and `yd` (Yandex Disk)
- Remote base path: `syncerman/` (subfolders for each scenario)
- Working directory: `./tmp/complex/`
- Binary location: `./bin/syncerman`

**Safety Constraints:**
- All remote operations restricted to `syncerman/` subfolder
- Each scenario uses isolated remote paths (e.g., `syncerman/scenario1/`)
- Cleanup procedures included for safe re-use

## Scenario Summary

| Scenario | Name | Purpose | Complexity | Time Estimate |
|----------|------|---------|------------|---------------|
| SCENARIO1 | Basic Linear Sync | Tests fundamental sync chain: local → gd → yd → local2 | ⭐⭐ | 5-10 min |
| SCENARIO2 | Updates, Deletes, Conflicts | Tests file modifications, additions, deletions, and conflict resolution | ⭐⭐⭐ | 10-15 min |
| SCENARIO3 | First-Run & State Recovery | Tests automatic resync handling, critical error recovery | ⭐⭐⭐ | 10-15 min |
| SCENARIO4 | Multi-Directional Multi-Path | Tests complex configuration with multiple independent sync paths | ⭐⭐⭐⭐ | 15-20 min |
| SCENARIO5 | Dry-Run Mode | Verifies dry-run doesn't make actual changes | ⭐⭐ | 5-10 min |
| SCENARIO6 | Error Handling & Edge Cases | Tests invalid configs, missing files, permission errors, etc. | ⭐⭐⭐⭐ | 10-15 min |

## Detailed Scenario Descriptions

### SCENARIO1: Basic Linear Synchronization

**Purpose:**
Validate the fundamental sync functionality through all three storage providers in a linear chain.

**Test Flow:**
1. Create local file structure with nested directories
2. Initialize configuration with three-way sync: local → gd → yd → local2
3. Execute sync and verify all locations have identical content
4. Run second sync to verify idempotency (no changes needed)
5. Complete cleanup

**Key Verification Points:**
- Files present in all 4 locations
- File contents identical across all locations
- Directory structure preserved
- No missing or unexpected files
- Second sync completes without errors

**File:** `tests/SCENARIO1.md`

---

### SCENARIO2: File Update, Delete, and Conflict Resolution

**Purpose:**
Validate sync behavior with complex file operations including modifications, additions, deletions, and conflict resolution.

**Test Flow:**
1. Initial sync of base file structure
2. Modify existing files in local
3. Add new files to local
4. Delete files from local
5. Sync changes across all locations
6. Create conflicting modifications in local and local2
7. Sync and verify conflict resolution
8. Final verification

**Key Verification Points:**
- Modified files propagate correctly
- New files propagate correctly
- Deleted files removed from all locations
- Conflict resolution works (may show conflict files)
- File counts match (except for conflict files)
- No unexpected file appearance/disappearance

**File:** `tests/SCENARIO2.md`

---

### SCENARIO3: First-Run (Resync) and State Recovery

**Purpose:**
Test automatic detection and handling of first-run scenarios, resync flag application, and recovery from critical errors.

**Test Flow:**
1. Create sync configuration without bisync state files
2. Execute sync and verify first-run error detection
3. Verify automatic --resync flag application
4. Confirm bisync state files created
5. Test normal sync with state present
6. Delete state files to simulate critical error
7. Verify auto-recovery with resync
8. Test manual resync configuration
9. Final state verification

**Key Verification Points:**
- First-run error pattern detected correctly
- Automatic resync retry successful
- State files created after first sync
- Normal sync works without resync
- Critical error recovery successful
- Manual resync configuration works
- State files persist across syncs

**File:** `tests/SCENARIO3.md`

---

### SCENARIO4: Multi-Directional Sync with Multiple Paths

**Purpose:**
Test complex configuration with multiple independent sync paths, including fan-out scenarios and isolated paths.

**Test Flow:**
1. Create three independent paths: docs, media, other
2. Configure complex sync:
   - docs: local → gd AND local → yd → local2 (fan-out)
   - media: local → gd → yd → local2 (chain)
   - other: NOT synced (local only)
3. Execute sync and verify isolated behavior
4. Modify each path independently
5. Sync and verify isolation maintained
6. Test single-target sync
7. Final verification

**Key Verification Points:**
- All 6 sync tasks execute successfully
- Docs path: fans out to both remotes
- Media path: chains through all providers
- Other path stays local only
- Independent modifications sync correctly per path
- Single target sync works
- No cross-contamination between paths

**File:** `tests/SCENARIO4.md`

---

### SCENARIO5: Dry-Run Mode Verification

**Purpose:**
Thoroughly validate dry-run mode ensures no actual changes are made while showing what would happen.

**Test Flow:**
1. Create initial files
2. Dry-run sync (verify no changes)
3. Actual sync
4. Modify files
5. Dry-run again (verify unsynced)
6. Actual sync of modifications
7. Create conflict
8. Dry-run with conflict
9. Test short flag (-d)
10. Test with check commands
11. Verify dry-run output indicators
12. Test multiple dry-run scenarios

**Key Verification Points:**
- Dry-run shows sync preview
- No files transferred during dry-run
- File counts unchanged during dry-run
- File contents unchanged during dry-run
- Actual sync after dry-run works
- Short flag (-d) works identically
- Dry-run output contains appropriate indicators
- Modifications not applied during dry-run
- Conflicts not resolved during dry-run

**File:** `tests/SCENARIO5.md`

---

### SCENARIO6: Error Handling and Edge Cases

**Purpose:**
Test robust error handling, invalid configurations, and edge cases to ensure graceful degradation.

**Test Flow:**
1. Missing configuration file
2. Invalid YAML syntax
3. Missing required fields
4. Invalid provider names
5. Invalid destination formats
6. Missing source directories
7. Empty configuration
8. Invalid CLI arguments
9. Mixed valid/invalid providers
10. Permission errors (if applicable)
11. Simultaneous sync attempts (concurrency)

**Key Verification Points:**
- Appropriate error messages for each case
- Non-zero exit codes for errors
- Clear indication of what went wrong
- Graceful handling of edge cases
- Partial validation (e.g., some invalid, some valid)
- Concurrency handling (locks or busy errors)

**File:** `tests/SCENARIO6.md`

---

## Execution Prompts

### Run All Scenarios

To execute all scenarios in sequence with cleanup between each:

```
Please execute all 6 battle testing scenarios in order:

1. SCENARIO1: Basic Linear Synchronization
2. SCENARIO2: File Update, Delete, and Conflict Resolution
3. SCENARIO3: First-Run and State Recovery
4. SCENARIO4: Multi-Directional Multi-Path Sync
5. SCENARIO5: Dry-Run Mode Verification
6. SCENARIO6: Error Handling and Edge Cases

For each scenario:
1. Follow the "Full Script (Ready to Execute)" section from the scenario file
2. Capture and report the output
3. Verify all success criteria are met
4. Perform cleanup between scenarios

Report the overall results with pass/fail status for each scenario.
```

### Run Specific Scenarios

#### Run SCENARIO1 Only

```
Please execute SCENARIO1: Basic Linear Synchronization.

Follow the "Full Script (Ready to Execute)" section from tests/SCENARIO1.md.

Report the output and verify all success criteria are met.
```

#### Run SCENARIO2 Only

```
Please execute SCENARIO2: File Update, Delete, and Conflict Resolution.

Follow the "Full Script (Ready to Execute)" section from tests/SCENARIO2.md.

Report the output and verify all success criteria are met.
```

#### Run SCENARIO3 Only

```
Please execute SCENARIO3: First-Run and State Recovery.

Follow the "Full Script (Ready to Execute)" section from tests/SCENARIO3.md.

Report the output and verify all success criteria are met.
```

#### Run SCENARIO4 Only

```
Please execute SCENARIO4: Multi-Directional Multi-Path Sync.

Follow the "Full Script (Ready to Execute)" section from tests/SCENARIO4.md.

Report the output and verify all success criteria are met.
```

#### Run SCENARIO5 Only

```
Please execute SCENARIO5: Dry-Run Mode Verification.

Follow the "Full Script (Ready to Execute)" section from tests/SCENARIO5.md.

Report the output and verify all success criteria are met.
```

#### Run SCENARIO6 Only

```
Please execute SCENARIO6: Error Handling and Edge Cases.

Follow the "Full Script (Ready to Execute)" section from tests/SCENARIO6.md.

Report the output and verify all success criteria are met.
```

### Run Subsets of Scenarios

#### Run Core Scenarios (1, 2, 3)

```
Please execute the core functionality scenarios:
1. SCENARIO1: Basic Linear Synchronization
2. SCENARIO2: File Update, Delete, and Conflict Resolution
3. SCENARIO3: First-Run and State Recovery

For each scenario, follow the full script and verify success criteria.
Report results for all three scenarios.
```

#### Run Advanced Scenarios (4, 5, 6)

```
Please execute the advanced testing scenarios:
1. SCENARIO4: Multi-Directional Multi-Path Sync
2. SCENARIO5: Dry-Run Mode Verification
3. SCENARIO6: Error Handling and Edge Cases

For each scenario, follow the full script and verify success criteria.
Report results for all three scenarios.
```

### Continuous Testing Prompt

```
For automated testing execution, run the following command sequence:

```bash
cd /home/llm/agents/takopi/syncerman

# Run all scenarios in order with error handling
for i in {1..6}; do
  echo "=========================================="
  echo "Running SCENARIO$i"
  echo "=========================================="
  
  # Extract and execute the full script from the scenario file
  bash -c "$(sed -n '/^## Full Script (Ready to Execute)/,/^## Success Criteria/p' tests/SCENARIO$i.md | sed '1d;$d')"
  
  RESULT=$?
  if [ $RESULT -eq 0 ]; then
    echo "✓ SCENARIO$i PASSED"
  else
    echo "✗ SCENARIO$i FAILED with exit code $RESULT"
  fi
  echo ""
done

echo "=========================================="
echo "All scenarios completed"
echo "=========================================="
```

This will execute all scenarios sequentially, capturing results and indicating pass/fail status for each.
```

## Test Safety and Cleanup

### Before Testing

Ensure rclone providers are configured and accessible:

```bash
# Verify providers exist
rclone listremotes

# Expected output:
# gd:
# yd:
```

### After Testing

!! DO NOT purge any data automaticaly before, during or after executing scenarios. !!

MANUAL Cleanup all test data:

```bash
# Remove local test directories
rm -rf ./tmp/complex

# Clean remote paths
# rclone purge gd:syncerman/scenario1/ --quiet
# rclone purge gd:syncerman/scenario2/ --quiet
# rclone purge gd:syncerman/scenario3/ --quiet
# rclone purge gd:syncerman/scenario4/ --quiet
# rclone purge gd:syncerman/scenario5/ --quiet
# rclone purge gd:syncerman/scenario6/ --quiet

# rclone purge yd:syncerman/scenario1/ --quiet
# rclone purge yd:syncerman/scenario2/ --quiet
# rclone purge yd:syncerman/scenario3/ --quiet
# rclone purge yd:syncerman/scenario4/ --quiet
# rclone purge yd:syncerman/scenario5/ --quiet
# rclone purge yd:syncerman/scenario6/ --quiet

# Clean bisync state files
rm -rf ~/.cache/rclone/bisync/*scenario*

# Re-create empty test directory
mkdir -p ./tmp/complex
```

## Expected Total Test Duration

- SCENARIO1: 5-10 minutes
- SCENARIO2: 10-15 minutes
- SCENARIO3: 10-15 minutes
- SCENARIO4: 15-20 minutes
- SCENARIO5: 5-10 minutes
- SCENARIO6: 10-15 minutes

**Total:** 55-85 minutes (~1-1.5 hours) for complete test suite

## Success Criteria Summary

Overall battle testing is considered successful if:

1. **All scenarios execute without critical errors** (non-zero exit codes expected in SCENARIO6 for error cases)
2. **File synchronization works correctly** across local, gd, yd, and local2
3. **First-run detection and automatic resync** function properly
4. **Dry-run mode** never makes actual changes
5. **Error handling** provides clear error messages and appropriate exit codes
6. **Complex multi-path configurations** work correctly
7. **Conflict resolution** works (may produce conflict files)
8. **Idempotency** maintained (repeated syncs don't cause issues)

## Test Environment Requirements

Before running scenarios, ensure:

- [ ] Built syncerman binary exists at `./bin/syncerman`
- [ ] rclone is installed and in PATH
- [ ] rclone has `gd` provider configured with access to `gd:syncerman/`
- [ ] rclone has `yd` provider configured with access to `yd:syncerman/`
- [ ] Read/write permissions for `./tmp/complex/` directory
- [ ] Network connectivity to access gd and yd providers
- [ ] Sufficient storage quota on both providers

## Reporting Format

For each scenario run, document:

```markdown
## Scenario Results: SCENARIO<N>

**Status:** PASSED / FAILED

**Exit Code:** <exit code (0 for success scenarios, non-zero for scenario 6 errors)>

**Output Summary:**
<brief summary of key output>

**Verification Results:**
- [ ] Success criterion 1
- [ ] Success criterion 2
- [ ] ... etc

**Issues Found:**
<list any issues or unexpected behavior>

**Recommendations:**
<any recommendations for fixes or improvements>
```

---

**Document Version:** 1.0
**Date Created:** 2025-03-15
**Author:** Autonomous Coding Agent
**Status:** Ready for Execution
