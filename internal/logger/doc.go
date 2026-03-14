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
//	logger.SetVerbose(true)  // Enables all debug output
//
// Quiet Mode
//
// The quiet flag (--quiet, -q) suppresses all standard log output. When quiet mode is
// enabled, the logger level is set to QUIET and no messages are written except those
// that may bypass the logging system. This mode is useful for scripts where only exit
// codes matter.
//
// Usage:
//	logger.SetQuiet(true)  // Suppresses all log output
//
// Dynamic Level Setting
//
// The logger level can be set directly for fine-grained control:
//
//	logger.SetLevel(LevelInfo)   // Show INFO, WARN, ERROR
//	logger.SetLevel(LevelWarn)   // Show WARN, ERROR only
//	logger.SetLevel(LevelQuiet)  // Suppress all output
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
//	logger.SetOutput(os.Stdout)
//	logger.SetOutput(logFile)
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
//	    SetLevel(level LogLevel)
//	    SetOutput(w io.Writer)
//	    GetLevel() LogLevel
//	    SetVerbose(verbose bool)
//	    SetQuiet(quiet bool)
//	}
//
// ConsoleLogger Implementation
//
// ConsoleLogger is the default implementation of the Logger interface. It provides
// structured logging to console output with the standard "[LEVEL] message" format.
//
// Basic Usage:
//
//	log := logger.NewConsoleLogger()
//	log.Info("Starting sync operation")
//	log.Debug("Connecting to remote endpoint")
//	log.Warn("Rate limit approaching")
//	log.Error("Connection failed")
//
// Configuring Verbosity:
//
//	log := logger.NewConsoleLogger()
//	log.SetVerbose(true)  // Enable debug output
//	log.SetQuiet(true)    // Suppress all output
//
// Example Programmatic Configuration:
//
//	func initLogger(verbose, quiet bool) *logger.ConsoleLogger {
//	    log := logger.NewConsoleLogger()
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
//	SetLevel(LevelDebug) → Show: DEBUG, INFO, WARN, ERROR
//	SetLevel(LevelInfo)  → Show: INFO, WARN, ERROR
//	SetLevel(LevelWarn)  → Show: WARN, ERROR
//	SetLevel(LevelError) → Show: ERROR
//	SetLevel(LevelQuiet) → Show: (nothing)
//
// Integration with CLI Flags
//
// In typical CLI usage, the logger is configured based on command-line flags:
//
//   --verbose: Equivalent to SetLevel(LevelDebug)
//   --quiet:   Equivalent to SetLevel(LevelQuiet)
//
// When both flags are specified, verbose takes precedence.
