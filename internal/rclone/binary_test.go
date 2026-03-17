package rclone

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRcloneBinary(t *testing.T) {
	originalPath := os.Getenv("PATH")

	testCases := []struct {
		name        string
		setupFunc   func() func()
		wantErr     bool
		errContains string
	}{
		{
			name: "finds rclone in PATH",
			setupFunc: func() func() {
				tempDir := t.TempDir()
				binaryPath := filepath.Join(tempDir, "rclone")
				if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho 'rclone'\n"), 0o755); err != nil {
					t.Fatalf("Failed to create fake rclone: %v", err)
				}
				_ = os.Setenv("PATH", tempDir+string(filepath.ListSeparator)+originalPath)
				_ = os.Unsetenv(RcloneEnvVar)
				return func() {
					_ = os.Setenv("PATH", originalPath)
					_ = os.Remove(binaryPath)
				}
			},
			wantErr: false,
		},
		{
			name: "uses custom path from env var",
			setupFunc: func() func() {
				tempDir := t.TempDir()
				binaryPath := filepath.Join(tempDir, "rclone-custom")
				if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho 'rclone'\n"), 0o755); err != nil {
					t.Fatalf("Failed to create fake rclone: %v", err)
				}
				_ = os.Setenv(RcloneEnvVar, binaryPath)
				return func() {
					_ = os.Unsetenv(RcloneEnvVar)
					_ = os.Remove(binaryPath)
				}
			},
			wantErr: false,
		},
		{
			name: "custom path with relative directory",
			setupFunc: func() func() {
				tempDir := t.TempDir()
				binaryPath := filepath.Join(tempDir, "rclone")
				if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho 'rclone'\n"), 0o755); err != nil {
					t.Fatalf("Failed to create fake rclone: %v", err)
				}
				absPath, _ := filepath.Abs(binaryPath)
				_ = os.Setenv(RcloneEnvVar, absPath)
				return func() {
					_ = os.Unsetenv(RcloneEnvVar)
					_ = os.Remove(binaryPath)
				}
			},
			wantErr: false,
		},
		{
			name: "custom path does not exist",
			setupFunc: func() func() {
				_ = os.Setenv(RcloneEnvVar, "/nonexistent/path/to/rclone")
				return func() {
					_ = os.Unsetenv(RcloneEnvVar)
				}
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "not found in PATH",
			setupFunc: func() func() {
				emptyDir := t.TempDir()
				_ = os.Setenv("PATH", emptyDir)
				_ = os.Unsetenv(RcloneEnvVar)
				return func() {
					_ = os.Setenv("PATH", originalPath)
				}
			},
			wantErr: true,
		},
		{
			name: "custom path with simple name (no separator)",
			setupFunc: func() func() {
				originalWd, _ := os.Getwd()
				originalPath := os.Getenv("PATH")
				testDir := t.TempDir()
				testBinary := filepath.Join(testDir, "rclone")
				if err := os.WriteFile(testBinary, []byte("#!/bin/sh\necho 'rclone'\n"), 0o755); err != nil {
					t.Fatalf("Failed to create fake rclone: %v", err)
				}
				if err := os.Chdir(testDir); err != nil {
					t.Fatalf("Failed to change directory: %v", err)
				}
				_ = os.Setenv("PATH", "")
				_ = os.Setenv(RcloneEnvVar, "rclone")
				return func() {
					_ = os.Chdir(originalWd)
					_ = os.Setenv("PATH", originalPath)
					_ = os.Unsetenv(RcloneEnvVar)
					_ = os.Remove(testBinary)
				}
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := tc.setupFunc()
			defer cleanup()

			path, err := FindRcloneBinary()

			if tc.wantErr {
				if err == nil {
					t.Errorf("FindRcloneBinary() expected error, got nil")
					return
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("FindRcloneBinary() error = %v, want error containing %q", err, tc.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("FindRcloneBinary() unexpected error: %v", err)
					return
				}
				if path == "" {
					t.Errorf("FindRcloneBinary() returned empty path")
				}
			}
		})
	}
}

func TestConfigFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		cleanup func()
		wantErr bool
	}{
		{
			name: "valid rclone in PATH",
			setup: func() {
				_ = os.Unsetenv(RcloneEnvVar)
			},
			cleanup: nil,
			wantErr: false,
		},
		{
			name: "custom env var set",
			setup: func() {
				tempDir := t.TempDir()
				binaryPath := filepath.Join(tempDir, "rclone-custom")
				if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho 'rclone'\n"), 0o755); err != nil {
					t.Fatalf("Failed to create fake rclone: %v", err)
				}
				_ = os.Setenv(RcloneEnvVar, binaryPath)
			},
			cleanup: nil,
			wantErr: false,
		},
		{
			name: "rclone not found - error case",
			setup: func() {
				originalPath := os.Getenv("PATH")
				emptyDir := t.TempDir()
				_ = os.Setenv("PATH", emptyDir)
				_ = os.Unsetenv(RcloneEnvVar)
				t.Cleanup(func() {
					_ = os.Setenv("PATH", originalPath)
				})
			},
			cleanup: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer func() { _ = os.Unsetenv(RcloneEnvVar) }()

			config, err := ConfigFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && config == nil {
				t.Errorf("ConfigFromEnv() returned nil config")
			}
			if tt.wantErr && config != nil {
				t.Errorf("ConfigFromEnv() should return nil config on error, got %+v", config)
			}
		})
	}
}

func skipIfNoRclone(t *testing.T) {
	if _, err := exec.LookPath(RcloneBinaryName); err != nil {
		t.Skip("rclone binary not found in PATH, skipping test")
	}
}
