package cmd

import (
	"context"
	"fmt"
	"strings"

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
When called with a job ID argument, syncs only that specific job.

Examples:
  syncerman sync                  # Sync all targets
  syncerman sync backup-docs      # Sync only job with ID 'backup-docs'
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
	return syncSingleJob(ctx, log, engine, cfg, args[0], opts)
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

func syncSingleJob(ctx context.Context, log *logger.ConsoleLogger, engine *sync.Engine, cfg *config.Config, jobID string, opts sync.SyncOptions) error {
	targets, err := findAndValidateJobTargets(log, engine, cfg, jobID)
	if err != nil {
		return err
	}

	if err := engine.Prepare(ctx, cfg, opts); err != nil && !opts.Quiet {
		log.Error("Failed to prepare directories: %v", err)
	}

	results := make([]*sync.SyncResult, 0, len(targets))
	for i, target := range targets {
		result, err := engine.Run(ctx, *target, opts)
		if err != nil {
			return wrapError(exitCodeRcloneError, fmt.Errorf("sync failed for target %d: %w", i+1, err), "")
		}
		results = append(results, result)

		if !result.Success {
			return wrapError(exitCodeRcloneError, fmt.Errorf("sync target %d failed: %v", i+1, result.Error), "")
		}
	}

	return reportResults(engine, results, opts, log)
}

func findAndValidateJobTargets(log *logger.ConsoleLogger, engine *sync.Engine, cfg *config.Config, jobID string) ([]*sync.SyncTarget, error) {
	targets, err := engine.ExpandTargets(cfg, jobID)
	if err != nil {
		log.Error("Failed to expand targets for job %s: %v", jobID, err)
		return nil, wrapError(exitCodeConfigError, err, "")
	}

	if len(targets) > 0 {
		return targets, nil
	}

	// No targets found for jobID
	if strings.Contains(jobID, ":") {
		log.Error("Invalid argument: use job ID instead of provider:path format (%s)", jobID)
		return nil, wrapError(exitCodeGeneralError, fmt.Errorf("invalid job ID format: %s", jobID), "")
	}

	log.Error("Job %s not found in configuration or has no active targets", jobID)
	return nil, wrapError(exitCodeConfigError, fmt.Errorf("job %s not found", jobID), "")
}
