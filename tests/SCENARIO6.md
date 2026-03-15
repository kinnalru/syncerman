# SCENARIO6: Error Handling and Edge Cases

## Overview
Tests error handling, invalid configurations, and edge cases:
- Invalid configuration files
- Invalid paths and providers
- Missing directories and remotes
- Permission errors
- Network timeouts simulation
- Invalid command-line arguments

## Prerequisites
- rclone configured with `gd` and `yd` providers
- Access to syncerman binary at `/home/llm/agents/takopi/syncerman/bin/syncerman`
- Ability to create invalid configurations and break things intentionally

## Test Setup

```bash
#!/bin/bash
# Define test paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"
CONFIG_FILE="$TEST_DIR/scenario6.yaml"

# Create directories
mkdir -p "$LOCAL_DIR"
mkdir -p "$LOCAL2_DIR"

# Clean remote paths
rclone purge "gd:syncerman/scenario6/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario6/" --quiet 2>/dev/null || true
```

## Test Execution

### Test Case 1: Missing Configuration File

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"

echo "=== Test Case 1: Missing configuration file ==="
echo ""

echo "Attempting sync with non-existent config file..."
$SYNCERMAN_BIN sync --config "$TEST_DIR/nonexistent.yaml" 2>&1 | tee "$TEST_DIR/test1_output.txt"

EXIT_CODE=$?
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -ne 0 ]; then
  echo "OK: Failed as expected"
else
  echo "ERROR: Should have failed"
fi
```

**Expected Results:**
- Command fails
- Exit code non-zero (likely 5 for file not found)
- Error message about missing configuration file

### Test Case 2: Invalid YAML Syntax

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
INVALID_CONFIG="$TEST_DIR/invalid_yaml.yaml"

echo "=== Test Case 2: Invalid YAML syntax ==="
echo ""

echo "Creating invalid YAML file..."
cat > "$INVALID_CONFIG" << 'EOF'
local:
  '/path':
    - to: 'gd:path'
      invalid syntax here
EOF

echo "Attempting check..."
$SYNCERMAN_BIN check --config "$INVALID_CONFIG" 2>&1 | tee "$TEST_DIR/test2_output.txt"

EXIT_CODE=$?
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -ne 0 ]; then
  echo "OK: Failed as expected"
else
  echo "ERROR: Should have failed"
fi

# Check for YAML error in output
if grep -qi "yaml\|parse\|syntax" "$TEST_DIR/test2_output.txt"; then
  echo "OK: YAML/parse error detected"
else
  echo "WARNING: No YAML error message found"
fi
```

**Expected Results:**
- Configuration validation fails
- Exit code non-zero (likely 2 for config error)
- Error message about YAML parsing

### Test Case 3: Missing Required Field (to)

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
INVALID_CONFIG="$TEST_DIR/missing_to.yaml"

echo "=== Test Case 3: Missing required 'to' field ==="
echo ""

echo "Creating config without 'to' field..."
cat > "$INVALID_CONFIG" << 'EOF'
local:
  '/home/llm/agents/takopi/syncerman/tmp/complex/scenario6/local':
    -
      args: []
EOF

echo "Attempting check..."
$SYNCERMAN_BIN check --config "$INVALID_CONFIG" 2>&1 | tee "$TEST_DIR/test3_output.txt"

EXIT_CODE=$?
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -ne 0 ]; then
  echo "OK: Failed as expected"
else
  echo "ERROR: Should have failed"
fi

# Check for validation error
if grep -qi "validation\|required\|missing" "$TEST_DIR/test3_output.txt"; then
  echo "OK: Validation error detected"
else
  echo "WARNING: No validation error message found"
fi
```

**Expected Results:**
- Configuration validation fails
- Exit code non-zero (likely 2 for config error)
- Error message about missing required field

### Test Case 4: Invalid Provider Name

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
INVALID_CONFIG="$TEST_DIR/invalid_provider.yaml"

echo "=== Test Case 4: Invalid provider name ==="
echo ""

echo "Creating config with non-existent provider..."
cat > "$INVALID_CONFIG" << EOF
nonexistent:
  '$TEST_DIR/local':
    -
      to: 'gd:syncerman/scenario6/'
EOF

echo "Attempting check..."
$SYNCERMAN_BIN check --config "$INVALID_CONFIG" 2>&1 | tee "$TEST_DIR/test4_output.txt"

EXIT_CODE=$?
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -ne 0 ]; then
  echo "OK: Failed as expected"
else
  echo "ERROR: Should have failed"
fi

# Check for provider error
if grep -qi "provider\|not found\|configured" "$TEST_DIR/test4_output.txt"; then
  echo "OK: Provider error detected"
else
  echo "WARNING: No provider error message found"
fi
```

