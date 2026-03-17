package rclone

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	syncerman_errors "gitlab.com/kinnalru/syncerman/internal/errors"
	"gitlab.com/kinnalru/syncerman/internal/logger"
)

// Config holds rclone command configuration for executing rclone commands.
// It provides the path to the rclone binary that will be used for command execution.
// Binary path discovery and verification is handled by FindRcloneBinary in binary.go.
// An empty Config defaults to using "rclone" as the binary name.
type Config struct {
	BinaryPath string `json:"binaryPath" yaml:"binaryPath"`
}

// NewConfig creates a new Config with default settings for rclone command execution.
// It initializes the BinaryPath field to "rclone", which is the default binary name.
//
// Returns:
//   - *Config: A pointer to a new Config instance with default binary path set to "rclone"
//
// Usage:
//
//	Typically used at application startup to create a basic configuration.
//	For environment-based configuration discovery, use ConfigFromEnv instead.
func NewConfig() *Config {
	return &Config{
		BinaryPath: "rclone",
	}
}

// Result represents the output of an rclone command execution.
// It captures the exit status and output streams for analyzing command execution results.
//
// Fields:
//   - ExitCode: The command exit code (0 indicates success, non-zero indicates failure)
//   - Stdout:   Standard output from the command execution
//   - Stderr:   Standard error output from the command execution
//   - Combined: Concatenation of stdout and stderr for full output access
//
// Usage:
//
//	Use this struct to analyze command execution results, check for success,
//	and inspect output for debugging or parsing command responses.
type Result struct {
	ExitCode int    `json:"exitCode" yaml:"exitCode"`
	Stdout   string `json:"stdout" yaml:"stdout"`
	Stderr   string `json:"stderr" yaml:"stderr"`
	Combined string `json:"combined" yaml:"combined"`
}

// Success returns true if the command exited with code 0.
//
// Returns:
//   - bool: true if ExitCode is 0, false otherwise
func (r *Result) Success() bool {
	return r.ExitCode == 0
}

// Error returns an error if the command failed, nil otherwise.
// If the command failed (non-zero exit code), it returns an error containing
// the exit code and stderr output for debugging.
//
// Returns:
//   - error: nil if command succeeded, otherwise an error with exit code and stderr
func (r *Result) Error() error {
	if r.Success() {
		return nil
	}
	return fmt.Errorf("rclone command failed with exit code %d: %s", r.ExitCode, r.Stderr)
}

// Executor defines the interface for executing rclone commands.
//
// Purpose:
//
//	Provides an abstraction layer for executing rclone commands, allowing for testing
//	and mock implementations. This interface enables dependency injection and makes
//	the codebase more testable and flexible.
//
// Method:
//
//	Run - Executes an rclone command with the provided arguments and returns the result
//
// Parameters:
//
//	ctx - Context for controlling cancellation and timeouts
//	args - Variable number of string arguments to pass to the rclone command
//
// Returns:
//
//	*Result - The execution result containing exit code, stdout, stderr, and combined output
//	error - An error if the execution fails, nil otherwise
//
// Design:
//
//	This interface allows for easy testing with mock implementations and supports
//	different logging strategies through concrete implementations.
type Executor interface {
	Run(ctx context.Context, args ...string) (*Result, error)
}

// ExecutorImpl is the default implementation of Executor.
//
// Purpose:
//
//	Implements the Executor interface for actual rclone command execution using
//	the exec.CommandContext function. This is the production implementation that
//	executes real rclone commands on the system.
//
// Fields:
//
//	config - Rclone configuration containing the binary path to execute
//	logger - Logger instance for recording command execution details and debugging
//
// Design:
//
//	This struct holds the necessary dependencies for executing rclone commands,
//	with the configuration determining the binary path and the logger providing
//	output capture functionality.
type ExecutorImpl struct {
	config *Config
	logger logger.Logger
}

// NewExecutor creates a new Executor with the given config.
//
// Purpose:
//
//	Factory function to create a new Executor instance configured with the
//	provided rclone configuration. This is the recommended way to create an
//	Executor for production use.
//
// Parameters:
//
//	config - Configuration containing the rclone binary path; should include
//	         the full path to the rclone executable or use system path lookup
//
// Returns:
//
//	Executor - A new Executor instance ready for executing rclone commands
//
// Default:
//
//	Uses a new ConsoleLogger instance for logging command execution details
//
// Usage:
//
//	Typically used at application initialization to create the main executor
//	that will be used throughout the application.
func NewExecutor(config *Config) Executor {
	return &ExecutorImpl{
		config: config,
		logger: logger.NewConsoleLogger(),
	}
}

