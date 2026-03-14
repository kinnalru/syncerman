package sync

import (
	"context"
	"fmt"
	"strings"

	"syncerman/internal/config"
	"syncerman/internal/rclone"
)

// CreateDestinationDirectories creates all destination directories from configuration.
// It identifies unique destination paths and attempts to create them.
// Already-existing directories are handled gracefully (not an error).
func (e *Engine) CreateDestinationDirectories(ctx context.Context, config *config.Config, options SyncOptions) error {
	targets, err := e.ExpandTargets(config)
	if err != nil {
		return fmt.Errorf("failed to expand targets for directory creation: %w", err)
	}

	uniquePaths := make(map[string]struct{})
	for _, target := range targets {
		uniquePaths[target.Destination.To] = struct{}{}
	}

	if len(uniquePaths) == 0 {
		e.logger.Info("No destinations to create")
		return nil
	}

	if options.DryRun || e.dryRun {
		e.logger.Info("Skipping directory creation in dry-run mode")
		e.logger.Debug("Would create %d destination directories: %v", len(uniquePaths), e.mapKeys(uniquePaths))
		return nil
	}

	e.logger.Info("Creating %d destination directories...", len(uniquePaths))
	for path := range uniquePaths {
		if e.logger != nil {
			e.logger.Debug("Creating directory: %s", path)
		}

		err := rclone.Mkdir(ctx, e.rclone, path)
		if err != nil {
			e.logger.Error("Failed to create directory %s: %v", path, err)
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	e.logger.Info("Successfully created %d destination directories", len(uniquePaths))
	return nil
}

// Prepare performs pre-sync operations like directory creation.
// This is part of the SyncEngine interface.
func (e *Engine) Prepare(ctx context.Context, config *config.Config, options SyncOptions) error {
	return e.CreateDestinationDirectories(ctx, config, options)
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
	if strings.Contains(to, ":") {
		// Split on first colon only to handle paths that might contain colons
		parts := strings.SplitN(to, ":", 2)
		if len(parts) == 2 {
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
