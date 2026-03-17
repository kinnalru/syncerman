package logger

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConsoleLogger_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
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
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

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
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

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
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

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
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.Info("formatted %s %d", "message", 42)

	if !strings.Contains(buf.String(), "formatted message 42") {
		t.Errorf("expected formatted message, got %q", buf.String())
	}
}

func TestConsoleLogger_SetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
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
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
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
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
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

func TestConsoleLogger_VerboseRestoresLevel(t *testing.T) {
	logger := NewConsoleLogger()
	logger.SetLevel(LevelWarn)

	initialLevel := logger.GetLevel()
	if initialLevel != LevelWarn {
		t.Errorf("expected initial level to be LevelWarn, got %v", initialLevel)
	}

	logger.SetVerbose(true)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("expected level to be LevelDebug when verbose, got %v", logger.GetLevel())
	}

	logger.SetVerbose(false)
	if logger.GetLevel() != LevelWarn {
		t.Errorf("expected level to be restored to LevelWarn, got %v", logger.GetLevel())
	}
}

func TestConsoleLogger_QuietRestoresLevel(t *testing.T) {
	logger := NewConsoleLogger()
	logger.SetLevel(LevelWarn)

	initialLevel := logger.GetLevel()
	if initialLevel != LevelWarn {
		t.Errorf("expected initial level to be LevelWarn, got %v", initialLevel)
	}

	logger.SetQuiet(true)
	if logger.GetLevel() != LevelQuiet {
		t.Errorf("expected level to be LevelQuiet when quiet, got %v", logger.GetLevel())
	}

	logger.SetQuiet(false)
	if logger.GetLevel() != LevelWarn {
		t.Errorf("expected level to be restored to LevelWarn, got %v", logger.GetLevel())
	}
}

func TestConsoleLogger_VerboseQuietMutuallyExclusive(t *testing.T) {
	logger := NewConsoleLogger()

	logger.SetQuiet(true)
	logger.SetVerbose(true)

	if logger.quiet {
		t.Errorf("expected quiet to be false after setting verbose")
	}
	if logger.GetLevel() != LevelDebug {
		t.Errorf("expected level to be LevelDebug, got %v", logger.GetLevel())
	}

	logger.SetQuiet(true)

	if logger.verbose {
		t.Errorf("expected verbose to be false after setting quiet")
	}
	if logger.GetLevel() != LevelQuiet {
		t.Errorf("expected level to be LevelQuiet, got %v", logger.GetLevel())
	}
}

func TestConsoleLogger_VerboseRestoresToDefault(t *testing.T) {
	logger := NewConsoleLogger()

	defaultLevel := logger.GetLevel()
	if defaultLevel != LevelInfo {
		t.Errorf("expected default level to be LevelInfo, got %v", defaultLevel)
	}

	logger.SetVerbose(true)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("expected level to be LevelDebug when verbose, got %v", logger.GetLevel())
	}

	logger.SetVerbose(false)
	if logger.GetLevel() != LevelInfo {
		t.Errorf("expected level to be restored to default LevelInfo, got %v", logger.GetLevel())
	}
}

func TestConsoleLogger_QuietRestoresToDefault(t *testing.T) {
	logger := NewConsoleLogger()

	defaultLevel := logger.GetLevel()
	if defaultLevel != LevelInfo {
		t.Errorf("expected default level to be LevelInfo, got %v", defaultLevel)
	}

	logger.SetQuiet(true)
	if logger.GetLevel() != LevelQuiet {
		t.Errorf("expected level to be LevelQuiet when quiet, got %v", logger.GetLevel())
	}

	logger.SetQuiet(false)
	if logger.GetLevel() != LevelInfo {
		t.Errorf("expected level to be restored to default LevelInfo, got %v", logger.GetLevel())
	}
}
func BenchmarkConsoleLogger_LogWithArgs(b *testing.B) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		b.Fatalf("failed to set output: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test message with value %d", i)
	}
}

func BenchmarkConsoleLogger_LogWithoutArgs(b *testing.B) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		b.Fatalf("failed to set output: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test message without args")
	}
}

