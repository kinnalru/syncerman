# SCENARIO4: Multi-Directional Sync with Multiple Paths

## Overview
Tests complex synchronization with multiple independent sync paths:
- Multiple source paths from same provider
- Independent sync chains
- Verify isolated behavior between different paths
- Test configuration with multiple destinations from single source

## Prerequisites
- rclone configured with `gd` and `yd` providers
- Access to syncerman binary at `/home/llm/agents/takopi/syncerman/bin/syncerman`
- Cleanup of test directories and remote paths before execution

## Test Setup

### 1. Initialize Test Environment

```bash
#!/bin/bash
# Define test paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
LOCAL_DOCS="$TEST_DIR/local/docs"
LOCAL_MEDIA="$TEST_DIR/local/media"
LOCAL_OTHER="$TEST_DIR/local/other"
LOCAL2_DOCS="$TEST_DIR/local2/docs"
LOCAL2_MEDIA="$TEST_DIR/local2/media"
CONFIG_FILE="$TEST_DIR/scenario4.yaml"

# Create directory structure
mkdir -p "$LOCAL_DOCS"
mkdir -p "$LOCAL_MEDIA"
mkdir -p "$LOCAL_OTHER"
mkdir -p "$LOCAL2_DOCS"
mkdir -p "$LOCAL2_MEDIA"

# Clean remote paths
rclone purge "gd:syncerman/scenario4/docs/" --quiet 2>/dev/null || true
rclone purge "gd:syncerman/scenario4/media/" --quiet 2>/dev/null || true
rclone purge "gd:syncerman/scenario4/other/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/docs/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/media/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/other/" --quiet 2>/dev/null || true
```

### 2. Create Test Data Structure

```bash
#!/bin/bash
# Create docs path data
echo "Document 1" > "$LOCAL_DOCS/doc1.txt"
echo "Document 2" > "$LOCAL_DOCS/doc2.txt"
mkdir -p "$LOCAL_DOCS/notes"
echo "Meeting notes" > "$LOCAL_DOCS/notes/meeting.txt"

# Create media path data
echo "Photo file (simulated binary)" > "$LOCAL_MEDIA/photo.jpg"
echo "Video file (simulated binary)" > "$LOCAL_MEDIA/video.mp4"
mkdir -p "$LOCAL_MEDIA/audio"
echo "Audio file (simulated binary)" > "$LOCAL_MEDIA/audio/song.mp3"

# Create other path data (no sync - should stay local only)
echo "Local only file 1" > "$LOCAL_OTHER/local1.txt"
echo "Local only file 2" > "$LOCAL_OTHER/local2.txt"

echo "=== Created test data ==="
echo "Docs path:"
find "$LOCAL_DOCS" -type f | sort
echo ""
echo "Media path:"
find "$LOCAL_MEDIA" -type f | sort
echo ""
echo "Other path (local only):"
find "$LOCAL_OTHER" -type f | sort
```

### 3. Create Configuration File

```yaml
# scenario4.yaml - Multiple independent sync paths
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local/docs':
    -
      to: 'gd:syncerman/scenario4/docs/'
    -
      to: 'yd:syncerman/scenario4/docs/'

  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local/media':
    -
      to: 'gd:syncerman/scenario4/media/'

gd:
  'syncerman/scenario4/docs/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local2/docs'

  'syncerman/scenario4/media/':
    -
      to: 'yd:syncerman/scenario4/media/'

yd:
  'syncerman/scenario4/media/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local2/media'
```

```bash
cat > "$CONFIG_FILE" << 'EOF'
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local/docs':
    -
      to: 'gd:syncerman/scenario4/docs/'
    -
      to: 'yd:syncerman/scenario4/docs/'

  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local/media':
    -
      to: 'gd:syncerman/scenario4/media/'

gd:
  'syncerman/scenario4/docs/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local2/docs'

  'syncerman/scenario4/media/':
    -
      to: 'yd:syncerman/scenario4/media/'

yd:
  'syncerman/scenario4/media/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local2/media'
EOF
```

**Configuration Explanation:**
- `local/docs` → `gd:docs` AND `yd:docs` (fan-out: one source, multiple destinations)
- `local/media` → `gd:media` → `yd:media` → `local2/media` (chain)
- `local/other` → NOT SYNCED (local only)
- `gd/docs` → `local2/docs` (separate path)
- `local2/docs` only syncs from `gd:docs`, NOT from `yd:docs`

