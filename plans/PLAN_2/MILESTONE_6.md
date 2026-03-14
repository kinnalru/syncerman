# Milestone 6: Document internal/cmd/*

## Goal
Enhance documentation for CLI command package, improving detailed command handler documentation for all CLI commands and handlers.

## Context

Current state analysis:
- internal/cmd/ has detailed doc.go
- Command handlers need thorough documentation
- Flag binding and validation require comments
- Command execution flow needs explanation

Reference documents:
- `guides/OVERALL.md:112-249` - CLI reference and examples
- `guides/OVERALL.md:112` - Global flags
- `guides/OVERALL.md:124-231` - Command definitions

## Tasks

### Task 6.1: Enhance root.go doc.go
Update root command documentation:
- Application overview
- Global flags (--config, --dry-run, --verbose, --quiet)
- Configuration file discovery logic
- Default behaviors

### Task 6.2: Document Execute function
Add function documentation for Execute:
- Purpose - main entry point for CLI execution
- Cobra command registration
- Error handling
- Exit codes

### Task 6.3: Document PreRun function (root command)
Add function documentation for PreRun:
- Purpose - pre-execution setup
- Configuration loading
- Logger initialization
- Early validation

### Task 6.4: Enhance sync_cmd.go doc.go
Update sync command documentation:
- Purpose - sync all or specific targets
- Usage patterns (all targets vs single target)
- Target format specifications
- Reference to OVERALL.md:125-237

### Task 6.5: Document syncAll function
Add function documentation for syncAll:
- Purpose - handler for "sync" command (sync all targets)
- Flow: config validation → remote verification → sync execution
- Error handling and reporting

### Task 6.6: Document syncTarget function
Add function documentation for syncTarget:
- Purpose - handler for "sync <provider:path>" command
- Target parameter parsing
- Provider and path extraction
- Target validation
- Error cases (invalid format, target not found)

### Task 6.7: Enhance check_cmd.go doc.go
Update check command documentation:
- Purpose - validation commands (config, remotes)
- Available subcommands
- Early detection of configuration issues
- Reference to OVERALL.md:164-204

### Task 6.8: Document checkConfig function
Add function documentation for checkConfig:
- Purpose - handler for "check config" command
- Validation steps:
  - YAML syntax check
  - Configuration structure validation
  - Provider name validation
  - Source path validation
  - Destination format validation
- Output format and error messages

### Task 6.9: Document checkRemotes function
Add function documentation for checkRemotes:
- Purpose - handler for "check remotes" command
- Verification steps:
  - Rclone binary check
  - Remote existence verification
  - Connection testing
- Error reporting

### Task 6.10: Document flag binding
Add comments for:
- Global flag binding and usage
- Command-specific flags
- Flag validation logic
- Flag conflicts and dependencies

### Task 6.11: Document error handling
Add documentation for:
- Error message formatting
- Exit code determination
- Verbose vs quiet output modes
- User-friendly error suggestions
