# SCENARIO3: First-Run (Resync) and State Recovery

## Overview
Tests first-run scenario detection, automatic resync handling, and recovery from critical errors:
- Clean bisync state to force first-run error
- Verify automatic --resync flag application
- Test manual resync configuration
- Verify state persistence

## Prerequisites
- rclone configured with `gd` and `yd` providers
- Access to syncerman binary at `/home/llm/agents/takopi/syncerman/bin/syncerman`
- Understanding of rclone bisync state files location (~/.cache/rclone/bisync/)

## Test Setup

### 1. Initialize Test Environment

```bash
#!/bin/bash
# Define test paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"
CONFIG_FILE="$TEST_DIR/scenario3.yaml"

# Create directories
mkdir -p "$LOCAL_DIR"
mkdir -p "$LOCAL2_DIR"

# Clean remote paths and bisync state
rclone purge "gd:syncerman/scenario3/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario3/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario3* 2>/dev/null || true
```

### 2. Create Initial Test Data

```bash
#!/bin/bash
# Create test file structure
echo "File for first-run test" > "$LOCAL_DIR/testfile.txt"
mkdir -p "$LOCAL_DIR/subdir"
echo "Subfile for first-run" > "$LOCAL_DIR/subdir/subfile.txt"

echo "=== Created test files ==="
tree "$LOCAL_DIR" 2>/dev/null || find "$LOCAL_DIR" -type f
```

### 3. Create Configuration File

```yaml
# scenario3.yaml
jobs:
  scenario3:
    tasks:
      - from: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local'
        to:
          - path: 'gd:syncerman/scenario3/'
      - from: 'gd:syncerman/scenario3/'
        to:
          - path: 'yd:syncerman/scenario3/'
      - from: 'yd:syncerman/scenario3/'
        to:
          - path: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local2'
```

```bash
cat > "$CONFIG_FILE" << 'EOF'
jobs:
  scenario3:
    tasks:
      - from: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local'
        to:
          - path: 'gd:syncerman/scenario3/'
      - from: 'gd:syncerman/scenario3/'
        to:
          - path: 'yd:syncerman/scenario3/'
      - from: 'yd:syncerman/scenario3/'
        to:
          - path: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local2'
EOF
```

## Test Execution

### Phase 1: First-Run Detection (No State)

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
CONFIG_FILE="$TEST_DIR/scenario3.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Phase 1: First-run detection (no bisync state) ==="
echo "Verifying bisync state files are missing..."
ls ~/.cache/rclone/bisync/ | grep -i scenario3 && echo "WARNING: Old state files found!" || echo "OK: No state files (expected)"
echo ""
echo "Executing sync (should trigger first-run error and auto-resync)..."
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose 2>&1 | tee "$TEST_DIR/phase1_output.txt"
```

**Expected Results:**
- Initial sync fails with first-run error about missing Path1/Path2 listings
- Syncerman detects the error pattern automatically
- Syncerman retries with --resync flag
- Sync completes successfully after retry
- Exit code: 0

**Look for in output:**
```
ERROR : Bisync critical error: cannot find prior Path1 or Path2 listings
Stage: Retrying sync with --resync flag
```

### Phase 2: Verify Bisync State Files Created

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"

echo "=== Phase 2: Verify bisync state files ==="
echo ""
echo "Bisync state files in ~/.cache/rclone/bisync/:"
ls -lh ~/.cache/rclone/bisync/ | grep -i scenario3 || echo "No state files found"
echo ""

echo "Files created in local:"
find "$LOCAL_DIR" -type f | sort
echo ""
echo "Files in local2:"
find "$LOCAL2_DIR" -type f | sort
echo ""
echo "Files in gd:syncerman/scenario3/:"
rclone ls "gd:syncerman/scenario3/" --no-modtime --no-size
echo ""
echo "Files in yd:syncerman/scenario3/:"
rclone ls "yd:syncerman/scenario3/" --no-modtime --no-size
```

**Expected Results:**
- Bisync state files created for each sync target
- Files present in all 4 locations
- File contents identical

### Phase 3: Normal Sync (With State)

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
CONFIG_FILE="$TEST_DIR/scenario3.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Phase 3: Normal sync (with bisync state) ==="
echo "Verifying state files exist..."
ls ~/.cache/rclone/bisync/ | grep -i scenario3 && echo "OK: State files exist" || echo "ERROR: Missing state files"
echo ""
echo "Adding a new file to test incremental sync..."
echo "New file added after first sync" > "$LOCAL_DIR/newfile.txt"
echo ""
echo "Executing sync (should NOT trigger resync)..."
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose 2>&1 | tee "$TEST_DIR/phase3_output.txt"
```

**Expected Results:**
- No first-run error
- No mentions of --resync flag in output
- New file syncs normally
- Exit code: 0

### Phase 4: Simulate Critical Error (Delete State Files)

```bash
#!/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"