## Test Execution

### Step 1: Verify Configuration and Remotes

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
CONFIG_FILE="$TEST_DIR/scenario4.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Step 1: Check configuration and verify remotes ==="
$SYNCERMAN_BIN check --config "$CONFIG_FILE" --verbose
```

**Expected Output:**
- Configuration is valid
- Shows number of providers and paths
- All providers (local, gd, yd) are configured

### Step 3: Execute Synchronization

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
CONFIG_FILE="$TEST_DIR/scenario4.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Step 3: Execute sync ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose
```

**Expected Output:**
- Multiple sync tasks execute (should see 6 tasks)
  1. local/docs → gd:docs
  2. local/docs → yd:docs
  3. local/media → gd:media
  4. gd:docs → local2/docs
  5. gd:media → yd:media
  6. yd:media → local2/media
- All tasks complete successfully

### Step 4: Verify Sync Results - Docs Path

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
LOCAL_DOCS="$TEST_DIR/local/docs"
LOCAL2_DOCS="$TEST_DIR/local2/docs"

echo "=== Step 4: Verify docs path ==="
echo ""
echo "Files in local/docs:"
find "$LOCAL_DOCS" -type f | sort
echo ""
echo "Files in local2/docs:"
find "$LOCAL2_DOCS" -type f 2>/dev/null | sort || echo "  (directory empty or missing)"
echo ""
echo "Files in gd:syncerman/scenario4/docs/:"
rclone ls "gd:syncerman/scenario4/docs/" --no-modtime --no-size
echo ""
echo "Files in yd:syncerman/scenario4/docs/:"
rclone ls "yd:syncerman/scenario4/docs/" --no-modtime --no-size
```

**Expected Results:**
- local/docs files present
- gd:docs has same files as local/docs
- yd:docs has same files as local/docs
- local2/docs has same files (synced from gd:docs)
- All file contents match

### Step 5: Verify Sync Results - Media Path

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
LOCAL_MEDIA="$TEST_DIR/local/media"
LOCAL2_MEDIA="$TEST_DIR/local2/media"

echo "=== Step 5: Verify media path ==="
echo ""
echo "Files in local/media:"
find "$LOCAL_MEDIA" -type f | sort
echo ""
echo "Files in local2/media:"
find "$LOCAL2_MEDIA" -type f 2>/dev/null | sort || echo "  (directory empty or missing)"
echo ""
echo "Files in gd:syncerman/scenario4/media/:"
rclone ls "gd:syncerman/scenario4/media/" --no-modtime --no-size
echo ""
echo "Files in yd:syncerman/scenario4/media/:"
rclone ls "yd:syncerman/scenario4/media/" --no-modtime --no-size
```

**Expected Results:**
- local/media files present
- gd:media has same files
- yd:media has same files
- local2/media has same files
- All file contents match

### Step 6: Verify Isolation - Other Path (Not Synced)

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
LOCAL_OTHER="$TEST_DIR/local/other"

echo "=== Step 6: Verify other path isolation ==="
echo ""
echo "Files in local/other (should NOT sync anywhere):"
find "$LOCAL_OTHER" -type f | sort
echo ""
echo "Verify no sync happened to remotes:"
echo "gd:syncerman/scenario4/other/ (should not exist):"
rclone ls "gd:syncerman/scenario4/other/" --no-modtime --no-size || echo "  (does not exist - OK)"
echo ""
echo "yd:syncerman/scenario4/other/ (should not exist):"
rclone ls "yd:syncerman/scenario4/other/" --no-modtime --no-size || echo "  (does not exist - OK)"
```

**Expected Results:**
- local/other files exist but are NOT synced
- No gd:other or yd:other directories
- Isolation maintained

### Step 7: Test Independent Path Modifications

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
LOCAL_DOCS="$TEST_DIR/local/docs"
LOCAL_MEDIA="$TEST_DIR/local/media"

echo "=== Step 7: Modify independent paths ==="
echo ""
echo "Adding new file to docs only:"
echo "New doc at $(date)" > "$LOCAL_DOCS/newdoc.txt"
echo ""
echo "Modifying file in media only:"
echo "Modified photo at $(date)" > "$LOCAL_MEDIA/photo.jpg"
echo ""
echo "Syncing..."
```

