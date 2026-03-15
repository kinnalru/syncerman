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
// For local provider, directory creation is skipped as rclone cannot create local directories.
func (e *Engine) CreateAllDirectories(ctx context.Context, config *config.Config, options SyncOptions) error {
	targets, err := e.ExpandTargets(config)
	if err != nil {
		return fmt.Errorf("failed to expand targets for directory creation: %w", err)
	}

	sourcePaths := make(map[string]struct{})
	destPaths := make(map[string]struct{})

	for _, target := range targets {
		sourceRemote := FormatRemote(target.Provider, target.SourcePath)
		if target.Provider != localProvider {
			sourcePaths[sourceRemote] = struct{}{}
		}
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

	for path := range sourcePaths {
		if e.logger != nil {
			e.logger.Debug("Creating source directory: %s", path)
		}

		err := rclone.Mkdir(ctx, e.rclone, path)
		if err != nil {
			e.logger.Error("Failed to create source directory %s: %v", path, err)
			return fmt.Errorf("failed to create source directory %s: %w", path, err)
		}
	}

	for path := range destPaths {
		if e.logger != nil {
			e.logger.Debug("Creating destination directory: %s", path)
		}

		err := rclone.Mkdir(ctx, e.rclone, path)
		if err != nil {
			e.logger.Error("Failed to create destination directory %s: %v", path, err)
			return fmt.Errorf("failed to create destination directory %s: %w", path, err)
		}
	}

	e.logger.Info("Successfully created %d source and %d destination directories", len(sourcePaths), len(destPaths))
	return nil
}

// Prepare performs pre-sync operations like directory creation.
// This is part of the SyncEngine interface.
func (e *Engine) Prepare(ctx context.Context, config *config.Config, options SyncOptions) error {
	return e.CreateAllDirectories(ctx, config, options)
}

// mapKeys returns all keys from a map as a slice.
// Used internally for logging unique destination paths.
func (e *Engine) mapKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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