func TestConsoleLogger_SetOutputHandlesNilWriter(t *testing.T) {
	logger := NewConsoleLogger()

	err := logger.SetOutput(nil)
	if err == nil {
		t.Errorf("expected error when setting nil output, got nil")
	}

	expectedErrMsg := "output writer cannot be nil"
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("expected error message %q, got %q", expectedErrMsg, err.Error())
	}

	// Verify that the default output is still in place
	buf := &bytes.Buffer{}
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set valid output: %v", err)
	}

	logger.Info("test message")
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("expected 'test message' in output, got %q", buf.String())
	}
}

func TestConsoleLogger_FormatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		args     []interface{}
		expected string
	}{
		{
			name:     "empty message with no args",
			msg:      "",
			args:     nil,
			expected: "[INFO] \n",
		},
		{
			name:     "message with empty args slice",
			msg:      "test",
			args:     []interface{}{},
			expected: "[INFO] test\n",
		},
		{
			name:     "message with special characters",
			msg:      "special: %s %v %%",
			args:     []interface{}{"value", 42},
			expected: "[INFO] special: value 42 %\n",
		},
		{
			name:     "message with nil arg",
			msg:      "value: %v",
			args:     []interface{}{nil},
			expected: "[INFO] value: <nil>\n",
		},
		{
			name:     "message with multiple args",
			msg:      "%s %d %v %t",
			args:     []interface{}{"str", 42, 3.14, true},
			expected: "[INFO] str 42 3.14 true\n",
		},
		{
			name:     "message with large number",
			msg:      "large: %d",
			args:     []interface{}{9999999999},
			expected: "[INFO] large: 9999999999\n",
		},
		{
			name:     "message with floating point",
			msg:      "float: %.2f",
			args:     []interface{}{3.14159},
			expected: "[INFO] float: 3.14\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewConsoleLogger()
			if err := logger.SetOutput(buf); err != nil {
				t.Fatalf("failed to set output: %v", err)
			}

			logger.Info(tt.msg, tt.args...)

			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestConsoleLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name          string
		level         LogLevel
		logDebug      bool
		logInfo       bool
		logWarn       bool
		logError      bool
		logQuiet      bool
		expectedMatch string
	}{
		{
			name:          "LevelDebug logs all",
			level:         LevelDebug,
			logDebug:      true,
			logInfo:       true,
			logWarn:       true,
			logError:      true,
			logQuiet:      true,
			expectedMatch: "DEBUG",
		},
		{
			name:          "LevelInfo logs info and above",
			level:         LevelInfo,
			logDebug:      false,
			logInfo:       true,
			logWarn:       true,
			logError:      true,
			logQuiet:      true,
			expectedMatch: "INFO",
		},
		{
			name:          "LevelWarn logs warn and above",
			level:         LevelWarn,
			logDebug:      false,
			logInfo:       false,
			logWarn:       true,
			logError:      true,
			logQuiet:      true,
			expectedMatch: "WARN",
		},
		{
			name:          "LevelError logs only errors",
			level:         LevelError,
			logDebug:      false,
			logInfo:       false,
			logWarn:       false,
			logError:      true,
			logQuiet:      true,
			expectedMatch: "ERROR",
		},
		{
			name:          "LevelQuiet logs nothing",
			level:         LevelQuiet,
			logDebug:      false,
			logInfo:       false,
			logWarn:       false,
			logError:      false,
			logQuiet:      true,
			expectedMatch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewConsoleLogger()
			if err := logger.SetOutput(buf); err != nil {
				t.Fatalf("failed to set output: %v", err)
			}
			logger.SetLevel(tt.level)

			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")

			output := buf.String()

			if tt.logDebug && !strings.Contains(output, "DEBUG") {
				t.Errorf("expected DEBUG in output")
			}
			if !tt.logDebug && strings.Contains(output, "DEBUG") {
				t.Errorf("did not expect DEBUG in output")
			}
			if tt.logInfo && !strings.Contains(output, "INFO") {
				t.Errorf("expected INFO in output")
			}
			if !tt.logInfo && strings.Contains(output, "INFO") {
				t.Errorf("did not expect INFO in output")
			}
			if tt.logWarn && !strings.Contains(output, "WARN") {
				t.Errorf("expected WARN in output")
			}
			if !tt.logWarn && strings.Contains(output, "WARN") {
				t.Errorf("did not expect WARN in output")
			}
			if tt.logError && !strings.Contains(output, "ERROR") {
				t.Errorf("expected ERROR in output")
			}
			if !tt.logError && strings.Contains(output, "ERROR") {
				t.Errorf("did not expect ERROR in output")
			}
			if !tt.logQuiet && output != "" {
				t.Errorf("expected no output in quiet mode")
			}
		})
	}
}

