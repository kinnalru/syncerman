package sync

import (
	"context"
	"fmt"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"
)

// RunSync executes a single sync operation for the given target.
// It builds the rclone bisync command with appropriate flags and executes it.
func (e *Engine) RunSync(ctx context.Context, target SyncTarget, options SyncOptions) (*SyncResult, error) {
	taskName := fmt.Sprintf("%s:%s → %s", target.Provider, target.SourcePath, StripProviderHash(target.Destination.To))

	e.logger.StageInfo("Stage: Starting sync task")
	e.logger.TargetInfo("Target: %s", taskName)

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

	maxRetries := 0
	if !target.Resync {
		maxRetries = 1
	}

	handler := NewFirstRunHandler(maxRetries, e.logger)
	cmdResult, retryCount, err := handler.Handle(ctx, e.rclone, args)

	result := &SyncResult{
		Target:     target,
		Success:    err == nil && cmdResult.ExitCode == 0,
		Error:      err,
		FirstRun:   retryCount > 0,
		RetryCount: retryCount,
	}

	if err != nil {
		e.logger.Error("Stage: Sync failed for %s", taskName)
		return nil, err
	}

	e.logger.StageInfo("Stage: Sync completed successfully")
	e.logger.TargetInfo("Target: %s", taskName)
	return result, nil
}

// RunAll executes all sync operations defined in the configuration.
// It processes targets sequentially and stops on first error, returning all results so far.
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
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		result, err := e.RunSync(ctx, *target, options)
		if err != nil {
			return results, fmt.Errorf("sync failed for target %d: %w", i+1, err)
		}

		results[i] = result

		if !result.Success {
			return results, fmt.Errorf("sync target %d failed: %v", i+1, result.Error)
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
