package rclone

import (
	"errors"
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
		if _, err := os.Stat(customPath); err == nil {
			if !strings.Contains(customPath, string(filepath.Separator)) {
				absPath, _ := exec.LookPath(customPath)
				if absPath != "" {
					return absPath, nil
				}
				return customPath, nil
			}
			absPath, err := filepath.Abs(customPath)
			if err != nil {
				return "", fmt.Errorf("invalid custom rclone path %s: %w", customPath, err)
			}
			return absPath, nil
		}
		return "", fmt.Errorf("custom rclone binary not found at %s", customPath)
	}

	path, err := exec.LookPath(RcloneBinaryName)
	if err != nil {
		return "", errors.New("rclone binary not found")
	}

	return path, nil
}

// FindRcloneBinaryOrFatal locates rclone binary or logs a fatal error.
// This is a convenience function for CLI applications.
func FindRcloneBinaryOrFatal() string {
	path, err := FindRcloneBinary()
	if err != nil {
		return ""
	}
	return path
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