func TestConsoleLogger_GetLevel(t *testing.T) {
	tests := []struct {
		name            string
		setLevel        LogLevel
		setVerbose      bool
		setQuiet        bool
		expectedLevel   LogLevel
		expectedVerbose bool
		expectedQuiet   bool
	}{
		{
			name:            "default level is Info",
			setLevel:        LevelInfo,
			setVerbose:      false,
			setQuiet:        false,
			expectedLevel:   LevelInfo,
			expectedVerbose: false,
			expectedQuiet:   false,
		},
		{
			name:            "set level to Debug",
			setLevel:        LevelDebug,
			setVerbose:      false,
			setQuiet:        false,
			expectedLevel:   LevelDebug,
			expectedVerbose: false,
			expectedQuiet:   false,
		},
		{
			name:            "set level to Debug then Verbose",
			setLevel:        LevelDebug,
			setVerbose:      true,
			setQuiet:        false,
			expectedLevel:   LevelDebug,
			expectedVerbose: true,
			expectedQuiet:   false,
		},
		{
			name:            "set level to Warn then Verbose",
			setLevel:        LevelWarn,
			setVerbose:      true,
			setQuiet:        false,
			expectedLevel:   LevelDebug,
			expectedVerbose: true,
			expectedQuiet:   false,
		},
		{
			name:            "set level to Warn then Quiet",
			setLevel:        LevelWarn,
			setVerbose:      false,
			setQuiet:        true,
			expectedLevel:   LevelQuiet,
			expectedVerbose: false,
			expectedQuiet:   true,
		},
		{
			name:            "set Quiet then Verbose",
			setLevel:        LevelInfo,
			setVerbose:      true,
			setQuiet:        true,
			expectedLevel:   LevelDebug,
			expectedVerbose: true,
			expectedQuiet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewConsoleLogger()
			logger.SetLevel(tt.setLevel)

			if tt.setQuiet {
				logger.SetQuiet(true)
			}
			if tt.setVerbose {
				logger.SetVerbose(true)
			}

			level := logger.GetLevel()
			if level != tt.expectedLevel {
				t.Errorf("expected level %v, got %v", tt.expectedLevel, level)
			}
		})
	}
}

func TestConsoleLogger_ConcurrentAccess(t *testing.T) {
	t.Run("parallel writes with synchronization", func(t *testing.T) {
		t.Parallel()
		buf := &bytes.Buffer{}
		logger := NewConsoleLogger()
		if err := logger.SetOutput(buf); err != nil {
			t.Fatalf("failed to set output: %v", err)
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		numGoroutines := 10
		messagesPerGoroutine := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					mu.Lock()
					logger.Info("goroutine %d message %d", id, j)
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		output := buf.String()
		expectedLines := numGoroutines * messagesPerGoroutine
		actualLines := strings.Count(output, "\n")

		if actualLines != expectedLines {
			t.Errorf("expected %d log lines, got %d", expectedLines, actualLines)
		}
	})
}

func TestConsoleLogger_NotThreadSafe(t *testing.T) {
	t.Run("unsynchronized concurrent access", func(t *testing.T) {
		t.Parallel()
		buf := &bytes.Buffer{}
		logger := NewConsoleLogger()
		if err := logger.SetOutput(buf); err != nil {
			t.Fatalf("failed to set output: %v", err)
		}

		var wg sync.WaitGroup
		numGoroutines := 10
		messagesPerGoroutine := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					logger.Info("goroutine %d message %d", id, j)
				}
			}(i)
		}

		wg.Wait()

		output := buf.String()
		actualLines := strings.Count(output, "\n")
		expectedLines := numGoroutines * messagesPerGoroutine

		if actualLines != expectedLines {
			t.Logf("WARNING: Expected %d log lines, got %d due to lack of thread safety (this is documented)", expectedLines, actualLines)
		}
	})
}

