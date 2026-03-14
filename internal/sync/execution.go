package sync

import (
	"context"
	"fmt"
	"syncerman/internal/config"
	"syncerman/internal/rclone"
)

// RunSync executes a single sync operation for the given target.
// It builds the rclone bisync command with appropriate flags and executes it.
func (e *Engine) RunSync(ctx context.Context, target SyncTarget, options SyncOptions) (*SyncResult, error) {
	if options.Verbose {
		e.logger.Info("Starting sync for %s:%s to %s", target.Provider, target.SourcePath, target.Destination.To)
	} else {
		e.logger.Debug("Starting sync for %s:%s to %s", target.Provider, target.SourcePath, target.Destination.To)
	}

	sourceRemote := FormatRemote(target.Provider, target.SourcePath)
	destRemote := target.Destination.To

	bisyncOptions := &rclone.BisyncOptions{
		Resync: target.Resync || options.DryRun || e.dryRun,
	}

	bisyncOptions.Args = append(bisyncOptions.Args, target.Destination.Args...)

	if options.DryRun || e.dryRun {
		bisyncOptions.DryRun = true
	}

	args := rclone.NewBisyncArgs(sourceRemote, destRemote, bisyncOptions)

	result := &SyncResult{
		Target:     target,
		Success:    false,
		Error:      nil,
		FirstRun:   false,
		RetryCount: 0,
	}

	maxRetries := 1
	if target.Resync {
		maxRetries = 1
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			result.RetryCount++

			e.logger.Info("Retrying sync for %s:%s (attempt %d/%d)", target.Provider, target.SourcePath, attempt+1, maxRetries+1)

			args = args.WithResync()
		}

		cmdResult, err := e.rclone.Run(ctx, args.Build()...)

		if err != nil {
			result.Error = fmt.Errorf("sync command failed: %w", err)

			if options.Verbose {
				e.logger.Error("Sync command error: %v", err)
			}

			return result, nil
		}

		if options.Verbose {
			e.logger.Debug("Sync output: %s", cmdResult.Combined)
		}

		if rclone.IsFirstRunError(cmdResult.Stderr) {
			if attempt == maxRetries {
				e.logger.Warn("First-run error detected but max retries exceeded for %s:%s", target.Provider, target.SourcePath)
				result.Error = fmt.Errorf("first-run error after %d retries: %s", maxRetries+1, cmdResult.Stderr)
				return result, nil
			}

			e.logger.Warn("First-run error detected for %s:%s, retrying with --resync", target.Provider, target.SourcePath)
			result.FirstRun = true
			continue
		}

		if cmdResult.ExitCode == 0 {
			result.Success = true
			e.logger.Info("Sync completed successfully for %s:%s to %s", target.Provider, target.SourcePath, destRemote)
			return result, nil
		}

		result.Error = fmt.Errorf("sync failed with exit code %d: %s", cmdResult.ExitCode, cmdResult.Stderr)
		e.logger.Error("Sync failed for %s:%s: %s", target.Provider, target.SourcePath, cmdResult.Stderr)
		return result, nil
	}

	result.Error = fmt.Errorf("sync aborted after %d attempts for %s:%s", maxRetries+1, target.Provider, target.SourcePath)
	return result, nil
}

// RunAll executes all sync operations defined in the configuration.
// It processes targets sequentially and returns aggregated results.
// Stops on first error unless ContinueOnError is set in options.
func (e *Engine) RunAll(ctx context.Context, config *config.Config, options SyncOptions) ([]*SyncResult, error) {
	targets, err := e.ExpandTargets(config)
	if err != nil {
		return nil, fmt.Errorf("failed to expand targets: %w", err)
	}

	if len(targets) == 0 {
		return []*SyncResult{}, nil
	}

	e.logger.Info("Starting sync for %d target(s)", len(targets))

	results := make([]*SyncResult, len(targets))

	for i, target := range targets {
		result, err := e.RunSync(ctx, *target, options)
		if err != nil {
			results[i] = result
			return results, fmt.Errorf("sync failed for target %d: %w", i+1, err)
		}

		results[i] = result

		if !result.Success {
			if options.Quiet {
				return results, fmt.Errorf("sync target %d failed: %w", i+1, result.Error)
			}
			return results, fmt.Errorf("sync target %d failed: %w", i+1, result.Error)
		}
	}

	successCount := 0
	firstRunCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		}
		if result.FirstRun {
			firstRunCount++
		}
	}

	e.logger.Info("Sync completed: %d/%d successful, %d first-runs", successCount, len(results), firstRunCount)

	return results, nil
}

// Run executes sync operations using the sync engine.
// This is the primary interface method implementation.
func (e *Engine) Run(ctx context.Context, target SyncTarget, options SyncOptions) (*SyncResult, error) {
	return e.RunSync(ctx, target, options)
}
