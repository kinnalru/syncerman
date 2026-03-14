package logger

import (
	"fmt"
	"io"
	"os"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelQuiet
)

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

type ConsoleLogger struct {
	level   LogLevel
	output  io.Writer
	quiet   bool
	verbose bool
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		level:  LevelInfo,
		output: os.Stderr,
	}
}

func (l *ConsoleLogger) format(level LogLevel, msg string, args ...interface{}) string {
	formatted := msg
	if len(args) > 0 {
		formatted = fmt.Sprintf(msg, args...)
	}
	return fmt.Sprintf("[%s] %s\n", level.String(), formatted)
}

func (l *ConsoleLogger) Debug(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelDebug {
		return
	}
	fmt.Fprint(l.output, l.format(LevelDebug, msg, args...))
}

func (l *ConsoleLogger) Info(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelInfo {
		return
	}
	fmt.Fprint(l.output, l.format(LevelInfo, msg, args...))
}

func (l *ConsoleLogger) Warn(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelWarn {
		return
	}
	fmt.Fprint(l.output, l.format(LevelWarn, msg, args...))
}

func (l *ConsoleLogger) Error(msg string, args ...interface{}) {
	if l.quiet || l.level > LevelError {
		return
	}
	fmt.Fprint(l.output, l.format(LevelError, msg, args...))
}

func (l *ConsoleLogger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *ConsoleLogger) SetOutput(w io.Writer) {
	l.output = w
}

func (l *ConsoleLogger) SetVerbose(verbose bool) {
	l.verbose = verbose
	if verbose {
		l.level = LevelDebug
	}
}

func (l *ConsoleLogger) SetQuiet(quiet bool) {
	l.quiet = quiet
	if quiet {
		l.level = LevelQuiet
	}
}

func (l *ConsoleLogger) GetLevel() LogLevel {
	return l.level
}
