package sync

import (
	"context"
	"fmt"
	"os"
	"testing"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncEngineFromConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := tmpDir + "/test-config.yaml"
	createTestConfigFile(t, configPath)

	engine, err := SyncEngineFromConfig(configPath, nil, nil)

	require.NoError(t, err)
	require.NotNil(t, engine)
	assert.NotNil(t, engine.config)
}

func TestSyncEngineFromConfig_InvalidPath(t *testing.T) {
	_, err := SyncEngineFromConfig("/nonexistent/config.yaml", nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestSyncEngineFromConfig_InvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := tmpDir + "/invalid-config.yaml"
	err = os.WriteFile(configPath, []byte("invalid: yaml: content:\n  - broken"), 0644)
	require.NoError(t, err)

	_, err = SyncEngineFromConfig(configPath, nil, nil)

	assert.Error(t, err)
}

func TestSyncEngineFromConfig_ValidationErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := tmpDir + "/invalid-config.yaml"
	err = os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	_, err = SyncEngineFromConfig(configPath, nil, nil)

	assert.Error(t, err)
}

func TestCreateAllDirectories_AllFailureScenarios(t *testing.T) {
	tests := []struct {
		name        string
		results     []*rclone.Result
		errors      []error
		expectError bool
	}{
		{
			name: "single directory creation",
			results: []*rclone.Result{
				{ExitCode: 0, Stdout: "", Stderr: ""},
			},
			errors:      []error{nil},
			expectError: false,
		},
		{
			name: "multiple directory creation",
			results: []*rclone.Result{
				{ExitCode: 0, Stdout: "", Stderr: ""},
				{ExitCode: 0, Stdout: "", Stderr: ""},
				{ExitCode: 0, Stdout: "", Stderr: ""},
			},
			errors:      []error{nil, nil, nil},
			expectError: false,
		},
		{
			name: "failure on second directory",
			results: []*rclone.Result{
				{ExitCode: 0, Stdout: "", Stderr: ""},
				{ExitCode: 1, Stdout: "", Stderr: "permission denied"},
			},
			errors:      []error{nil, assert.AnError},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{
				results: tt.results,
				errors:  tt.errors,
			}

			cfg := config.NewConfig()
			for i := 0; i < len(tt.results); i++ {
				providerName := fmt.Sprintf("gdrive%d", i)
				cfg.Jobs = append(cfg.Jobs, config.Job{
					ID:      providerName,
					Name:    providerName,
					Enabled: true,
					Tasks: []config.Task{
						{From: providerName + ":docs", Enabled: true, To: []config.Destination{{Path: fmt.Sprintf("s3:backup/docs%d", i)}}},
					},
				})
			}

			engine := NewEngine(cfg, mockExec, &mockLogger{})
			ctx := context.Background()
			err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createTestConfigFile(t *testing.T, path string) {
	content := `jobs:
  job1:
    tasks:
      - from: gdrive:docs
        to:
          - path: s3:backup/docs
  job2:
    tasks:
      - from: /data
        to:
          - path: /backup/data`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}
