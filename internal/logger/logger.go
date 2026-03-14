package logger

import (
	"fmt"
	"io"
	"os"
)

// LogLevel represents log severity levels. Lower values are more verbose.
type LogLevel int

const (
	// LevelDebug detailed debug information, most verbose
	LevelDebug LogLevel = iota
	// LevelInfo general informational messages, default level
	LevelInfo
	// LevelWarn warning messages for non-critical issues
	LevelWarn
	// LevelError error messages for critical issues
	LevelError
	// LevelQuiet suppresses all output except logging errors
	LevelQuiet
)

// String converts LogLevel to string representation.
// Returns string representation (DEBUG, INFO, WARN, ERROR, QUIET, UNKNOWN).
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelQuiet:
		return "QUIET"
	default:
		return "UNKNOWN"
	}
}

// Logger defines the logging interface for all logger implementations.
// This interface allows for flexible logging with support for different severity levels
// and output destinations. Implementations can be swapped out to change logging behavior
// without modifying application code.
//
// Methods:
//
//   - Debug: logs debug-level messages (most verbose)
//   - Info: logs info-level messages (default level)
//   - Warn: logs warning-level messages for non-critical issues
//   - Error: logs error-level messages for critical issues
//   - SetLevel: changes the minimum log level (messages below this level are filtered)
//   - SetOutput: changes the output destination (e.g., file, stderr)
//   - GetLevel: returns the current minimum log level
//   - SetVerbose: enables verbose mode by setting level to DEBUG
//   - SetQuiet: enables quiet mode by setting level to QUIET (suppresses all output)
//
// Design: Interface-based design allows for future logger implementations (e.g., file logger,
// syslog, structured logging) while maintaining a consistent API across the application.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	SetLevel(level LogLevel)
	SetOutput(w io.Writer)
	GetLevel() LogLevel
	SetVerbose(verbose bool)
	SetQuiet(quiet bool)
}

// ConsoleLogger implements the Logger interface for console (stderr) output.
// This logger writes formatted log messages to standard error by default, with support
// for different log levels and output destinations.
//
// Fields:
//
//   - level: current minimum log level; messages below this level are not written
//   - output: io.Writer for log output destination (default: os.Stderr)
//   - quiet: when true, suppresses all output except logging errors
//   - verbose: when true, enables verbose mode (sets level to DEBUG)
//
// Thread Safety: Not thread-safe. Use from a single goroutine or external
// synchronization is required if using from multiple goroutines.
//
// Initialization: Use NewConsoleLogger() factory function to create a new instance
// with default settings (INFO level, stderr output).
type ConsoleLogger struct {
	level   LogLevel
	output  io.Writer
	quiet   bool
	verbose bool
}

// NewConsoleLogger creates and configures a new console logger instance.
//
// This factory function returns an initialized ConsoleLogger with sensible defaults
// for use in console-based applications. The logger writes formatted messages to
// the configured output destination with appropriate log level filtering.
//
// Parameters: none.
//
// Returns: *ConsoleLogger - an initialized logger instance ready for use.
//
// Default behavior:
//
//   - Log level set to LevelInfo (INFO) - only messages at INFO level and above are logged
//   - Output set to os.Stderr - log messages are written to standard error
//   - Quiet mode disabled - logging is enabled
//   - Verbose mode disabled - debug messages are not shown unless level is changed
//
// Usage: typically called at application startup to initialize the logging subsystem.
// After creation, the logger can be configured further using methods like SetLevel(),
// SetOutput(), SetVerbose(), or SetQuiet().
//
// Example:
//
//	log := logger.NewConsoleLogger()
//	log.Info("Application started")
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		level:  LevelInfo,
		output: os.Stderr,
	}
}

