package rclone

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	syncerman_errors "gitlab.com/kinnalru/syncerman/internal/errors"
	"gitlab.com/kinnalru/syncerman/internal/logger"
)

func TestResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		want     bool
	}{
		{
			name:     "successful command",
			exitCode: 0,
			want:     true,
		},
		{
			name:     "failed command",
			exitCode: 1,
			want:     false,
		},
		{
			name:     "high exit code",
			exitCode: 127,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{ExitCode: tt.exitCode}
			if got := r.Success(); got != tt.want {
				t.Errorf("Result.Success() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_Error(t *testing.T) {
	tests := []struct {
		name    string
		result  *Result
		wantErr bool
	}{
		{
			name:    "no error",
			result:  &Result{ExitCode: 0},
			wantErr: false,
		},
		{
			name:    "has error",
			result:  &Result{ExitCode: 1, Stderr: "error message"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Error()
			if (err != nil) != tt.wantErr {
				t.Errorf("Result.Error() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewExecutor(t *testing.T) {
	config := NewConfig()
	exec := NewExecutor(config)

	if exec == nil {
		t.Fatal("NewExecutor() returned nil")
	}

	if impl, ok := exec.(*ExecutorImpl); !ok || impl.config != config {
		t.Error("NewExecutor() did not create expected executor type or config mismatch")
	}
}

func TestNewExecutorWithLogger(t *testing.T) {
	config := NewConfig()
	log := logger.NewConsoleLogger()
	exec := NewExecutorWithLogger(config, log)

	if exec == nil {
		t.Fatal("NewExecutorWithLogger() returned nil")
	}

	if impl, ok := exec.(*ExecutorImpl); !ok || impl.config != config {
		t.Error("NewExecutorWithLogger() did not create expected executor type or config mismatch")
	}
}

func TestExecutorImpl_Run_Success(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := createTestBinary(t, tempDir, "success", 0)

	config := &Config{BinaryPath: binaryPath}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)

	result, err := exec.Run(context.Background(), "test")

	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
		return
	}

	if !result.Success() {
		t.Errorf("Run() result success = false, want true")
	}

	if !strings.Contains(result.Stdout, "success") {
		t.Errorf("Run() stdout = %q, want to contain 'success'", result.Stdout)
	}
}

func TestExecutorImpl_Run_Failure(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := createTestBinaryWithStderr(t, tempDir, "error", 1)

	config := &Config{BinaryPath: binaryPath}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)

	result, err := exec.Run(context.Background(), "test")

	if err == nil {
		t.Error("Run() expected error, got nil")
		return
	}

	if result.Success() {
		t.Errorf("Run() result success = true, want false")
	}

	if !syncerman_errors.IsRcloneError(err) {
		t.Errorf("Run() error type is not RcloneError")
	}
}

func TestExecutorImpl_Run_ContextCancelled(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := createSlowBinary(t, tempDir)

	config := &Config{BinaryPath: binaryPath}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	result, err := exec.Run(ctx, "test")

	if err == nil {
		t.Error("Run() expected error for cancelled context, got nil")
		return
	}

	if ctx.Err() == nil {
		t.Error("Run() context was not cancelled")
	}

	_ = result
}

func TestExecutorImpl_Run_BinaryNotFound(t *testing.T) {
	config := &Config{BinaryPath: "/nonexistent/binary/path"}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)

	_, err := exec.Run(context.Background(), "test")

	if err == nil {
		t.Error("Run() expected error for non-existent binary, got nil")
	}

	if !strings.Contains(err.Error(), "failed to start command") {
		t.Errorf("Run() error = %v, want error about failed to start", err)
	}
}

func TestExecutorImpl_Run_WithArgs(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := createTestBinary(t, tempDir, "Args: arg1 arg2 arg3", 0)

	config := &Config{BinaryPath: binaryPath}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)

	result, err := exec.Run(context.Background(), "arg1", "arg2", "arg3")

	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
		return
	}

	if !strings.Contains(result.Stdout, "arg1 arg2 arg3") {
		t.Errorf("Run() stdout = %q, want to contain 'arg1 arg2 arg3'", result.Stdout)
	}
}

func TestExecutorImpl_Run_Timeout(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := createSlowBinary(t, tempDir)

	config := &Config{BinaryPath: binaryPath}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := exec.Run(ctx, "test")

	if err == nil {
		t.Error("Run() expected error for timeout, got nil")
	}

	if result == nil {
		t.Error("Run() result is nil on timeout")
	}

	_ = result
}

func skipIfNoEcho(t *testing.T) {
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("echo binary not found, skipping test")
	}
}

func TestExecutorImpl_Run_RealEcho(t *testing.T) {
	skipIfNoEcho(t)

	config := &Config{BinaryPath: "echo"}
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LevelQuiet)
	exec := NewExecutorWithLogger(config, log)

	result, err := exec.Run(context.Background(), "test", "message")

	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	if !result.Success() {
		t.Errorf("Run() result success = false, want true")
	}

	expectedOutput := "test message\n"
	if result.Stdout != expectedOutput {
		t.Errorf("Run() stdout = %q, want %q", result.Stdout, expectedOutput)
	}

	if !strings.Contains(result.Combined, "test message") {
		t.Errorf("Run() combined = %q, want to contain 'test message'", result.Combined)
	}
}

func TestExecutorImpl_extractExitCode(t *testing.T) {
	config := NewConfig()
	executor := NewExecutor(config)
	impl, ok := executor.(*ExecutorImpl)
	if !ok {
		t.Fatal("Failed to get ExecutorImpl")
	}

	t.Run("nil ProcessState", func(t *testing.T) {
		cmd := exec.Command("echo", "test")
		exitCode := impl.extractExitCode(cmd)
		if exitCode != 1 {
			t.Errorf("extractExitCode() should return 1 for nil ProcessState, got %d", exitCode)
		}
	})
}
