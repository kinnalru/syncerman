package sync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"syncerman/internal/config"
	"syncerman/internal/rclone"
)

type mockExecutor struct {
	results []*rclone.Result
	errors  []error
	index   int
}

func (m *mockExecutor) Run(ctx context.Context, args ...string) (*rclone.Result, error) {
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

type mockLogger struct {
	info     []string
	warn     []string
	errorLog []string
	debugLog []string
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.info = append(m.info, msg)
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	m.warn = append(m.warn, msg)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.errorLog = append(m.errorLog, msg)
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.debugLog = append(m.debugLog, msg)
}

func TestRunSync_Success(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	target := SyncTarget{
		Provider:    "gdrive",
		SourcePath:  "docs",
		Destination: config.Destination{To: "s3:backup/docs"},
		Resync:      false,
	}

	ctx := context.Background()
	result, err := engine.RunSync(ctx, target, SyncOptions{Verbose: false})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.False(t, result.FirstRun)
	assert.Equal(t, 0, result.RetryCount)
	assert.Nil(t, result.Error)
}

func TestRunSync_FirstRunError(t *testing.T) {
	firstRunError := "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames\n"

	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 1, Stdout: "", Stderr: firstRunError},
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	target := SyncTarget{
		Provider:    "gdrive",
		SourcePath:  "docs",
		Destination: config.Destination{To: "s3:backup/docs"},
		Resync:      false,
	}

	ctx := context.Background()
	result, err := engine.RunSync(ctx, target, SyncOptions{Verbose: false})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.True(t, result.FirstRun)
	assert.Equal(t, 1, result.RetryCount)
	assert.Nil(t, result.Error)
	assert.Len(t, log.warn, 1)
}

func TestRunSync_DryRun(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "dry-run complete", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	target := SyncTarget{
		Provider:    "gdrive",
		SourcePath:  "docs",
		Destination: config.Destination{To: "s3:backup/docs"},
		Resync:      false,
	}

	ctx := context.Background()
	result, err := engine.RunSync(ctx, target, SyncOptions{DryRun: true, Verbose: false})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.False(t, result.FirstRun)
	assert.Equal(t, 0, result.RetryCount)
}

func TestRunSync_CommandFailure(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 1, Stdout: "", Stderr: "permission denied"},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	target := SyncTarget{
		Provider:    "gdrive",
		SourcePath:  "docs",
		Destination: config.Destination{To: "s3:backup/docs"},
		Resync:      false,
	}

	ctx := context.Background()
	result, err := engine.RunSync(ctx, target, SyncOptions{Verbose: false})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.False(t, result.FirstRun)
	assert.Equal(t, 0, result.RetryCount)
	assert.NotNil(t, result.Error)
}

func TestRunAll_SingleTarget(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{
			{To: "s3:backup/docs"},
		},
	})

	ctx := context.Background()
	results, err := engine.RunAll(ctx, cfg, SyncOptions{Verbose: false})

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.True(t, results[0].Success)
}

func TestRunAll_MultipleTargets(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

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

	ctx := context.Background()
	results, err := engine.RunAll(ctx, cfg, SyncOptions{Verbose: false})

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)
}

func TestRunAll_StopOnError(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
			{ExitCode: 1, Stdout: "", Stderr: "failed"},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

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

	ctx := context.Background()
	results, err := engine.RunAll(ctx, cfg, SyncOptions{Verbose: false})

	require.Error(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.Contains(t, err.Error(), "sync target 2 failed")
}

func TestRunAll_Verbose(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{
			{To: "s3:backup/docs"},
		},
	})

	ctx := context.Background()
	results, err := engine.RunAll(ctx, cfg, SyncOptions{Verbose: true, DryRun: false})

	require.NoError(t, err)
	assert.False(t, results[0].FirstRun)
}

func TestRunSync_ContextCancellation(t *testing.T) {
	mockExec := &mockExecutor{}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	target := SyncTarget{
		Provider:    "gdrive",
		SourcePath:  "docs",
		Destination: config.Destination{To: "s3:backup/docs"},
		Resync:      false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	time.AfterFunc(10*time.Millisecond, func() {
		cancel()
	})

	_, err := engine.RunSync(ctx, target, SyncOptions{Verbose: false})

	if err != nil {
		assert.Contains(t, err.Error(), "command failed")
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

func TestRunSync_CustomArgs(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	target := SyncTarget{
		Provider:   "gdrive",
		SourcePath: "docs",
		Destination: config.Destination{
			To:     "s3:backup/docs",
			Args:   []string{"--exclude", "*.tmp"},
			Resync: false,
		},
		Resync: false,
	}

	ctx := context.Background()
	result, err := engine.RunSync(ctx, target, SyncOptions{Verbose: false})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
}