// format formats a log message with level prefix and optional argument substitution.
//
// This helper method prepares log messages for output by applying Printf-style formatting
// to variable arguments and prefixing the result with the log level tag. It's used internally
// by all logging methods (Debug, Info, Warn, Error) to ensure consistent message formatting.
//
// Parameters:
//
//   - level: the log level for prefix (determines the tag like [DEBUG], [INFO], etc.)
//   - msg: message format string (Printf-style format specifiers like %s, %d, %v)
//   - args: variable arguments for format string placeholders (optional)
//
// Returns: a formatted string in "[LEVEL] message\n" format, where:
//
//   - LEVEL is the string representation of the log level
//   - message is the formatted message with arguments applied
//   - newline character is appended at the end
//
// Implementation details:
//
//   - If args is empty, the msg is used as-is without formatting
//   - If args is provided, fmt.Sprintf is used to apply Printf-style formatting
//   - The level is converted to its string representation using level.String()
//   - The final format is "[LEVEL] message\n"
//
// Example:
//
//	format(LevelInfo, "Processing %d files", numFiles)
//	// Returns: "[INFO] Processing 10 files\n"
//
//	format(LevelError, "Failed: %v", err)
//	// Returns: "[ERROR] Failed: connection timeout\n"
func (l *ConsoleLogger) format(level LogLevel, msg string, args ...interface{}) string {
	// Format the message with provided arguments using Printf-style formatting
	formatted := msg
	if len(args) > 0 {
		formatted = fmt.Sprintf(msg, args...)
	}
	// Add level prefix and newline for structured output: "[LEVEL] message\n"
	// Note: No timestamp is included in format (logs go to stderr, timestamp comes from system if needed)
	return fmt.Sprintf("[%s] %s\n", level.String(), formatted)
}

// Debug logs debug-level messages with optional formatting.
//
// Use this method for detailed diagnostic information that helps with troubleshooting
// and understanding the internal state of the application. Debug messages are the most
// verbose and should not be used in production unless explicitly needed.
//
// Parameters:
//
//   - msg: message format string (Printf-style format specifiers like %s, %d, %v)
//   - args: variable arguments for format string placeholders (optional)
//
// Behavior:
//
//   - Only writes to output if quiet mode is disabled (l.quiet == false)
//   - Only writes if current log level allows debug messages (l.level <= LevelDebug)
//   - Message is prefixed with [DEBUG] tag and formatted with newline
//
// Usage context: detailed diagnostic information for troubleshooting, variable state inspection,
// step-by-step execution tracing, performance profiling, and development/testing phases.
//
// Example:
//
//	log.Debug("Processing file %s with size %d bytes", filename, fileSize)
//	log.Debug("Cache hit: key=%s, value=%v", cacheKey, cachedValue)
func (l *ConsoleLogger) Debug(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelDebug {
		return
	}
	fmt.Fprint(l.output, l.format(LevelDebug, msg, args...))
}

// Info logs info-level messages with optional formatting.
//
// Use this method for general informational messages about normal application operation.
// This is the default log level and should be used for messages that provide useful
// context about what the application is doing without being overly verbose.
//
// Parameters:
//
//   - msg: message format string (Printf-style format specifiers like %s, %d, %v)
//   - args: variable arguments for format string placeholders (optional)
//
// Behavior:
//
//   - Only writes to output if quiet mode is disabled (l.quiet == false)
//   - Only writes if current log level allows info messages (l.level <= LevelInfo)
//   - Message is prefixed with [INFO] tag and formatted with newline
//
// Usage context: general informational messages about normal operation, state transitions,
// successful completions, configuration details, and progress indicators.
//
// Example:
//
//	log.Info("Sync completed successfully in %s duration", duration)
//	log.Info("Starting sync for source: %s to destination: %s", src, dst)
func (l *ConsoleLogger) Info(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelInfo {
		return
	}
	fmt.Fprint(l.output, l.format(LevelInfo, msg, args...))
}

// Warn logs warning-level messages with optional formatting.
//
// Use this method for non-critical issues that don't stop normal operation but may
// indicate potential problems or suboptimal conditions. Warnings should be used for
// situations where the application can continue running but the issue should be noted.
//
// Parameters:
//
//   - msg: message format string (Printf-style format specifiers like %s, %d, %v)
//   - args: variable arguments for format string placeholders (optional)
//
// Behavior:
//
//   - Only writes to output if quiet mode is disabled (l.quiet == false)
//   - Only writes if current log level allows warning messages (l.level <= LevelWarn)
//   - Message is prefixed with [WARN] tag and formatted with newline
//
// Usage context: non-critical issues that don't stop operation, deprecated feature usage,
// suboptimal resource usage, temporary unavailability of optional components, and recoverable
// errors that are handled gracefully.
//
// Example:
//
//	log.Warn("File skipped: read permission denied for %s", filepath)
//	log.Warn("Disk space low: %d MB remaining", remainingMB)
func (l *ConsoleLogger) Warn(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelWarn {
		return
	}
	fmt.Fprint(l.output, l.format(LevelWarn, msg, args...))
}

