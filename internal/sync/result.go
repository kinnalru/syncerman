package sync

import (
	"fmt"
	"strings"
)

const (
	exitCodeSuccess = 0
	exitCodeFailure = 1
)

// SyncReport represents aggregated statistics and details of sync operations.
type SyncReport struct {
	TotalTargets     int          // Total number of sync targets processed
	SuccessCount     int          // Number of successful syncs
	FailureCount     int          // Number of failed syncs
	FirstRunCount    int          // Number of first-run retried syncs
	FirstRunTargets  []SyncTarget // List of targets that had first-runs
	FailedTargets    []SyncTarget // List of targets that failed
	SucceededTargets []SyncTarget // List of targets that succeeded
	HasErrors        bool         // Whether any errors occurred
	ExitCode         int          // Recommended exit code
	Errors           []error      // Collection of errors from failed syncs
}

func (e *Engine) CollectResults(results []*SyncResult) *SyncReport {
	nonNilResults := []*SyncResult{}

	for _, result := range results {
		if result != nil {
			nonNilResults = append(nonNilResults, result)
		}
	}

	report := newEmptyReport()
	report.TotalTargets = len(nonNilResults)

	for _, result := range nonNilResults {
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

// calculateExitCode determines the appropriate exit code based on sync results.
// Returns exitCodeSuccess for success, exitCodeFailure for any failures.
func (r *SyncReport) calculateExitCode() int {
	if r.FailureCount > 0 {
		return exitCodeFailure
	}
	return exitCodeSuccess
}

func (r *SyncReport) Format(verbose bool) string {
	var builder strings.Builder

	r.formatHeader(&builder)
	r.formatSuccessfulTargets(&builder, verbose)
	r.formatFailedTargets(&builder, verbose)
	r.formatFirstRunTargets(&builder, verbose)
	fmt.Fprintf(&builder, "Exit code: %d\n", r.ExitCode)

	return builder.String()
}

func (r *SyncReport) formatHeader(builder *strings.Builder) {
	builder.WriteString("=== Sync Summary ===\n")
	fmt.Fprintf(builder, "Total targets:   %d\n", r.TotalTargets)
	fmt.Fprintf(builder, "Successful:     %d\n", r.SuccessCount)

	if r.FailureCount > 0 {
		fmt.Fprintf(builder, "Failed:         %d\n", r.FailureCount)
	}

	if r.FirstRunCount > 0 {
		fmt.Fprintf(builder, "First-runs:     %d\n", r.FirstRunCount)
	}

	builder.WriteString("\n")
}

func (r *SyncReport) formatSuccessfulTargets(builder *strings.Builder, verbose bool) {
	if r.SuccessCount > 0 && verbose {
		builder.WriteString("=== Successful Targets ===\n")
		for i, target := range r.SucceededTargets {
			fmt.Fprintf(builder, "%d. %s:%s -> %s\n", i+1, target.Provider, target.SourcePath, target.Destination.To)
		}
		builder.WriteString("\n")
	}
}

func (r *SyncReport) formatFailedTargets(builder *strings.Builder, verbose bool) {
	if r.FailureCount > 0 && verbose {
		builder.WriteString("=== Failed Targets ===\n")
		for i, target := range r.FailedTargets {
			fmt.Fprintf(builder, "%d. %s:%s -> %s\n", i+1, target.Provider, target.SourcePath, target.Destination.To)

			if i < len(r.Errors) {
				fmt.Fprintf(builder, "   Error: %v\n", r.Errors[i])
			}
		}
		builder.WriteString("\n")
	}
}

func (r *SyncReport) formatFirstRunTargets(builder *strings.Builder, verbose bool) {
	if r.FirstRunCount > 0 && verbose {
		builder.WriteString("=== First-Runs ===\n")
		for i, target := range r.FirstRunTargets {
			fmt.Fprintf(builder, "%d. %s:%s -> %s\n", i+1, target.Provider, target.SourcePath, target.Destination.To)
		}
		builder.WriteString("\n")
	}
}

func (r *SyncReport) FormatError() string {
	if !r.HasErrors {
		return "sync completed successfully"
	}

	errorsStr := make([]string, len(r.Errors))
	for i, err := range r.Errors {
		errorsStr[i] = err.Error()
	}

	return fmt.Sprintf("sync failed with %d error(s):\n%s", r.FailureCount, strings.Join(errorsStr, "\n---\n"))
}

// NewReport creates a new SyncReport from sync results.
func NewReport(results []*SyncResult) *SyncReport {
	engine := NewEngine(nil, nil, nil)
	return engine.CollectResults(results)
}

func newEmptyReport() *SyncReport {
	return &SyncReport{
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
}

func AggregateReport(reports []*SyncReport) *SyncReport {
	combined := newEmptyReport()

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
