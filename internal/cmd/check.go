package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"syncerman/internal/config"
	"syncerman/internal/rclone"
	"syncerman/internal/sync"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configuration and remotes",
	Long: `Check validates configuration and verifies rclone remotes.
	
Use 'check config' to validate configuration file.
Use 'check remotes' to verify rclone remote configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var checkConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Check configuration file validity",
	Long: `Config validates configuration file and checks that all
configured providers and destinations are valid.`,
	Run: func(cmd *cobra.Command, args []string) {
		doCheckConfig(cmd)
	},
}

var checkRemotesCmd = &cobra.Command{
	Use:   "remotes",
	Short: "Check rclone remotes configuration",
	Long: `Remotes checks that all providers in configuration
are properly configured in rclone.`,
	Run: func(cmd *cobra.Command, args []string) {
		doCheckRemotes(cmd)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkConfigCmd)
	checkCmd.AddCommand(checkRemotesCmd)
}

func doCheckConfig(cmd *cobra.Command) {
	log := GetLogger()

	configPath := getConfigPath()
	if configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: No configuration file found (use --config to specify)")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		log.Error("Configuration validation failed: %v", err)
		os.Exit(1)
	}

	executor := rclone.NewExecutor(rclone.NewConfig())
	engine := sync.NewEngine(cfg, executor, log)
	ctx := context.Background()

	if err := engine.Validate(ctx, cfg); err != nil {
		log.Error("Target validation failed: %v", err)
		os.Exit(1)
	}

	log.Info("Configuration is valid")
	providers := cfg.GetProviders()
	log.Info("Found %d provider(s):", len(providers))
	for provider, paths := range providers {
		log.Info("  %s: %d path(s)", provider, len(paths))
	}
}

func doCheckRemotes(cmd *cobra.Command) {
	log := GetLogger()
	ctx := context.Background()

	executor := rclone.NewExecutor(rclone.NewConfig())
	engine := sync.NewEngine(nil, executor, log)

	configPath := getConfigPath()
	if configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: No configuration file found (use --config to specify)")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	providers := cfg.GetProviders()
	allValid := true

	log.Info("Checking rclone remotes...")
	for provider := range providers {
		exists, err := engine.ProviderExists(ctx, provider)
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
		os.Exit(1)
	}
}
