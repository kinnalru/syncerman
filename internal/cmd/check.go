package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/kinnalru/syncerman/internal/sync"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configuration and remotes",
	Long: `Check validates configuration file and verifies rclone remotes.

This command verifies:
  - Configuration file syntax and structure
  - Jobs, tasks, and path configurations
  - Destination format and required fields
  - All providers used in tasks exist in rclone configuration
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
			log.Error("Configuration validation failed:")
			if valErrs, ok := err.(sync.ValidationErrors); ok {
				for _, e := range valErrs {
					log.Error("  - %v", e)
				}
			} else {
				log.Error("  %v", err)
			}
			return wrapError(exitCodeValidationError, fmt.Errorf("configuration validation failed"), "")
		}

		log.Info("Configuration is valid")
		jobs := cfg.GetJobs()
		log.Info("Found %d job(s):", len(jobs))

		providerNames := make(map[string]bool)
		for _, job := range jobs {
			if !job.Enabled {
				continue
			}
			var numTasks int
			for _, task := range job.Tasks {
				if task.Enabled {
					numTasks++
					provider, _, err := sync.ParseRemote(task.From)
					if err == nil && provider != "local" {
						providerNames[provider] = true
					}
				}
			}
			log.Info("  %s (%s): %d active task(s)", job.Name, job.ID, numTasks)
		}

		if len(providerNames) > 0 {
			allValid := true
			log.Info("Checking rclone remotes...")
			for providerName := range providerNames {
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

			if !allValid {
				return wrapError(exitCodeValidationError, fmt.Errorf("one or more remotes checks failed"), "")
			}
		}

		log.Info("All checks passed")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
