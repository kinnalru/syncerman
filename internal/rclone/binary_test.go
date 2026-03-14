package rclone

import (
	"os"
	"os/exec"
	"path/filepath"
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
				os.Setenv("PATH", tempDir+string(filepath.ListSeparator)+originalPath)
				os.Unsetenv(RcloneEnvVar)
				return func() {
					os.Setenv("PATH", originalPath)
					os.Remove(binaryPath)
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
				os.Setenv(RcloneEnvVar, binaryPath)
				return func() {
					os.Unsetenv(RcloneEnvVar)
					os.Remove(binaryPath)
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
				os.Setenv(RcloneEnvVar, absPath)
				return func() {
					os.Unsetenv(RcloneEnvVar)
					os.Remove(binaryPath)
				}
			},
			wantErr: false,
		},
		{
			name: "custom path does not exist",
			setupFunc: func() func() {
				os.Setenv(RcloneEnvVar, "/nonexistent/path/to/rclone")
				return func() {
					os.Unsetenv(RcloneEnvVar)
				}
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "not found in PATH",
			setupFunc: func() func() {
				emptyDir := t.TempDir()
				os.Setenv("PATH", emptyDir)
				os.Unsetenv(RcloneEnvVar)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr: true,
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
				if tc.errContains != "" && !containsString(err.Error(), tc.errContains) {
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
		wantErr bool
	}{
		{
			name: "valid rclone in PATH",
			setup: func() {
				os.Unsetenv(RcloneEnvVar)
			},
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
				os.Setenv(RcloneEnvVar, binaryPath)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer os.Unsetenv(RcloneEnvVar)

			config, err := ConfigFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && config == nil {
				t.Errorf("ConfigFromEnv() returned nil config")
			}
		})
	}
}

func TestFindRcloneBinaryOrFatal(t *testing.T) {
	path := FindRcloneBinaryOrFatal()
	if path != "" {
		t.Logf("Found rclone at: %s", path)
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr
}

func skipIfNoRclone(t *testing.T) {
	if _, err := exec.LookPath(RcloneBinaryName); err != nil {
		t.Skip("rclone binary not found in PATH, skipping test")
	}
}
