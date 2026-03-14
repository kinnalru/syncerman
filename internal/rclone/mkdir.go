package rclone

import (
	"context"
	"fmt"
	"strings"
)

// Mkdir creates a directory on specified remote:path using rclone mkdir.
//
// This function executes the "rclone mkdir" command to create a single directory
// at the specified remote path. It handles various error conditions gracefully,
// including treating an existing directory as a successful operation rather than
// an error. This behavior aligns with idempotent operations where creating an
// already-existent resource is not an error.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - executor: Executor interface to run rclone commands
//   - remotePath: The remote path where the directory should be created (e.g., "remote:path/to/dir")
//
// Returns:
//   - error: nil if directory was created or already exists, formatted error message on failure
//
// Behavior:
//   - Treats existing directory as success (returns nil) regardless of rclone's error message
//   - Returns formatted error on failure with path and error details
//
// Implementation details:
//   - Uses "rclone mkdir" command which creates a single directory
//   - Does NOT create parent directories automatically - parent must exist
//   - Error output is parsed to detect existing directories using isDirectoryExistsError()
//
// Error cases:
//   - Empty path: returns "remote path cannot be empty"
//   - Permission denied: returns formatted error from rclone
//   - Remote error: returns formatted error with stderr output
//   - Parent directory doesn't exist: returns error from rclone (not treated as success)
func Mkdir(ctx context.Context, executor Executor, remotePath string) error {
	if remotePath == "" {
		return fmt.Errorf("remote path cannot be empty")
	}

	result, err := executor.Run(ctx, "mkdir", remotePath)
	if err != nil {
		// If rclone returned an error but the stderr indicates directory already exists,
		// treat it as success (idempotent behavior)
		if result != nil && isDirectoryExistsError(result.Stderr) {
			return nil
		}
		stderr := ""
		if result != nil {
			stderr = result.Stderr
		}
		return fmt.Errorf("failed to create directory %s: %s", remotePath, stderr)
	}

	// Check exit code for cases where rclone returns non-zero exit code
	// but the executor didn't return an error
	if result.ExitCode == 0 {
		return nil
	}

	// Even with non-zero exit code, check if directory exists to treat as success
	if isDirectoryExistsError(result.Stderr) {
		return nil
	}

	return fmt.Errorf("failed to create directory %s: %s", remotePath, result.Stderr)
}

// isDirectoryExistsError checks if the error output indicates the directory already exists.
//
// This function parses rclone's stderr output to determine if a directory creation
// failure was due to the directory already existing. rclone may report an existing
// directory in various ways depending on the remote type (S3, Dropbox, Google Drive, etc.)
// and backend implementation. This function accommodates multiple possible error messages.
//
// Parameters:
//   - stderr: The stderr output from rclone command execution
//
// Returns:
//   - bool: true if stderr indicates directory already exists, false otherwise
//
// Implementation details:
//   - rclone reports existing directory in various ways across different remotes
//   - Performs case-insensitive matching to handle different error message formats
//   - Checks for multiple patterns: "already exists", "file exists", "path already exists"
//   - Returns false for empty stderr strings (no error to parse)
//
// Case-insensitive matching:
//   - Converts stderr to lowercase for matching
//   - Ensures pattern matching works regardless of rclone's output casing
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

// CreatePath creates a directory structure on the specified remote path with parent directory support.
//
// This function attempts to create a directory at the specified remote path. It is designed
// to handle cases where parent directories may or may not exist, and provides more
// informative error messages about parent directory issues. Note that not all rclone
// remotes support parent directory creation (mkdir -p behavior), so this function
// provides better error messaging for such cases.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - executor: Executor interface to run rclone commands
//   - remotePath: The remote path where the directory should be created (e.g., "remote:path/to/dir")
//
// Returns:
//   - error: nil if directory was created or already exists, formatted error message on failure
//
// Behavior:
//   - Creates directory at the specified path
//   - Attempts to handle parent directory issues with specific error messages
//   - Treats existing directory as success (returns nil)
//   - Provides enhanced error messages when parent directory doesn't exist
//
// Implementation details:
//   - Uses "rclone mkdir" command (same as Mkdir function)
//   - Requires rclone to support creating parent directories for full functionality
//   - Uses similar error detection logic to Mkdir
//   - Additional check for parent directory not found errors using isParentDirNotFoundError()
//
// Note: Some rclone remotes don't support parent directory creation
//   - S3: Supports parent directory creation
//   - Local filesystem: Supports parent directory creation
//   - Some cloud providers: May not support nested directory creation
//   - Check specific rclone backend documentation for details
func CreatePath(ctx context.Context, executor Executor, remotePath string) error {
	if remotePath == "" {
		return fmt.Errorf("remote path cannot be empty")
	}

	result, err := executor.Run(ctx, "mkdir", remotePath)
	if err != nil {
		// If directory already exists, treat as success
		if result != nil && isDirectoryExistsError(result.Stderr) {
			return nil
		}
		// Provide specific error message if parent directory doesn't exist
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
//
// This function parses rclone's stderr output to determine if a directory creation
// failure was due to the parent directory not existing. This is useful for providing
// more informative error messages to users, helping them understand that they need
// to create parent directories first or use a function that supports parent directory creation.
//
// Parameters:
//   - stderr: The stderr output from rclone command execution
//
// Returns:
//   - bool: true if stderr indicates parent directory doesn't exist, false otherwise
//
// Implementation details:
//   - Checks for multiple error patterns indicating parent directory issues
//   - Pattern list includes: "parent directory", "no such file", "not found", "directory not found"
//   - Performs case-insensitive matching to handle different error message formats
//   - Returns false for empty stderr strings (no error to parse)
//
// Usage:
//   - Used by CreatePath function to provide better error messages
//   - Helps distinguish between general failures and specific parent directory issues
//   - Enables users to take corrective action (create parent directories)
//
// Case-insensitive matching:
//   - Converts stderr to lowercase for pattern matching
//   - Ensures pattern matching works regardless of rclone's output casing
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
