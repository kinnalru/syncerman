package cmd

import (
	"fmt"

	"syncerman/internal/rclone"
	"syncerman/internal/sync"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configuration and remotes",
	Long: `Check validates configuration and verifies rclone remotes.

Use 'check config' to validate configuration file.
Use 'check remotes' to verify rclone remote configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		return nil
	},
}

var checkConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Check configuration file validity",
	Long: `Config validates configuration file and checks that all
configured providers and destinations are valid.

This command verifies:
  - Configuration file syntax and structure
  - Provider and path configurations
  - Destination format and required fields
  - Provider existence in rclone configuration (if not --skip-remotes)

Examples:
  syncerman check config
  syncerman check config --config /path/to/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doCheckConfig(cmd)
	},
}

var checkRemotesCmd = &cobra.Command{
	Use:   "remotes",
	Short: "Check rclone remotes configuration",
	Long: `Remotes checks that all providers in configuration
are properly configured in rclone.

This command:
  - Lists all providers from configuration file
  - Verifies each provider exists in rclone
  - Reports OK for each valid provider
  - Reports NOT FOUND for missing providers

Exit codes:
  0 - All providers configured in rclone
  1 - One or more providers not found

Examples:
  syncerman check remotes
  syncerman check remotes --verbose`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doCheckRemotes(cmd)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkConfigCmd)
	checkCmd.AddCommand(checkRemotesCmd)
}

func doCheckConfig(cmd *cobra.Command) error {
	log := GetLogger()

	cfg, err := loadAndValidateConfig()
	if err != nil {
		return err
	}

	executor := rclone.NewExecutor(rclone.NewConfig())
	engine := sync.NewEngine(cfg, executor, log)
	ctx, cancel := GetConfig().CreateContext()
	defer cancel()

	if err := engine.Validate(ctx, cfg); err != nil {
		log.Error("Target validation failed: %v", err)
		return &ExitCodeError{Code: exitCodeValidationError, Err: err}
	}

	log.Info("Configuration is valid")
	providers := cfg.GetProviders()
	log.Info("Found %d provider(s):", len(providers))
	for provider, paths := range providers {
		log.Info("  %s: %d path(s)", provider, len(paths))
	}

	return nil
}

func doCheckRemotes(cmd *cobra.Command) error {
	log := GetLogger()
	ctx, cancel := GetConfig().CreateContext()
	defer cancel()

	cfg, err := loadAndValidateConfig()
	if err != nil {
		return err
	}

	providers := cfg.GetProviders()
	allValid := true

	log.Info("Checking rclone remotes...")
	for provider := range providers {
		executor := rclone.NewExecutor(rclone.NewConfig())
		engine := sync.NewEngine(nil, executor, log)

		exists, err := engine.RemoteProviderExists(ctx, provider)
		if err != nil {
			log.Error("Failed to verify provider %s: %v", provider, err)
			allValid = false
			continue
		}

		if exists {
			log.Info("  %s: OK", provider)
		} else {
			log.Error("  %s: NOT FOUND", provider)
			allValid = false
		}
	}

	if allValid {
		log.Info("All providers are configured in rclone")
	} else {
		return &ExitCodeError{Code: exitCodeValidationError, Err: fmt.Errorf("one or more providers not found")}
	}

	return nil
}

type ExitCodeError struct {
	Code int
	Err  error
}

func (e *ExitCodeError) Error() string {
	return e.Err.Error()
}

func (e *ExitCodeError) ExitCode() int {
	return e.Code
}
