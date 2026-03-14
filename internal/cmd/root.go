package cmd

import (
	"fmt"
	"os"

	"syncerman/internal/logger"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dryRun  bool
	verbose bool
	quiet   bool
	log     logger.Logger
)

var rootCmd = &cobra.Command{
	Use:   "syncerman",
	Short: "Synchronizing targets using rclone",
	Long: `Syncerman is a CLI application for synchronizing targets
(sources and destinations) based on rclone CLI.

It provides a simple configuration-driven approach to manage
bidirectional synchronization between different storage providers.

Features:
  - Bidirectional sync using rclone bisync
  - Declarative YAML configuration
  - Dry-run mode for safe testing
  - First-run error detection and auto-retry
  - Automatic destination directory creation

Usage:
  syncerman [command] [flags]

Available Commands:
  sync       Synchronize targets from configuration
  check      Check configuration and verify rclone remotes
  version     Print version number

Global Flags:
  -c, --config string   Path to configuration file (default: ./syncerman.yaml)
  -d, --dry-run        Dry run mode (show what would be done)
  -v, --verbose         Verbose output
  -q, --quiet          Quiet mode (suppress output)

Examples:
  # Sync all targets from configuration
  syncerman sync

  # Sync specific target
  syncerman sync gdrive:docs

  # Check configuration validity
  syncerman check config

  # Check rclone remotes
  syncerman check remotes

  # Dry-run with verbose output
  syncerman sync --dry-run --verbose`,
	Version: "0.1.0",
}

// Execute is the main entry point for CLI execution.
// It calls the root cobra command's Execute method which parses
// arguments and runs the appropriate command. Cobra handles error
// display and return values automatically.
//
// Returns:
//
//	error: Any error returned by cobra's Execute method
func Execute() error {
	return rootCmd.Execute()
}

// init initializes the root command by registering cobra initialization and
// binding global command-line flags.
//
// This function is called automatically by Go when the package is loaded.
// It registers the initConfig function to be called before command execution
// and binds the following persistent flags (available to all subcommands):
//
// Flags:
//
//	-c, --config string
//	    Path to configuration file. If not specified, defaults to searching
//	    for ./syncerman.yaml or ./syncerman.yml in the current directory.
//
//	-d, --dry-run
//	    Dry run mode. When enabled, shows what operations would be performed
//	    without actually executing any changes.
//
//	-v, --verbose
//	    Verbose output mode. Enables detailed logging including debug messages.
//
//	-q, --quiet
//	    Quiet mode. Suppresses all non-error output. Cannot be used with --verbose.
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run mode (show what would be done)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (suppress output)")
}

// initConfig initializes the logger instance based on command-line flags
// and validates flag configuration.
//
// This function is called by cobra before command execution (registered via
// cobra.OnInitialize in init()). It performs the following operations:
//
//  1. Creates a new console logger instance
//  2. Validates that --verbose and --quiet flags are not used together (fatal error)
//  3. Sets logger verbosity based on flags:
//     - If --verbose: enables verbose mode with debug output
//     - If --quiet: enables quiet mode (only errors)
//     - Default: normal mode without debug output
//  4. Logs the configuration file path if explicitly specified via --config
//
// Error Handling:
//   - Exits with code 1 if both --verbose and --quiet flags are set
func initConfig() {
	log = logger.NewConsoleLogger()

	if verbose && quiet {
		fmt.Fprintln(os.Stderr, "Error: cannot use both --verbose and --quiet")
		os.Exit(1)
	}

	if verbose {
		log.SetVerbose(true)
	} else if quiet {
		log.SetQuiet(true)
	}

	if cfgFile != "" {
		log.Debug("Using config file: %s", cfgFile)
	}
}

// GetLogger returns the logger instance, creating one if necessary.
//
// This function implements lazy initialization pattern for the logger.
// If the logger has not been initialized (e.g., called before initConfig),
// it creates a new default console logger instance.
//
// Returns:
//
//	logger.Logger: The logger instance for use by other packages
func GetLogger() logger.Logger {
	if log == nil {
		log = logger.NewConsoleLogger()
	}
	return log
}

// GetConfigFile returns the configuration file path from the --config flag.
//
// If the --config flag was not specified, returns an empty string. The caller
// should handle discovery of default configuration files (syncerman.yaml/yml)
// if needed.
//
// Returns:
//
//	string: The configuration file path, or empty string if not specified
func GetConfigFile() string {
	return cfgFile
}

// IsDryRun returns the state of the --dry-run flag.
//
// When true, the application should only show what operations would be
// performed without actually executing any changes.
//
// Returns:
//
//	bool: True if dry-run mode is enabled, false otherwise
func IsDryRun() bool {
	return dryRun
}

// IsVerbose returns the state of the --verbose flag.
//
// When true, the application should output detailed information including
// debug messages during execution. Cannot be used with --quiet.
//
// Returns:
//
//	bool: True if verbose mode is enabled, false otherwise
func IsVerbose() bool {
	return verbose
}

// IsQuiet returns the state of the --quiet flag.
//
// When true, the application should suppress all non-error output.
// Cannot be used with --verbose.
//
// Returns:
//
//	bool: True if quiet mode is enabled, false otherwise
func IsQuiet() bool {
	return quiet
}
