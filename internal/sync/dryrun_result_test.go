package sync

import (
	"fmt"
	"testing"

	"syncerman/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectResults_SuccessOnly(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)

	assert.Equal(t, 2, report.TotalTargets)
	assert.Equal(t, 2, report.SuccessCount)
	assert.Equal(t, 0, report.FailureCount)
	assert.Equal(t, 0, report.FirstRunCount)
	assert.False(t, report.HasErrors)
	assert.Equal(t, 0, report.ExitCode)
}

func TestCollectResults_WithFailures(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    false,
			Error:      assert.AnError,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)

	assert.Equal(t, 2, report.TotalTargets)
	assert.Equal(t, 1, report.SuccessCount)
	assert.Equal(t, 1, report.FailureCount)
	assert.Equal(t, 0, report.FirstRunCount)
	assert.True(t, report.HasErrors)
	assert.Equal(t, 1, report.ExitCode)
}

func TestCollectResults_WithFirstRun(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   true,
			RetryCount: 1,
		},
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)

	assert.Equal(t, 2, report.TotalTargets)
	assert.Equal(t, 2, report.SuccessCount)
	assert.Equal(t, 0, report.FailureCount)
	assert.Equal(t, 1, report.FirstRunCount)
	assert.False(t, report.HasErrors)
	assert.Equal(t, 0, report.ExitCode)
}

func TestCollectResults_NilValues(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
		nil,
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)

	assert.Equal(t, 2, report.TotalTargets)
	assert.Equal(t, 2, report.SuccessCount)
}

func TestReportFormat_NonVerbose(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	output := report.Format(false)

	assert.Contains(t, output, "=== Sync Summary ===")
	assert.Contains(t, output, "Total targets:   1")
	assert.Contains(t, output, "Successful:     1")
	assert.Contains(t, output, "Exit code: 0")
	assert.NotContains(t, output, "=== Successful Targets ===")
}

func TestReportFormat_Verbose(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	output := report.Format(true)

	assert.Contains(t, output, "=== Sync Summary ===")
	assert.Contains(t, output, "=== Successful Targets ===")
	assert.Contains(t, output, "1. gdrive:docs -> s3:backup/docs")
}

func TestReportFormat_VerboseWithFailures(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    false,
			Error:      fmt.Errorf("permission denied"),
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	output := report.Format(true)

	assert.Contains(t, output, "=== Failed Targets ===")
	assert.Contains(t, output, "1. local:/data -> gdrive:data")
	assert.Contains(t, output, "permission denied")
}

func TestReportFormat_VerboseWithFirstRun(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   true,
			RetryCount: 1,
		},
	}

	report := engine.CollectResults(results)
	output := report.Format(true)

	assert.Contains(t, output, "=== First-Run Errors ===")
	assert.Contains(t, output, "1. gdrive:docs -> s3:backup/docs")
}

func TestReportFormatError_NoErrors(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	errMsg := report.FormatError()

	assert.Equal(t, "sync completed successfully", errMsg)
}

func TestReportFormatError_WithErrors(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    false,
			Error:      fmt.Errorf("permission denied"),
			FirstRun:   false,
			RetryCount: 0,
		},
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "photos", Destination: config.Destination{To: "s3:backup/photos"}},
			Success:    false,
			Error:      fmt.Errorf("network timeout"),
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	errMsg := report.FormatError()

	assert.Contains(t, errMsg, "sync failed with 2 error(s)")
	assert.Contains(t, errMsg, "permission denied")
	assert.Contains(t, errMsg, "network timeout")
}

func TestCalculateExitCode_Success(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	exitCode := report.GetExitCodeField()

	assert.Equal(t, 0, exitCode)
}

func TestCalculateExitCode_Failure(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    false,
			Error:      assert.AnError,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	exitCode := report.GetExitCodeField()

	assert.Equal(t, 1, exitCode)
}

func TestCalculateExitCode_FirstRunNoFailure(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   true,
			RetryCount: 1,
		},
	}

	report := engine.CollectResults(results)
	exitCode := report.GetExitCodeField()

	assert.Equal(t, 0, exitCode)
}

func TestHasErrorsField_True(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    false,
			Error:      assert.AnError,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	assert.True(t, report.HasErrorsField())
}

func TestHasErrorsField_False(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := engine.CollectResults(results)
	assert.False(t, report.HasErrorsField())
}

func TestNewReport(t *testing.T) {
	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := NewReport(results)

	require.NotNil(t, report)
	assert.Equal(t, 1, report.TotalTargets)
	assert.Equal(t, 1, report.SuccessCount)
}

func TestAggregateReport(t *testing.T) {
	report1 := NewReport([]*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   false,
			RetryCount: 0,
		},
	})

	report2 := NewReport([]*SyncResult{
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    false,
			Error:      assert.AnError,
			FirstRun:   false,
			RetryCount: 0,
		},
	})

	combined := AggregateReport([]*SyncReport{report1, report2})

	assert.Equal(t, 2, combined.TotalTargets)
	assert.Equal(t, 1, combined.SuccessCount)
	assert.Equal(t, 1, combined.FailureCount)
	assert.Equal(t, 0, combined.FirstRunCount)
	assert.True(t, combined.HasErrors)
	assert.Equal(t, 1, combined.ExitCode)
}
