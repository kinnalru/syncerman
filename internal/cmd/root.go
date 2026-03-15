package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"syncerman/internal/logger"

	"github.com/spf13/cobra"
)

const (
	exitCodeSuccess         = 0
	exitCodeGeneralError    = 1
	exitCodeConfigError     = 2
	exitCodeRcloneError     = 3
	exitCodeValidationError = 4
	exitCodeFileNotFound    = 5
	defaultContextTimeout   = 30 * time.Minute
)

type CommandConfig struct {
	ConfigFile string
	DryRun     bool
	Verbose    bool
	Quiet      bool
	Logger     *logger.ConsoleLogger
}

func (c *CommandConfig) GetLogger() *logger.ConsoleLogger {
	if c.Logger == nil {
		c.Logger = logger.NewConsoleLogger()
	}
	return c.Logger
}

func (c *CommandConfig) InitLogger() error {
	c.Logger = logger.NewConsoleLogger()

	if c.Verbose && c.Quiet {
		return fmt.Errorf("cannot use both --verbose and --quiet")
	}

	if c.Verbose {
		c.Logger.SetVerbose(true)
	} else if c.Quiet {
		c.Logger.SetQuiet(true)
	}

	if c.ConfigFile != "" {
		c.Logger.Debug("Using config file: %s", c.ConfigFile)
	}

	return nil
}

func (c *CommandConfig) CreateContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultContextTimeout)
}

func NewCommandConfig() *CommandConfig {
	return &CommandConfig{}
}

var commandConfig *CommandConfig

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

  # Check configuration validity and rclone remotes
  syncerman check

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

func init() {
	commandConfig = NewCommandConfig()
	cobra.OnInitialize(initCommandConfig)

	rootCmd.PersistentFlags().StringVarP(&commandConfig.ConfigFile, "config", "c", "", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&commandConfig.DryRun, "dry-run", "d", false, "Dry run mode (show what would be done)")
	rootCmd.PersistentFlags().BoolVarP(&commandConfig.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&commandConfig.Quiet, "quiet", "q", false, "Quiet mode (suppress output)")
}

func initCommandConfig() {
	if err := commandConfig.InitLogger(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(exitCodeGeneralError)
	}
}

func GetConfig() *CommandConfig {
	return commandConfig
}

func GetLogger() *logger.ConsoleLogger {
	return commandConfig.GetLogger()
}

func GetConfigFile() string {
	return commandConfig.ConfigFile
}

func IsDryRun() bool {
	return commandConfig.DryRun
}

func IsVerbose() bool {
	return commandConfig.Verbose
}

func IsQuiet() bool {
	return commandConfig.Quiet
}

func discoverConfigPath() string {
	if commandConfig.ConfigFile != "" {
		return commandConfig.ConfigFile
	}

	defaultPaths := []string{
		"./syncerman.yaml",
		"./syncerman.yml",
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
