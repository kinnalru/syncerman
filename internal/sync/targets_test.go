package sync

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockListRemotesExecutor struct {
	remotes []string
	error   error
}

func (m *mockListRemotesExecutor) Run(ctx context.Context, args ...string) (*rclone.Result, error) {
	if m.error != nil {
		return nil, m.error
	}

	if len(args) > 0 && args[0] == "listremotes" {
		stdout := strings.Join(m.remotes, "\n")
		return &rclone.Result{
			ExitCode: 0,
			Stdout:   stdout,
			Stderr:   "",
		}, nil
	}

	return &rclone.Result{
		ExitCode: 0,
		Stdout:   "",
		Stderr:   "",
	}, nil
}

func TestRemoteProviderExists(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		remotes     []string
		resultError error
		wantExists  bool
		wantError   bool
	}{
		{
			name:       "local provider always exists",
			provider:   "local",
			remotes:    []string{},
			wantExists: true,
			wantError:  false,
		},
		{
			name:       "remote provider exists",
			provider:   "gdrive",
			remotes:    []string{"gdrive", "s3", "dropbox"},
			wantExists: true,
			wantError:  false,
		},
		{
			name:       "remote provider does not exist",
			provider:   "s3",
			remotes:    []string{"gdrive", "dropbox"},
			wantExists: false,
			wantError:  false,
		},
		{
			name:        "error checking provider",
			provider:    "invalid",
			remotes:     []string{},
			resultError: fmt.Errorf("rclone error"),
			wantExists:  false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := &mockListRemotesExecutor{
				remotes: tt.remotes,
				error:   tt.resultError,
			}

			engine := NewEngine(nil, exec, nil)
			exists, err := engine.RemoteProviderExists(context.Background(), tt.provider)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

func TestValidateTargets_NoProviders(t *testing.T) {
	engine := NewEngine(nil, nil, nil)
	cfg := config.NewConfig()

	err := engine.ValidateTargets(context.Background(), cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no jobs configured")
}

func TestValidateTargets_EmptyProvider(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "empty_job",
		Name:    "empty_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: ":path", Enabled: true, To: []config.Destination{
				{Path: "dest"},
			}},
		},
	})

	exec := &mockListRemotesExecutor{remotes: []string{"local", "gdrive"}}
	engine := NewEngine(nil, exec, nil)
	err := engine.ValidateTargets(context.Background(), cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider name cannot be empty")
}

func TestValidateTargets_ProviderNotFound(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "unknown_job",
		Name:    "unknown_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "unknown:path", Enabled: true, To: []config.Destination{
				{Path: "dest"},
			}},
		},
	})

	exec := &mockListRemotesExecutor{remotes: []string{"local", "gdrive"}}
	engine := NewEngine(nil, exec, nil)
	err := engine.ValidateTargets(context.Background(), cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in rclone configuration")
}

func TestValidateTargets_VerificationError(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "gdrive_job",
		Name:    "gdrive_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "gdrive:path", Enabled: true, To: []config.Destination{
				{Path: "dest"},
			}},
		},
	})

	exec := &mockListRemotesExecutor{error: fmt.Errorf("rclone connection failed")}
	engine := NewEngine(nil, exec, nil)
	err := engine.ValidateTargets(context.Background(), cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify provider")
}

func TestValidateTargets_SuccessWithMixedProviders(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "gdrive_job",
		Name:    "gdrive_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "gdrive:docs", Enabled: true, To: []config.Destination{
				{Path: "s3:backup/docs"},
			}},
		},
	})
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "local_job",
		Name:    "local_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "/data", Enabled: true, To: []config.Destination{
				{Path: "gdrive:data"},
			}},
		},
	})

	exec := &mockListRemotesExecutor{remotes: []string{"local", "gdrive", "s3"}}
	engine := NewEngine(nil, exec, nil)
	err := engine.ValidateTargets(context.Background(), cfg)

	assert.NoError(t, err)
}

func TestExpandTargets_EmptySourcePath(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "gdrive_job",
		Name:    "gdrive_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "", Enabled: true, To: []config.Destination{
				{Path: "dest"},
			}},
		},
	})

	engine := NewEngine(nil, nil, nil)
	targets, err := engine.ExpandTargets(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source 'from' cannot be empty")
	assert.Nil(t, targets)
}

func TestExpandTargets_NoDestinations(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "gdrive_job",
		Name:    "gdrive_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "gdrive:docs", Enabled: true, To: []config.Destination{}},
		},
	})

	engine := NewEngine(nil, nil, nil)
	targets, err := engine.ExpandTargets(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no destinations configured")
	assert.Nil(t, targets)
}

func TestExpandTargets_EmptyDestination(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "gdrive_job",
		Name:    "gdrive_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "gdrive:docs", Enabled: true, To: []config.Destination{
				{Path: ""},
			}},
		},
	})

	engine := NewEngine(nil, nil, nil)
	targets, err := engine.ExpandTargets(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destination 'path' cannot be empty")
	assert.Nil(t, targets)
}

func TestExpandTargets_MultipleDestinations(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "gdrive_job",
		Name:    "gdrive_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "gdrive:docs", Enabled: true, To: []config.Destination{
				{Path: "s3:backup/docs"},
				{Path: "dropbox:backup/docs"},
				{Path: "local:/backup/docs"},
			}},
		},
	})

	engine := NewEngine(nil, nil, nil)
	targets, err := engine.ExpandTargets(cfg)

	require.NoError(t, err)
	assert.Len(t, targets, 3)
}

