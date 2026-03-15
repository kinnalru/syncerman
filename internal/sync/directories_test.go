package sync

import (
	"context"
	"testing"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMkdirExecutor struct {
	results []*rclone.Result
	errors  []error
	index   int
}

func (m *mockMkdirExecutor) Run(ctx context.Context, args ...string) (*rclone.Result, error) {
	if m.index >= len(m.results) {
		return &rclone.Result{ExitCode: 0, Stdout: "", Stderr: ""}, nil
	}

	result := m.results[m.index]
	m.index++

	err := error(nil)
	if m.index-1 < len(m.errors) {
		err = m.errors[m.index-1]
	}

	return result, err
}

func TestCreateAllDirectories_Success(t *testing.T) {
	mockExec := &mockMkdirExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "", Stderr: ""},
			{ExitCode: 0, Stdout: "", Stderr: ""},
		},
	}

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{
			{To: "s3:backup/docs"},
		},
	})
	cfg.AddProvider("local", config.PathMap{
		"/data": []config.Destination{
			{To: "gdrive:data"},
		},
	})

	engine := NewEngine(cfg, mockExec, nil)
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	require.NoError(t, err)
}

func TestCreateAllDirectories_DryRun(t *testing.T) {
	mockExec := &mockMkdirExecutor{}

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{
			{To: "s3:backup/docs"},
		},
	})

	engine := NewEngine(cfg, mockExec, nil)
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{DryRun: true})

	require.NoError(t, err)
}

func TestCreateAllDirectories_Error(t *testing.T) {
	mockExec := &mockMkdirExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "", Stderr: ""},
			{ExitCode: 1, Stdout: "", Stderr: "permission denied"},
		},
		errors: []error{
			nil,
			assert.AnError,
		},
	}

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{
			{To: "s3:backup/docs"},
			{To: "s3:backup/docs2"},
		},
	})

	engine := NewEngine(cfg, mockExec, nil)
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory")
}

func TestCreateAllDirectories_NoDestinations(t *testing.T) {
	mockExec := &mockMkdirExecutor{}

	cfg := config.NewConfig()

	engine := NewEngine(cfg, mockExec, nil)
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	assert.Error(t, err)
}

func TestExtractDestinationPathFromTo(t *testing.T) {
	tests := []struct {
		to   string
		want string
	}{
		{
			to:   "gdrive:docs",
			want: "docs",
		},
		{
			to:   "/local/path",
			want: "/local/path",
		},
		{
			to:   "s3:bucket/path/to/file",
			want: "bucket/path/to/file",
		},
		{
			to:   "invalid format without colon",
			want: "invalid format without colon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.to, func(t *testing.T) {
			got := ExtractDestinationPathFromTo(tt.to)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateDestinationPaths_Valid(t *testing.T) {
	targets := []*SyncTarget{
		{
			Provider:    "gdrive",
			SourcePath:  "docs",
			Destination: config.Destination{To: "s3:backup/docs"},
		},
		{
			Provider:    "local",
			SourcePath:  "/data",
			Destination: config.Destination{To: "/backup/data"},
		},
	}

	engine := NewEngine(nil, nil, nil)
	err := engine.ValidateDestinationPaths(targets)

	require.NoError(t, err)
}

func TestValidateDestinationPaths_InvalidFormat(t *testing.T) {
	targets := []*SyncTarget{
		{
			Provider:    "gdrive",
			SourcePath:  "docs",
			Destination: config.Destination{To: ":emptyprovider"},
		},
	}

	engine := NewEngine(nil, nil, nil)
	err := engine.ValidateDestinationPaths(targets)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider name cannot be empty")
}

func TestValidateDestinationPaths_EmptyProvider(t *testing.T) {
	targets := []*SyncTarget{
		{
			Provider:    "gdrive",
			SourcePath:  "docs",
			Destination: config.Destination{To: ":path"},
		},
	}

	engine := NewEngine(nil, nil, nil)
	err := engine.ValidateDestinationPaths(targets)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider name cannot be empty")
}

func TestPrepare_Directories(t *testing.T) {
	mockExec := &mockMkdirExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "", Stderr: ""},
		},
	}

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{
			{To: "s3:backup/docs"},
		},
	})

	engine := NewEngine(cfg, mockExec, nil)
	ctx := context.Background()

	err := engine.Prepare(ctx, cfg, SyncOptions{})

	require.NoError(t, err)
}

func TestMapKeys(t *testing.T) {
	m := map[string]struct{}{
		"key1": {},
		"key2": {},
		"key3": {},
	}

	engine := NewEngine(nil, nil, nil)
	keys := engine.mapKeys(m)

	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "key3")
}

func TestMapKeys_Empty(t *testing.T) {
	m := map[string]struct{}{}

	engine := NewEngine(nil, nil, nil)
	keys := engine.mapKeys(m)

	assert.Len(t, keys, 0)
}

func TestCreateAllDirectories_DryRunViaEngine(t *testing.T) {
	mockExec := &mockMkdirExecutor{}

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{
			{To: "s3:backup/docs"},
		},
	})

	engine := NewEngine(cfg, mockExec, nil)
	engine.SetDryRun(true)
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	require.NoError(t, err)
}

func TestCreateAllDirectories_CreatesSourceDirectories(t *testing.T) {
	mockExec := &mockMkdirExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "", Stderr: ""},
			{ExitCode: 0, Stdout: "", Stderr: ""},
		},
	}

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"source_folder": []config.Destination{
			{To: "s3:backup/source_folder"},
		},
	})

	engine := NewEngine(cfg, mockExec, nil)
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	require.NoError(t, err)
}

func TestCreateAllDirectories_CreatesLocalSource(t *testing.T) {
	mockExec := &mockMkdirExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "", Stderr: ""},
			{ExitCode: 0, Stdout: "", Stderr: ""},
		},
	}

	cfg := config.NewConfig()
	cfg.AddProvider("local", config.PathMap{
		"local_folder": []config.Destination{
			{To: "gdrive:backup/local_folder"},
		},
	})

	engine := NewEngine(cfg, mockExec, nil)
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	require.NoError(t, err)
	require.Equal(t, 2, mockExec.index, "should call mkdir for both local source and destination")
}

func TestCreateAllDirectories_BothSourceAndDestination(t *testing.T) {
	mockExec := &mockMkdirExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "", Stderr: ""},
			{ExitCode: 0, Stdout: "", Stderr: ""},
		},
	}

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"source_folder": []config.Destination{
			{To: "s3:backup/source_folder"},
		},
	})

	engine := NewEngine(cfg, mockExec, nil)
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	require.NoError(t, err)
	require.Equal(t, 2, mockExec.index, "should call mkdir for both source and destination")
}
