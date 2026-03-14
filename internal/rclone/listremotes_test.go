package rclone

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"syncerman/internal/logger"
)

func TestListRemotes(t *testing.T) {
	testCases := []struct {
		name        string
		output      string
		exitCode    int
		wantRemotes []string
		wantErr     bool
		setupBinary bool
	}{
		{
			name:        "multiple remotes",
			output:      "gdrive:\nydisk:\ns3:\n",
			exitCode:    0,
			wantRemotes: []string{"gdrive", "ydisk", "s3"},
			wantErr:     false,
			setupBinary: true,
		},
		{
			name:        "single remote",
			output:      "gdrive:\n",
			exitCode:    0,
			wantRemotes: []string{"gdrive"},
			wantErr:     false,
			setupBinary: true,
		},
		{
			name:        "no remotes",
			output:      "",
			exitCode:    0,
			wantRemotes: []string{},
			wantErr:     false,
			setupBinary: true,
		},
		{
			name:        "remotes with spaces and newlines",
			output:      "  gdrive:\n  ydisk:\n\n",
			exitCode:    0,
			wantRemotes: []string{"gdrive", "ydisk"},
			wantErr:     false,
			setupBinary: true,
		},
		{
			name:        "rclone command fails",
			output:      "error: rclone not configured\n",
			exitCode:    1,
			wantRemotes: nil,
			wantErr:     true,
			setupBinary: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.setupBinary {
				t.Skip("requires test binary")
			}

			tempDir := t.TempDir()
			binaryPath := filepath.Join(tempDir, "test-listremotes")
			content := "#!/bin/sh\necho '" + tc.output + "'\nexit " + string(rune('0'+tc.exitCode)) + "\n"
			if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
				t.Fatalf("Failed to create test binary: %v", err)
			}

			config := &Config{BinaryPath: binaryPath}
			exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
			exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

			ctx := context.Background()
			remotes, err := ListRemotes(ctx, exec)

			if (err != nil) != tc.wantErr {
				t.Errorf("ListRemotes() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr {
				return
			}

			if len(remotes) != len(tc.wantRemotes) {
				t.Errorf("ListRemotes() got %d remotes, want %d", len(remotes), len(tc.wantRemotes))
			}

			for i, want := range tc.wantRemotes {
				if i >= len(remotes) || remotes[i] != want {
					t.Errorf("ListRemotes()[%d] = %q, want %q", i, remotes[i], want)
				}
			}
		})
	}
}

func TestRemoteExists(t *testing.T) {
	testCases := []struct {
		name        string
		output      string
		remoteName  string
		wantExists  bool
		wantErr     bool
		setupBinary bool
	}{
		{
			name:        "remote exists",
			output:      "gdrive:\nydisk:\ns3:\n",
			remoteName:  "gdrive",
			wantExists:  true,
			wantErr:     false,
			setupBinary: true,
		},
		{
			name:        "remote does not exist",
			output:      "gdrive:\nydisk:\ns3:\n",
			remoteName:  "dropbox",
			wantExists:  false,
			wantErr:     false,
			setupBinary: true,
		},
		{
			name:        "no remotes configured",
			output:      "",
			remoteName:  "gdrive",
			wantExists:  false,
			wantErr:     false,
			setupBinary: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.setupBinary {
				t.Skip("requires test binary")
			}

			tempDir := t.TempDir()
			binaryPath := filepath.Join(tempDir, "test-remoteexists")
			content := "#!/bin/sh\necho '" + tc.output + "'\nexit 0\n"
			if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
				t.Fatalf("Failed to create test binary: %v", err)
			}

			config := &Config{BinaryPath: binaryPath}
			exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
			exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

			ctx := context.Background()
			exists, err := RemoteExists(ctx, exec, tc.remoteName)

			if (err != nil) != tc.wantErr {
				t.Errorf("RemoteExists() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if exists != tc.wantExists {
				t.Errorf("RemoteExists() = %v, want %v", exists, tc.wantExists)
			}
		})
	}
}

func TestListRemotes_ColonStripping(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "test-colon")
	content := "#!/bin/sh\necho 'remote1:'\necho 'remote2:'\necho 'remote3:'\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	config := &Config{BinaryPath: binaryPath}
	exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
	exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

	ctx := context.Background()
	remotes, err := ListRemotes(ctx, exec)

	if err != nil {
		t.Fatalf("ListRemotes() unexpected error: %v", err)
	}

	for _, remote := range remotes {
		if strings.HasSuffix(remote, ":") {
			t.Errorf("ListRemotes() remote %q should not have trailing colon", remote)
		}
	}
}

func TestListRemotes_RealRclone(t *testing.T) {
	skipIfNoRclone(t)

	config, err := ConfigFromEnv()
	if err != nil {
		t.Skipf("Skipping real rclone test: %v", err)
	}

	exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
	exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

	ctx := context.Background()
	remotes, err := ListRemotes(ctx, exec)

	if err != nil {
		t.Errorf("ListRemotes() with real rclone failed: %v", err)
		return
	}

	t.Logf("Found %d remotes: %v", len(remotes), remotes)
}

func TestListRemotes_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "test-cancel")
	content := "#!/bin/sh\nsleep 10\necho 'gdrive:'\n"
	if err := os.WriteFile(binaryPath, []byte(content), 0o755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	config := &Config{BinaryPath: binaryPath}
	exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
	exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		cancel()
	}()

	_, err := ListRemotes(ctx, exec)

	if err == nil {
		t.Error("ListRemotes() expected error for cancelled context, got nil")
	}
}