func TestExpandTargets_NoValidTargets(t *testing.T) {
	cfg := config.NewConfig()

	engine := NewEngine(nil, nil, nil)
	targets, err := engine.ExpandTargets(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid sync targets")
	assert.Nil(t, targets)
}

func TestExpandTargets_HasResyncFlag(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "gdrive_job",
		Name:    "gdrive_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "gdrive:docs", Enabled: true, To: []config.Destination{
				{Path: "s3:backup/docs", Resync: true},
			}},
		},
	})

	engine := NewEngine(nil, nil, nil)
	targets, err := engine.ExpandTargets(cfg)

	require.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.True(t, targets[0].Resync)
}

func TestValidate(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Jobs = append(cfg.Jobs, config.Job{
		ID:      "local_job",
		Name:    "local_job",
		Enabled: true,
		Tasks: []config.Task{
			{From: "/data", Enabled: true, To: []config.Destination{
				{Path: "dest"},
			}},
		},
	})

	exec := &mockListRemotesExecutor{remotes: []string{"local", "gdrive"}}
	engine := NewEngine(nil, exec, nil)
	err := engine.Validate(context.Background(), cfg)

	assert.NoError(t, err)
}

func TestFormatRemote(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		path     string
		want     string
	}{
		{
			name:     "local provider",
			provider: "local",
			path:     "/path/to/file",
			want:     "/path/to/file",
		},
		{
			name:     "remote provider",
			provider: "gdrive",
			path:     "docs",
			want:     "gdrive:docs",
		},
		{
			name:     "remote provider with nested path",
			provider: "s3",
			path:     "bucket/path/to/file",
			want:     "s3:bucket/path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRemote(tt.provider, tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseRemote(t *testing.T) {
	tests := []struct {
		name         string
		remote       string
		wantProvider string
		wantPath     string
		wantError    bool
		errorMsg     string
	}{
		{
			name:         "local path",
			remote:       "/path/to/file",
			wantProvider: "local",
			wantPath:     "/path/to/file",
			wantError:    false,
		},
		{
			name:         "remote path",
			remote:       "gdrive:docs",
			wantProvider: "gdrive",
			wantPath:     "docs",
			wantError:    false,
		},
		{
			name:         "remote with nested path",
			remote:       "s3:bucket/path/to/file",
			wantProvider: "s3",
			wantPath:     "bucket/path/to/file",
			wantError:    false,
		},
		{
			name:         "empty provider",
			remote:       ":path",
			wantProvider: "",
			wantPath:     "",
			wantError:    true,
			errorMsg:     "provider name cannot be empty",
		},
		{
			name:         "empty path",
			remote:       "gdrive:",
			wantProvider: "",
			wantPath:     "",
			wantError:    true,
			errorMsg:     "path cannot be empty",
		},
		{
			name:         "invalid format multiple colons handled",
			remote:       "gdrive:docs:nested",
			wantProvider: "gdrive",
			wantPath:     "docs:nested",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, path, err := ParseRemote(tt.remote)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantProvider, provider)
				assert.Equal(t, tt.wantPath, path)
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		fmt.Errorf("error 1"),
		fmt.Errorf("error 2"),
	}

	errMsg := errs.Error()
	assert.Contains(t, errMsg, "validation errors")
	assert.Contains(t, errMsg, "error 1")
	assert.Contains(t, errMsg, "error 2")
}

func TestValidationErrors_Empty(t *testing.T) {
	errs := ValidationErrors{}
	errMsg := errs.Error()
	assert.Empty(t, errMsg)
}

func TestStripProviderHash(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "valid path with hash suffix",
			path:     "gd{ABC}:path/",
			expected: "gd:path/",
		},
		{
			name:     "valid path with hash suffix long hash",
			path:     "gdrive{YRXYK123}:documents/work",
			expected: "gdrive:documents/work",
		},
		{
			name:     "path without hash",
			path:     "gd:path/",
			expected: "gd:path/",
		},
		{
			name:     "path with multiple colons",
			path:     "gd:path/to:file",
			expected: "gd:path/to:file",
		},
		{
			name:     "path with multiple colons and hash",
			path:     "gd{ABC}:path/to:file",
			expected: "gd:path/to:file",
		},
		{
			name:     "empty string",
			path:     "",
			expected: "",
		},
		{
			name:     "invalid format missing hash",
			path:     "gd{}",
			expected: "gd{}",
		},
		{
			name:     "invalid format no colon after hash",
			path:     "gd{ABC}path",
			expected: "gd{ABC}path",
		},
		{
			name:     "invalid format hash without provider",
			path:     "{ABC}:path",
			expected: "{ABC}:path",
		},
		{
			name:     "local path (no colon)",
			path:     "/local/path/to/file",
			expected: "/local/path/to/file",
		},
		{
			name:     "mixed case hash",
			path:     "s3{AbCdEf}:bucket/path",
			expected: "s3:bucket/path",
		},
		{
			name:     "path with only hash and colon",
			path:     "gd{ABC}:",
			expected: "gd:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripProviderHash(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
