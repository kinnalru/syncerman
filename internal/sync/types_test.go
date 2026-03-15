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

func TestCreateDestinationDirectories_AllFailureScenarios(t *testing.T) {
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
				cfg.AddProvider(providerName, config.PathMap{
					"docs": []config.Destination{{To: fmt.Sprintf("s3:backup/docs%d", i)}},
				})
			}

			engine := NewEngine(cfg, mockExec, nil)
			ctx := context.Background()
			err := engine.CreateDestinationDirectories(ctx, cfg, SyncOptions{})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDestinationPaths_AllScenarios(t *testing.T) {
	tests := []struct {
		name          string
		targets       []*SyncTarget
		expectError   bool
		errorContains string
	}{
		{
			name: "valid targets with local provider",
			targets: []*SyncTarget{
				{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "/backup/data"}},
			},
			expectError: false,
		},
		{
			name: "valid targets with remote provider",
			targets: []*SyncTarget{
				{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			},
			expectError: false,
		},
		{
			name: "valid targets with mixed providers",
			targets: []*SyncTarget{
				{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "/backup/data"}},
				{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
				{Provider: "s3", SourcePath: "bucket/path", Destination: config.Destination{To: "dropbox:backup/path"}},
			},
			expectError: false,
		},
		{
			name: "empty remote format",
			targets: []*SyncTarget{
				{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: ""}},
			},
			expectError:   true,
			errorContains: "empty path",
		},
		{
			name: "empty path in remote",
			targets: []*SyncTarget{
				{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "gdrive:"}},
			},
			expectError:   true,
			errorContains: "path cannot be empty",
		},
		{
			name: "empty provider in remote",
			targets: []*SyncTarget{
				{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: ":path"}},
			},
			expectError:   true,
			errorContains: "provider name cannot be empty",
		},
		{
			name: "local path with colon (not a remote)",
			targets: []*SyncTarget{
				{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "C:/data"}},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(nil, nil, nil)
			err := engine.ValidateDestinationPaths(tt.targets)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createTestConfigFile(t *testing.T, path string) {
	content := `gdrive:
  docs:
    - to: s3:backup/docs
local:
  /data:
    - to: /backup/data`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}
