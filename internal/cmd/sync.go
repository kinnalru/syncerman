package cmd

import (
	"context"
	"fmt"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/logger"
	"gitlab.com/kinnalru/syncerman/internal/sync"

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
   -c, --config string   Path to configuration file (default: .syncerman.yml)
   -d, --dry-run        Dry run mode (show what would be done)
   -v, --verbose        Verbose output
   -q, --quiet          Quiet mode (suppress output)`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSync(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	log := GetLogger()

	cfg, err := loadAndValidateConfig()
	if err != nil {
		return err
	}

	opts := sync.SyncOptions{
		DryRun:  IsDryRun(),
		Verbose: IsVerbose(),
		Quiet:   IsQuiet(),
	}

	engine := createEngine(cfg)
	ctx, cancel := GetConfig().CreateContext()
	defer cancel()

	if len(args) == 0 {
		return syncAllTargets(ctx, engine, cfg, opts, log)
	}
	return syncSingleTarget(ctx, log, engine, cfg, args[0], opts)
}

func syncAllTargets(ctx context.Context, engine *sync.Engine, cfg *config.Config, opts sync.SyncOptions, log *logger.ConsoleLogger) error {
	if err := engine.Prepare(ctx, cfg, opts); err != nil && !opts.Quiet {
		log.Error("Failed to prepare directories: %v", err)
	}

	results, err := engine.RunAll(ctx, cfg, opts)
	if err != nil {
		report := sync.NewReport(results)
		if !opts.Quiet {
			log.Error("%s", report.FormatError())
		}
		return wrapError(exitCodeRcloneError, err, "")
	}

	return reportResults(engine, results, opts, log)
}

func reportResults(engine *sync.Engine, results []*sync.SyncResult, opts sync.SyncOptions, log *logger.ConsoleLogger) error {
	report := sync.NewReport(results)
	if opts.Verbose || !opts.Quiet {
		log.Info("%s", report.Format(opts.Verbose))
	}

	if report.ExitCode != 0 {
		return wrapError(report.ExitCode, fmt.Errorf("sync completed with exit code %d", report.ExitCode), "")
	}

	return nil
}

func syncSingleTarget(ctx context.Context, log *logger.ConsoleLogger, engine *sync.Engine, cfg *config.Config, targetArg string, opts sync.SyncOptions) error {
	target, err := findAndValidateTarget(log, engine, cfg, targetArg)
	if err != nil {
		return err
	}

	if err := engine.Prepare(ctx, cfg, opts); err != nil && !opts.Quiet {
		log.Error("Failed to prepare directories: %v", err)
	}

	result, err := engine.Run(ctx, *target, opts)
	if err != nil {
		return wrapError(exitCodeRcloneError, err, "")
	}

	return reportResults(engine, []*sync.SyncResult{result}, opts, log)
}

func findAndValidateTarget(log *logger.ConsoleLogger, engine *sync.Engine, cfg *config.Config, targetArg string) (*sync.SyncTarget, error) {
	provider, path, err := sync.ParseRemote(targetArg)
	if err != nil {
		log.Error("Invalid target format: %v (expected: provider:path)", err)
		return nil, wrapError(exitCodeGeneralError, err, "")
	}

	targets, err := engine.ExpandTargets(cfg)
	if err != nil {
		log.Error("Failed to expand targets: %v", err)
		return nil, wrapError(exitCodeConfigError, err, "")
	}

	for _, target := range targets {
		if target.Provider == provider && target.SourcePath == path {
			return target, nil
		}
	}

	log.Error("Target %s:%s not found in configuration", provider, path)
	return nil, wrapError(exitCodeConfigError, fmt.Errorf("target %s:%s not found", provider, path), "")
}
