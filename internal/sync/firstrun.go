package sync

import (
	"context"
	"fmt"

	"syncerman/internal/rclone"
)

// FirstRunHandler provides specialized handling for first-run errors.
type FirstRunHandler struct {
	maxRetries int
	logger     Logger
}

// NewFirstRunHandler creates a new handler with given max retries.
func NewFirstRunHandler(maxRetries int, log Logger) *FirstRunHandler {
	if maxRetries < 0 {
		maxRetries = 1
	}

	return &FirstRunHandler{
		maxRetries: maxRetries,
		logger:     log,
	}
}

// Handle attempts to run a sync, detecting and retrying first-run errors.
// It uses IsFirstRunError from rclone package to detect specific error pattern.
// Returns final result after all retries are exhausted.
func (h *FirstRunHandler) Handle(ctx context.Context, exec rclone.Executor, args *rclone.BisyncArgs) (*rclone.Result, int, error) {
	retries := 0

	for {
		cmdResult, err := exec.Run(ctx, args.Build()...)

		if err != nil {
			h.logger.Error("Sync command failed: %v", err)
			return nil, retries, fmt.Errorf("sync command failed: %w", err)
		}

		if cmdResult.ExitCode == 0 {
			h.logger.Info("Sync completed successfully")
			return cmdResult, retries, nil
		}

		if !rclone.IsFirstRunError(cmdResult.Stderr) {
			h.logger.Error("Sync failed with exit code %d: %s", cmdResult.ExitCode, cmdResult.Stderr)
			return cmdResult, retries, fmt.Errorf("sync failed: %s", cmdResult.Stderr)
		}

		retries++

		if retries > h.maxRetries {
			h.logger.Error("First-run error detected but max retries exceeded")
			return cmdResult, retries - 1, fmt.Errorf("first-run error after %d retries: %s", h.maxRetries+1, cmdResult.Stderr)
		}

		h.logger.Warn("First-run error detected, retrying with --resync (attempt %d/%d)", retries, h.maxRetries+1)
		args = args.WithResync()
	}
}

// ShouldRetry determines if a sync operation should be retried based on error.
// Currently only retries first-run errors.
func (h *FirstRunHandler) ShouldRetry(stderr string) bool {
	return rclone.IsFirstRunError(stderr)
}

// ExtractFirstRunError extracts first-run error details from stderr.
// Returns error with first-run details if pattern matches, nil otherwise.
func ExtractFirstRunError(stderr string) error {
	if !rclone.IsFirstRunError(stderr) {
		return nil
	}

	err := rclone.ParseFirstRunError(stderr)
	if err == nil {
		return fmt.Errorf("first-run error detected but could not parse details: %s", stderr)
	}

	return fmt.Errorf("first-run sync error detected - %s. Paths: %v", err.Message, err.Paths)
}

// IsFirstRunSyncError checks if result indicates a first-run error.
// This is a convenience wrapper around rclone.IsFirstRunError.
func IsFirstRunSyncError(result *rclone.Result) bool {
	if result == nil {
		return false
	}

	return rclone.IsFirstRunError(result.Stderr)
}

// DefaultFirstRunHandler creates a handler with default settings.
// Default max retries is 1, meaning one initial attempt + one retry.
func DefaultFirstRunHandler(log Logger) *FirstRunHandler {
	return NewFirstRunHandler(1, log)
}
