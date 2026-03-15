---
title: "Milestone 5: CLI Commands Implementation"
status: "completed"
---

# Milestone 5: CLI Commands Implementation

## Goal

Implement all CLI command variants using Cobra framework with proper flag handling and command structure.

## Context

Build CLI interface for Syncerman application (OVERALL.md: lines 125-204):
- Commands required:
  - `sync [flags]` - Sync all targets from configuration
  - `sync <provider>:<path> [flags]` - Sync specific target
  - `check config [flags]` - Check configuration validity
  - `check remotes [flags]` - Check rclone remotes
- Global flags: --config, --dry-run, --verbose, --quiet
- Cobra-based CLI structure
- Must use previously implemented: config, rclone, and sync packages

## Tasks

### 5.1: Define CLI Root Command Structure

Create root Cobra command structure:
- Define root command with "syncerman" as command name
- Add version command (`syncerman version`)
- Add help command (built-in to Cobra)
- Add global flags: --config, --dry-run, --verbose, --quiet
- Implement flag persistence across subcommands
- Set up logger integration with global flags

### 5.2: Implement Sync Command

Create `sync` command with variants:
- `sync` subcommand: Sync all targets from configuration
  - Load config from --config path (default: "./syncerman.yaml")
  - Use sync.Engine.RunAll() to execute all targets
  - Apply global flags (dry-run, verbose, quiet)
  - Display formatted report from sync results
  - Return appropriate exit code based on results
- `sync <target>` variant: Sync specific provider:path target
  - Parse target argument as "provider:path" format
  - Use sync.Engine.Run() to execute single target
  - Apply global flags
  - Display single result with formatted output
  - Return appropriate exit code

### 5.3: Implement Check Config Command

Create `check config` command:
- Load configuration file from --config path
- Validate configuration using sync.Engine.Validate()
- Validate targets using sync.Engine.ValidateTargets()
- Display validation results with clear messaging
- List all valid providers and paths
- Show any validation errors with context
- Return exit code: 0 if valid, 1 if invalid

### 5.4: Implement Check Remotes Command

Create `check remotes` command:
- List all configured rclone remotes
- Use rclone.ListRemotes() to get remote list
- Check if providers in config exist in rclone
- Display valid and missing providers
- Use sync.Engine.RemoteProviderExists() for verification
- Return exit code: 0 if all providers exist, 1 if any missing

### 5.5: Add Configuration File Defaults

Implement default configuration handling:
- Check for default config file path: "./syncerman.yaml"
- Check alternative path: "$HOME/.syncerman/syncerman.yaml"
- Use first found valid path if --config not specified
- Create error if no config file found when required
- Default to verbose=false, dry-run=false, quiet=false

### 5.6: Implement Flag Integration

Connect CLI flags to sync engine options:
- Map --dry-run to SyncOptions.DryRun
- Map --verbose to SyncOptions.Verbose
- Map --quiet to SyncOptions.Quiet
- Support --config for custom config path
- Pass flags as SyncOptions to sync Engine methods
- Ensure flag overrides work correctly with engine.SetDryRun()

### 5.7: Add Command Help and Usage

Add comprehensive command documentation:
- Add descriptions for each command
- Add examples in command help
- Document command variants
- Document global flags with help text
- Add error messages with actionable guidance
- Provide clear success/failure feedback

### 5.8: Implement Main Entry Point

Create main.go application entry:
- Initialize logger
- Parse CLI arguments with Cobra
- Execute appropriate command
- Handle PANIC-level errors gracefully
- Ensure clean shutdown with appropriate exit codes
- Set up graceful error handling for all commands

### 5.9: Write Unit Tests for CLI Commands

Create comprehensive tests for CLI:
- Test root command structure and global flags
- Test sync command with all flags combinations
- Test sync <target> variant with various inputs
- Test check config command with valid/invalid configs
- Test check remotes command
- Test default configuration file handling
- Test error messages and exit codes
- Mock sync engine and rclone executor where needed
- Target 85%+ test coverage for cmd package

### 5.10: Update Package Documentation

Ensure proper documentation of CLI:
- Add godoc comments to all exported functions
- Document command structure and variants
- Document global flags and their behavior
- Add usage examples for each command
- Document error handling strategies
- Add package-level documentation with CLI workflow overview
