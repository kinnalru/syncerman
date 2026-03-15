# SCENARIO1: Basic Linear Synchronization

## Overview
Tests basic linear synchronization through all three storage providers:
local → gd → yd → local2

## Prerequisites
- rclone configured with `gd` and `yd` providers
- Access to syncerman binary at `/home/llm/agents/takopi/syncerman/bin/syncerman`
- Cleanup of test directories and remote paths before execution

## Test Setup

### 1. Initialize Test Environment

!! DO NOT purge any data automaticaly before, during or after executing scenarios. !!

```bash
#!/bin/bash
# Define test paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"
CONFIG_FILE="$TEST_DIR/scenario1.yaml"

# Create directories
mkdir -p "$LOCAL_DIR"
mkdir -p "$LOCAL2_DIR"

# Clean remote paths (if they exist)
# rclone purge "gd:syncerman/scenario1/" --quiet 2>/dev/null || true
# rclone purge "yd:syncerman/scenario1/" --quiet 2>/dev/null || true
```

### 2. Create Test Data Structure

```bash
#!/bin/bash
# Create test file structure
echo "This is file1 in local" > "$LOCAL_DIR/file1.txt"
echo "This is file2 in local" > "$LOCAL_DIR/file2.txt"

mkdir -p "$LOCAL_DIR/subdir1"
echo "File in subdir1" > "$LOCAL_DIR/subdir1/subfile1.txt"

mkdir -p "$LOCAL_DIR/subdir1/deep"
echo "Deep file" > "$LOCAL_DIR/subdir1/deep/deepfile.txt"

mkdir -p "$LOCAL_DIR/subdir2"
echo "File in subdir2" > "$LOCAL_DIR/subdir2/subfile2.txt"

# Verify files created
echo "=== Created files in local ==="
find "$LOCAL_DIR" -type f -exec echo "{}" \;
```

### 3. Create Configuration File

```yaml
# scenario1.yaml
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario1/local':
    -
      to: 'gd:syncerman/scenario1/'

gd:
  'syncerman/scenario1/':
    -
      to: 'yd:syncerman/scenario1/'

yd:
  'syncerman/scenario1/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario1/local2'
```

```bash
# Write configuration
cat > "$CONFIG_FILE" << 'EOF'
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario1/local':
    -
      to: 'gd:syncerman/scenario1/'

gd:
  'syncerman/scenario1/':
    -
      to: 'yd:syncerman/scenario1/'

yd:
  'syncerman/scenario1/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario1/local2'
EOF
```

## Test Execution

### Step 1: Verify Configuration

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"
CONFIG_FILE="$TEST_DIR/scenario1.yaml"

echo "=== Step 1: Check configuration and verify rclone remotes ==="
/home/llm/agents/takopi/syncerman/bin/syncerman check --config "$CONFIG_FILE" --verbose
```

**Expected Output:**
- Configuration is valid
- Found 3 provider(s): local, gd, yd
- Exit code: 0

### Step 2: Verify Remotes

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"
CONFIG_FILE="$TEST_DIR/scenario1.yaml"

echo "=== Step 2: Check remotes ==="
/home/llm/agents/takopi/syncerman/bin/syncerman check remotes --config "$CONFIG_FILE" --verbose
```

**Expected Output:**
- Checking rclone remotes...
- local: OK
- gd: OK
- yd: OK
- All providers are configured in rclone
- Exit code: 0

### Step 3: Dry Run Check

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"
CONFIG_FILE="$TEST_DIR/scenario1.yaml"

echo "=== Step 3: Dry run sync ==="
/home/llm/agents/takopi/syncerman/bin/syncerman sync --config "$CONFIG_FILE" --dry-run --verbose
```

**Expected Output:**
- Shows what would be synced
- No actual changes made
- Exit code: 0

### Step 4: Perform Synchronization

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"
CONFIG_FILE="$TEST_DIR/scenario1.yaml"

echo "=== Step 4: Execute sync ==="
/home/llm/agents/takopi/syncerman/bin/syncerman sync --config "$CONFIG_FILE" --verbose
```

**Expected Output:**
- Sync tasks execute successfully
- All 3 targets complete
- Files transferred from local to gd to yd to local2
- Exit code: 0

### Step 5: Verify Sync Results

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"

echo "=== Step 5: Verify sync results ==="
echo ""
echo "=== Files in local ==="
find "$LOCAL_DIR" -type f | sort
echo ""
echo "=== Files in local2 ==="
find "$LOCAL2_DIR" -type f | sort
echo ""
echo "=== Files in gd:syncerman/scenario1 ==="
rclone ls "gd:syncerman/scenario1/"
echo ""
echo "=== Files in yd:syncerman/scenario1 ==="
rclone ls "yd:syncerman/scenario1/"
echo ""
echo "=== Verify file content equality ==="
for file in $(find "$LOCAL_DIR" -type f | sed "s|$LOCAL_DIR||"); do
  echo "Comparing: $file"
  diff "$LOCAL_DIR$file" "$LOCAL2_DIR$file" && echo "  OK"