func TestConsoleLogger_ConcurrentConfigChanges(t *testing.T) {
	t.Run("parallel config changes", func(t *testing.T) {
		t.Parallel()
		logger := NewConsoleLogger()

		var wg sync.WaitGroup
		numGoroutines := 5
		operationsPerGoroutine := 50

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					switch j % 4 {
					case 0:
						logger.SetVerbose(true)
					case 1:
						logger.SetVerbose(false)
					case 2:
						logger.SetQuiet(true)
					case 3:
						logger.SetQuiet(false)
					}
				}
			}(i)
		}

		wg.Wait()

		level := logger.GetLevel()
		if level < LevelDebug || level > LevelQuiet {
			t.Errorf("level %v is out of valid range", level)
		}
	})
}

func TestConsoleLogger_NewConsoleLoggerDefaults(t *testing.T) {
	logger := NewConsoleLogger()

	if logger.GetLevel() != LevelInfo {
		t.Errorf("expected default level LevelInfo, got %v", logger.GetLevel())
	}

	buf := &bytes.Buffer{}
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	output := buf.String()
	if strings.Contains(output, "DEBUG") {
		t.Errorf("default level should not show DEBUG messages")
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("default level should show INFO messages")
	}
	if !strings.Contains(output, "WARN") {
		t.Errorf("default level should show WARN messages")
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("default level should show ERROR messages")
	}
}

func TestConsoleLogger_MultipleLevelTransitions(t *testing.T) {
	logger := NewConsoleLogger()

	logger.SetLevel(LevelWarn)
	if logger.GetLevel() != LevelWarn {
		t.Errorf("expected level LevelWarn, got %v", logger.GetLevel())
	}

	logger.SetVerbose(true)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("expected level LevelDebug after verbose, got %v", logger.GetLevel())
	}

	logger.SetQuiet(true)
	if logger.GetLevel() != LevelQuiet {
		t.Errorf("expected level LevelQuiet after quiet, got %v", logger.GetLevel())
	}

	logger.SetVerbose(true)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("expected level LevelDebug after verbose, got %v", logger.GetLevel())
	}

	logger.SetQuiet(false)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("expected level LevelDebug (verbose mode still active), got %v", logger.GetLevel())
	}

	logger.SetVerbose(false)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("expected level LevelDebug (previousLevel is Quiet), got %v", logger.GetLevel())
	}
}

func TestConsoleLogger_OutputReplacement(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	logger := NewConsoleLogger()

	if err := logger.SetOutput(buf1); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.Info("message to buffer1")

	if !strings.Contains(buf1.String(), "message to buffer1") {
		t.Errorf("expected message in buffer1")
	}

	if err := logger.SetOutput(buf2); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.Info("message to buffer2")

	if strings.Contains(buf1.String(), "message to buffer2") {
		t.Errorf("second message should not be in buffer1")
	}
	if !strings.Contains(buf2.String(), "message to buffer2") {
		t.Errorf("expected message in buffer2")
	}
}

func TestConsoleLogger_PerformanceWithLevels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.SetLevel(LevelError)

	start := time.Now()
	for i := 0; i < 10000; i++ {
		logger.Debug("debug %d", i)
		logger.Info("info %d", i)
		logger.Warn("warn %d", i)
	}
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Logf("filtered logging took %v for 30000 operations", elapsed)
	}

	buf.Reset()
	logger.SetLevel(LevelDebug)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		logger.Debug("debug %d", i)
	}
	elapsed = time.Since(start)

	if elapsed > 500*time.Millisecond {
		t.Logf("unfiltered logging took %v for 10000 operations", elapsed)
	}
}

func TestLogLevel_StringUnknown(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"level -1", LogLevel(-1), "UNKNOWN"},
		{"level 100", LogLevel(100), "UNKNOWN"},
		{"level MAX_INT", LogLevel(2147483647), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.level.String())
			}
		})
	}
}

func TestConsoleLogger_Command(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.Command("rclone sync /src /dst")

	output := buf.String()
	if !strings.Contains(output, "rclone sync /src /dst") {
		t.Errorf("expected command in output, got %q", output)
	}
	if !strings.Contains(output, "\033[36m") {
		t.Errorf("expected cyan color code in output")
	}
}

