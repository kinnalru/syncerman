package logger

// Package logger provides structured logging utilities for Syncerman.
//
// Logging Architecture
//
// The logger package implements a structured logging system with multiple severity levels,
// designed for CLI applications. It provides a flexible logging interface with support for
// different verbosity modes and output destinations.
//
// See guides/OVERALL.md:24 for an overview of the logging system architecture within Syncerman.
//
// Log Levels
//
// The logger supports five log levels, ordered by severity:
//
//	DEBUG (0): Detailed diagnostic information for troubleshooting and development.
//	            Shows granular details about program execution flow and internal state.
//
//	INFO (1):  General informational messages about normal program operation.
//	            Useful for tracking progress and confirming successful operations.
//
//	WARN (2):  Warning messages for potentially harmful situations that do not prevent
//	            program execution. Indicates issues that may need attention.
//
//	ERROR (3): Error messages for serious problems that occurred during execution.
//	            Used for failures, exceptions, and critical issues.
//
//	QUIET (4): Special level that suppresses all log output. When active, only critical
//	            errors that bypass the logging system may be displayed.
//
// Message Filtering
//
// A log message at level L is displayed only if:
//   - The logger's level is <= L (less severe)
//   - The quiet mode is not enabled
//
// For example, with LevelInfo set, DEBUG messages are suppressed while INFO, WARN, and
// ERROR messages are shown.
//
// Configuration Options
//
// Verbose Mode
//
// The verbose flag (--verbose, -v) enables detailed logging by setting the logger level
// to DEBUG. This provides maximum visibility into program execution. When verbose mode
// is active, all log messages (DEBUG, INFO, WARN, ERROR) are output.
//
// Usage:
//	log.SetVerbose(true)  // Enables all debug output
//
// Quiet Mode
//
// The quiet flag (--quiet, -q) suppresses all standard log output. When quiet mode is
// enabled, the logger level is set to QUIET and no messages are written except those
// that may bypass the logging system. This mode is useful for scripts where only exit
// codes matter.
//
// Usage:
//	log.SetQuiet(true)  // Suppresses all log output
//
// Dynamic Level Setting
//
// The logger level can be set directly for fine-grained control:
//
//	log.SetLevel(LevelInfo)   // Show INFO, WARN, ERROR
//	log.SetLevel(LevelWarn)   // Show WARN, ERROR only
//	log.SetLevel(LevelQuiet)  // Suppress all output
//
// Output Format and Destination
//
// Output Format
//
// All log messages follow the format:
//
//	[LEVEL] message
//
// where LEVEL is the uppercase log level name (DEBUG, INFO, WARN, ERROR).
//
// The format supports message formatting with Printf-style arguments:
//
//	logger.Info("Syncing %d files", count)
//	logger.Warn("Retry %d of %d", attempts, maxAttempts)
//
// Output Destination
//
// By default, logs are written to standard error (stderr). This follows CLI convention
// to separate informational output from program results. The output destination can be
// changed to any io.Writer:
//
//	if err := log.SetOutput(os.Stdout); err != nil {
//	    handle error
//	}
//	if err := log.SetOutput(logFile); err != nil {
//	    handle error
//	}
//
// Logger Interface
//
// The Logger interface defines the contract for logger implementations:
//
//	type Logger interface {
//	    Debug(msg string, args ...interface{})
//	    Info(msg string, args ...interface{})
//	    Warn(msg string, args ...interface{})
//	    Error(msg string, args ...interface{})
//	    Command(cmd string)
//	    CombinedOutput(output string)
//	    StageInfo(msg string, args ...interface{})
//	    TargetInfo(msg string, args ...interface{})
//	}
//
// The Configurable interface defines configuration methods:
//
//	type Configurable interface {
//	    SetLevel(level LogLevel)
//	    SetOutput(w io.Writer) error
//	    GetLevel() LogLevel
//	    SetVerbose(verbose bool)
//	    SetQuiet(quiet bool)
//	}
//
// ConsoleLogger Implementation
//
// ConsoleLogger is the default implementation of the Logger interface. It provides
// structured logging to console output with the standard "[LEVEL] message" format.
// ConsoleLogger also implements the Configurable interface for configuration support.
//
// Basic Usage:
//
//	log := NewConsoleLogger()
//	log.Info("Starting sync operation")
//	log.Debug("Connecting to remote endpoint")
//	log.Warn("Rate limit approaching")
//	log.Error("Connection failed")
//
// Specialized Logging Methods:
//
//	log.Command("rclone bisync source: dest:")
//	log.CombinedOutput("Synced 5 files")
//	log.StageInfo("Processing stage 1")
//	log.TargetInfo("Target: gdrive:documents")
//
// Configuring Verbosity:
//
//	log := NewConsoleLogger()
//	log.SetVerbose(true)  // Enable debug output
//	log.SetQuiet(true)    // Suppress all output
//
// Example Programmatic Configuration:
//
//	func initLogger(verbose, quiet bool) *ConsoleLogger {
//	    log := NewConsoleLogger()
//	    if verbose {
//	        log.SetVerbose(true)
//	    } else if quiet {
//	        log.SetQuiet(true)
//	    }
//	    return log
//	}
//
// Log Level Hierarchical Filtering
//
// The logger suppresses messages at lower severity levels when a higher level is set:
//
//	log.SetLevel(LevelDebug) → Show: DEBUG, INFO, WARN, ERROR
//	log.SetLevel(LevelInfo)  → Show: INFO, WARN, ERROR
//	log.SetLevel(LevelWarn)  → Show: WARN, ERROR
//	log.SetLevel(LevelError) → Show: ERROR
//	log.SetLevel(LevelQuiet) → Show: (nothing)
//
// Integration with CLI Flags
//
// In typical CLI usage, the logger is configured based on command-line flags:
//
//   --verbose: Equivalent to SetLevel(LevelDebug)
//   --quiet:   Equivalent to SetLevel(LevelQuiet)
//
// When both flags are specified, verbose takes precedence.
//
// Specialized Logging Methods
//
// The logger provides specialized methods for specific use cases within Syncerman:
//
//	Command(cmd string)
//	    Logs the rclone command being executed with cyan color.
//	    Useful for tracking command execution in verbose mode.
//
//	CombinedOutput(output string)
//	    Logs the combined stdout/stderr output from rclone commands.
//	    Filters out informational lines (e.g., "Elapsed time:", "Checks:").
//	    Displays the filtered output in green color.
//
//	StageInfo(msg string, args ...interface{})
//	    Logs important stage or milestone messages with bold formatting.
//	    Used for highlighting significant synchronization stages.
//
//	TargetInfo(msg string, args ...interface{})
//	    Logs target-specific messages with normal brightness.
//	    Used for standard target information without emphasis.
//
// Color Constants
//
// The logger uses ANSI color codes for console output. Available colors:
//
//	colorReset ("\033[0m")
//	    Resets all color and style attributes to terminal defaults.
//
//	colorRed ("\033[31m")
//	    Red text, used for error indicators (not currently used in log methods).
//
//	colorGreen ("\033[32m")
//	    Green text, used for CombinedOutput to display command results.
//
//	colorGray ("\033[90m")
//	    Gray/dim text, used for indentation in formatted blocks.
//
//	colorCyan ("\033[36m")
//	    Cyan text, used for Command method to highlight command execution.
//
//	colorBold ("\033[1m")
//	    Bold text, used for StageInfo to emphasize important messages.
//
// Note: colorYellow, colorBlue, and colorDim have been removed from the API.