**Expected Results:**
- Remote check fails
- Exit code non-zero (likely 4 for validation error)
- Error message about provider not configured

### Test Case 5: Invalid Destination Path Format

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
INVALID_CONFIG="$TEST_DIR/invalid_dest.yaml"

echo "=== Test Case 5: Invalid destination path format ==="
echo ""

echo "Creating config with invalid destination..."
cat > "$INVALID_CONFIG" << EOF
local:
  '$TEST_DIR/local':
    -
      to: 'invalid_format_no_colon'

gd:
  'syncerman/scenario6/':
    -
      to: 'also_invalid'
EOF

echo "Attempting check..."
$SYNCERMAN_BIN check --config "$INVALID_CONFIG" 2>&1 | tee "$TEST_DIR/test5_output.txt"

EXIT_CODE=$?
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -ne 0 ]; then
  echo "OK: Failed as expected"
else
  echo "ERROR: Should have failed"
fi

# Check for format error
if grep -qi "format\|destination\|path" "$TEST_DIR/test5_output.txt"; then
  echo "OK: Format error detected"
else
  echo "WARNING: No format error message found"
fi
```

**Expected Results:**
- Configuration validation may pass or fail
- If passes, sync will fail later
- Better if caught during validation

### Test Case 6: Missing Source Directory

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
MISSING_DIR_CONFIG="$TEST_DIR/missing_dir.yaml"

echo "=== Test Case 6: Missing source directory ==="
echo ""

echo "Creating config with missing source directory..."
cat > "$MISSING_DIR_CONFIG" << EOF
local:
  '$TEST_DIR/nonexistent_directory':
    -
      to: 'gd:syncerman/scenario6/'

gd:
  'syncerman/scenario6/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario6/local2'
EOF

echo "Creating test file in local2 (to test reverse sync)..."
echo "Test file" > "$TEST_DIR/local2/test.txt"

echo "Attempting sync..."
$SYNCERMAN_BIN sync --config "$MISSING_DIR_CONFIG" 2>&1 | tee "$TEST_DIR/test6_output.txt"

EXIT_CODE=$?
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -ne 0 ]; then
  echo "OK: Failed as expected"
else
  echo "ERROR: Should have failed"
fi

# Check for directory error
if grep -qi "directory\|not found\|no such" "$TEST_DIR/test6_output.txt"; then
  echo "OK: Directory error detected"
else
  echo "WARNING: No directory error message found"
fi
```

**Expected Results:**
- Sync fails
- Exit code non-zero
- Error message about missing directory

### Test Case 7: Empty Configuration

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
EMPTY_CONFIG="$TEST_DIR/empty.yaml"

echo "=== Test Case 7: Empty configuration ==="
echo ""

echo "Creating empty config file..."
touch "$EMPTY_CONFIG"

echo "Attempting sync with empty config..."
$SYNCERMAN_BIN sync --config "$EMPTY_CONFIG" 2>&1 | tee "$TEST_DIR/test7_output.txt"

EXIT_CODE=$?
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -eq 0 ]; then
  echo "OK: Empty config is valid (no targets to sync)"
else
  echo "INFO: Empty config may be invalid (depends on implementation)"
fi
```

**Expected Results:**
- May pass with "no targets" message
- Or may fail with "invalid config"
- Both behaviors are acceptable

### Test Case 8: Invalid Command-Line Arguments

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=== Test Case 8: Invalid command-line arguments ==="
echo ""

echo "Subtest 8.1: Invalid flag..."
$SYNCERMAN_BIN sync --invalid-flag 2>&1 | head -5
echo ""

echo "Subtest 8.2: Too many arguments..."
$SYNCERMAN_BIN sync target1 target2 target3 2>&1 | head -5
echo ""

echo "Subtest 8.3: Invalid command..."
$SYNCERMAN_BIN nonexistent_command 2>&1 | head -5
```

**Expected Results:**
- Invalid flag shows error
- Too many arguments shows error
- Invalid command shows help/error

### Test Case 9: Multiple Providers, One Invalid

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
MIXED_CONFIG="$TEST_DIR/mixed_providers.yaml"

echo "=== Test Case 9: Mixed valid and invalid providers ==="
echo ""

echo "Creating config with both valid and invalid providers..."
cat > "$MIXED_CONFIG" << EOF
local:
  '$TEST_DIR/local':
    -
      to: 'gd:syncerman/scenario6/'

invalid_provider:
  '$TEST_DIR/local':
    -
      to: 'yd:syncerman/scenario6/'

gd:
  'syncerman/scenario6/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario6/local2'
EOF

echo "Attempting check..."
$SYNCERMAN_BIN check --config "$MIXED_CONFIG" 2>&1 | tee "$TEST_DIR/test9_output.txt"

