package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

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

const (
	logPrefix       = "["
	logSuffix       = "] "
	nilWriterErrMsg = "output writer cannot be nil"

	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

var levelStrings = [...]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelQuiet: "QUIET",
}

// String converts LogLevel to string representation.
// Returns string representation (DEBUG, INFO, WARN, ERROR, QUIET, UNKNOWN).
func (l LogLevel) String() string {
	if int(l) >= 0 && int(l) < len(levelStrings) {
		return levelStrings[l]
	}
	return "UNKNOWN"
}

// Logger defines the logging interface for all logger implementations.
// This interface focuses only on logging operations without configuration concerns.
// Implementations can be swapped out to change logging behavior without modifying
// application code.
//
// Methods:
//
//   - Debug: logs debug-level messages (most verbose)
//   - Info: logs info-level messages (default level)
//   - Warn: logs warning-level messages for non-critical issues
//   - Error: logs error-level messages for critical issues
//   - Command: logs command execution with cyan color at INFO level
//   - CombinedOutput: logs combined stdout/stderr output at INFO level with distinct color
//   - Output: logs multi-line output blocks (deprecated, use CombinedOutput)
//   - ErrorOutput: logs error output blocks with red color (deprecated, use CombinedOutput)
//   - StageInfo: logs stage messages with bold highlighting
//   - TargetInfo: logs target messages with normal brightness
//
// Design: Small, focused interface following Go best practices. Configuration
// operations are separated into the Configurable interface.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Command(cmd string)
	CombinedOutput(output string)
	Output(output string)
	ErrorOutput(output string)
	StageInfo(msg string, args ...interface{})
	TargetInfo(msg string, args ...interface{})
}

// Configurable defines configuration methods for logger implementations.
// This interface allows dynamic configuration of log level and output destination.
//
// Methods:
//
//   - SetLevel: changes the minimum log level (messages below this level are filtered)
//   - SetOutput: changes the output destination (e.g., file, stderr)
//   - GetLevel: returns the current minimum log level
//   - SetVerbose: enables verbose mode by setting level to DEBUG
//   - SetQuiet: enables quiet mode by setting level to QUIET (suppresses all output)
//
// Design: Separate interface for configuration concerns, following interface
// segregation principle. Consumers that only need logging can depend on Logger,
// while those needing configuration can depend on both Logger and Configurable.
type Configurable interface {
	SetLevel(level LogLevel)
	SetOutput(w io.Writer) error
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
	level         LogLevel
	output        io.Writer
	quiet         bool
	verbose       bool
	previousLevel LogLevel
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
//   - Previous level initialized to LevelInfo - used for restoring level after mode changes
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
		level:         LevelInfo,
		previousLevel: LevelInfo,
		output:        os.Stderr,
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
//   - levelStr: the string representation of the log level for prefix (e.g., "DEBUG", "INFO")
//   - msg: message format string (Printf-style format specifiers like %s, %d, %v)
//   - args: variable arguments for format string placeholders (optional)
//
// Implementation details:
//
//   - Uses sync.Pool to reuse buffers and reduce allocations
//   - If args is empty, the msg is used as-is without formatting
//   - If args is provided, fmt.Fprintf is used to apply Printf-style formatting
//   - The final format is "[LEVEL] message\n"
//   - After use, buffer is returned to the pool
//
// Example:
//
//	format("INFO", "Processing %d files", numFiles)
//	// Writes: "[INFO] Processing 10 files\n"
//
//	format("ERROR", "Failed: %v", err)
//	// Writes: "[ERROR] Failed: connection timeout\n"
func (l *ConsoleLogger) format(levelStr string, msg string, args ...interface{}) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.WriteString(logPrefix)
	buf.WriteString(levelStr)
	buf.WriteString(logSuffix)

	if len(args) > 0 {
		fmt.Fprintf(buf, msg, args...)
	} else {
		buf.WriteString(msg)
	}
	buf.WriteByte('\n')

	_, _ = buf.WriteTo(l.output)
}

