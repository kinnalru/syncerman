package rclone

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"syncerman/internal/logger"
)

func TestMkdir(t *testing.T) {
	testCases := []struct {
		name        string
		remotePath  string
		exitCode    int
		stderr      string
		wantErr     bool
		errContains string
		setupBinary bool
	}{
		{
			name:        "successful creation",
			remotePath:  "gdrive:backups",
			exitCode:    0,
			stderr:      "",
			wantErr:     false,
			setupBinary: true,
		},
		{
			name:        "directory already exists",
			remotePath:  "gdrive:backups",
			exitCode:    1,
			stderr:      "Error: directory already exists: backups",
			wantErr:     false,
			setupBinary: true,
		},
		{
			name:        "parent directory not found",
			remotePath:  "gdrive:parent/child",
			exitCode:    1,
			stderr:      "Error: parent directory not found: parent",
			wantErr:     true,
			errContains: "failed to create directory",
			setupBinary: true,
		},
		{
			name:        "invalid remote",
			remotePath:  "invalid:remote:path",
			exitCode:    1,
			stderr:      "Failed to mkdir invalid:remote/path: remote not found",
			wantErr:     true,
			errContains: "failed to create directory",
			setupBinary: true,
		},
		{
			name:        "empty path",
			remotePath:  "",
			exitCode:    0,
			stderr:      "",
			wantErr:     true,
			errContains: "remote path cannot be empty",
			setupBinary: false,
		},
		{
			name:        "permission denied",
			remotePath:  "read-only:dir",
			exitCode:    1,
			stderr:      "Error: permission denied",
			wantErr:     true,
			errContains: "failed to create directory",
			setupBinary: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.setupBinary {
			} else {
				tempDir := t.TempDir()
				binaryPath := filepath.Join(tempDir, "test-mkdir")
				if tc.stderr != "" {
					if tc.exitCode != 0 {
						binaryPath = createMkdirBinary(t, tempDir, tc.exitCode, tc.stderr)
					} else {
						binaryPath = createSuccessBinary(t, tempDir)
					}
				} else {
					binaryPath = createSuccessBinary(t, tempDir)
				}

				config := &Config{BinaryPath: binaryPath}
				exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
				exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

				ctx := context.Background()
				err := Mkdir(ctx, exec, tc.remotePath)

				if (err != nil) != tc.wantErr {
					t.Errorf("Mkdir() error = %v, wantErr %v", err, tc.wantErr)
					return
				}

				if tc.wantErr && tc.errContains != "" {
					if err == nil {
						t.Errorf("Mkdir() expected error containing %q, got nil", tc.errContains)
						return
					}
					if !strings.Contains(err.Error(), tc.errContains) {
						t.Errorf("Mkdir() error = %v, want error containing %q", err, tc.errContains)
					}
				}
			}
		})
	}
}

func TestIsDirectoryExistsError(t *testing.T) {
	testCases := []struct {
		name     string
		stderr   string
		wantTrue bool
	}{
		{
			name:     "already exists",
			stderr:   "Error: directory already exists: backups",
			wantTrue: true,
		},
		{
			name:     "file exists",
			stderr:   "Error: file exists: path/to/file",
			wantTrue: true,
		},
		{
			name:     "path already exists",
			stderr:   "Error: path already exists: /some/path",
			wantTrue: true,
		},
		{
			name:     "empty stderr",
			stderr:   "",
			wantTrue: false,
		},
		{
			name:     "other error",
			stderr:   "Error: permission denied",
			wantTrue: false,
		},
		{
			name:     "case insensitive",
			stderr:   "Error: DIRECTORY ALREADY EXISTS",
			wantTrue: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isDirectoryExistsError(tc.stderr)
			if got != tc.wantTrue {
				t.Errorf("isDirectoryExistsError() = %v, want %v", got, tc.wantTrue)
			}
		})
	}
}

func TestIsParentDirNotFoundError(t *testing.T) {
	testCases := []struct {
		name     string
		stderr   string
		wantTrue bool
	}{
		{
			name:     "parent directory",
			stderr:   "Error: parent directory not found: parent",
			wantTrue: true,
		},
		{
			name:     "no such file",
			stderr:   "Error: no such file or directory",
			wantTrue: true,
		},
		{
			name:     "not found",
			stderr:   "Error: directory not found",
			wantTrue: true,
		},
		{
			name:     "empty stderr",
			stderr:   "",
			wantTrue: false,
		},
		{
			name:     "other error",
			stderr:   "Error: permission denied",
			wantTrue: false,
		},
		{
			name:     "case insensitive",
			stderr:   "Error: PARENT DIRECTORY NOT FOUND",
			wantTrue: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isParentDirNotFoundError(tc.stderr)
			if got != tc.wantTrue {
				t.Errorf("isParentDirNotFoundError() = %v, want %v", got, tc.wantTrue)
			}
		})
	}
}

func TestCreatePath(t *testing.T) {
	testCases := []struct {
		name        string
		remotePath  string
		exitCode    int
		stderr      string
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful creation",
			remotePath: "gdrive:backups",
			exitCode:   0,
			stderr:     "",
			wantErr:    false,
		},
		{
			name:       "directory already exists",
			remotePath: "gdrive:backups",
			exitCode:   1,
			stderr:     "Error: directory already exists: backups",
			wantErr:    false,
		},
		{
			name:        "parent directory not found",
			remotePath:  "gdrive:parent/child",
			exitCode:    1,
			stderr:      "Error: parent directory not found: parent",
			wantErr:     true,
			errContains: "parent directory does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			var binaryPath string

			if tc.exitCode == 0 {
				binaryPath = createSuccessBinary(t, tempDir)
			} else {
				binaryPath = createMkdirBinary(t, tempDir, tc.exitCode, tc.stderr)
			}

			config := &Config{BinaryPath: binaryPath}
			exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
			exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

			ctx := context.Background()
			err := CreatePath(ctx, exec, tc.remotePath)

			if (err != nil) != tc.wantErr {
				t.Errorf("CreatePath() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr && tc.errContains != "" {
				if err == nil {
					t.Errorf("CreatePath() expected error containing %q, got nil", tc.errContains)
					return
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("CreatePath() error = %v, want error containing %q", err, tc.errContains)
				}
			}
		})
	}
}

func TestMkdir_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := createSlowBinary(t, tempDir)

	config := &Config{BinaryPath: binaryPath}
	exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
	exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Mkdir(ctx, exec, "gdrive:test")
	if err == nil {
		t.Error("Mkdir() expected error for cancelled context, got nil")
	}
}

func TestEmptyPath(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := createSuccessBinary(t, tempDir)

	config := &Config{BinaryPath: binaryPath}
	exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
	exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

	ctx := context.Background()

	err := Mkdir(ctx, exec, "")
	if err == nil {
		t.Error("Mkdir() expected error for empty path, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "remote path cannot be empty") {
		t.Errorf("Mkdir() error = %v, want error containing 'remote path cannot be empty'", err)
	}
}

func createSuccessBinary(t *testing.T, tempDir string) string {
	binaryPath := filepath.Join(tempDir, "test-success")
	content := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

func createMkdirBinary(t *testing.T, tempDir string, exitCode int, stderr string) string {
	binaryPath := filepath.Join(tempDir, "test-mkdir")
	content := "#!/bin/sh\necho '" + stderr + "' >&2\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}

func createSlowBinary(t *testing.T, tempDir string) string {
	binaryPath := filepath.Join(tempDir, "test-slow")
	content := "#!/bin/sh\nsleep 10\nexit 0\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binaryPath
}
