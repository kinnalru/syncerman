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

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize targets from configuration or single target",
	Long: `Sync executes bidirectional synchronization using rclone bisync.

When called without arguments, syncs all targets from configuration file.
When called with a target argument (provider:path), syncs only that specific target.

Examples:
  syncerman sync                  # Sync all targets
  syncerman sync gdrive:docs       # Sync gdrive:docs only
  syncerman sync --dry-run        # Show what would be synced
  syncerman sync --verbose         # Show detailed output

Global Flags:
  -c, --config string   Path to configuration file (default: ./syncerman.yaml)
  -d, --dry-run        Dry run mode (show what would be done)
  -v, --verbose         Verbose output
  -q, --quiet          Quiet mode (suppress output)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runSync(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

// runSync is the handler for the sync command.
// It loads configuration and executes synchronization operations for either
// all targets or a single target based on command arguments.
//
// Parameters:
//
//	cmd: The cobra command instance
//	args: Command arguments - optionally contains a target in provider:path format
//
// Workflow:
//  1. Get configuration file path (from flags or default discovery)
//  2. Load and parse the configuration file
//  3. Create sync options from global flags (dry-run, verbose, quiet)
//  4. Initialize rclone executor and sync engine
//  5. Branch based on arguments:
//     - No arguments: sync all targets from configuration
//     - One argument: sync the specified target
//  6. Use report exit code for final exit status
//
// Error Handling:
//   - Exits with code 1 if no configuration file found
//   - Exits with code 1 if configuration fails to load
//   - Delegates to syncAllTargets/syncSingleTarget for sync errors
//
// Usage:
//
//	syncerman sync              # Sync all targets
//	syncerman sync gdrive:docs  # Sync specific target
func runSync(cmd *cobra.Command, args []string) {
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

	opts := sync.SyncOptions{
		DryRun:  IsDryRun(),
		Verbose: IsVerbose(),
		Quiet:   IsQuiet(),
	}

	executor := rclone.NewExecutor(rclone.NewConfig())
	engine := sync.NewEngine(cfg, executor, log)
	ctx := context.Background()

	if len(args) == 0 {
		syncAllTargets(ctx, engine, cfg, opts)
	} else {
		syncSingleTarget(ctx, engine, args[0], opts)
	}
}

// syncAllTargets synchronizes all targets defined in the configuration file.
//
// This function prepares directories, runs all sync operations, and reports results.
// It continues processing all targets even if individual syncs fail, unless quiet mode
// is enabled which may change error handling behavior.
//
// Parameters:
//
//	ctx: Context for cancellation and timeout handling
//	engine: The sync engine that executes sync operations
//	cfg: The configuration containing all targets to sync
//	opts: Sync options including dry-run, verbose, and quiet flags
//
// Flow:
//  1. Prepare directories for all targets (logs errors unless quiet)
//  2. Run sync operations for all targets
//  3. On error, collect and format error report (unless quiet)
//  4. On success, collect and format results (unless quiet)
//  5. Exit with code from report if non-zero
//
// Error Handling:
//   - Directory preparation errors are logged but don't stop execution (unless quiet)
//   - Sync errors cause exit with code 1 after collecting results
//   - Exit code from report is used for final status
//
// Report Formatting:
//   - Detailed output if --verbose flag is set
//   - Summary output in normal mode (unless --quiet)
//   - No output in quiet mode
func syncAllTargets(ctx context.Context, engine *sync.Engine, cfg *config.Config, opts sync.SyncOptions) {
	log := GetLogger()
	if err := engine.Prepare(ctx, cfg, opts); err != nil && !opts.Quiet {
		log.Error("Failed to prepare directories: %v", err)
	}

	results, err := engine.RunAll(ctx, cfg, opts)
	if err != nil {
		report := engine.CollectResults(results)
		if !opts.Quiet {
			fmt.Fprintln(os.Stderr, report.FormatError())
		}
		os.Exit(1)
	}

	report := engine.CollectResults(results)
	if opts.Verbose || !opts.Quiet {
		fmt.Println(report.Format(opts.Verbose))
	}

	if report.ExitCode != 0 {
		os.Exit(report.ExitCode)
	}
}

// syncSingleTarget synchronizes a specific target specified by provider:path format.
//
// This function validates the target format, searches for matching targets in the
// configuration, and executes a single sync operation for the found target.
//
// Parameters:
//
//	ctx: Context for cancellation and timeout handling
//	engine: The sync engine that executes sync operations
//	targetArg: Target specification in provider:path format (e.g., "gdrive:docs")
//	opts: Sync options including dry-run, verbose, and quiet flags
//
// Workflow:
//  1. Parse target argument to extract provider and path components
//  2. Load configuration file
//  3. Expand all targets from configuration
//  4. Search for matching target by provider and path
//  5. Prepare directories (logs errors unless quiet)
//  6. Execute sync operation for the found target
//  7. Collect and format results (unless quiet)
//  8. Exit with code from report if non-zero
//
// Target Validation:
//   - Exits with code 1 if target format is invalid (not provider:path)
//   - Exits with code 1 if target not found in configuration
//
// Error Handling:
//   - Configuration load errors: exit with code 1
//   - Target expansion errors: exit with code 1
//   - Target not found: exit with code 1
//   - Directory preparation errors: logged (unless quiet)
//   - Sync errors: exit with code 1
//   - Exit code from report is used for final status
//
// Usage:
//
//	syncerman sync gdrive:docs
//	syncerman sync dropbox:shared --dry-run
func syncSingleTarget(ctx context.Context, engine *sync.Engine, targetArg string, opts sync.SyncOptions) {
	log := GetLogger()
	provider, path, err := sync.ParseRemote(targetArg)
	if err != nil {
		log.Error("Invalid target format: %v (expected: provider:path)", err)
		os.Exit(1)
	}

	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	targets, err := engine.ExpandTargets(cfg)
	if err != nil {
		log.Error("Failed to expand targets: %v", err)
		os.Exit(1)
	}

	var found *sync.SyncTarget
	for _, target := range targets {
		if target.Provider == provider && target.SourcePath == path {
			found = target
			break
		}
	}

	if found == nil {
		log.Error("Target %s:%s not found in configuration", provider, path)
		os.Exit(1)
	}

	if err := engine.Prepare(ctx, cfg, opts); err != nil && !opts.Quiet {
		log.Error("Failed to prepare directories: %v", err)
	}

	result, err := engine.Run(ctx, *found, opts)
	if err != nil {
		log.Error("Sync failed: %v", err)
		os.Exit(1)
	}

	report := engine.CollectResults([]*sync.SyncResult{result})
	if opts.Verbose || !opts.Quiet {
		fmt.Println(report.Format(opts.Verbose))
	}

	if report.ExitCode != 0 {
		os.Exit(report.ExitCode)
	}
}

// getConfigPath discovers the configuration file path using a priority order.
//
// This function searches for the configuration file in the following order:
//
//  1. From the global --config/-c flag if specified
//  2. Current directory: ./syncerman.yaml
//  3. Current directory: ./syncerman.yml
//
// Returns:
//
//	string: The first valid configuration file path found, or empty string if none exist
//
// Error Cases:
//   - Returns empty string if no configuration file is found in any location
//
// Usage:
//
//	This function is called by sync and check commands to locate the configuration
//	file before loading. Callers should handle the empty string case by prompting
//	the user to specify the path via --config flag.
func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
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
