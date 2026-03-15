package sync

import (
	"context"
	"testing"
	"time"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	firstRunError := "ERROR: cannot find prior Path1 or Path2 listings...here are filenames\n"

	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 1, Stdout: "", Stderr: firstRunError, Combined: firstRunError},
			{ExitCode: 0, Stdout: "synced", Stderr: "", Combined: "synced"},
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
	containsFirstRunRetry := false
	for _, msg := range log.stage {
		if msg == "Stage: First-run detected, retrying with --resync (attempt %d/%d)" {
			containsFirstRunRetry = true
			break
		}
	}
	assert.True(t, containsFirstRunRetry, "Expected log message about first-run retry missing")
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
			{ExitCode: 1, Stdout: "", Stderr: "permission denied", Combined: "permission denied"},
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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync failed")
	assert.Nil(t, result)
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
			{ExitCode: 0, Stdout: "synced", Stderr: "", Combined: "synced"},
			{ExitCode: 1, Stdout: "", Stderr: "failed", Combined: "failed"},
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

	assert.Error(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.Nil(t, results[1])
	assert.Contains(t, err.Error(), "sync failed for target 2")
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

func TestSetDryRun(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)
	assert.False(t, engine.dryRun)

	engine.SetDryRun(true)
	assert.True(t, engine.dryRun)

	engine.SetDryRun(false)
	assert.False(t, engine.dryRun)
}

func TestRunSync_DryRunViaEngine(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "dry-run complete", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)
	engine.SetDryRun(true)

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
}

func TestRunSync_DryRunOptionsOverridesEngine(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "dry-run complete", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)
	engine.SetDryRun(false)

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
}

func TestRunAll_DryRun(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "dry-run", Stderr: ""},
			{ExitCode: 0, Stdout: "dry-run", Stderr: ""},
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
	results, err := engine.RunAll(ctx, cfg, SyncOptions{DryRun: true, Verbose: false})

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)
}

func TestRunAll_DryRunViaEngine(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "dry-run", Stderr: ""},
			{ExitCode: 0, Stdout: "dry-run", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)
	engine.SetDryRun(true)

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

func TestRunAll_PreservesConfigurationOrder(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced local to gd", Stderr: ""},
			{ExitCode: 0, Stdout: "synced gd to yd", Stderr: ""},
			{ExitCode: 0, Stdout: "synced yd to local2", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	cfg := config.NewConfig()
	cfg.AddProvider("local", config.PathMap{
		"/path/to/local": []config.Destination{
			{To: "gd:syncerman/scenario1/"},
		},
	})
	cfg.AddProvider("gd", config.PathMap{
		"syncerman/scenario1/": []config.Destination{
			{To: "yd:syncerman/scenario1/"},
		},
	})
	cfg.AddProvider("yd", config.PathMap{
		"syncerman/scenario1/": []config.Destination{
			{To: "/path/to/local2"},
		},
	})

	ctx := context.Background()
	results, err := engine.RunAll(ctx, cfg, SyncOptions{Verbose: false})

	require.NoError(t, err)
	assert.Len(t, results, 3)

	expectedSequence := []struct {
		provider string
		path     string
		dest     string
	}{
		{"local", "/path/to/local", "gd:syncerman/scenario1/"},
		{"gd", "syncerman/scenario1/", "yd:syncerman/scenario1/"},
		{"yd", "syncerman/scenario1/", "/path/to/local2"},
	}

	for i, expected := range expectedSequence {
		if i >= len(results) {
			t.Fatalf("Expected result at index %d, but only %d results returned", i, len(results))
		}
		assert.Equal(t, expected.provider, results[i].Target.Provider)
		assert.Equal(t, expected.path, results[i].Target.SourcePath)
		assert.Equal(t, expected.dest, results[i].Target.Destination.To)
	}
}
