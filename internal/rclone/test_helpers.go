package rclone

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// CreateTestBinary creates a test binary script that outputs content and exits with specified code.
// Returns the path to the created binary.
//
// Parameters:
//   - t: Testing instance for error handling
//   - tempDir: Temporary directory where binary will be created
//   - output: Stdout content to echo
//   - exitCode: Exit code for the binary (0-255)
//
// Returns:
//   - string: Path to the created binary executable
func CreateTestBinary(t *testing.T, tempDir string, output string, exitCode int) string {
	binaryPath := filepath.Join(tempDir, "test-binary")
	content := "#!/bin/sh\necho '" + output + "'\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

// CreateTestBinaryWithStderr creates a test binary that outputs to stderr and exits with specified code.
//
// Parameters:
//   - t: Testing instance for error handling
//   - tempDir: Temporary directory where binary will be created
//   - stderr: Stderr content to echo
//   - exitCode: Exit code for the binary (0-255)
//
// Returns:
//   - string: Path to the created binary executable
func CreateTestBinaryWithStderr(t *testing.T, tempDir string, stderr string, exitCode int) string {
	binaryPath := filepath.Join(tempDir, "test-binary")
	content := "#!/bin/sh\necho '" + stderr + "' >&2\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

// CreateSuccessBinary creates a test binary that exits successfully.
//
// Parameters:
//   - t: Testing instance for error handling
//   - tempDir: Temporary directory where binary will be created
//
// Returns:
//   - string: Path to the created binary executable
func CreateSuccessBinary(t *testing.T, tempDir string) string {
	binaryPath := filepath.Join(tempDir, "test-success")
	content := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

// CreateSlowBinary creates a test binary that sleeps before exiting (for timeout tests).
//
// Parameters:
//   - t: Testing instance for error handling
//   - tempDir: Temporary directory where binary will be created
//
// Returns:
//   - string: Path to the created binary executable
func CreateSlowBinary(t *testing.T, tempDir string) string {
	binaryPath := filepath.Join(tempDir, "test-slow")
	content := "#!/bin/sh\nsleep 10\nexit 0\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

// ContainsString checks if a string contains a substring.
//
// Parameters:
//   - s: The string to search in
//   - substr: The substring to search for
//
// Returns:
//   - bool: true if substr is found in s, false otherwise
func ContainsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
