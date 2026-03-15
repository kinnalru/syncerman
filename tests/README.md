# Battle Testing Suite

This directory contains comprehensive battle testing scenarios for the Syncerman CLI application.

## Quick Start

To run all scenarios:

```bash
cd /home/llm/agents/takopi/syncerman
```

Then use the prompt from [COMPLEX_BATTLE_TESING.md](./COMPLEX_BATTLE_TESING.md#execution-prompts) - "Run All Scenarios"

## Available Scenarios

| # | Scenario | Complexity | Duration | Description |
|---|----------|------------|----------|-------------|
| 1 | [Basic Linear Sync](./SCENARIO1.md) | ⭐⭐ | 5-10 min | Test fundamental sync chain: local → gd → yd → local2 |
| 2 | [Updates, Deletes, Conflicts](./SCENARIO2.md) | ⭐⭐⭐ | 10-15 min | Test file modifications, additions, deletions, and conflict resolution |
| 3 | [First-Run & State Recovery](./SCENARIO3.md) | ⭐⭐⭐ | 10-15 min | Test automatic resync handling and critical error recovery |
| 4 | [Multi-Directional Multi-Path](./SCENARIO4.md) | ⭐⭐⭐⭐ | 15-20 min | Test complex configuration with multiple independent sync paths |
| 5 | [Dry-Run Mode](./SCENARIO5.md) | ⭐⭐ | 5-10 min | Verify dry-run doesn't make actual changes |
| 6 | [Error Handling & Edge Cases](./SCENARIO6.md) | ⭐⭐⭐⭐ | 10-15 min | Test invalid configs, missing files, permission errors, etc. |

## Running Individual Scenarios

Each scenario file contains a complete, ready-to-execute script at the end. To run a scenario:

1. Open the scenario file (e.g., `SCENARIO1.md`)
2. Scroll to "Full Script (Ready to Execute)" section
3. Copy and execute the bash script

Or use one of the execution prompts from [COMPLEX_BATTLE_TESING.md](./COMPLEX_BATTLE_TESING.md#run-specific-scenarios).

## Test Environment

- **Providers:** `gd` (Google Drive) and `yd` (Yandex Disk)
- **Remote base path:** `syncerman/`
- **Working directory:** `../tmp/complex/`
- **Binary location:** `../bin/syncerman`

**Important:** All remote operations are restricted to `syncerman/` subfolders to ensure safety.

## Pre-Flight Checklist

Before running any scenario:

- [ ] Syncerman binary built: `make build` or `make all`
- [ ] rclone installed: `which rclone`
- [ ] rclone providers configured: `rclone listremotes`
- [ ] Network connectivity available
- [ ] Sufficient storage quota on providers

## Cleanup

After testing, clean up all test data:

```bash
# Remove local test directories
rm -rf ../tmp/complex

# Clean remote paths (for each scenario 1-6)
for i in {1..6}; do
  rclone purge gd:syncerman/scenario$i/ --quiet
  rclone purge yd:syncerman/scenario$i/ --quiet
done

# Clean bisync state files
rm -rf ~/.cache/rclone/bisync/*scenario*
```

## Test Coverage

Each scenario validates specific aspects of Syncerman:

### SCENARIO1
- ✓ Basic sync chain execution
- ✓ File and directory synchronization
- ✓ Content verification across all locations
- ✓ Idempotency (repeated syncs)

### SCENARIO2
- ✓ File modifications propagate
- ✓ New files propagate
- ✓ Deleted files removed
- ✓ Conflict resolution
- ✓ Multi-destination isolation

### SCENARIO3
- ✓ First-run error detection
- ✓ Automatic --resync retry
- ✓ Bisync state file creation
- ✓ Critical error recovery
- ✓ Manual resync configuration

### SCENARIO4
- ✓ Multiple sync paths
- ✓ Fan-out configurations
- ✓ Independent path isolation
- ✓ Single-target sync
- ✓ Complex configuration validation

### SCENARIO5
- ✓ Dry-run preview
- ✓ No actual changes during dry-run
- ✓ Dry-run with various operation types
- ✓ Short flag (-d) support
- ✓ Dry-run output indicators

### SCENARIO6
- ✓ Invalid configuration detection
- ✓ Missing files/paths handling
- ✓ Invalid provider detection
- ✓ Error message clarity
- ✓ Graceful degradation
- ✓ Concurrency handling

## Expected Total Duration

Running all scenarios sequentially: **55-85 minutes** (~1-1.5 hours)

## Documentation

- [Overall Test Plan](./COMPLEX_BATTLE_TESING.md) - Comprehensive testing plan with execution prompts
- [Project Overview](../guides/OVERALL.md) - Syncerman application documentation
- [Coding Workflow](../guides/PLANING.md) - Agent workflow guidelines

## Safety Notes

- Each scenario uses isolated remote paths (e.g., `syncerman/scenario1/`)
- All scripts include cleanup procedures
- Dry-run is tested separately (SCENARIO5) to ensure safety
- Error scenarios (SCENARIO6) do not modify production data

---

**Status:** Ready for Execution
**Version:** 1.0
**Last Updated:** 2025-03-15
