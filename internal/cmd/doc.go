package cmd

// Package cmd provides CLI command definitions and execution logic for Syncerman.
//
// It uses the cobra framework to define the root command and subcommands for
// synchronization operations, configuration checking, and rclone verification.
//
// Commands:
//
//   syncerman sync [target] [flags]
//       Executes bidirectional synchronization using rclone bisync.
//       Without arguments, syncs all targets from configuration file.
//       With target argument (provider:path), syncs only that specific target.
//
//   syncerman check config [flags]
//       Validates configuration file and checks that all configured
//       providers and destinations are valid.
//
//   syncerman check remotes [flags]
//       Verifies that all providers in configuration are properly
//       configured in rclone.
//
//   syncerman version
//       Prints version number of Syncerman.
//
// Global Flags:
//
//   -c, --config string
//       Path to configuration file (default: ./syncerman.yaml or ./syncerman.yml)
//
//   -d, --dry-run
//       Dry run mode - show what would be done without making changes
//
//   -v, --verbose
//       Verbose output - show detailed information during execution
//
//   -q, --quiet
//       Quiet mode - suppress non-error output
//
// Usage Examples:
//
//   # Sync all targets from configuration
//   syncerman sync
//
//   # Sync specific target with dry-run
//   syncerman sync gdrive:docs --dry-run
//
//   # Check configuration validity
//   syncerman check config
//
//   # Check rclone remotes with verbose output
//   syncerman check remotes --verbose
//
// Command Exit Codes:
//
//   0 - Success (all operations completed)
//   1 - Error (configuration invalid, sync failed, or remote not found)
//
// Error Handling:
//
// All commands use structured error messages and return appropriate exit codes.
// Errors are printed to stderr with actionable guidance.
// Use --verbose flag for more detailed error information.
