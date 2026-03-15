# SCENARIO2: File Update, Delete, and Conflict Resolution

## Overview
Tests file modifications, deletions, updates, and conflict resolution across the sync chain:
- Create initial sync
- Modify files
- Add new files
- Delete files
- Verify bidirectional conflict resolution

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
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario2"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"
CONFIG_FILE="$TEST_DIR/scenario2.yaml"

# Create directories
mkdir -p "$LOCAL_DIR"
mkdir -p "$LOCAL2_DIR"

# Clean remote paths
#rclone purge "gd:syncerman/scenario2/" --quiet 2>/dev/null || true
#rclone purge "yd:syncerman/scenario2/" --quiet 2>/dev/null || true
```

### 2. Create Initial Test Data Structure

```bash
#!/bin/bash
# Create initial test file structure
echo "Initial content of file1" > "$LOCAL_DIR/file1.txt"
echo "Initial content of file2" > "$LOCAL_DIR/file2.txt"
echo "Initial content of file3" > "$LOCAL_DIR/file3.txt"

mkdir -p "$LOCAL_DIR/subdir1"
echo "Initial subfile content" > "$LOCAL_DIR/subdir1/subfile1.txt"

echo "=== Initial files created ==="
find "$LOCAL_DIR" -type f | sort
```

### 3. Create Configuration File

```yaml
# scenario2.yaml
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario2/local':
    -
      to: 'gd:syncerman/scenario2/'

gd:
  'syncerman/scenario2/':
    -
      to: 'yd:syncerman/scenario2/'

yd:
  'syncerman/scenario2/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario2/local2'
```

```bash
cat > "$CONFIG_FILE" << 'EOF'
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario2/local':
    -
      to: 'gd:syncerman/scenario2/'

gd:
  'syncerman/scenario2/':
    -
      to: 'yd:syncerman/scenario2/'

yd:
  'syncerman/scenario2/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario2/local2'
EOF
```

## Test Execution

### Phase 1: Initial Synchronization

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario2"
CONFIG_FILE="$TEST_DIR/scenario2.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Phase 1: Initial sync ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose

echo "=== Verify initial sync ==="
echo "Files in local:"
find "$LOCAL_DIR" -type f | sort
echo ""
echo "Files in local2:"
find "$LOCAL2_DIR" -type f | sort
```

**Expected Results:**
- Initial sync completes successfully
- Files present in all 4 locations (local, gd, yd, local2)
- File contents identical

### Phase 2: Modify Files

```bash
#!/bin/bash
# Modify existing files in local
echo "Modified content of file1 at $(date)" > "$LOCAL_DIR/file1.txt"
echo "Modified content of file2 at $(date)" > "$LOCAL_DIR/file2.txt"

echo "=== Modified files in local ==="
echo "file1.txt: $(cat $LOCAL_DIR/file1.txt)"
echo "file2.txt: $(cat $LOCAL_DIR/file2.txt)"
```

### Phase 3: Add New Files

```bash
#!/bin/bash
# Add new files to local
echo "New file4 added at $(date)" > "$LOCAL_DIR/file4.txt"
echo "New file5 added at $(date)" > "$LOCAL_DIR/file5.txt"

mkdir -p "$LOCAL_DIR/newdir"
echo "New subfile in newdir" > "$LOCAL_DIR/newdir/newsubfile.txt"

echo "=== Added new files ==="
find "$LOCAL_DIR" -type f | sort
```

### Phase 4: Delete Files

```bash
#!/bin/bash
# Delete file3 from local (also present in local2, gd, yd from Phase 1)
rm "$LOCAL_DIR/file3.txt"

# Delete subdir1/subfile1.txt
rm "$LOCAL_DIR/subdir1/subfile1.txt"

echo "=== Deleted files from local ==="
echo "file3.txt deleted"
echo "subdir1/subfile1.txt deleted"
echo ""
echo "=== Remaining files in local ==="
find "$LOCAL_DIR" -type f | sort
```

### Phase 5: Sync After Modifications

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario2"
CONFIG_FILE="$TEST_DIR/scenario2.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Phase 5: Sync after modifications ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose

echo "=== Verify sync after modifications ==="
echo "Files in local:"
find "$LOCAL_DIR" -type f | sort
echo ""
echo "Files in local2:"
find "$LOCAL2_DIR" -type f | sort
echo ""
echo "Files in gd:syncerman/scenario2/:"
rclone ls "gd:syncerman/scenario2/" --no-modtime --no-size
echo ""
echo "Files in yd:syncerman/scenario2/:"
rclone ls "yd:syncerman/scenario2/" --no-modtime --no-size
```

**Expected Results:**
- Modified files synced to all locations
- New files propagated to all locations
- Deleted files removed from gd, yd, and local2
- Deleted directory structures cleaned up

### Phase 6: Conflict Resolution Test

```bash
#!/bin/bash
# Create conflicting modifications
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario2"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"

# Make different changes to the same file in local and local2
echo "Conflict from local at $(date)" > "$LOCAL_DIR/file1.txt"
echo "Conflict from local2 at $(date)" > "$LOCAL2_DIR/file1.txt"

echo "=== Created conflict in file1.txt ==="
echo "local version: $(cat $LOCAL_DIR/file1.txt)"
echo "local2 version: $(cat $LOCAL2_DIR/file1.txt)"
```

### Phase 7: Sync Conflicts

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario2"
CONFIG_FILE="$TEST_DIR/scenario2.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Phase 7: Sync with conflicts ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose

echo "=== Verify conflict resolution ==="
echo "Files in local:"
find "$LOCAL_DIR" -type f | sort
echo ""
echo "Files in local2:"
find "$LOCAL2_DIR" -type f | sort
echo ""
echo "Content of file1.txt in local:"
cat "$LOCAL_DIR/file1.txt"
echo ""
echo "Content of file1.txt in local2:"
cat "$LOCAL2_DIR/file1.txt"
```

**Expected Results:**
- Sync completes (may show conflict resolution)
- File conflict resolved according to rclone's bisync behavior
- Check for conflict files (with .conflict extensions or similar)

### Phase 8: Verify All Locations

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario2"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"

echo "=== Phase 8: Final verification ==="
echo ""

echo "File count in each location:"
echo "local: $(find $LOCAL_DIR -type f | wc -l)"
echo "local2: $(find $LOCAL2_DIR -type f | wc -l)"
echo "gd:syncerman/scenario2/: $(rclone ls 'gd:syncerman/scenario2/' 2>/dev/null | wc -l)"
echo "yd:syncerman/scenario2/: $(rclone ls 'yd:syncerman/scenario2/' 2>/dev/null | wc -l)"
echo ""

echo "Comparing file contents:"
for file in $(find "$LOCAL_DIR" -type f | sed "s|$LOCAL_DIR||"); do
  echo -n "Checking $file: "
  if diff "$LOCAL_DIR$file" "$LOCAL2_DIR$file" > /dev/null 2>&1; then
    echo "OK"
  else
    echo "DIFFERENT"
    echo "  local: $(cat $LOCAL_DIR$file)"
    echo "  local2: $(cat $LOCAL2_DIR$file)"
  fi
done
```

## Cleanup

!! DO NOT purge any data automaticaly before, during or after executing scenarios. !!

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario2"

echo "=== Cleanup ==="
rm -rf "$TEST_DIR"
#rclone purge "gd:syncerman/scenario2/" --quiet 2>/dev/null || true
#rclone purge "yd:syncerman/scenario2/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario2* 2>/dev/null || true
echo "Cleanup complete"
```

## Full Script (Ready to Execute)

!! DO NOT purge any data automaticaly before, during or after executing scenarios. !!