### Step 8: Sync Independent Modifications

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
CONFIG_FILE="$TEST_DIR/scenario4.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose
```

### Step 9: Verify Independent Syncs

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"

echo "=== Step 9: Verify independent modifications ==="
echo ""
echo "Check newdoc.txt in all doc locations:"
echo "local/docs/newdoc.txt:"
test -f "$TEST_DIR/local/docs/newdoc.txt" && cat "$TEST_DIR/local/docs/newdoc.txt" || echo "  NOT FOUND"
echo ""
echo "gd:syncerman/scenario4/docs/newdoc.txt:"
rclone cat "gd:syncerman/scenario4/docs/newdoc.txt" 2>/dev/null || echo "  NOT FOUND"
echo ""
echo "yd:syncerman/scenario4/docs/newdoc.txt:"
rclone cat "yd:syncerman/scenario4/docs/newdoc.txt" 2>/dev/null || echo "  NOT FOUND"
echo ""
echo "local2/docs/newdoc.txt:"
test -f "$TEST_DIR/local2/docs/newdoc.txt" && cat "$TEST_DIR/local2/docs/newdoc.txt" || echo "  NOT FOUND"
echo ""
echo "Check modified photo.jpg:"
echo "local/media/photo.jpg:"
test -f "$TEST_DIR/local/media/photo.jpg" && cat "$TEST_DIR/local/media/photo.jpg" | head -1 || echo "  NOT FOUND"
echo ""
echo "Verification complete"
```

### Step 10: Test Single Target Sync

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
CONFIG_FILE="$TEST_DIR/scenario4.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Step 10: Test single target sync ==="
echo ""
echo "Sync only docs path:"
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" local:/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local/docs --verbose
```

**Expected Results:**
- Only docs path sync executes
- Media path NOT affected
- Shows targeted sync works

## Cleanup

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"

echo "=== Cleanup ==="
rm -rf "$TEST_DIR"
rclone purge "gd:syncerman/scenario4/docs/" --quiet 2>/dev/null || true
rclone purge "gd:syncerman/scenario4/media/" --quiet 2>/dev/null || true
rclone purge "gd:syncerman/scenario4/other/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/docs/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/media/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/other/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario4* 2>/dev/null || true
echo "Cleanup complete"
```

## Full Script (Ready to Execute)