EXIT_CODE=$?
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -ne 0 ]; then
  echo "OK: Failed due to invalid provider"
else
  echo "WARNING: Should detect invalid provider"
fi

# Check for specific provider errors
if grep -i "invalid_provider" "$TEST_DIR/test9_output.txt" | grep -qi "not found"; then
  echo "OK: Invalid provider detected"
fi
```

**Expected Results:**
- Should report invalid_provider as not found
- May report gd as OK
- Overall exit code non-zero

### Test Case 10: Sync to Read-Only Destination (if applicable)

```bash
#!/bin/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
READONLY_CONFIG="$TEST_DIR/readonly.yaml"

echo "=== Test Case 10: Sync with read-only restrictions ==="
echo ""
echo "Note: This test creates a scenario where sync might fail"
echo "due to permissions or read-only configuration."
echo ""

cat > "$READONLY_CONFIG" << EOF
local:
  '$TEST_DIR/local':
    -
      to: 'gd:syncerman/scenario6/'

gd:
  'syncerman/scenario6/':
    -
      to: 'yd:syncerman/scenario6/'

yd:
  'syncerman/scenario6/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario6/local2'
EOF

echo "Creating test file..."
echo "Read-only test file" > "$TEST_DIR/local/readonly_test.txt"
echo ""

echo "Attempting sync (should work unless permission issues)..."
$SYNCERMAN_BIN sync --config "$READONLY_CONFIG" --verbose 2>&1 | tee "$TEST_DIR/test10_output.txt"

EXIT_CODE=$?
echo ""
echo "Exit code: $EXIT_CODE"

if [ $EXIT_CODE -eq 0 ]; then
  echo "OK: Sync completed (no permission restrictions)"
else
  echo "INFO: Sync failed (permission or other error)"
  if grep -qi "permission\|denied" "$TEST_DIR/test10_output.txt"; then
    echo "Permission error detected"
  fi
fi
```

**Expected Results:**
- May pass (no restrictions)
- May fail with permission error
- Either is acceptable depending on environment

### Test Case 11: Simultaneous Sync Attempts

```bash
#!/bash
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
SIMUL_CONFIG="$TEST_DIR/simultaneous.yaml"

echo "=== Test Case 11: Simultaneous sync attempts ==="
echo ""

cat > "$SIMUL_CONFIG" << EOF
local:
  '$TEST_DIR/local':
    -
      to: 'gd:syncerman/scenario6/'

gd:
  'syncerman/scenario6/':
    -
      to: 'yd:syncerman/scenario6/'

yd:
  'syncerman/scenario6/':
    -
      to: '/home/llm/agents/takopi/syncerman/tmp/complex/scenario6/local2'
EOF

echo "Creating test file..."
echo "Simultaneous test file" > "$TEST_DIR/local/simultaneous.txt"
echo ""

echo "Starting simultaneous syncs (background processes)..."
$SYNCERMAN_BIN sync --config "$SIMUL_CONFIG" --verbose > "$TEST_DIR/sync1.log" 2>&1 &
PID1=$!

$SYNCERMAN_BIN sync --config "$SIMUL_CONFIG" --verbose > "$TEST_DIR/sync2.log" 2>&1 &
PID2=$!

echo "Waiting for processes to complete..."
wait $PID1
EXIT1=$?
wait $PID2
EXIT2=$?

echo ""
echo "Sync 1 exit code: $EXIT1"
echo "Sync 2 exit code: $EXIT2"

if [ $EXIT1 -ne 0 ] || [ $EXIT2 -ne 0 ]; then
  echo "INFO: At least one sync failed (expected for simultaneous attempts)"
  echo "Checking for lock or conflict errors..."
  grep -i "lock\|busy\|conflict" "$TEST_DIR/sync1.log" "$TEST_DIR/sync2.log" && echo "Found lock/conflict errors"
else
  echo "OK: Both syncs completed (rclone may handle concurrency)"
fi
```

**Expected Results:**
- May handle concurrency gracefully
- Or may fail with lock/busy errors
- Either is acceptable

## Cleanup

```bash
#!/bin/bash
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"

