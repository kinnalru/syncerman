package sync

import (
	"context"
	"fmt"

	"gitlab.com/kinnalru/syncerman/internal/rclone"
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

		if cmdResult == nil {
			h.logger.StageInfo("Stage: Command execution failed")
			return nil, retries, fmt.Errorf("sync command failed: %w", err)
		}

		if cmdResult.Combined != "" {
			h.logger.CombinedOutput(cmdResult.Combined)
		}

		if cmdResult.ExitCode == 0 {
			if retries == 0 {
				h.logger.Info("Result: Command completed successfully")
			} else {
				h.logger.Info("Result: Command completed successfully after %d retry(ies)", retries)
			}
			return cmdResult, retries, nil
		}

		isFirstRun := rclone.IsFirstRunError(cmdResult.Combined)
		if !isFirstRun {
			h.logger.Error("Result: Command failed with exit code %d", cmdResult.ExitCode)
			return cmdResult, retries, fmt.Errorf("sync failed: %s", cmdResult.Combined)
		}

		retries++

		if retries > h.maxRetries {
			h.logger.Error("Result: First-run error but max retries exceeded")
			return cmdResult, retries, fmt.Errorf("first-run error after %d retries: %s", h.maxRetries+1, cmdResult.Stderr)
		}

		h.logger.StageInfo("Stage: First-run detected, retrying with --resync (attempt %d/%d)", retries, h.maxRetries+1)
		args = args.WithResync()
	}
}

// DefaultFirstRunHandler creates a handler with default settings.
// Default max retries is minMaxRetries, meaning one initial attempt + one retry.
func DefaultFirstRunHandler(log Logger) *FirstRunHandler {
	return NewFirstRunHandler(minMaxRetries, log)
}