func (l *ConsoleLogger) formatBlock(color, title string, lines []string) {
	if l.quiet || l.level > LevelDebug {
		return
	}
	if len(lines) == 0 {
		return
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	if title != "" {
		buf.WriteString(color)
		buf.WriteString(title)
		buf.WriteString(colorReset)
		buf.WriteByte('\n')
	}

	for _, line := range lines {
		if line != "" {
			buf.WriteString(colorGray)
			buf.WriteString("  ")
			buf.WriteString(colorReset)
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	buf.WriteByte('\n')
	_, _ = buf.WriteTo(l.output)
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
	l.format(levelStrings[LevelDebug], msg, args...)
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
	l.format(levelStrings[LevelInfo], msg, args...)
}

func (l *ConsoleLogger) StageInfo(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelInfo {
		return
	}
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.WriteString(logPrefix)
	buf.WriteString(levelStrings[LevelInfo])
	buf.WriteString(logSuffix)
	buf.WriteString(colorBold)

	if len(args) > 0 {
		fmt.Fprintf(buf, msg, args...)
	} else {
		buf.WriteString(msg)
	}

	buf.WriteString(colorReset)
	buf.WriteByte('\n')

	_, _ = buf.WriteTo(l.output)
}

func (l *ConsoleLogger) TargetInfo(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelInfo {
		return
	}
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.WriteString(logPrefix)
	buf.WriteString(levelStrings[LevelInfo])
	buf.WriteString(logSuffix)

	if len(args) > 0 {
		fmt.Fprintf(buf, msg, args...)
	} else {
		buf.WriteString(msg)
	}

	buf.WriteByte('\n')

	_, _ = buf.WriteTo(l.output)
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
	l.format(levelStrings[LevelWarn], msg, args...)
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
	l.format(levelStrings[LevelError], msg, args...)
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
// Returns:
//
//   - error: nil on success, error if the writer is nil
//
// Implementation details:
//
//   - Any io.Writer can be used (os.Stdout, os.Stderr, files, buffers, etc.)
//   - The writer is used directly without additional buffering
//   - Thread safety depends on the underlying writer's implementation
//   - Writer is validated to ensure it is not nil before accepting it
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
//	if err := log.SetOutput(logFile); err != nil {
//		panic(err)
//	}
//
//	// Redirect to stdout
//	if err := log.SetOutput(os.Stdout); err != nil {
//		panic(err)
//	}
//
// Note: It is the caller's responsibility to close the writer when it's no longer needed
// (e.g., closing files, flushing buffers).
func (l *ConsoleLogger) SetOutput(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("%s", nilWriterErrMsg)
	}
	l.output = w
	return nil
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
//   - When verbose is false: restores the previous log level that was saved before
//     enabling verbose mode (to LevelInfo if no previous level was saved)
//   - The verbose flag is stored in the l.verbose field
//   - Enabling verbose mode also clears the quiet flag (quiet and verbose are mutually exclusive)
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
//	// Disable verbose mode (restores previous level)
//	log.SetVerbose(false)
func (l *ConsoleLogger) SetVerbose(verbose bool) {
	l.verbose = verbose
	if verbose {
		if l.level != LevelDebug {
			l.previousLevel = l.level
		}
		l.level = LevelDebug
		l.quiet = false
	} else {
		if l.level == LevelDebug && l.previousLevel != LevelQuiet {
			l.level = l.previousLevel
		}
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
//   - When quiet is true: saves the current level as previous level and sets level to LevelQuiet
//   - When quiet is false: restores the previous log level that was saved before
//     enabling quiet mode (to LevelInfo if no previous level was saved)
//   - The quiet flag is checked before writing any log message in all logging methods
//   - Enabling quiet mode also clears the verbose flag (quiet and verbose are mutually exclusive)
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
//	// Disable quiet mode (restores previous level)
//	log.SetQuiet(false)
//	log.Info("Logging is now restored")
func (l *ConsoleLogger) SetQuiet(quiet bool) {
	l.quiet = quiet
	if quiet {
		if l.level != LevelQuiet {
			l.previousLevel = l.level
		}
		l.level = LevelQuiet
		l.verbose = false
	} else {
		if l.level == LevelQuiet && l.previousLevel != LevelQuiet {
			l.level = l.previousLevel
		}
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
// by SetLevel(), SetVerbose(), or SetQuiet() calls. When SetVerbose(true) or SetQuiet(true)
// are called, the previous level is stored internally and restored when the mode is disabled.
// The returned value does not directly reflect the quiet or verbose flags, but their
// effects on the log level.
func (l *ConsoleLogger) GetLevel() LogLevel {
	return l.level
}

// GetPreviousLevel returns the saved previous log level.
//
// This method allows inspection of the previously saved log level, which is used
// internally by SetVerbose() and SetQuiet() to restore the level when those modes
// are disabled. This is primarily useful for testing purposes.
//
// Returns: LogLevel - the previously stored log level value.
func (l *ConsoleLogger) GetPreviousLevel() LogLevel {
	return l.previousLevel
}

func (l *ConsoleLogger) InfoBlock(title string, lines []string) {
	if l.quiet || l.level > LevelInfo {
		return
	}
	l.formatBlock("", title, lines)
}

func (l *ConsoleLogger) DebugBlock(title string, lines []string) {
	if l.quiet || l.level > LevelDebug {
		return
	}
	l.formatBlock("", title, lines)
}

func (l *ConsoleLogger) Command(cmd string) {
	if l.quiet || l.level > LevelInfo {
		return
	}
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.WriteString(colorCyan)
	buf.WriteString("→ ")
	buf.WriteString(cmd)
	buf.WriteString(colorReset)
	buf.WriteByte('\n')

	_, _ = buf.WriteTo(l.output)
}

func (l *ConsoleLogger) Output(output string) {
	if l.quiet || l.level > LevelDebug || output == "" {
		return
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	l.formatBlock(colorGray, "", lines)
}

func (l *ConsoleLogger) CombinedOutput(output string) {
	if l.quiet || l.level > LevelInfo || output == "" {
		return
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "Elapsed time:") || strings.Contains(line, "Checks:") {
			continue
		}
		filtered = append(filtered, line)
	}
	l.formatBlock(colorGreen, "", filtered)
}

func (l *ConsoleLogger) ErrorOutput(output string) {
	if l.quiet || l.level > LevelDebug || output == "" {
		return
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	l.formatBlock(colorRed, "", lines)
}
