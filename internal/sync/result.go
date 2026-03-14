package sync

import (
	"fmt"
	"strings"
)

// SyncReport represents aggregated statistics and details of sync operations.
type SyncReport struct {
	TotalTargets     int          // Total number of sync targets processed
	SuccessCount     int          // Number of successful syncs
	FailureCount     int          // Number of failed syncs
	FirstRunCount    int          // Number of first-run retried syncs
	FirstRunTargets  []SyncTarget // List of targets that had first-run errors
	FailedTargets    []SyncTarget // List of targets that failed
	SucceededTargets []SyncTarget // List of targets that succeeded
	HasErrors        bool         // Whether any errors occurred
	ExitCode         int          // Recommended exit code
	Errors           []error      // Collection of errors from failed syncs
}

// CollectResults aggregates sync results into a comprehensive report.
// It calculates statistics and generates exit codes based on results.
func (e *Engine) CollectResults(results []*SyncResult) *SyncReport {
	report := &SyncReport{
		TotalTargets:     len(results),
		SuccessCount:     0,
		FailureCount:     0,
		FirstRunCount:    0,
		FirstRunTargets:  []SyncTarget{},
		FailedTargets:    []SyncTarget{},
		SucceededTargets: []SyncTarget{},
		HasErrors:        false,
		ExitCode:         0,
		Errors:           []error{},
	}

	for _, result := range results {
		if result == nil {
			continue
		}

		if result.Success {
			report.SuccessCount++
			report.SucceededTargets = append(report.SucceededTargets, result.Target)
			if result.FirstRun {
				report.FirstRunCount++
				report.FirstRunTargets = append(report.FirstRunTargets, result.Target)
			}
		} else {
			report.FailureCount++
			report.FailedTargets = append(report.FailedTargets, result.Target)
			report.HasErrors = true

			if result.Error != nil {
				report.Errors = append(report.Errors, fmt.Errorf("target %s:%s -> %s: %v",
					result.Target.Provider, result.Target.SourcePath, result.Target.Destination.To, result.Error))
			}
		}
	}

	report.ExitCode = report.calculateExitCode()

	return report
}

func (r *SyncReport) calculateExitCode() int {
	if !r.HasErrorsField() {
		return 0
	}

	if r.FailureCount == 0 {
		return 0
	}

	if r.FirstRunCount > 0 && r.FailureCount == 0 {
		return 0
	}

	return 1
}

// Format generates human-readable summary of sync report.
// The verbosity level determines how much detail is included.
func (r *SyncReport) Format(verbose bool) string {
	var builder strings.Builder

	builder.WriteString("=== Sync Summary ===\n")
	builder.WriteString(fmt.Sprintf("Total targets:   %d\n", r.TotalTargets))
	builder.WriteString(fmt.Sprintf("Successful:     %d\n", r.SuccessCount))

	if r.FailureCount > 0 {
		builder.WriteString(fmt.Sprintf("Failed:         %d\n", r.FailureCount))
	}

	if r.FirstRunCount > 0 {
		builder.WriteString(fmt.Sprintf("First-runs:     %d\n", r.FirstRunCount))
	}

	builder.WriteString("\n")

	if r.SuccessCount > 0 && verbose {
		builder.WriteString("=== Successful Targets ===\n")
		for i, target := range r.SucceededTargets {
			builder.WriteString(fmt.Sprintf("%d. %s:%s -> %s\n", i+1, target.Provider, target.SourcePath, target.Destination.To))
		}
		builder.WriteString("\n")
	}

	if r.FailureCount > 0 && verbose {
		builder.WriteString("=== Failed Targets ===\n")
		for i, target := range r.FailedTargets {
			builder.WriteString(fmt.Sprintf("%d. %s:%s -> %s\n", i+1, target.Provider, target.SourcePath, target.Destination.To))

			idx := r.FailureCount - 1
			for j, result := range r.Errors {
				if j == idx {
					builder.WriteString(fmt.Sprintf("   Error: %v\n", result))
					break
				}
			}
		}
		builder.WriteString("\n")
	}

	if r.FirstRunCount > 0 && verbose {
		builder.WriteString("=== First-Run Errors ===\n")
		for i, target := range r.FirstRunTargets {
			builder.WriteString(fmt.Sprintf("%d. %s:%s -> %s\n", i+1, target.Provider, target.SourcePath, target.Destination.To))
			builder.WriteString(fmt.Sprintf("   Retried with --resync (attempt %d)\n", target.Resync+1))
		}
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("Exit code: %d\n", r.ExitCode))

	return builder.String()
}

// FormatError generates detailed error message from sync report.
func (r *SyncReport) FormatError() string {
	if !r.HasErrorsField() {
		return "sync completed successfully"
	}

	errorsStr := make([]string, len(r.Errors))
	for i, err := range r.Errors {
		errorsStr[i] = err.Error()
	}

	return fmt.Sprintf("sync failed with %d error(s):\n%s", r.FailureCount, strings.Join(errorsStr, "\n---\n"))
}

// HasErrorsField returns true if any sync operations failed.
func (r *SyncReport) HasErrorsField() bool {
	return r.HasErrorsField()
}

// GetExitCodeField returns recommended exit code for this report.
func (r *SyncReport) GetExitCodeField() int {
	return r.ExitCode
}

// NewReport creates a new SyncReport from sync results.
func NewReport(results []*SyncResult) *SyncReport {
	engine := NewEngine(nil, nil, nil)
	return engine.CollectResults(results)
}

// AggregateReport combines multiple reports into a single summary.
func AggregateReport(reports []*SyncReport) *SyncReport {
	combined := &SyncReport{
		TotalTargets:     0,
		SuccessCount:     0,
		FailureCount:     0,
		FirstRunCount:    0,
		FirstRunTargets:  []SyncTarget{},
		FailedTargets:    []SyncTarget{},
		SucceededTargets: []SyncTarget{},
		HasErrors:        false,
		ExitCode:         0,
		Errors:           []error{},
	}

	for _, report := range reports {
		combined.TotalTargets += report.TotalTargets
		combined.SuccessCount += report.SuccessCount
		combined.FailureCount += report.FailureCount
		combined.FirstRunCount += report.FirstRunCount
		combined.FirstRunTargets = append(combined.FirstRunTargets, report.FirstRunTargets...)
		combined.FailedTargets = append(combined.FailedTargets, report.FailedTargets...)
		combined.SucceededTargets = append(combined.SucceededTargets, report.SucceededTargets...)
		combined.HasErrors = combined.HasErrors || report.HasErrors
		combined.Errors = append(combined.Errors, report.Errors...)
	}

	combined.ExitCode = combined.calculateExitCode()

	return combined
}