func TestConsoleLogger_CommandLevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		shouldLog bool
	}{
		{"LevelDebug allows Command", LevelDebug, true},
		{"LevelInfo allows Command", LevelInfo, true},
		{"LevelWarn suppresses Command", LevelWarn, false},
		{"LevelError suppresses Command", LevelError, false},
		{"LevelQuiet suppresses Command", LevelQuiet, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewConsoleLogger()
			if err := logger.SetOutput(buf); err != nil {
				t.Fatalf("failed to set output: %v", err)
			}
			logger.SetLevel(tt.level)

			logger.Command("test command")

			output := buf.String()
			hasOutput := strings.Contains(output, "test command")
			if hasOutput != tt.shouldLog {
				t.Errorf("expected output: %v, got output: %v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestConsoleLogger_CommandQuietMode(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
	logger.SetQuiet(true)

	logger.Command("test command")

	if buf.String() != "" {
		t.Errorf("quiet mode should suppress Command output, got %q", buf.String())
	}
}

func TestConsoleLogger_StageInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.StageInfo("Starting sync phase")

	output := buf.String()
	if !strings.Contains(output, "Starting sync phase") {
		t.Errorf("expected message in output, got %q", output)
	}
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("expected INFO tag in output")
	}
	if !strings.Contains(output, "\033[1m") {
		t.Errorf("expected bold color code in output")
	}
}

func TestConsoleLogger_StageInfoWithArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.StageInfo("Processing %d files", 42)

	output := buf.String()
	if !strings.Contains(output, "Processing 42 files") {
		t.Errorf("expected formatted message, got %q", output)
	}
	if !strings.Contains(output, "\033[1m") {
		t.Errorf("expected bold color code in output")
	}
}

func TestConsoleLogger_StageInfoLevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		shouldLog bool
	}{
		{"LevelDebug allows StageInfo", LevelDebug, true},
		{"LevelInfo allows StageInfo", LevelInfo, true},
		{"LevelWarn suppresses StageInfo", LevelWarn, false},
		{"LevelError suppresses StageInfo", LevelError, false},
		{"LevelQuiet suppresses StageInfo", LevelQuiet, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewConsoleLogger()
			if err := logger.SetOutput(buf); err != nil {
				t.Fatalf("failed to set output: %v", err)
			}
			logger.SetLevel(tt.level)

			logger.StageInfo("test message")

			output := buf.String()
			hasOutput := strings.Contains(output, "test message")
			if hasOutput != tt.shouldLog {
				t.Errorf("expected output: %v, got output: %v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestConsoleLogger_StageInfoQuietMode(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
	logger.SetQuiet(true)

	logger.StageInfo("test message")

	if buf.String() != "" {
		t.Errorf("quiet mode should suppress StageInfo output, got %q", buf.String())
	}
}

func TestConsoleLogger_TargetInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.TargetInfo("Syncing from /src")

	output := buf.String()
	if !strings.Contains(output, "Syncing from /src") {
		t.Errorf("expected message in output, got %q", output)
	}
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("expected INFO tag in output")
	}
	if strings.Contains(output, "\033[1m") {
		t.Errorf("should not contain bold color code in output")
	}
}

func TestConsoleLogger_TargetInfoWithArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.TargetInfo("Target: %s at %s", "remote:bucket", "main")

	output := buf.String()
	if !strings.Contains(output, "Target: remote:bucket at main") {
		t.Errorf("expected formatted message, got %q", output)
	}
}

