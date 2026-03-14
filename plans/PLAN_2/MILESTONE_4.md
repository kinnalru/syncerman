# Milestone 4: Document internal/rclone/*

## Goal
Enhance documentation for rclone integration package, focusing on complex types, command execution, and verification functions.

## Context

Current state analysis:
- internal/rclone/ has basic doc.go
- Complex types (RcloneConfig, SyncConfig) need full documentation
- Command execution and output parsing require detailed comments
- Error handling and verification logic needs explanation

Reference documents:
- `guides/OVERALL.md:250-337` - Rclone integration details
- `guides/OVERALL.md:256-284` - Rclone bisync command template

## Tasks

### Task 4.1: Enhance rclone.go doc.go
Update package documentation to explain:
- Rclone integration architecture
- Command execution model
- Output parsing strategy
- Error detection and handling

### Task 4.2: Document RcloneConfig struct
Add comprehensive documentation for RcloneConfig:
- Purpose - rclone command configuration
- Binary path and verification
- Default settings
- Connection handling

### Task 4.3: Document SyncConfig struct
Add comprehensive documentation for SyncConfig:
- Purpose - sync operation configuration
- Source and destination format
- Command-line arguments
- Resync flag behavior

### Task 4.4: Document NewRclone function
Add function documentation for NewRclone:
- Purpose - creates new RcloneConfig instance
- Path verification logic
- Error cases - binary not found
- Default initialization

### Task 4.5: Document VerifyRemote function
Add function documentation for VerifyRemote:
- Purpose - verifies remote exists in rclone config
- Parameters - provider name
- Returns - error if remote not found
- Implementation details (rclone listremotes)
- Error messages

### Task 4.6: Document VerifyPath function
Add function documentation for VerifyPath:
- Purpose - creates destination path if needed
- Parameters - provider and path
- Error cases - permissions, invalid path
- Implementation (rclone mkdir)

### Task 4.7: Document ExecuteBisync function
Add function documentation for ExecuteBisync:
- Purpose - executes rclone bisync command
- Parameters - SyncConfig with all details
- Returns - stdout, stderr, error
- Command construction details
- Error handling approach

### Task 4.8: Document BuildBisyncCommand function
Add function documentation for BuildBisyncCommand:
- Purpose - constructs rclone bisync command string
- Parameters - SyncConfig structure
- Returns - complete command string
- Argument construction logic
- Base options from OVERALL.md:256-267

### Task 4.9: Document IsFirstRunError function
Add function documentation for IsFirstRunError:
- Purpose - detects first-run error pattern
- Parameters - error or string to check
- Returns - true if first-run error detected
- Regular expression pattern
- Reference to OVERALL.md:317-326

### Task 4.10: Add inline comments for complex logic
Document complex implementation details:
- Command string construction and escaping
- Output parsing logic
- Error pattern matching
- Remote format validation
