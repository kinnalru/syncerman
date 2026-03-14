package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestConsoleLogger_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	logger.SetOutput(buf)
	logger.SetVerbose(true)

	logger.Debug("test message")

	if !strings.Contains(buf.String(), "DEBUG") {
		t.Errorf("expected DEBUG in output, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("expected 'test message' in output, got %q", buf.String())
	}
}

func TestConsoleLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	logger.SetOutput(buf)

	logger.Info("test message")

	if !strings.Contains(buf.String(), "INFO") {
		t.Errorf("expected INFO in output, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("expected 'test message' in output, got %q", buf.String())
	}
}

func TestConsoleLogger_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	logger.SetOutput(buf)

	logger.Warn("test warning")

	if !strings.Contains(buf.String(), "WARN") {
		t.Errorf("expected WARN in output, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "test warning") {
		t.Errorf("expected 'test warning' in output, got %q", buf.String())
	}
}

func TestConsoleLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	logger.SetOutput(buf)

	logger.Error("test error")

	if !strings.Contains(buf.String(), "ERROR") {
		t.Errorf("expected ERROR in output, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "test error") {
		t.Errorf("expected 'test error' in output, got %q", buf.String())
	}
}

func TestConsoleLogger_WithArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	logger.SetOutput(buf)

	logger.Info("formatted %s %d", "message", 42)

	if !strings.Contains(buf.String(), "formatted message 42") {
		t.Errorf("expected formatted message, got %q", buf.String())
	}
}

func TestConsoleLogger_SetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	logger.SetOutput(buf)
	logger.SetLevel(LevelError)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Error("error message")

	output := buf.String()
	if strings.Contains(output, "DEBUG") {
		t.Errorf("should not log DEBUG when level is ERROR")
	}
	if strings.Contains(output, "INFO") {
		t.Errorf("should not log INFO when level is ERROR")
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("should log ERROR when level is ERROR")
	}
}

func TestConsoleLogger_Quiet(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	logger.SetOutput(buf)
	logger.SetQuiet(true)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	if buf.String() != "" {
		t.Errorf("quiet mode should produce no output, got %q", buf.String())
	}
}

func TestConsoleLogger_Verbose(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	logger.SetOutput(buf)
	logger.SetVerbose(true)

	logger.Debug("debug message")

	if !strings.Contains(buf.String(), "DEBUG") {
		t.Errorf("verbose mode should allow DEBUG logs, got %q", buf.String())
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelQuiet, "QUIET"},
	}

	for _, tt := range tests {
		if tt.level.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.level.String())
		}
	}
}
