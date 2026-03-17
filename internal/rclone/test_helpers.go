package rclone

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"gitlab.com/kinnalru/syncerman/internal/logger"
)

// createTestBinary creates a test binary script that outputs content and exits with specified code.
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
func createTestBinary(t *testing.T, tempDir string, output string, exitCode int) string {
	binaryPath := filepath.Join(tempDir, "test-binary")
	content := "#!/bin/sh\necho '" + output + "'\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

// createTestBinaryWithStderr creates a test binary that outputs to stderr and exits with specified code.
//
// Parameters:
//   - t: Testing instance for error handling
//   - tempDir: Temporary directory where binary will be created
//   - stderr: Stderr content to echo
//   - exitCode: Exit code for the binary (0-255)
//
// Returns:
//   - string: Path to the created binary executable
func createTestBinaryWithStderr(t *testing.T, tempDir string, stderr string, exitCode int) string {
	binaryPath := filepath.Join(tempDir, "test-binary")
	content := "#!/bin/sh\necho '" + stderr + "' >&2\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

// createSuccessBinary creates a test binary that exits successfully.
//
// Parameters:
//   - t: Testing instance for error handling
//   - tempDir: Temporary directory where binary will be created
//
// Returns:
//   - string: Path to the created binary executable
func createSuccessBinary(t *testing.T, tempDir string) string {
	binaryPath := filepath.Join(tempDir, "test-success")
	content := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

// createSlowBinary creates a test binary that sleeps before exiting (for timeout tests).
//
// Parameters:
//   - t: Testing instance for error handling
//   - tempDir: Temporary directory where binary will be created
//
// Returns:
//   - string: Path to the created binary executable
func createSlowBinary(t *testing.T, tempDir string) string {
	binaryPath := filepath.Join(tempDir, "test-slow")
	content := "#!/bin/sh\nsleep 10\nexit 0\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

// setupTestExecutor creates a configured executor for testing.
// This consolidates common test setup code and provides a consistent way
// to create test executors with quiet logging.
//
// Parameters:
//   - t: Testing instance
//   - binaryPath: Path to the test binary executable
//
// Returns:
//   - context.Context: Background context for testing
//   - Executor: Configured executor with quiet logger
func setupTestExecutor(t *testing.T, binaryPath string) (context.Context, Executor) {
	config := &Config{BinaryPath: binaryPath}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)
	ctx := context.Background()
	return ctx, exec
}
