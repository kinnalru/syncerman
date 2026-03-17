package cmd

import (
	"fmt"

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
  syncerman check --config /path/to/.syncerman.yml
  syncerman check --verbose`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := GetLogger()

		cfg, err := loadAndValidateConfig()
		if err != nil {
			return err
		}

		engine := createEngine(cfg)
		ctx, cancel := GetConfig().CreateContext()
		defer cancel()

		if err := engine.Validate(ctx, cfg); err != nil {
			log.Error("Target validation failed: %v", err)
			return wrapError(exitCodeValidationError, err, "")
		}

		log.Info("Configuration is valid")
		providers := cfg.GetProviders()
		log.Info("Found %d provider(s):", len(providers))
		for _, provider := range providers {
			log.Info("  %s: %d path(s)", provider.Name, len(provider.Data))
		}

		allValid := true
		log.Info("Checking rclone remotes...")
		for _, provider := range providers {
			providerName := provider.Name

			exists, err := engine.RemoteProviderExists(ctx, providerName)
			if err != nil {
				log.Error("Failed to verify provider %s: %v", providerName, err)
				allValid = false
				continue
			}

			if exists {
				log.Info("  %s: OK", providerName)
			} else {
				log.Error("  %s: NOT FOUND", providerName)
				allValid = false
			}
		}

		if allValid {
			log.Info("All checks passed")
		} else {
			return wrapError(exitCodeValidationError, fmt.Errorf("one or more checks failed"), "")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
