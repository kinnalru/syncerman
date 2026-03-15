package sync

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReport_SuccessOnly(t *testing.T) {
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

	report := NewReport(results)

	assert.Equal(t, 2, report.TotalTargets)
	assert.Equal(t, 2, report.SuccessCount)
	assert.Equal(t, 0, report.FailureCount)
	assert.Equal(t, 0, report.FirstRunCount)
	assert.False(t, report.HasErrors)
	assert.Equal(t, 0, report.ExitCode)
}

func TestNewReport_WithFailures(t *testing.T) {
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

	report := NewReport(results)

	assert.Equal(t, 2, report.TotalTargets)
	assert.Equal(t, 1, report.SuccessCount)
	assert.Equal(t, 1, report.FailureCount)
	assert.Equal(t, 0, report.FirstRunCount)
	assert.True(t, report.HasErrors)
	assert.Equal(t, 1, report.ExitCode)
}

func TestNewReport_WithFirstRun(t *testing.T) {

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

	report := NewReport(results)

	assert.Equal(t, 2, report.TotalTargets)
	assert.Equal(t, 2, report.SuccessCount)
	assert.Equal(t, 0, report.FailureCount)
	assert.Equal(t, 1, report.FirstRunCount)
	assert.False(t, report.HasErrors)
	assert.Equal(t, 0, report.ExitCode)
}

func TestNewReport_NilValues(t *testing.T) {

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

	report := NewReport(results)

	assert.Equal(t, 2, report.TotalTargets)
	assert.Equal(t, 2, report.SuccessCount)
}

func TestReportFormat_NonVerbose(t *testing.T) {

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
	output := report.Format(false)

	assert.Contains(t, output, "=== Sync Summary ===")
	assert.Contains(t, output, "Total targets:   1")
	assert.Contains(t, output, "Successful:     1")
	assert.Contains(t, output, "Exit code: 0")
	assert.NotContains(t, output, "=== Successful Targets ===")
}

func TestReportFormat_Verbose(t *testing.T) {

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
	output := report.Format(true)

	assert.Contains(t, output, "=== Sync Summary ===")
	assert.Contains(t, output, "=== Successful Targets ===")
	assert.Contains(t, output, "1. gdrive:docs -> s3:backup/docs")
}

func TestReportFormat_VerboseWithFailures(t *testing.T) {

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

	report := NewReport(results)
	output := report.Format(true)

	assert.Contains(t, output, "=== Failed Targets ===")
	assert.Contains(t, output, "1. local:/data -> gdrive:data")
	assert.Contains(t, output, "permission denied")
}

func TestReportFormat_VerboseWithFirstRun(t *testing.T) {

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   true,
			RetryCount: 1,
		},
	}

	report := NewReport(results)
	output := report.Format(true)

	assert.Contains(t, output, "=== First-Runs ===")
	assert.Contains(t, output, "1. gdrive:docs -> s3:backup/docs")
}

func TestReportFormatError_NoErrors(t *testing.T) {

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
	errMsg := report.FormatError()

	assert.Equal(t, "sync completed successfully", errMsg)
}

func TestReportFormatError_WithErrors(t *testing.T) {

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

	report := NewReport(results)
	errMsg := report.FormatError()

	assert.Contains(t, errMsg, "sync failed with 2 error(s)")
	assert.Contains(t, errMsg, "permission denied")
	assert.Contains(t, errMsg, "network timeout")
}

func TestCalculateExitCode_Success(t *testing.T) {

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

	assert.Equal(t, 0, report.ExitCode)
}

func TestCalculateExitCode_Failure(t *testing.T) {

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    false,
			Error:      assert.AnError,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := NewReport(results)

	assert.Equal(t, 1, report.ExitCode)
}

func TestCalculateExitCode_FirstRunNoFailure(t *testing.T) {

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			Success:    true,
			Error:      nil,
			FirstRun:   true,
			RetryCount: 1,
		},
	}

	report := NewReport(results)

	assert.Equal(t, 0, report.ExitCode)
}

