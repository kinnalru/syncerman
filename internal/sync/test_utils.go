package sync

import (
	"context"

	"gitlab.com/kinnalru/syncerman/internal/rclone"
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
	command  []string
	stage    []string
	target   []string
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

func (m *mockLogger) Command(cmd string) {
	m.command = append(m.command, cmd)
}

func (m *mockLogger) Output(output string)         {}
func (m *mockLogger) ErrorOutput(output string)    {}
func (m *mockLogger) CombinedOutput(output string) {}
func (m *mockLogger) StageInfo(msg string, args ...interface{}) {
	m.stage = append(m.stage, msg)
}
func (m *mockLogger) TargetInfo(msg string, args ...interface{}) {
	m.target = append(m.target, msg)
}
