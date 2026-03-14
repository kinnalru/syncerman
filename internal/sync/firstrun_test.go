package sync

import (
	"context"
	"testing"

	"syncerman/internal/rclone"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandle_Success(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	handler := NewFirstRunHandler(1, &mockLogger{})
	ctx := context.Background()
	args := rclone.NewBisyncArgs("local:src", "gdrive:dst", nil)

	result, _, err := handler.Handle(ctx, mockExec, args)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
}

func TestHandle_FirstRunRetry(t *testing.T) {
	firstRunError := "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames\n"

	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 1, Stdout: "", Stderr: firstRunError},
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	handler := NewFirstRunHandler(1, log)
	ctx := context.Background()
	args := rclone.NewBisyncArgs("local:src", "gdrive:dst", nil)

	result, retries, err := handler.Handle(ctx, mockExec, args)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, 1, retries)
	assert.Len(t, log.warn, 1)
}

func TestHandle_MaxRetriesExceeded(t *testing.T) {
	firstRunError := "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames\n"

	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 1, Stdout: "", Stderr: firstRunError},
			{ExitCode: 1, Stdout: "", Stderr: firstRunError},
		},
	}

	log := &mockLogger{}
	handler := NewFirstRunHandler(1, log)
	ctx := context.Background()
	args := rclone.NewBisyncArgs("local:src", "gdrive:dst", nil)

	result, _, err := handler.Handle(ctx, mockExec, args)

	require.Error(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, err.Error(), "first-run error after 2 retries")
	assert.Len(t, log.errorLog, 1)
}

func TestHandle_NonFirstRunError(t *testing.T) {
	genericError := "ERROR: permission denied\n"

	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 1, Stdout: "", Stderr: genericError},
		},
	}

	handler := NewFirstRunHandler(1, &mockLogger{})
	ctx := context.Background()
	args := rclone.NewBisyncArgs("local:src", "gdrive:dst", nil)

	result, _, err := handler.Handle(ctx, mockExec, args)

	require.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ExitCode)
	assert.Contains(t, err.Error(), "sync failed")
}

type errorMockExecutor struct {
	error error
}

func (e *errorMockExecutor) Run(ctx context.Context, args ...string) (*rclone.Result, error) {
	return nil, e.error
}

func TestHandle_CommandError(t *testing.T) {
	mockExec := &errorMockExecutor{error: assert.AnError}

	handler := NewFirstRunHandler(1, &mockLogger{})
	ctx := context.Background()
	args := rclone.NewBisyncArgs("local:src", "gdrive:dst", nil)

	result, _, err := handler.Handle(ctx, mockExec, args)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "command failed")
}

func TestShouldRetry(t *testing.T) {
	handler := NewFirstRunHandler(1, &mockLogger{})

	firstRunError := "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames\n"
	genericError := "ERROR: permission denied\n"

	assert.True(t, handler.ShouldRetry(firstRunError))
	assert.False(t, handler.ShouldRetry(genericError))
}

func TestExtractFirstRunError(t *testing.T) {
	firstRunError := "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames\n"
	genericError := "ERROR: permission denied\n"

	t.Run("first run error", func(t *testing.T) {
		err := ExtractFirstRunError(firstRunError)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "first-run")
	})

	t.Run("generic error", func(t *testing.T) {
		err := ExtractFirstRunError(genericError)
		require.NoError(t, err)
	})
}

func TestIsFirstRunSyncError(t *testing.T) {
	firstRunResult := &rclone.Result{
		ExitCode: 1,
		Stdout:   "",
		Stderr:   "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames\n",
	}

	genericResult := &rclone.Result{
		ExitCode: 1,
		Stdout:   "",
		Stderr:   "ERROR: permission denied\n",
	}

	t.Run("first run result", func(t *testing.T) {
		assert.True(t, IsFirstRunSyncError(firstRunResult))
	})

	t.Run("generic result", func(t *testing.T) {
		assert.False(t, IsFirstRunSyncError(genericResult))
	})

	t.Run("nil result", func(t *testing.T) {
		assert.False(t, IsFirstRunSyncError(nil))
	})
}

func TestDefaultFirstRunHandler(t *testing.T) {
	log := &mockLogger{}
	handler := DefaultFirstRunHandler(log)

	assert.Equal(t, 1, handler.maxRetries)
	assert.Same(t, log, handler.logger)
}

func TestHandle_WithZeroMaxRetries(t *testing.T) {
	handler := NewFirstRunHandler(0, &mockLogger{})

	firstRunError := "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames\n"

	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 1, Stdout: "", Stderr: firstRunError},
		},
	}

	ctx := context.Background()
	args := rclone.NewBisyncArgs("local:src", "gdrive:dst", nil)

	result, _, err := handler.Handle(ctx, mockExec, args)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "first-run error after 1 retries")
	assert.NotNil(t, result)
}