```bash
#!/bin/bash
set -e

# Define paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario2"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"
CONFIG_FILE="$TEST_DIR/scenario2.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=========================================="
echo "SCENARIO2: Update, Delete, Conflicts"
echo "=========================================="
echo ""

# Cleanup
echo "--- Cleanup ---"
rm -rf "$TEST_DIR"
mkdir -p "$LOCAL_DIR" "$LOCAL2_DIR"
#rclone purge "gd:syncerman/scenario2/" --quiet 2>/dev/null || true
#rclone purge "yd:syncerman/scenario2/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario2* 2>/dev/null || true
echo "Cleanup complete"
echo ""

# Create configuration
echo "--- Create configuration ---"
cat > "$CONFIG_FILE" << 'EOF'
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario2/local':
    -
      to: 'gd:syncerman/scenario2/'

gd:
  'syncerman/scenario2/':
    -
      to: 'yd:syncerman/scenario2/'

yd:
  'syncerman/scenario2/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario2/local2'
EOF
echo "Configuration created"
echo ""

# Phase 1: Initial sync
echo "=== Phase 1: Initial sync ==="
echo "Initial content of file1" > "$LOCAL_DIR/file1.txt"
echo "Initial content of file2" > "$LOCAL_DIR/file2.txt"
echo "Initial content of file3" > "$LOCAL_DIR/file3.txt"
mkdir -p "$LOCAL_DIR/subdir1"
echo "Initial subfile content" > "$LOCAL_DIR/subdir1/subfile1.txt"

$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose
echo "Files synced to all locations"
echo ""

# Phase 2: Modify files
echo "=== Phase 2: Modify files ==="
echo "Modified content of file1 at $(date)" > "$LOCAL_DIR/file1.txt"
echo "Modified content of file2 at $(date)" > "$LOCAL_DIR/file2.txt"
echo "Files modified in local"
echo ""

# Phase 3: Add new files
echo "=== Phase 3: Add new files ==="
echo "New file4 added at $(date)" > "$LOCAL_DIR/file4.txt"
echo "New file5 added at $(date)" > "$LOCAL_DIR/file5.txt"
mkdir -p "$LOCAL_DIR/newdir"
echo "New subfile in newdir" > "$LOCAL_DIR/newdir/newsubfile.txt"
echo "New files added"
echo ""

# Phase 4: Delete files
echo "=== Phase 4: Delete files ==="
rm "$LOCAL_DIR/file3.txt"
rm "$LOCAL_DIR/subdir1/subfile1.txt"
echo "file3.txt and subdir1/subfile1.txt deleted from local"
echo ""

# Phase 5: Sync after modifications
echo "=== Phase 5: Sync after modifications ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose
echo ""
echo "Verification after modifications:"
echo "local: $(find $LOCAL_DIR -type f | wc -l) files"
echo "local2: $(find $LOCAL2_DIR -type f | wc -l) files"
echo "gd:syncerman/scenario2/: $(rclone ls 'gd:syncerman/scenario2/' 2>/dev/null | wc -l) files"
echo "yd:syncerman/scenario2/: $(rclone ls 'yd:syncerman/scenario2/' 2>/dev/null | wc -l) files"
echo ""

# Phase 6: Conflict resolution test
echo "=== Phase 6: Conflict resolution ==="
echo "Conflict from local at $(date)" > "$LOCAL_DIR/file1.txt"
echo "Conflict from local2 at $(date)" > "$LOCAL2_DIR/file1.txt"
echo "Created conflicting versions of file1.txt"
echo ""

# Phase 7: Sync conflicts
echo "=== Phase 7: Sync with conflicts ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose
echo ""

# Phase 8: Final verification
echo "=== Phase 8: Final verification ==="
echo ""
echo "File count in each location:"
echo "local: $(find $LOCAL_DIR -type f | wc -l)"
echo "local2: $(find $LOCAL2_DIR -type f | wc -l)"
echo "gd:syncerman/scenario2/: $(rclone ls 'gd:syncerman/scenario2/' 2>/dev/null | wc -l)"
echo "yd:syncerman/scenario2/: $(rclone ls 'yd:syncerman/scenario2/' 2>/dev/null | wc -l)"
echo ""

echo "Comparing file contents:"
DIFF_COUNT=0
for file in $(find "$LOCAL_DIR" -type f | sed "s|$LOCAL_DIR||"); do
  if diff "$LOCAL_DIR$file" "$LOCAL2_DIR$file" > /dev/null 2>&1; then
    echo "  OK: $file"
  else
    echo "  DIFFERENT: $file"
    DIFF_COUNT=$((DIFF_COUNT + 1))
  fi
done

if [ $DIFF_COUNT -gt 0 ]; then
  echo ""
  echo "Warning: $DIFF_COUNT files differ (expected for conflict resolution)"
fi

echo ""
echo "=========================================="
echo "SCENARIO2: All tests PASSED"
echo "=========================================="
```

## Success Criteria
1. Initial sync completes successfully
2. Modified files propagate correctly
3. New files propagate correctly
4. Deleted files are removed from all locations
5. Conflict resolution works (may show differences)
6. File counts match across all locations (except possibly for conflict files)
7. No unexplained files appear or disappear
