package rclone

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	syncerman_errors "syncerman/internal/errors"
	"syncerman/internal/logger"
)

// Config holds rclone configuration options.
type Config struct {
	BinaryPath string `json:"binaryPath" yaml:"binaryPath"`
}

// NewConfig creates a new Config with default settings.
func NewConfig() *Config {
	return &Config{
		BinaryPath: "rclone",
	}
}

// Result represents the output of an rclone command execution.
type Result struct {
	ExitCode int    `json:"exitCode" yaml:"exitCode"`
	Stdout   string `json:"stdout" yaml:"stdout"`
	Stderr   string `json:"stderr" yaml:"stderr"`
	Combined string `json:"combined" yaml:"combined"`
}

// Success returns true if the command exited with code 0.
func (r *Result) Success() bool {
	return r.ExitCode == 0
}

// Error returns an error if the command failed, nil otherwise.
func (r *Result) Error() error {
	if r.Success() {
		return nil
	}
	return fmt.Errorf("rclone command failed with exit code %d: %s", r.ExitCode, r.Stderr)
}

// Remote represents a configured rclone remote.
type Remote struct {
	Name   string            `json:"name" yaml:"name"`
	Config map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
}

// Executor defines the interface for executing rclone commands.
type Executor interface {
	Run(ctx context.Context, args ...string) (*Result, error)
}

// ExecutorImpl is the default implementation of Executor.
type ExecutorImpl struct {
	config *Config
	logger logger.Logger
}

// NewExecutor creates a new Executor with the given config.
func NewExecutor(config *Config) Executor {
	return &ExecutorImpl{
		config: config,
		logger: logger.NewConsoleLogger(),
	}
}

// NewExecutorWithLogger creates a new Executor with the given config and logger.
func NewExecutorWithLogger(config *Config, log logger.Logger) Executor {
	return &ExecutorImpl{
		config: config,
		logger: log,
	}
}

// Run executes an rclone command with the given arguments.
func (e *ExecutorImpl) Run(ctx context.Context, args ...string) (*Result, error) {
	if e.logger != nil {
		e.logger.Debug("Executing rclone command: %s %s", e.config.BinaryPath, strings.Join(args, " "))
	}

	cmd := exec.CommandContext(ctx, e.config.BinaryPath, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		if ctx.Err() != nil {
			return nil, syncerman_errors.NewRcloneError("command cancelled by context", ctx.Err())
		}
		return nil, syncerman_errors.NewRcloneError("failed to start command", err)
	}

	if err := cmd.Wait(); err != nil {
		exitCode := 1
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
			if exitCode == -1 {
				exitCode = 1
			}
		}

		stdoutStr := stdoutBuf.String()
		stderrStr := stderrBuf.String()

		if ctx.Err() != nil {
			return &Result{
				ExitCode: exitCode,
				Stdout:   stdoutStr,
				Stderr:   stderrStr,
				Combined: stdoutStr + stderrStr,
			}, syncerman_errors.NewRcloneError("command cancelled by context", ctx.Err())
		}

		return &Result{
			ExitCode: exitCode,
			Stdout:   stdoutStr,
			Stderr:   stderrStr,
			Combined: stdoutStr + stderrStr,
		}, syncerman_errors.NewRcloneError("rclone command failed", err)
	}

	return &Result{
		ExitCode: 0,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Combined: stdoutBuf.String() + stderrBuf.String(),
	}, nil
}