func TestHasErrorsField_True(t *testing.T) {

	results := []*SyncResult{
		{
			Target:     SyncTarget{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			Success:    false,
			Error:      assert.AnError,
			FirstRun:   false,
			RetryCount: 0,
		},
	}

	report := NewReport(results)
	assert.True(t, report.HasErrors)
}

func TestHasErrorsField_False(t *testing.T) {

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
	assert.False(t, report.HasErrors)
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

func TestNewEmptyReport(t *testing.T) {
	report := newEmptyReport()

	assert.NotNil(t, report)
	assert.Equal(t, 0, report.TotalTargets)
	assert.Equal(t, 0, report.SuccessCount)
	assert.Equal(t, 0, report.FailureCount)
	assert.Equal(t, 0, report.FirstRunCount)
	assert.Len(t, report.FirstRunTargets, 0)
	assert.Len(t, report.FailedTargets, 0)
	assert.Len(t, report.SucceededTargets, 0)
	assert.False(t, report.HasErrors)
	assert.Equal(t, 0, report.ExitCode)
	assert.Len(t, report.Errors, 0)
}

func TestCalculateExitCode_AllScenarios(t *testing.T) {
	tests := []struct {
		name             string
		successCount     int
		failureCount     int
		expectedExitCode int
	}{
		{
			name:             "all success",
			successCount:     5,
			failureCount:     0,
			expectedExitCode: 0,
		},
		{
			name:             "all failures",
			successCount:     0,
			failureCount:     3,
			expectedExitCode: 1,
		},
		{
			name:             "mixed results",
			successCount:     2,
			failureCount:     1,
			expectedExitCode: 1,
		},
		{
			name:             "empty results",
			successCount:     0,
			failureCount:     0,
			expectedExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &SyncReport{
				SuccessCount: tt.successCount,
				FailureCount: tt.failureCount,
			}

			exitCode := report.calculateExitCode()
			assert.Equal(t, tt.expectedExitCode, exitCode)
		})
	}
}

func TestFormatHeader(t *testing.T) {
	tests := []struct {
		name             string
		report           *SyncReport
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "basic report",
			report: &SyncReport{
				TotalTargets:  5,
				SuccessCount:  5,
				FailureCount:  0,
				FirstRunCount: 0,
			},
			shouldContain: []string{
				"=== Sync Summary ===",
				"Total targets:   5",
				"Successful:     5",
			},
			shouldNotContain: []string{
				"Failed:",
				"First-runs:",
			},
		},
		{
			name: "report with failures",
			report: &SyncReport{
				TotalTargets:  5,
				SuccessCount:  3,
				FailureCount:  2,
				FirstRunCount: 0,
			},
			shouldContain: []string{
				"=== Sync Summary ===",
				"Total targets:   5",
				"Successful:     3",
				"Failed:         2",
			},
			shouldNotContain: []string{
				"First-runs:",
			},
		},
		{
			name: "report with first-runs",
			report: &SyncReport{
				TotalTargets:  5,
				SuccessCount:  5,
				FailureCount:  0,
				FirstRunCount: 2,
			},
			shouldContain: []string{
				"=== Sync Summary ===",
				"Total targets:   5",
				"Successful:     5",
				"First-runs:     2",
			},
			shouldNotContain: []string{
				"Failed:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var builder strings.Builder
			tt.report.formatHeader(&builder)
			output := builder.String()

			for _, s := range tt.shouldContain {
				assert.Contains(t, output, s)
			}

			for _, s := range tt.shouldNotContain {
				assert.NotContains(t, output, s)
			}
		})
	}
}

func TestFormatSuccessfulTargets(t *testing.T) {
	report := &SyncReport{
		SuccessCount: 2,
		SucceededTargets: []SyncTarget{
			{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
		},
	}

	var builder strings.Builder
	report.formatSuccessfulTargets(&builder, true)
	output := builder.String()

	assert.Contains(t, output, "=== Successful Targets ===")
	assert.Contains(t, output, "1. gdrive:docs -> s3:backup/docs")
	assert.Contains(t, output, "2. local:/data -> gdrive:data")
}

func TestFormatSuccessfulTargets_NonVerbose(t *testing.T) {
	report := &SyncReport{
		SuccessCount: 2,
		SucceededTargets: []SyncTarget{
			{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
		},
	}

	var builder strings.Builder
	report.formatSuccessfulTargets(&builder, false)
	output := builder.String()

	assert.Empty(t, output)
}

func TestFormatFailedTargets(t *testing.T) {
	report := &SyncReport{
		FailureCount: 2,
		FailedTargets: []SyncTarget{
			{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
			{Provider: "gdrive", SourcePath: "photos", Destination: config.Destination{To: "s3:backup/photos"}},
		},
		Errors: []error{
			fmt.Errorf("permission denied"),
			fmt.Errorf("network timeout"),
		},
	}

	var builder strings.Builder
	report.formatFailedTargets(&builder, true)
	output := builder.String()

	assert.Contains(t, output, "=== Failed Targets ===")
	assert.Contains(t, output, "1. local:/data -> gdrive:data")
	assert.Contains(t, output, "   Error: permission denied")
	assert.Contains(t, output, "2. gdrive:photos -> s3:backup/photos")
	assert.Contains(t, output, "   Error: network timeout")
}

func TestFormatFailedTargets_NonVerbose(t *testing.T) {
	report := &SyncReport{
		FailureCount: 2,
		FailedTargets: []SyncTarget{
			{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
		},
	}

	var builder strings.Builder
	report.formatFailedTargets(&builder, false)
	output := builder.String()

	assert.Empty(t, output)
}

func TestFormatFirstRunTargets(t *testing.T) {
	report := &SyncReport{
		FirstRunCount: 2,
		FirstRunTargets: []SyncTarget{
			{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
			{Provider: "local", SourcePath: "/data", Destination: config.Destination{To: "gdrive:data"}},
		},
	}

	var builder strings.Builder
	report.formatFirstRunTargets(&builder, true)
	output := builder.String()

	assert.Contains(t, output, "=== First-Runs ===")
	assert.Contains(t, output, "1. gdrive:docs -> s3:backup/docs")
	assert.Contains(t, output, "2. local:/data -> gdrive:data")
}

func TestFormatFirstRunTargets_NonVerbose(t *testing.T) {
	report := &SyncReport{
		FirstRunCount: 2,
		FirstRunTargets: []SyncTarget{
			{Provider: "gdrive", SourcePath: "docs", Destination: config.Destination{To: "s3:backup/docs"}},
		},
	}

	var builder strings.Builder
	report.formatFirstRunTargets(&builder, false)
	output := builder.String()

	assert.Empty(t, output)
}

func TestRun_Interface(t *testing.T) {
	mockExec := &mockExecutor{
		results: []*rclone.Result{
			{ExitCode: 0, Stdout: "synced", Stderr: ""},
		},
	}

	log := &mockLogger{}
	engine := NewEngine(nil, mockExec, log)

	target := SyncTarget{
		Provider:    "gdrive",
		SourcePath:  "docs",
		Destination: config.Destination{To: "s3:backup/docs"},
		Resync:      false,
	}

	ctx := context.Background()
	result, err := engine.Run(ctx, target, SyncOptions{Verbose: false})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
}