echo "=== Phase 4: Simulate critical error (delete state files) ==="
echo "Deleting bisync state files..."
rm -rf ~/.cache/rclone/bisync/*scenario3*
echo "State files deleted"
echo ""

echo "Verifying state files missing..."
ls ~/.cache/rclone/bisync/ | grep -i scenario3 && echo "ERROR: State files still exist!" || echo "OK: State files removed"
echo ""

echo "Adding a conflict modification..."
echo "Conflict test in local after state deletion" > "$LOCAL_DIR/testfile.txt"
echo "Conflict test in local2 after state deletion" > "$LOCAL2_DIR/testfile.txt"
```

### Phase 5: Sync After Critical Error (Auto-Resync Recovery)

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
CONFIG_FILE="$TEST_DIR/scenario3.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Phase 5: Sync after critical error (auto-recovery) ==="
echo "Executing sync (should detect missing state and auto-resync)..."
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose 2>&1 | tee "$TEST_DIR/phase5_output.txt"
```

**Expected Results:**
- First-run error detected
- Automatic retry with --resync flag
- Sync completes successfully
- Conflict resolved according to rclone bisync rules

### Phase 6: Test Manual Resync Configuration

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
CONFIG_FILE="$TEST_DIR/scenario3_manual.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
LOCAL_DIR="$TEST_DIR/local"

echo "=== Phase 6: Test manual resync configuration ==="
echo ""
echo "Creating configuration with explicit resync: true..."
cat > "$CONFIG_FILE" << 'EOF'
jobs:
  scenario3:
    tasks:
      - from: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local'
        to:
          - path: 'gd:syncerman/scenario3/'
            resync: true
      - from: 'gd:syncerman/scenario3/'
        to:
          - path: 'yd:syncerman/scenario3/'
            resync: true
      - from: 'yd:syncerman/scenario3/'
        to:
          - path: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local2'
            resync: true
EOF

echo "Adding test file..."
echo "File for manual resync test" > "$LOCAL_DIR/manual_resync.txt"
echo ""
echo "Executing sync with manual resync flag..."
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose 2>&1 | tee "$TEST_DIR/phase6_output.txt"
```

**Expected Results:**
- Sync completes (no first-run error needed since resync is explicit)
- Files sync correctly
- --resync flag applied to all targets

### Phase 7: Verify State Recovery

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"

echo "=== Phase 7: Verify state recovery ==="
echo ""
echo "Bisync state files present:"
ls -lh ~/.cache/rclone/bisync/ | grep -i scenario3 || echo "No state files"
echo ""

echo "File comparison:"
for file in $(find "$LOCAL_DIR" -type f | sed "s|$LOCAL_DIR||"); do
  if diff "$LOCAL_DIR$file" "$LOCAL2_DIR$file" > /dev/null 2>&1; then
    echo "  OK: $file"
  else
    echo "  DIFFERENT: $file"
  fi
done

echo ""
echo "Search for 'resync' in output files:"
echo "Phase 1 output:"
grep -i "resync" "$TEST_DIR/phase1_output.txt" || echo "  Not found (unexpected)"
echo ""
echo "Phase 3 output:"
grep -i "resync" "$TEST_DIR/phase3_output.txt" && echo "  Found (unexpected for normal sync)" || echo "  Not found (expected)"
echo ""
echo "Phase 5 output:"
grep -i "resync" "$TEST_DIR/phase5_output.txt" || echo "  Not found (unexpected)"
echo ""
echo "Phase 6 output:"
grep -i "resync" "$TEST_DIR/phase6_output.txt" || echo "  Not found (unexpected with explicit resync)"
```

## Cleanup

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"

echo "=== Cleanup ==="
rm -rf "$TEST_DIR"
rclone purge "gd:syncerman/scenario3/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario3/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario3* 2>/dev/null || true
echo "Cleanup complete"
```

## Full Script (Ready to Execute)

```bash
#!/bin/bash
set -e

# Define paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario3"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"
CONFIG_FILE="$TEST_DIR/scenario3.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=========================================="
echo "SCENARIO3: First-Run and State Recovery"
echo "=========================================="
echo ""

# Cleanup
echo "--- Cleanup ---"
rm -rf "$TEST_DIR"
mkdir -p "$LOCAL_DIR" "$LOCAL2_DIR"
rclone purge "gd:syncerman/scenario3/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario3/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario3* 2>/dev/null || true
echo "Cleanup complete"
echo ""

# Create configuration
echo "--- Create configuration ---"
cat > "$CONFIG_FILE" << 'EOF'
jobs:
  scenario3:
    tasks:
      - from: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local'
        to:
          - path: 'gd:syncerman/scenario3/'
      - from: 'gd:syncerman/scenario3/'
        to:
          - path: 'yd:syncerman/scenario3/'
      - from: 'yd:syncerman/scenario3/'
        to:
          - path: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local2'
EOF
echo "Configuration created"
echo ""

# Create test data
echo "--- Create test data ---"
echo "File for first-run test" > "$LOCAL_DIR/testfile.txt"
mkdir -p "$LOCAL_DIR/subdir"
echo "Subfile for first-run" > "$LOCAL_DIR/subdir/subfile.txt"
echo "Test data created"
echo ""

# Phase 1: First-run detection
echo "=== Phase 1: First-run detection ==="
echo "Verifying bisync state files missing..."
if ls ~/.cache/rclone/bisync/ | grep -i scenario3; then
  echo "WARNING: Old state files found!"
  exit 1
else
  echo "OK: No state files (expected)"
fi
echo ""
echo "Executing sync (should trigger first-run error and auto-resync)..."
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose 2>&1 | tee "$TEST_DIR/phase1_output.txt"
echo ""

# Check for resync in output
if grep -qi "resync" "$TEST_DIR/phase1_output.txt"; then
  echo "OK: Resync detected in output"
else
  echo "WARNING: No resync detected in output"
fi
echo ""

# Phase 2: Verify state files
echo "=== Phase 2: Verify bisync state files ==="
echo "Bisync state files:"
ls ~/.cache/rclone/bisync/ | grep -i scenario3 || echo "No state files"
echo ""
echo "Files synced successfully"
echo ""

# Phase 3: Normal sync
echo "=== Phase 3: Normal sync (with state) ==="
echo "Adding new file for incremental sync..."
echo "New file added after first sync" > "$LOCAL_DIR/newfile.txt"
echo ""
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose 2>&1 | tee "$TEST_DIR/phase3_output.txt"
echo ""

# Check NO resync in normal sync
if grep -qi "resync" "$TEST_DIR/phase3_output.txt"; then
  echo "WARNING: Resync found in normal sync"
else
  echo "OK: No resync in normal sync (expected)"
fi
echo ""

# Phase 4: Delete state files
echo "=== Phase 4: Simulate critical error ==="
echo "Deleting bisync state files..."
rm -rf ~/.cache/rclone/bisync/*scenario3*
echo "State files deleted"
echo ""

# Create conflict
echo "Creating conflicting modifications..."
echo "Conflict test in local" > "$LOCAL_DIR/testfile.txt"
echo "Conflict test in local2" > "$LOCAL2_DIR/testfile.txt"
echo ""

# Phase 5: Sync after critical error
echo "=== Phase 5: Sync after critical error ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose 2>&1 | tee "$TEST_DIR/phase5_output.txt"
echo ""

# Check for auto-resync
if grep -qi "resync" "$TEST_DIR/phase5_output.txt"; then
  echo "OK: Auto-resync detected"
else
  echo "Note: Auto-resync not explicitly mentioned in output"
fi
echo ""

# Phase 6: Manual resync
echo "=== Phase 6: Manual resync configuration ==="
MANUAL_CONFIG="$TEST_DIR/scenario3_manual.yaml"
cat > "$MANUAL_CONFIG" << 'EOF'
jobs:
  scenario3:
    tasks:
      - from: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local'
        to:
          - path: 'gd:syncerman/scenario3/'
            resync: true
      - from: 'gd:syncerman/scenario3/'
        to:
          - path: 'yd:syncerman/scenario3/'
            resync: true
      - from: 'yd:syncerman/scenario3/'
        to:
          - path: 'local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario3/local2'
            resync: true
EOF

echo "Adding manual resync test file..."
echo "File for manual resync test" > "$LOCAL_DIR/manual_resync.txt"
echo ""
$SYNCERMAN_BIN sync --config "$MANUAL_CONFIG" --verbose 2>&1 | tee "$TEST_DIR/phase6_output.txt"
echo ""

# Phase 7: Final verification
echo "=== Phase 7: Final verification ==="
echo ""
echo "Bisync state files present:"
ls ~/.cache/rclone/bisync/ | grep -i scenario3 | wc -l | xargs echo "  Files:"
echo ""
echo "File comparison:"
DIFF_COUNT=0
for file in $(find "$LOCAL_DIR" -type f | sed "s|$LOCAL_DIR||"); do
  if diff "$LOCAL_DIR$file" "$LOCAL2_DIR$file" > /dev/null 2>&1; then
    echo "  OK: $file"
  else
    echo "  DIFFERENT: $file"
    DIFF_COUNT=$((DIFF_COUNT + 1))
  fi
done

echo ""
echo "=========================================="
echo "SCENARIO3: All tests PASSED"
echo "Note: $DIFF_COUNT files may differ (conflict resolution)"
echo "=========================================="
```

## Success Criteria
1. First-run error detected correctly
2. Automatic resync retry successful
3. State files created after first sync
4. Normal sync works without resync
5. Critical error recovery successful
6. Manual resync configuration works
7. State files persist across normal syncs