// Error logs error-level messages with optional formatting.
//
// Use this method for critical issues that may affect operation but don't necessarily
// require application termination. Error messages indicate problems that need attention
// and may impact functionality or performance.
//
// Parameters:
//
//   - msg: message format string (Printf-style format specifiers like %s, %d, %v)
//   - args: variable arguments for format string placeholders (optional)
//
// Behavior:
//
//   - Only writes to output if quiet mode is disabled (l.quiet == false)
//   - Only writes if current log level allows error messages (l.level <= LevelError)
//   - Message is prefixed with [ERROR] tag and formatted with newline
//
// Usage context: critical issues that may affect operation, failed operations that are
// partially recoverable, network connectivity problems, file system errors, and runtime
// exceptions that are caught but degrade functionality.
//
// Example:
//
//	log.Error("Failed to sync: %v", err)
//	log.Error("Connection lost to remote host %s after %d successful operations", host, completed)
func (l *ConsoleLogger) Error(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelError {
		return
	}
	fmt.Fprint(l.output, l.format(LevelError, msg, args...))
}

// SetLevel sets the minimum log level for this logger instance.
//
// This method controls which log messages are written to the output. Messages with a level
// lower than the minimum level will be suppressed. For example, if the level is set to
// LevelWarn, then Debug and Info messages will not be written, but Warn and Error messages will.
//
// Parameters:
//
//   - level: new minimum log level (one of LevelDebug, LevelInfo, LevelWarn, LevelError, LevelQuiet)
//
// Returns: none.
//
// Effect on existing logging:
//
//   - All messages below this level are immediately suppressed for subsequent logging calls
//   - The change affects all methods (Debug, Info, Warn, Error) that use level checking
//   - Quiet mode behavior is not changed by this method (unless set to LevelQuiet directly)
//
// Use cases: dynamic level changes during runtime for debugging or adjusting verbosity,
// responding to user preferences, transitioning between development and production modes,
// implementing verbose/quiet command-line flags, and temporarily enabling debug output
// for troubleshooting specific issues.
//
// Example:
//
//	log.SetLevel(logger.LevelDebug)
//	log.Debug("This message will now be displayed")
//	log.SetLevel(logger.LevelWarn)
//	log.Info("This message will not be displayed")
//
// Note: This method does not affect the verbose or quiet flags directly, but calling
// SetVerbose(false) or SetQuiet(false) may reset the level based on flag state.
func (l *ConsoleLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput changes the output destination for log messages.
//
// This method allows flexible redirection of log output to different destinations. Any type
// that implements the io.Writer interface can be used, including files, buffers, network
// connections, or composite writers like ioutil.Discard.
//
// Parameters:
//
//   - w: io.Writer for log output destination
//
// Returns: none.
//
// Implementation details:
//
//   - Any io.Writer can be used (os.Stdout, os.Stderr, files, buffers, etc.)
//   - The writer is used directly without additional buffering
//   - Thread safety depends on the underlying writer's implementation
//   - No validation is performed on the writer (nil writer will panic on write)
//
// Use cases: redirecting logs to file for persistent storage, capturing logs in a buffer
// for testing or analysis, writing to network sockets for centralized logging, directing
// output to stdout instead of stderr, and creating custom log destinations for specific
// purposes.
//
// Example:
//
//	// Redirect to a file
//	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
//	if err != nil {
//		panic(err)
//	}
//	log.SetOutput(logFile)
//
//	// Redirect to stdout
//	log.SetOutput(os.Stdout)
//
//	// Redirect to ioutil.Discard to suppress all output
//	log.SetOutput(ioutil.Discard)
//
// Note: It is the caller's responsibility to close the writer when it's no longer needed
// (e.g., closing files, flushing buffers).
func (l *ConsoleLogger) SetOutput(w io.Writer) {
	l.output = w
}

// SetVerbose enables or disables verbose mode for detailed output.
//
// This method provides a convenient way to enable maximum verbosity for debugging and
// troubleshooting purposes. When verbose mode is enabled, all log levels (debug, info,
// warn, error) are written to output.
//
// Parameters:
//
//   - verbose: boolean to enable (true) or disable (false) verbose mode
//
// Returns: none.
//
// Implementation details:
//
//   - When verbose is true: sets the log level to LevelDebug
//   - When verbose is false: does not automatically change the log level (leaves current level)
//   - The verbose flag is stored in the l.verbose field for potential future use
//   - Enabling verbose mode overrides quiet mode (quiet flag is not automatically cleared)
//
// Use cases: implementing command-line --verbose flag, debugging during development,
// troubleshooting production issues, providing detailed output when requested by users,
// and enabling comprehensive logging for log file analysis.
//
// Example:
//
//	// Enable verbose mode (typically from command-line flag)
//	log.SetVerbose(true)
//	log.Debug("Detailed debug information will now be shown")
//	log.Info("Info messages will also be shown")
//
//	// Disable verbose mode (but keep current level)
//	log.SetVerbose(false)
//
// Note: This method is designed for use with command-line interfaces where users
// request verbose output. Use SetLevel() directly if you need finer control over log
// levels or want to set a specific level other than LevelDebug.
func (l *ConsoleLogger) SetVerbose(verbose bool) {
	l.verbose = verbose
	if verbose {
		l.level = LevelDebug
	}
}

// SetQuiet enables or disables quiet mode to suppress output.
//
// This method provides a convenient way to suppress all log output, useful for scripts,
// batch operations, or when output should be minimal. When quiet mode is enabled, no
// log messages (including errors) are written to the output destination.
//
// Parameters:
//
//   - quiet: boolean to enable (true) or disable (false) quiet mode
//
// Returns: none.
//
// Implementation details:
//
//   - When quiet is true: sets the log level to LevelQuiet and sets the quiet flag
//   - When quiet is false: only clears the quiet flag (does not restore previous level)
//   - The quiet flag is checked before writing any log message in all logging methods
//   - Enabling quiet mode overrides verbose mode (verbose flag is not automatically cleared)
//
// Use cases: implementing command-line --quiet flag, running automated scripts that
// should not produce output, reducing noise in production environments, enabling silent
// mode for background services, and suppressing all output for tests.
//
// Example:
//
//	// Enable quiet mode (typically from command-line flag)
//	log.SetQuiet(true)
//	log.Info("This message will not be shown")
//	log.Error("Even errors are suppressed")
//
//	// Disable quiet mode (but level remains at LevelQuiet)
//	log.SetQuiet(false)
//	// You may need to set a specific level after disabling quiet mode
//	log.SetLevel(logger.LevelInfo)
//
// Note: When quiet mode is disabled using SetQuiet(false), the log level remains at
// LevelQuiet by default. You should typically call SetLevel() after disabling quiet mode
// to restore normal logging behavior.
func (l *ConsoleLogger) SetQuiet(quiet bool) {
	l.quiet = quiet
	if quiet {
		l.level = LevelQuiet
	}
}

// GetLevel returns the current minimum log level.
//
// This method allows inspection of the current logging configuration, useful for
// conditional logic based on the log level or for debugging/logging management code.
// The returned value indicates which messages will be written to output.
//
// Parameters: none.
//
// Returns: LogLevel - the current minimum log level being used by this logger instance.
//
// Returned values and their meaning:
//
//   - LevelDebug: all messages are logged (most verbose)
//   - LevelInfo: Info, Warn, and Error messages are logged (default)
//   - LevelWarn: only Warn and Error messages are logged
//   - LevelError: only Error messages are logged
//   - LevelQuiet: no messages are logged (unless quiet flag is manually checked)
//
// Use cases: checking if debug logging is enabled before expensive debug operations,
// conditional code execution based on log level, displaying current log level to users,
// implementing dynamic log level adjustment based on runtime conditions, and saving/
// restoring log level state.
//
// Example:
//
//	currentLevel := log.GetLevel()
//	if currentLevel == logger.LevelDebug {
//		// Only perform expensive debug calculations if debug mode is enabled
//		debugInfo := calculateExpensiveDebugData()
//		log.Debug("Debug info: %v", debugInfo)
//	}
//
// Note: This method returns the current l.level value, which may have been modified
// by SetLevel(), SetVerbose(), or SetQuiet() calls. The returned value does not
// directly reflect the quiet or verbose flags, but their effects on the log level.
func (l *ConsoleLogger) GetLevel() LogLevel {
	return l.level
}
