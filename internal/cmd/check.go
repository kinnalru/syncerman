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
	Long: `Check validates configuration file and verifies rclone remotes.

This command verifies:
  - Configuration file syntax and structure
  - Provider and path configurations
  - Destination format and required fields
  - All provider names exist in rclone configuration
  - Rclone binary is accessible
  - Connection to each remote is possible

Examples:
  syncerman check
  syncerman check --config /path/to/config.yml
  syncerman check --verbose`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
			log.Info("All checks passed")
		} else {
			return &ExitCodeError{Code: exitCodeValidationError, Err: fmt.Errorf("one or more checks failed")}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
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
