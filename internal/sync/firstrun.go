package sync

import (
	"context"
	"fmt"

	"syncerman/internal/rclone"
)

const (
	minMaxRetries = 1
)

// FirstRunHandler provides specialized handling for first-run errors.
type FirstRunHandler struct {
	maxRetries int
	logger     Logger
}

// NewFirstRunHandler creates a new handler with given max retries.
func NewFirstRunHandler(maxRetries int, log Logger) *FirstRunHandler {
	if maxRetries < 0 {
		maxRetries = minMaxRetries
	}

	return &FirstRunHandler{
		maxRetries: maxRetries,
		logger:     log,
	}
}

// Handle attempts to run a sync, detecting and retrying first-run errors.
// It uses IsFirstRunError from rclone package to detect specific error pattern.
// Returns final result after all retries are exhausted.
// The retry count indicates how many times the sync was retried (0 for success on first try).
func (h *FirstRunHandler) Handle(ctx context.Context, exec rclone.Executor, args *rclone.BisyncArgs) (*rclone.Result, int, error) {
	retries := 0

	for {
		cmdResult, err := exec.Run(ctx, args.Build()...)

		// Command execution failed (not sync operation error)
		if err != nil {
			h.logger.Error("Sync command failed: %v", err)
			return nil, retries, fmt.Errorf("sync command failed: %w", err)
		}

		// Sync operation succeeded
		if cmdResult.ExitCode == 0 {
			h.logger.Info("Sync completed successfully")
			return cmdResult, retries, nil
		}

		// Check if this is a recoverable first-run error
		if !rclone.IsFirstRunError(cmdResult.Stderr) {
			h.logger.Error("Sync failed with exit code %d: %s", cmdResult.ExitCode, cmdResult.Stderr)
			return cmdResult, retries, fmt.Errorf("sync failed: %s", cmdResult.Stderr)
		}

		retries++

		if retries > h.maxRetries {
			h.logger.Error("First-run error detected but max retries exceeded")
			return cmdResult, retries, fmt.Errorf("first-run error after %d retries: %s", h.maxRetries+1, cmdResult.Stderr)
		}

		// First-run error: retry with --resync flag to initialize state files
		h.logger.Warn("First-run error detected, retrying with --resync (attempt %d/%d)", retries, h.maxRetries+1)
		args = args.WithResync()
	}
}

// DefaultFirstRunHandler creates a handler with default settings.
// Default max retries is minMaxRetries, meaning one initial attempt + one retry.
func DefaultFirstRunHandler(log Logger) *FirstRunHandler {
	return NewFirstRunHandler(minMaxRetries, log)
}