echo "=== Cleanup ==="
rm -rf "$TEST_DIR"
rclone purge "gd:syncerman/scenario6/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario6/" --quiet 2>/dev/null || true
rm -rf ~/.cache/rclone/bisync/*scenario6* 2>/dev/null || true
echo "Cleanup complete"
```

## Full Script (Ready to Execute)

```bash
#!/bin/bash

# Define paths
TEST_DIR="/home/llm/agents/takopi/syncerman/tmp/complex/scenario6"
LOCAL_DIR="$TEST_DIR/local"
LOCAL2_DIR="$TEST_DIR/local2"
SYNCERMAN_BIN="/home/llm/agents/takopi/syncerman/bin/syncerman"

echo "=========================================="
echo "SCENARIO6: Error Handling & Edge Cases"
echo "=========================================="
echo ""

# Cleanup and setup
echo "--- Setup ---"
rm -rf "$TEST_DIR"
mkdir -p "$LOCAL_DIR" "$LOCAL2_DIR"
rclone purge "gd:syncerman/scenario6/" --quiet 2>/dev/null || true
rclone purge "yd:syncerman/scenario6/" --quiet 2>/dev/null || true
echo "Setup complete"
echo ""

echo "Note: Some tests intentionally create invalid configurations."
echo "      Errors are expected and part of the test."
echo ""

# Test Case 1
echo "=== Test 1: Missing config file ==="
$SYNCERMAN_BIN sync --config "$TEST_DIR/nonexistent.yaml" 2>/dev/null
[ $? -ne 0 ] && echo "PASS: Failed as expected" || echo "FAIL: Should have failed"
echo ""

# Test Case 2
echo "=== Test 2: Invalid YAML ==="
cat > "$TEST_DIR/invalid.yaml" << 'EOF'
local:
  '/path':
    - to: 'gd:path'
      invalid syntax
EOF
$SYNCERMAN_BIN check --config "$TEST_DIR/invalid.yaml" 2>/dev/null
[ $? -ne 0 ] && echo "PASS: Failed as expected" || echo "FAIL: Should have failed"
echo ""

# Test Case 3
echo "=== Test 3: Missing 'to' field ==="
cat > "$TEST_DIR/missing_to.yaml" << EOF
local:
  '$LOCAL_DIR':
    -
      args: []
EOF
$SYNCERMAN_BIN check --config "$TEST_DIR/missing_to.yaml" 2>/dev/null
[ $? -ne 0 ] && echo "PASS: Failed as expected" || echo "FAIL: Should have failed"
echo ""

# Test Case 4
echo "=== Test 4: Invalid provider ==="
cat > "$TEST_DIR/invalid_provider.yaml" << EOF
nonexistent:
  '$LOCAL_DIR':
    -
      to: 'gd:syncerman/scenario6/'
EOF
$SYNCERMAN_BIN check --config "$TEST_DIR/invalid_provider.yaml" 2>/dev/null
[ $? -ne 0 ] && echo "PASS: Failed as expected" || echo "FAIL: Should have failed"
echo ""

# Test Case 5
echo "=== Test 5: Invalid dest format ==="
cat > "$TEST_DIR/invalid_dest.yaml" << EOF
local:
  '$LOCAL_DIR':
    -
      to: 'invalid_no_colon'
EOF
$SYNCERMAN_BIN check --config "$TEST_DIR/invalid_dest.yaml" 2>/dev/null
[ $? -ne 0 ] && echo "PASS: Failed as expected" || echo "INFO: May pass depending on validation"
echo ""

# Test Case 6
echo "=== Test 6: Missing source dir ==="
cat > "$TEST_DIR/missing_dir.yaml" << EOF
local:
  '$TEST_DIR/nonexistent':
    -
      to: 'gd:syncerman/scenario6/'
EOF
echo "test file" > "$LOCAL2_DIR/test.txt"
$SYNCERMAN_BIN sync --config "$TEST_DIR/missing_dir.yaml" 2>/dev/null
[ $? -ne 0 ] && echo "PASS: Failed as expected" || echo "INFO: May not fail depending on rclone behavior"
echo ""

# Test Case 7
echo "=== Test 7: Empty config ==="
touch "$TEST_DIR/empty.yaml"
$SYNCERMAN_BIN sync --config "$TEST_DIR/empty.yaml" 2>/dev/null
EXIT=$?
[ $EXIT -eq 0 ] && echo "PASS: Empty config accepted (or no sync)" || echo "INFO: Empty config may be invalid"
echo ""

# Test Case 8
echo "=== Test 8: Invalid CLI args ==="
$SYNCERMAN_BIN --invalid-flag 2>/dev/null
[ $? -ne 0 ] && echo "PASS: Invalid flag rejected" || echo "FAIL: Should reject invalid flag"
echo ""

echo ""
echo "=========================================="
echo "SCENARIO6: Error handling tests complete"
echo "Some failures are expected and correct"
echo "=========================================="
```

## Success Criteria
1. Missing configuration file detected
2. Invalid YAML syntax caught
3. Missing required fields detected
4. Invalid providers identified
5. Invalid path formats rejected (if validated)
6. Missing directories cause errors
7. Empty configuration handled
8. Invalid CLI arguments rejected
9. Mixed valid/invalid providers partially validated
10. Permission errors handled (if applicable)
11. Concurrency issues detected or handled gracefully
