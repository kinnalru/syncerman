package rclone

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// RcloneEnvVar is the environment variable name for custom rclone binary path
	RcloneEnvVar = "SYNCERMAN_RCLONE_PATH"
	// RcloneBinaryName is the default binary name
	RcloneBinaryName = "rclone"
)

// FindRcloneBinary locates the rclone binary on the system.
// It attempts to find the binary using the following discovery logic:
//
//  1. SYNCERMAN_RCLONE_PATH environment variable (custom binary path)
//  2. PATH environment variable search for "rclone" executable
//
// The SYNCERMAN_RCLONE_PATH environment variable can be used to specify
// a custom path to the rclone binary, useful for development or testing
// with different rclone versions.
//
// Returns:
//   - string: Absolute path to the rclone binary
//   - error:  Error if binary is not found or custom path does not exist
//
// Error cases:
//   - Custom path specified in SYNCERMAN_RCLONE_PATH does not exist
//   - rclone binary not found in PATH
func FindRcloneBinary() (string, error) {
	if customPath := os.Getenv(RcloneEnvVar); customPath != "" {
		if _, err := os.Stat(customPath); err != nil {
			return "", fmt.Errorf("custom rclone binary not found: %w", err)
		}

		absPath, err := resolvePath(customPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve custom rclone path: %w", err)
		}
		return absPath, nil
	}

	path, err := exec.LookPath(RcloneBinaryName)
	if err != nil {
		return "", fmt.Errorf("rclone binary not found in PATH: %w", err)
	}

	return path, nil
}

// resolvePath resolves a path to its absolute form.
// For simple paths without separators, it uses exec.LookPath to find the binary.
// For paths with separators, it uses filepath.Abs to resolve the absolute path.
//
// Parameters:
//   - path: the path to resolve
//
// Returns:
//   - string: the resolved absolute path
//   - error: error if resolution fails
func resolvePath(path string) (string, error) {
	if !strings.Contains(path, string(filepath.Separator)) {
		absPath, _ := exec.LookPath(path)
		if absPath != "" {
			return absPath, nil
		}
		return path, nil
	}
	return filepath.Abs(path)
}

// ConfigFromEnv loads rclone configuration from environment variables.
// It discovers the rclone binary path using the following order:
//
//  1. SYNCERMAN_RCLONE_PATH environment variable for a custom binary path
//  2. PATH environment variable search for the "rclone" executable
//
// This function provides automatic binary discovery suitable for production
// environments where the rclone binary location may vary.
//
// Returns:
//   - *Config: A pointer to a new Config instance with the discovered binary path
//   - error:  Error if rclone binary cannot be found
//
// Error cases:
//   - rclone binary not found in any of the discovery locations
func ConfigFromEnv() (*Config, error) {
	binaryPath, err := FindRcloneBinary()
	if err != nil {
		return nil, err
	}
	return &Config{
		BinaryPath: binaryPath,
	}, nil
}