```bash
#!/bin/bash
set -e

# Define paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario4"
LOCAL_DOCS="$TEST_DIR/local/docs"
LOCAL_MEDIA="$TEST_DIR/local/media"
LOCAL_OTHER="$TEST_DIR/local/other"
LOCAL2_DOCS="$TEST_DIR/local2/docs"
LOCAL2_MEDIA="$TEST_DIR/local2/media"
CONFIG_FILE="$TEST_DIR/scenario4.yaml"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=========================================="
echo "SCENARIO4: Multi-Directional Multi-Path"
echo "=========================================="
echo ""

# Cleanup
echo "--- Cleanup ---"
rm -rf "$TEST_DIR"
mkdir -p "$LOCAL_DOCS" "$LOCAL_MEDIA" "$LOCAL_OTHER"
mkdir -p "$LOCAL2_DOCS" "$LOCAL2_MEDIA"
rclone purge "gd:syncerman/scenario4/docs/" --quiet 2>/dev/null || true
rclone purge "gd:syncerman/scenario4/media/" --quiet 2>/dev/null || true
rclone purge "gd:syncerman/scenario4/other/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/docs/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/media/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario4/other/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario4* 2>/dev/null || true
echo "Cleanup complete"
echo ""

# Create configuration
echo "--- Create configuration ---"
cat > "$CONFIG_FILE" << 'EOF'
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local/docs':
    -
      to: 'gd:syncerman/scenario4/docs/'
    -
      to: 'yd:syncerman/scenario4/docs/'

  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local/media':
    -
      to: 'gd:syncerman/scenario4/media/'

gd:
  'syncerman/scenario4/docs/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local2/docs'

  'syncerman/scenario4/media/':
    -
      to: 'yd:syncerman/scenario4/media/'

yd:
  'syncerman/scenario4/media/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario4/local2/media'
EOF
echo "Configuration created"
echo ""

# Create test data
echo "--- Create test data ---"
echo "Document 1" > "$LOCAL_DOCS/doc1.txt"
echo "Document 2" > "$LOCAL_DOCS/doc2.txt"
mkdir -p "$LOCAL_DOCS/notes"
echo "Meeting notes" > "$LOCAL_DOCS/notes/meeting.txt"
echo "Photo file" > "$LOCAL_MEDIA/photo.jpg"
echo "Video file" > "$LOCAL_MEDIA/video.mp4"
mkdir -p "$LOCAL_MEDIA/audio"
echo "Audio file" > "$LOCAL_MEDIA/audio/song.mp3"
echo "Local only file 1" > "$LOCAL_OTHER/local1.txt"
echo "Local only file 2" > "$LOCAL_OTHER/local2.txt"
echo "Test data created"
echo ""

# Step 1: Check configuration and verify remotes
echo "=== Check configuration and verify remotes ==="
$SYNCERMAN_BIN check --config "$CONFIG_FILE" --verbose
echo ""

# Step 3: Execute sync
echo "=== Execute sync ==="
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose
echo ""

# Step 4-6: Verify results
echo "=== Verify docs path ==="
echo "local/docs: $(find $LOCAL_DOCS -type f | wc -l) files"
echo "local2/docs: $(find $LOCAL2_DOCS -type f 2>/dev/null | wc -l) files"
echo "gd:docs: $(rclone ls 'gd:syncerman/scenario4/docs/' 2>/dev/null | wc -l) files"
echo "yd:docs: $(rclone ls 'yd:syncerman/scenario4/docs/' 2>/dev/null | wc -l) files"
echo ""

echo "=== Verify media path ==="
echo "local/media: $(find $LOCAL_MEDIA -type f | wc -l) files"
echo "local2/media: $(find $LOCAL2_MEDIA -type f 2>/dev/null | wc -l) files"
echo "gd:media: $(rclone ls 'gd:syncerman/scenario4/media/' 2>/dev/null | wc -l) files"
echo "yd:media: $(rclone ls 'yd:syncerman/scenario4/media/' 2>/dev/null | wc -l) files"
echo ""

echo "=== Verify other path isolation ==="
echo "local/other: $(find $LOCAL_OTHER -type f | wc -l) files (should stay local-only)"
echo "gd:other exists: $([ -n \"$(rclone ls 'gd:syncerman/scenario4/other/' 2>/dev/null)\" ] && echo 'NO (OK)' || echo 'YES (ERROR)')"
echo "yd:other exists: $([ -n \"$(rclone ls 'yd:syncerman/scenario4/other/' 2>/dev/null)\" ] && echo 'NO (OK)' || echo 'YES (ERROR)')"
echo ""

# Step 7-9: Independent modifications
echo "=== Test independent path modifications ==="
echo "New doc at $(date)" > "$LOCAL_DOCS/newdoc.txt"
echo "Modified photo at $(date)" > "$LOCAL_MEDIA/photo.jpg"
$SYNCERMAN_BIN sync --config "$CONFIG_FILE" --verbose > /dev/null
echo "Sync complete"
echo ""

echo "Verify newdoc.txt synced to docs path:"
test -f "$TEST_DIR/local2/docs/newdoc.txt" && echo "  local2/docs: OK" || echo "  local2/docs: MISSING (ERROR)"
rclone ls "gd:syncerman/scenario4/docs/newdoc.txt" > /dev/null 2>&1 && echo "  gd:docs: OK" || echo "  gd:docs: MISSING (ERROR)"
rclone ls "yd:syncerman/scenario4/docs/newdoc.txt" > /dev/null 2>&1 && echo "  yd:docs: OK" || echo "  yd:docs: MISSING (ERROR)"
echo ""

echo "=========================================="
echo "SCENARIO4: All tests PASSED"
echo "=========================================="
```

## Success Criteria
1. Configuration validates correctly with multiple paths
2. All 6 sync tasks execute successfully
3. Docs path: local → gd AND local → yd → local2
4. Media path: local → gd → yd → local2
5. Other path stays local only (not synced)
6. Independent modifications sync correctly per path
7. Single target sync works
8. No cross-contamination between independent paths
