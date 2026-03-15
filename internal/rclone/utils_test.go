package rclone

import (
	"context"
	"strings"
	"testing"

	syncerman_errors "syncerman/internal/errors"
	"syncerman/internal/logger"
)

func TestParseRemoteLine(t *testing.T) {
	testCases := []struct {
		name       string
		line       string
		wantRemote string
		wantValid  bool
	}{
		{
			name:       "basic remote with colon",
			line:       "gdrive:",
			wantRemote: "gdrive",
			wantValid:  true,
		},
		{
			name:       "remote without colon",
			line:       "gdrive",
			wantRemote: "gdrive",
			wantValid:  true,
		},
		{
			name:       "remote with spaces",
			line:       "  gdrive:  ",
			wantRemote: "gdrive",
			wantValid:  true,
		},
		{
			name:       "empty line",
			line:       "",
			wantRemote: "",
			wantValid:  false,
		},
		{
			name:       "whitespace only",
			line:       "   ",
			wantRemote: "",
			wantValid:  false,
		},
		{
			name:       "complex remote name",
			line:       "my-remote-123:",
			wantRemote: "my-remote-123",
			wantValid:  true,
		},
		{
			name:       "remote with underscores",
			line:       "my_remote_1:",
			wantRemote: "my_remote_1",
			wantValid:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			remote, valid := parseRemoteLine(tc.line)

			if valid != tc.wantValid {
				t.Errorf("parseRemoteLine() valid = %v, want %v", valid, tc.wantValid)
			}

			if valid && remote != tc.wantRemote {
				t.Errorf("parseRemoteLine() remote = %q, want %q", remote, tc.wantRemote)
			}
		})
	}
}

func TestMatchesAnyPattern(t *testing.T) {
	testCases := []struct {
		name     string
		stderr   string
		patterns []string
		want     bool
	}{
		{
			name:     "matches first pattern",
			stderr:   "Error: directory already exists",
			patterns: []string{"already exists", "file not found"},
			want:     true,
		},
		{
			name:     "matches second pattern",
			stderr:   "Error: file not found",
			patterns: []string{"already exists", "file not found"},
			want:     true,
		},
		{
			name:     "matches none",
			stderr:   "Error: permission denied",
			patterns: []string{"already exists", "file not found"},
			want:     false,
		},
		{
			name:     "empty stderr",
			stderr:   "",
			patterns: []string{"test pattern"},
			want:     false,
		},
		{
			name:     "empty patterns",
			stderr:   "some error",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "case insensitive match",
			stderr:   "Error: DIRECTORY ALREADY EXISTS",
			patterns: []string{"already exists"},
			want:     true,
		},
		{
			name:     "partial match",
			stderr:   "The directory already exists on remote",
			patterns: []string{"already exists"},
			want:     true,
		},
		{
			name:     "multiple patterns one match",
			stderr:   "Error: no such file",
			patterns: []string{"already exists", "no such file", "permission denied"},
			want:     true,
		},
		{
			name:     "subpattern match",
			stderr:   "Error: parent directory not found",
			patterns: []string{"directory not found"},
			want:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesAnyPattern(tc.stderr, tc.patterns)
			if got != tc.want {
				t.Errorf("matchesAnyPattern() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExecuteMkdirCommand(t *testing.T) {
	testCases := []struct {
		name               string
		remotePath         string
		exitCode           int
		stderr             string
		customErrorHandler func(Result) error
		wantErr            bool
		errContains        string
	}{
		{
			name:       "successful creation",
			remotePath: "gdrive:backups",
			exitCode:   0,
			stderr:     "",
			wantErr:    false,
		},
		{
			name:       "directory exists - treated as success",
			remotePath: "gdrive:backups",
			exitCode:   1,
			stderr:     "Error: directory already exists: backups",
			wantErr:    false,
		},
		{
			name:       "file exists - treated as success",
			remotePath: "gdrive:backups",
			exitCode:   1,
			stderr:     "Error: file exists: backups",
			wantErr:    false,
		},
		{
			name:       "path already exists - treated as success",
			remotePath: "gdrive:backups",
			exitCode:   1,
			stderr:     "Error: path already exists: /some/path",
			wantErr:    false,
		},
		{
			name:        "parent directory error",
			remotePath:  "gdrive:parent/child",
			exitCode:    1,
			stderr:      "Error: parent directory not found: parent",
			wantErr:     true,
			errContains: "failed to create directory",
		},
		{
			name:        "empty path",
			remotePath:  "",
			exitCode:    0,
			stderr:      "",
			wantErr:     true,
			errContains: "remote path cannot be empty",
		},
		{
			name:       "custom error handler with wrap",
			remotePath: "gdrive:readonly/dir",
			exitCode:   1,
			stderr:     "Error: permission denied",
			customErrorHandler: func(result Result) error {
				if strings.Contains(result.Stderr, "permission denied") {
					return syncerman_errors.NewRcloneError(result.Stderr, nil)
				}
				return nil
			},
			wantErr:     true,
			errContains: "failed to create directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			binaryPath := CreateTestBinaryWithStderr(t, tempDir, tc.stderr, tc.exitCode)

			config := &Config{BinaryPath: binaryPath}
			log := logger.NewConsoleLogger()
			log.SetLevel(logger.LevelQuiet)
			exec := NewExecutorWithLogger(config, log)

			ctx := context.Background()
			err := executeMkdirCommand(ctx, exec, tc.remotePath, tc.customErrorHandler)

			if (err != nil) != tc.wantErr {
				t.Errorf("executeMkdirCommand() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr && tc.errContains != "" {
				if err == nil {
					t.Errorf("executeMkdirCommand() expected error containing %q, got nil", tc.errContains)
					return
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("executeMkdirCommand() error = %v, want error containing %q", err, tc.errContains)
				}
			}
		})
	}
}

func TestExecuteMkdirCommand_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := CreateSlowBinary(t, tempDir)

	config := &Config{BinaryPath: binaryPath}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		cancel()
	}()

	err := executeMkdirCommand(ctx, exec, "gdrive:test", nil)
	if err == nil {
		t.Error("executeMkdirCommand() expected error for cancelled context, got nil")
	}
}

func TestMatchesAnyPattern_Performance(t *testing.T) {
	stderr := strings.Repeat("a", 10000) + "target pattern" + strings.Repeat("b", 10000)
	patterns := []string{"not here", "also not here", "target pattern", "nor here"}

	for i := 0; i < 100; i++ {
		if !matchesAnyPattern(stderr, patterns) {
			t.Errorf("matchesAnyPattern() iteration %d failed", i)
		}
	}
}
