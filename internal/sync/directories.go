package sync

import (
	"context"
	"fmt"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"
)

// CreateAllDirectories creates all source and destination directories from configuration.
// It identifies unique source and destination paths and attempts to create them.
// Already-existing directories are handled gracefully (not an error).
func (e *Engine) CreateAllDirectories(ctx context.Context, config *config.Config, options SyncOptions) error {
	targets, err := e.ExpandTargets(config)
	if err != nil {
		return fmt.Errorf("failed to expand targets for directory creation: %w", err)
	}

	sourcePaths := make(map[string]struct{})
	destPaths := make(map[string]struct{})

	for _, target := range targets {
		sourceRemote := FormatRemote(target.Provider, target.SourcePath)
		sourcePaths[sourceRemote] = struct{}{}
		destPaths[target.Destination.To] = struct{}{}
	}

	totalPaths := len(sourcePaths) + len(destPaths)
	if totalPaths == 0 {
		e.logger.Info("No directories to create")
		return nil
	}

	if options.DryRun || e.dryRun {
		e.logger.Info("Ensuring %d directories exist (required by rclone even in dry-run mode)...", totalPaths)
	} else {
		e.logger.Info("Creating %d directories...", totalPaths)
	}

	if err := e.createDirectories(ctx, sourcePaths, "source"); err != nil {
		return err
	}

	if err := e.createDirectories(ctx, destPaths, "destination"); err != nil {
		return err
	}

	e.logger.Info("Successfully created %d source and %d destination directories", len(sourcePaths), len(destPaths))
	return nil
}

// Prepare performs pre-sync operations like directory creation.
// This is part of the SyncEngine interface.
func (e *Engine) Prepare(ctx context.Context, config *config.Config, options SyncOptions) error {
	return e.CreateAllDirectories(ctx, config, options)
}

func (e *Engine) createDirectories(ctx context.Context, paths map[string]struct{}, dirType string) error {
	for path := range paths {
		e.logger.Debug("Creating %s directory: %s", dirType, path)

		if err := rclone.Mkdir(ctx, e.rclone, path); err != nil {
			e.logger.Error("Failed to create %s directory %s: %v", dirType, path, err)
			return fmt.Errorf("failed to create %s directory %s: %w", dirType, path, err)
		}
	}
	return nil
}
