package sync

import (
	"context"

	"syncerman/internal/rclone"
)

type mockExecutor struct {
	results []*rclone.Result
	errors  []error
	index   int
}

func (m *mockExecutor) Run(ctx context.Context, args ...string) (*rclone.Result, error) {
	if m.index >= len(m.results) {
		return &rclone.Result{ExitCode: 0, Stdout: "", Stderr: ""}, nil
	}

	result := m.results[m.index]
	m.index++

	err := error(nil)
	if m.index-1 < len(m.errors) {
		err = m.errors[m.index-1]
	}

	return result, err
}

type mockLogger struct {
	info     []string
	warn     []string
	errorLog []string
	debugLog []string
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.info = append(m.info, msg)
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	m.warn = append(m.warn, msg)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.errorLog = append(m.errorLog, msg)
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.debugLog = append(m.debugLog, msg)
}
