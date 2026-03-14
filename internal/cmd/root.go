package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"syncerman/internal/logger"
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

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run mode (show what would be done)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (suppress output)")
}

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

func GetLogger() logger.Logger {
	if log == nil {
		log = logger.NewConsoleLogger()
	}
	return log
}

func GetConfigFile() string {
	return cfgFile
}

func IsDryRun() bool {
	return dryRun
}

func IsVerbose() bool {
	return verbose
}

func IsQuiet() bool {
	return quiet
}
