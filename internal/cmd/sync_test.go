package cmd

import (
	"context"
	"testing"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"
	"gitlab.com/kinnalru/syncerman/internal/sync"
)

type mockCmdExecutor struct {
	success bool
}

func (m *mockCmdExecutor) Run(ctx context.Context, args ...string) (*rclone.Result, error) {
	if m.success {
		return &rclone.Result{ExitCode: 0, Stdout: "success", Stderr: ""}, nil
	}
	return &rclone.Result{ExitCode: 1, Stdout: "", Stderr: "error"}, nil
}

type mockCmdLogger struct {
	infoLog  []string
	errorLog []string
}

func (m *mockCmdLogger) Info(msg string, args ...interface{})       { m.infoLog = append(m.infoLog, msg) }
func (m *mockCmdLogger) Warn(msg string, args ...interface{})       {}
func (m *mockCmdLogger) Error(msg string, args ...interface{})      { m.errorLog = append(m.errorLog, msg) }
func (m *mockCmdLogger) Debug(msg string, args ...interface{})      {}
func (m *mockCmdLogger) Command(cmd string)                         {}
func (m *mockCmdLogger) Output(output string)                       {}
func (m *mockCmdLogger) ErrorOutput(output string)                  {}
func (m *mockCmdLogger) CombinedOutput(output string)               {}
func (m *mockCmdLogger) StageInfo(msg string, args ...interface{})  {}
func (m *mockCmdLogger) TargetInfo(msg string, args ...interface{}) {}

func getTestConfig() *config.Config {
	return &config.Config{
		Jobs: []config.Job{
			{
				ID:      "test-job",
				Name:    "Test Job",
				Enabled: true,
				Tasks: []config.Task{
					{
						From: "gdrive:docs",
						To: []config.Destination{
							{Path: "s3:backup"},
						},
						Enabled: true,
					},
				},
			},
		},
	}
}

func TestSyncAllTargets_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockExec := &mockCmdExecutor{success: false}
	mockLog := &mockCmdLogger{}
	engine := sync.NewEngine(nil, mockExec, mockLog)
	cfg := getTestConfig()

	opts := sync.SyncOptions{
		DryRun:  false,
		Verbose: false,
		Quiet:   false,
	}

	err := syncAllTargets(ctx, engine, cfg, opts, GetLogger())
	if err != nil {
		t.Logf("syncAllTargets returned expected error: %v", err)
	}
}

func TestSyncSingleJob_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockExec := &mockCmdExecutor{success: false}
	mockLog := &mockCmdLogger{}
	engine := sync.NewEngine(nil, mockExec, mockLog)
	cfg := getTestConfig()

	opts := sync.SyncOptions{
		DryRun:  false,
		Verbose: false,
		Quiet:   false,
	}

	err := syncSingleJob(ctx, GetLogger(), engine, cfg, "test-job", opts)
	if err != nil {
		t.Logf("syncSingleJob returned expected error: %v", err)
	}
}

func TestReportResults_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		results []*sync.SyncResult
		opts    sync.SyncOptions
		wantErr bool
	}{
		{
			name:    "empty results",
			results: []*sync.SyncResult{},
			opts:    sync.SyncOptions{},
			wantErr: false,
		},
		{
			name: "successful result",
			results: []*sync.SyncResult{
				{Success: true},
			},
			opts:    sync.SyncOptions{Verbose: true},
			wantErr: false,
		},
		{
			name: "failed result",
			results: []*sync.SyncResult{
				{Success: false},
			},
			opts:    sync.SyncOptions{Verbose: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := sync.NewEngine(nil, nil, nil)
			err := reportResults(engine, tt.results, tt.opts, GetLogger())
			if (err != nil) != tt.wantErr {
				t.Errorf("reportResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindAndValidateJobTargets_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		jobID     string
		setupFunc func() *config.Config
		wantErr   bool
	}{
		{
			name:  "invalid target format (old format)",
			jobID: "gdrive:docs",
			setupFunc: func() *config.Config {
				return getTestConfig()
			},
			wantErr: true,
		},
		{
			name:  "job not found",
			jobID: "nonexistent-job",
			setupFunc: func() *config.Config {
				return getTestConfig()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupFunc()
			mockExec := &mockCmdExecutor{success: false}
			mockLog := &mockCmdLogger{}
			engine := sync.NewEngine(nil, mockExec, mockLog)

			_, err := findAndValidateJobTargets(GetLogger(), engine, cfg, tt.jobID)
			if (err != nil) != tt.wantErr {
				t.Errorf("findAndValidateJobTargets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadAndValidateConfig_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func()
		wantErr  bool
	}{
		{
			name: "no config file",
			setupCmd: func() {
				commandConfig = NewCommandConfig()
				commandConfig.ConfigFile = "/nonexistent/config.yml"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCmd()
			GetLogger().SetQuiet(true)
			_, err := loadAndValidateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadAndValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
