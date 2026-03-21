package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/errors"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [config_path]",
	Short: "Migrate configuration file to the new job-centric format",
	Long: `Migrates an existing configuration file from the legacy provider-to-path format 
to the current job-centric format.

It creates a backup of the original file as <config_path>.bak.
If no config_path is provided, it tries to use the default config file path.`,
	Example: `  # Migrate default configuration file (.syncerman.yml)
  syncerman migrate

  # Migrate specific configuration file
  syncerman migrate /path/to/old/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	log := GetLogger()

	var configPath string
	var err error

	if len(args) > 0 {
		configPath = args[0]
	} else {
		// Try to find the default config file or use the one specified by --config flag
		configPath, err = config.DiscoverConfigPath(commandConfig.ConfigFile)
		if err != nil {
			return errors.NewConfigError("failed to discover config path", err)
		}
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return errors.NewConfigError("failed to resolve absolute path", fmt.Errorf("failed to get absolute path for %s: %w", configPath, err))
	}

	log.Info("Updating configuration file: %s", absPath)

	err = config.MigrateOldConfig(absPath)
	if err != nil {
		log.Error("Failed to update configuration: %v", err)
		return errors.NewConfigError("failed to update configuration file", err)
	}

	log.Info("Configuration successfully updated.")
	log.Info("Original configuration backup saved as: %s.bak", absPath)

	return nil
}