func TestConsoleLogger_TargetInfoLevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		shouldLog bool
	}{
		{"LevelDebug allows TargetInfo", LevelDebug, true},
		{"LevelInfo allows TargetInfo", LevelInfo, true},
		{"LevelWarn suppresses TargetInfo", LevelWarn, false},
		{"LevelError suppresses TargetInfo", LevelError, false},
		{"LevelQuiet suppresses TargetInfo", LevelQuiet, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewConsoleLogger()
			if err := logger.SetOutput(buf); err != nil {
				t.Fatalf("failed to set output: %v", err)
			}
			logger.SetLevel(tt.level)

			logger.TargetInfo("test target")

			output := buf.String()
			hasOutput := strings.Contains(output, "test target")
			if hasOutput != tt.shouldLog {
				t.Errorf("expected output: %v, got output: %v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestConsoleLogger_TargetInfoQuietMode(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
	logger.SetQuiet(true)

	logger.TargetInfo("test target")

	if buf.String() != "" {
		t.Errorf("quiet mode should suppress TargetInfo output, got %q", buf.String())
	}
}

func TestConsoleLogger_CombinedOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
	logger.SetLevel(LevelDebug)

	output := "Line 1\nLine 2\nLine 3"
	logger.CombinedOutput(output)

	result := buf.String()
	if !strings.Contains(result, "Line 1") {
		t.Errorf("expected Line 1 in output, got %q", result)
	}
	if !strings.Contains(result, "Line 2") {
		t.Errorf("expected Line 2 in output, got %q", result)
	}
	if !strings.Contains(result, "Line 3") {
		t.Errorf("expected Line 3 in output, got %q", result)
	}
	if !strings.Contains(result, "\033[90m") {
		t.Errorf("expected gray color code in output")
	}
}

func TestConsoleLogger_CombinedOutputFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
	logger.SetLevel(LevelDebug)

	output := "Line 1\nElapsed time: 1.234s\nChecks: 123\nLine 4"
	logger.CombinedOutput(output)

	result := buf.String()
	if !strings.Contains(result, "Line 1") {
		t.Errorf("expected Line 1 in output, got %q", result)
	}
	if strings.Contains(result, "Elapsed time:") {
		t.Errorf("Elapsed time line should be filtered, got %q", result)
	}
	if strings.Contains(result, "Checks:") {
		t.Errorf("Checks line should be filtered, got %q", result)
	}
	if !strings.Contains(result, "Line 4") {
		t.Errorf("expected Line 4 in output, got %q", result)
	}
}

func TestConsoleLogger_CombinedOutputLevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		shouldLog bool
	}{
		{"LevelDebug allows CombinedOutput", LevelDebug, true},
		{"LevelInfo suppresses CombinedOutput", LevelInfo, false},
		{"LevelWarn suppresses CombinedOutput", LevelWarn, false},
		{"LevelError suppresses CombinedOutput", LevelError, false},
		{"LevelQuiet suppresses CombinedOutput", LevelQuiet, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewConsoleLogger()
			if err := logger.SetOutput(buf); err != nil {
				t.Fatalf("failed to set output: %v", err)
			}
			logger.SetLevel(tt.level)

			logger.CombinedOutput("test output")

			hasOutput := strings.Contains(buf.String(), "test output")
			if hasOutput != tt.shouldLog {
				t.Errorf("expected output: %v, got output: %v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestConsoleLogger_CombinedOutputQuietMode(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
	logger.SetQuiet(true)

	logger.CombinedOutput("test output")

	if buf.String() != "" {
		t.Errorf("quiet mode should suppress CombinedOutput output, got %q", buf.String())
	}
}

func TestConsoleLogger_CombinedOutputEmptyInput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}

	logger.CombinedOutput("")

	if buf.String() != "" {
		t.Errorf("empty input should produce no output, got %q", buf.String())
	}
}

func TestConsoleLogger_CombinedOutputWhitespaceOnly(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
	logger.SetLevel(LevelDebug)

	logger.CombinedOutput("   \n   ")

	output := buf.String()
	if output != "\n" {
		t.Errorf("whitespace-only input should produce only a newline, got %q", output)
	}
	if strings.Contains(output, " ") {
		t.Errorf("whitespace-only input should not contain spaces in output, got %q", output)
	}
}

func TestConsoleLogger_CombinedOutputTrimsEmptyLines(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger()
	if err := logger.SetOutput(buf); err != nil {
		t.Fatalf("failed to set output: %v", err)
	}
	logger.SetLevel(LevelDebug)

	logger.CombinedOutput("   \n\n   \n  \n")

	output := buf.String()
	if output != "\n" {
		t.Errorf("empty/whitespace-only lines should be trimmed, got %q", output)
	}
}