// NewExecutorWithLogger creates a new Executor with the given config and logger.
//
// Purpose:
//
//	Factory function to create a new Executor instance with custom logging support.
//	This allows for dependency injection of a specific logger implementation.
//
// Parameters:
//
//	config - Configuration containing the rclone binary path; should include
//	         the full path to the rclone executable or use system path lookup
//	log    - Custom logger implementation for recording command execution
//
// Returns:
//
//	Executor - A new Executor instance configured with the provided logger
//
// Use case:
//
//	Primary use cases include:
//	- Testing: Inject a mock or test logger for verification
//	- Custom logging: Use file-based, structured, or other specialized logging
//	- Integration testing: Capture and verify command execution in tests
func NewExecutorWithLogger(config *Config, log logger.Logger) Executor {
	return &ExecutorImpl{
		config: config,
		logger: log,
	}
}

// Run executes an rclone command with the given arguments.
//
// Purpose:
//
//	Executes a rclone command using the configured binary path and captures all
//	output streams (stdout, stderr) for analysis. Supports context cancellation
//	for timeout handling and graceful shutdown.
//
// Parameters:
//
//	ctx  - Context for controlling cancellation and timeouts; allows the caller
//	       to cancel long-running operations or enforce time limits
//	args - Variable number of string arguments to pass to the rclone command;
//	       should include the rclone subcommand and any additional flags
//
// Returns:
//
//	*Result - Contains the execution result with exit code, stdout, stderr,
//	          and combined output; ExitCode is 0 for success, non-zero for failure
//	error   - Returns a wrapped error if the command fails to start or is cancelled,
//	          nil if the command executes successfully (even with non-zero exit code)
//
// Implementation details:
//   - Uses exec.CommandContext for full context cancellation support
//   - Captures both stdout and stderr using separate buffers
//   - Logs the command execution if a logger is configured
//   - Returns full output (stdout, stderr, and combined) for debugging
//   - Properly handles exit code extraction from ProcessState
//
// Error handling:
//   - Context cancellation: Returns nil Result with RcloneError containing context error
//   - Command start failure: Returns nil Result with RcloneError for start errors
//   - Command execution failure: Returns Result with non-zero exit code and RcloneError
//   - Successful execution: Returns Result with exit code 0 and nil error
func (e *ExecutorImpl) Run(ctx context.Context, args ...string) (*Result, error) {
	if e.logger != nil {
		e.logger.Command(fmt.Sprintf("%s %s", e.config.BinaryPath, strings.Join(args, " ")))
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
		exitCode := e.extractExitCode(cmd)
		result := e.buildResultFromBuffers(stdoutBuf, stderrBuf)
		result.ExitCode = exitCode

		if ctx.Err() != nil {
			return result, syncerman_errors.NewRcloneError("command cancelled by context", ctx.Err())
		}

		return result, syncerman_errors.NewRcloneError("rclone command failed", err)
	}

	return e.buildResultFromBuffers(stdoutBuf, stderrBuf), nil
}

// buildResultFromBuffers creates a Result from stdout and stderr buffers.
//
// Parameters:
//   - stdoutBuf: buffer containing standard output
//   - stderrBuf: buffer containing standard error
//
// Returns: Result with exit code 0 and combined output
func (e *ExecutorImpl) buildResultFromBuffers(stdoutBuf, stderrBuf bytes.Buffer) *Result {
	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()
	return &Result{
		ExitCode: 0,
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		Combined: stdoutStr + stderrStr,
	}
}

// extractExitCode extracts the exit code from a completed command.
// Handles edge cases where ProcessState might be nil or exit code might be -1.
//
// Parameters:
//   - cmd: the executed command with ProcessState
//
// Returns: extracted exit code (defaults to 1 if unavailable or invalid)
func (e *ExecutorImpl) extractExitCode(cmd *exec.Cmd) int {
	if cmd.ProcessState == nil {
		return 1
	}
	exitCode := cmd.ProcessState.ExitCode()
	if exitCode == -1 {
		return 1
	}
	return exitCode
}
