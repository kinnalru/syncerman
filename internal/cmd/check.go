package cmd

import (
	"context"
	"fmt"
	"os"

	"syncerman/internal/config"
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
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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
	Run: func(cmd *cobra.Command, args []string) {
		doCheckConfig(cmd)
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
	Run: func(cmd *cobra.Command, args []string) {
		doCheckRemotes(cmd)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkConfigCmd)
	checkCmd.AddCommand(checkRemotesCmd)
}

// doCheckConfig validates the configuration file and its targets.
//
// This function performs comprehensive validation of the configuration file,
// checking both structure and semantic correctness. It validates the YAML
// structure, provider configurations, and ensures targets are properly defined.
//
// Parameters:
//
//	cmd: The cobra command instance
//
// Validation Steps:
//  1. Discover configuration file path (flags or default locations)
//  2. Load and parse the configuration file
//  3. Validate configuration structure (syntax, fields, types)
//  4. Validate targets and remotes through the sync engine
//
// Output:
//   - Success message if configuration is valid
//   - List of all providers with their path counts
//
// Error Handling:
//   - Exits with code 1 if no configuration file found
//   - Exits with code 1 if configuration fails to load
//   - Exits with code 1 if configuration structure validation fails
//   - Exits with code 1 if target/remote validation fails
//
// Usage:
//
//	syncerman check config
//	syncerman check config --config /path/to/config.yaml
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

// doCheckRemotes verifies that all providers in the configuration are properly
// configured in rclone.
//
// This function checks each provider from the configuration file against the
// rclone configuration to ensure they exist and are accessible. It reports
// the status of each provider and exits with an error if any are missing.
//
// Parameters:
//
//	cmd: The cobra command instance
//
// Verification Steps:
//  1. Discover and load configuration file
//  2. Extract all unique providers from configuration
//  3. Iterate through each provider and check if it exists in rclone
//  4. Report status for each provider (OK or NOT FOUND)
//
// Output:
//   - "Checking rclone remotes..." message
//   - For each provider: "  provider-name: OK" or "  provider-name: NOT FOUND"
//   - Final message: "All providers are configured in rclone" (if all valid)
//
// Error Handling:
//   - Continues checking other providers if one fails verification
//   - Logs error for each failed provider check
//   - Exits with code 1 at the end if any provider was not found or check failed
//
// Exit Codes:
//
//	0 - All providers are configured in rclone
//	1 - One or more providers not found or verification errors
//
// Usage:
//
//	syncerman check remotes
//	syncerman check remotes --config /path/to/config.yaml
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