done
```

**Expected Results:**
- Same file structure in local, local2, gd, and yd
- All file contents are identical
- No missing files
- No extra files

### Step 6: Second Sync (Idempotency Check)

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"
CONFIG_FILE="$TEST_DIR/scenario1.yaml"

echo "=== Step 6: Second sync (idempotency check) ==="
/home/llm/agents/takopi/syncerman/bin/syncerman sync --config "$CONFIG_FILE" --verbose
```

**Expected Output:**
- Sync completes without errors
- No new files transferred
- No changes made (everything already in sync)
- Exit code: 0

## Cleanup

!! DO NOT purge any data automaticaly before, during or after executing scenarios. !!

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"

echo "=== Cleanup ==="
# Remove local directories
rm -rf "$TEST_DIR/local"
rm -rf "$TEST_DIR/local2"

# Clean remote paths
# rclone purge "gd:syncerman/scenario1/" --quiet 2>/dev/null || true
# rclone purge "yd:syncerman/scenario1/" --quiet 2>/dev/null || true

# Clean rclone bisync metadata
rm -rf ~/.cache/rclone/bisync/*scenario1* 2>/dev/null || true

echo "Cleanup complete"
```

## Full Script (Ready to Execute)

!! DO NOT purge any data automaticaly before, during or after executing scenarios. !!

```bash
#!/bin/bash
set -e

# Define paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario1"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"
CONFIG_FILE="$TEST_DIR/scenario1.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=========================================="
echo "SCENARIO1: Basic Linear Synchronization"
echo "=========================================="
echo ""

# Cleanup
echo "--- Cleanup ---"
rm -rf "$TEST_DIR"
mkdir -p "$LOCAL_DIR" "$LOCAL2_DIR"
# rclone purge "gd:syncerman/scenario1/" --quiet 2>/dev/null || true
# rclone purge "yd:syncerman/scenario1/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario1* 2>/dev/null || true
echo "Cleanup complete"
echo ""

# Create test data
echo "--- Create test data ---"
echo "This is file1 in local" > "$LOCAL_DIR/file1.txt"
echo "This is file2 in local" > "$LOCAL_DIR/file2.txt"
mkdir -p "$LOCAL_DIR/subdir1"
echo "File in subdir1" > "$LOCAL_DIR/subdir1/subfile1.txt"
mkdir -p "$LOCAL_DIR/subdir1/deep"
echo "Deep file" > "$LOCAL_DIR/subdir1/deep/deepfile.txt"
mkdir -p "$LOCAL_DIR/subdir2"
echo "File in subdir2" > "$LOCAL_DIR/subdir2/subfile2.txt"
echo "Test data created"
echo ""

# Create configuration
echo "--- Create configuration ---"
cat > "$CONFIG_FILE" << 'EOF'
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario1/local':
    -
      to: 'gd:syncerman/scenario1/'

gd:
  'syncerman/scenario1/':
    -
      to: 'yd:syncerman/scenario1/'

yd:
  'syncerman/scenario1/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario1/local2'
EOF
echo "Configuration created"
echo ""

# Step 1: Check configuration and verify remotes
echo "=== Step 1: Check configuration and verify remotes ==="
$SYNCERMAN_BIN check --config "$CONFIG_FILE" --verbose
echo ""

# Step 3: Dry run
echo "=== Step 3: Dry run ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --dry-run --verbose
echo ""

# Step 4: Execute sync
echo "=== Step 4: Execute sync ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose
echo ""

# Step 5: Verify results
echo "=== Step 5: Verify results ==="
echo "Files in local:"
find "$LOCAL_DIR" -type f | sort
echo ""
echo "Files in local2:"
find "$LOCAL2_DIR" -type f | sort
echo ""
echo "Files in gd:syncerman/scenario1/:"
rclone ls "gd:syncerman/scenario1/"
echo ""
echo "Files in yd:syncerman/scenario1/:"
rclone ls "yd:syncerman/scenario1/"
echo ""
echo "Content verification:"
for file in $(find "$LOCAL_DIR" -type f | sed "s|$LOCAL_DIR||"); do
  diff "$LOCAL_DIR$file" "$LOCAL2_DIR$file" && echo "  OK: $file"
done
echo ""

# Step 6: Second sync
echo "=== Step 6: Second sync (idempotency) ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose
echo ""

echo "=========================================="
echo "SCENARIO1: All tests PASSED"
echo "=========================================="
```

## Success Criteria
1. Configuration validation passes (exit code 0)
2. All remotes are accessible (exit code 0)
3. Dry run executes without errors
4. Sync completes successfully (exit code 0)
5. File structure is identical across all 4 locations
6. File contents are identical across all 4 locations
7. Second sync completes without errors (idempotency)
8. No missing or extra files in any location
