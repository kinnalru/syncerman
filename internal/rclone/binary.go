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

// FindRcloneBinary locates the rclone binary by checking:
// 1. SYNCERMAN_RCLONE_PATH environment variable
// 2. PATH for an executable named "rclone"
//
// Returns the path to the rclone binary, or an error if not found.
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

// ConfigFromEnv loads rclone configuration from environment.
// Returns a Config with the binary path based on environment variables.
func ConfigFromEnv() (*Config, error) {
	binaryPath, err := FindRcloneBinary()
	if err != nil {
		return nil, err
	}
	return &Config{
		BinaryPath: binaryPath,
	}, nil
}
