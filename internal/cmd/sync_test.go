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

func TestSyncAllTargets_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockExec := &mockCmdExecutor{success: false}
	mockLog := &mockCmdLogger{}
	engine := sync.NewEngine(nil, mockExec, mockLog)
	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{{To: "s3:backup"}},
	})

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

func TestSyncSingleTarget_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockExec := &mockCmdExecutor{success: false}
	mockLog := &mockCmdLogger{}
	engine := sync.NewEngine(nil, mockExec, mockLog)
	cfg := config.NewConfig()
	cfg.AddProvider("gdrive", config.PathMap{
		"docs": []config.Destination{{To: "s3:backup"}},
	})

	opts := sync.SyncOptions{
		DryRun:  false,
		Verbose: false,
		Quiet:   false,
	}

	err := syncSingleTarget(ctx, GetLogger(), engine, cfg, "gdrive:docs", opts)
	if err != nil {
		t.Logf("syncSingleTarget returned expected error: %v", err)
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

func TestFindAndValidateTarget_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		targetArg string
		setupFunc func()
		wantErr   bool
	}{
		{
			name:      "invalid target format",
			targetArg: "invalid",
			setupFunc: func() {},
			wantErr:   true,
		},
		{
			name:      "target not found",
			targetArg: "notfound:path",
			setupFunc: func() {
				cfg := config.NewConfig()
				cfg.AddProvider("gdrive", config.PathMap{
					"docs": []config.Destination{{To: "s3:backup"}},
				})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			mockExec := &mockCmdExecutor{success: false}
			mockLog := &mockCmdLogger{}
			engine := sync.NewEngine(nil, mockExec, mockLog)
			cfg := config.NewConfig()
			cfg.AddProvider("gdrive", config.PathMap{
				"docs": []config.Destination{{To: "s3:backup"}},
			})

			_, err := findAndValidateTarget(GetLogger(), engine, cfg, tt.targetArg)
			if (err != nil) != tt.wantErr {
				t.Errorf("findAndValidateTarget() error = %v, wantErr %v", err, tt.wantErr)
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
