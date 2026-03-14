package rclone

import (
	"context"
	"fmt"
	"strings"
)

// Mkdir creates a directory on specified remote:path using rclone mkdir.
// If the directory already exists, the function returns nil (treats it as success).
// If an error occurs, returns a formatted error message with details about the failure.
func Mkdir(ctx context.Context, executor Executor, remotePath string) error {
	if remotePath == "" {
		return fmt.Errorf("remote path cannot be empty")
	}

	result, err := executor.Run(ctx, "mkdir", remotePath)
	if err != nil {
		if result != nil && isDirectoryExistsError(result.Stderr) {
			return nil
		}
		stderr := ""
		if result != nil {
			stderr = result.Stderr
		}
		return fmt.Errorf("failed to create directory %s: %s", remotePath, stderr)
	}

	if result.ExitCode == 0 {
		return nil
	}

	if isDirectoryExistsError(result.Stderr) {
		return nil
	}

	return fmt.Errorf("failed to create directory %s: %s", remotePath, result.Stderr)
}

// isDirectoryExistsError checks if the error output indicates the directory already exists.
// rclone may report an existing directory in various ways depending on the remote type.
func isDirectoryExistsError(stderr string) bool {
	if stderr == "" {
		return false
	}

	lowerStderr := strings.ToLower(stderr)

	alreadyExistsPatterns := []string{
		"already exists",
		"file exists",
		"path already exists",
	}

	for _, pattern := range alreadyExistsPatterns {
		if strings.Contains(lowerStderr, pattern) {
			return true
		}
	}

	return false
}

// CreatePath creates a directory structure on the specified remote path.
// This is a helper that creates the parent directories if they don't exist.
// Note: This requires rclone to support creating parent directories.
func CreatePath(ctx context.Context, executor Executor, remotePath string) error {
	if remotePath == "" {
		return fmt.Errorf("remote path cannot be empty")
	}

	result, err := executor.Run(ctx, "mkdir", remotePath)
	if err != nil {
		if result != nil && isDirectoryExistsError(result.Stderr) {
			return nil
		}
		if result != nil && isParentDirNotFoundError(result.Stderr) {
			return fmt.Errorf("parent directory does not exist for path %s: %s", remotePath, result.Stderr)
		}
		stderr := ""
		if result != nil {
			stderr = result.Stderr
		}
		return fmt.Errorf("failed to create directory %s: %s", remotePath, stderr)
	}

	if result.ExitCode == 0 {
		return nil
	}

	if isDirectoryExistsError(result.Stderr) {
		return nil
	}

	if isParentDirNotFoundError(result.Stderr) {
		return fmt.Errorf("parent directory does not exist for path %s: %s", remotePath, result.Stderr)
	}

	return fmt.Errorf("failed to create directory %s: %s", remotePath, result.Stderr)
}

// isParentDirNotFoundError checks if the error output indicates the parent directory does not exist.
func isParentDirNotFoundError(stderr string) bool {
	if stderr == "" {
		return false
	}

	lowerStderr := strings.ToLower(stderr)

	parentNotFoundPatterns := []string{
		"parent directory",
		"no such file",
		"not found",
		"directory not found",
	}

	for _, pattern := range parentNotFoundPatterns {
		if strings.Contains(lowerStderr, pattern) {
			return true
		}
	}

	return false
}
