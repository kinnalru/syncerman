package sync

import (
	"context"
	"fmt"
	"strings"

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
		if e.logger != nil {
			e.logger.Info("No directories to create")
		}
		return nil
	}

	if options.DryRun || e.dryRun {
		if e.logger != nil {
			e.logger.Info("Ensuring %d directories exist (required by rclone even in dry-run mode)...", totalPaths)
		}
	} else {
		if e.logger != nil {
			e.logger.Info("Creating %d directories...", totalPaths)
		}
	}

	if err := e.createDirectories(ctx, sourcePaths, "source"); err != nil {
		return err
	}

	if err := e.createDirectories(ctx, destPaths, "destination"); err != nil {
		return err
	}

	if e.logger != nil {
		e.logger.Info("Successfully created %d source and %d destination directories", len(sourcePaths), len(destPaths))
	}
	return nil
}

// Prepare performs pre-sync operations like directory creation.
// This is part of the SyncEngine interface.
func (e *Engine) Prepare(ctx context.Context, config *config.Config, options SyncOptions) error {
	return e.CreateAllDirectories(ctx, config, options)
}

// ExtractDestinationPathFromTo extracts the destination path from 'to' field.
// Handles both 'provider:path' and '/path' formats.
// For remote paths, returns the path portion after the colon.
// For local paths, returns the path as-is.
func ExtractDestinationPathFromTo(to string) string {
	if strings.Contains(to, remoteDelimiter) {
		// Split on first colon only to handle paths that might contain colons
		parts := strings.SplitN(to, remoteDelimiter, remoteSplitCount)
		if len(parts) == remoteSplitCount {
			return parts[1]
		}
	}
	return to
}

func (e *Engine) createDirectories(ctx context.Context, paths map[string]struct{}, dirType string) error {
	for path := range paths {
		if e.logger != nil {
			e.logger.Debug("Creating %s directory: %s", dirType, path)
		}

		if err := rclone.Mkdir(ctx, e.rclone, path); err != nil {
			if e.logger != nil {
				e.logger.Error("Failed to create %s directory %s: %v", dirType, path, err)
			}
			return fmt.Errorf("failed to create %s directory %s: %w", dirType, path, err)
		}
	}
	return nil
}

// ValidateDestinationPaths checks if all destination paths in targets are valid.
// Returns error if any destination path is malformed.
func (e *Engine) ValidateDestinationPaths(targets []*SyncTarget) error {
	for i, target := range targets {
		provider, path, err := ParseRemote(target.Destination.To)
		if err != nil {
			return fmt.Errorf("destination %d has invalid remote format '%s': %w", i+1, target.Destination.To, err)
		}

		if provider == "" {
			return fmt.Errorf("destination %d has empty provider: %s", i+1, target.Destination.To)
		}

		if path == "" {
			return fmt.Errorf("destination %d has empty path: %s", i+1, target.Destination.To)
		}
	}

	return nil
}
