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

	engine := NewEngine(cfg, mockExec, &mockLogger{})
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

	engine := NewEngine(cfg, mockExec, &mockLogger{})
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

	engine := NewEngine(cfg, mockExec, &mockLogger{})
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory")
}

func TestCreateAllDirectories_NoDestinations(t *testing.T) {
	mockExec := &mockMkdirExecutor{}

	cfg := config.NewConfig()

	engine := NewEngine(cfg, mockExec, &mockLogger{})
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	assert.Error(t, err)
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

	engine := NewEngine(cfg, mockExec, &mockLogger{})
	ctx := context.Background()

	err := engine.Prepare(ctx, cfg, SyncOptions{})

	require.NoError(t, err)
}

func TestCreateAllDirectories_DryRunViaEngine(t *testing.T) {
	mockExec := &mockMkdirExecutor{}

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{
			{To: "s3:backup/docs"},
		},
	})

	engine := NewEngine(cfg, mockExec, &mockLogger{})
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

	engine := NewEngine(cfg, mockExec, &mockLogger{})
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

	engine := NewEngine(cfg, mockExec, &mockLogger{})
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

	engine := NewEngine(cfg, mockExec, &mockLogger{})
	ctx := context.Background()

	err := engine.CreateAllDirectories(ctx, cfg, SyncOptions{})

	require.NoError(t, err)
	require.Equal(t, 2, mockExec.index, "should call mkdir for both source and destination")
}
