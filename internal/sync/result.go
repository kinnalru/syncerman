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

func collectResults(results []*SyncResult) *SyncReport {
	nonNilResults := filterNilResults(results)
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
				report.Errors = append(report.Errors, fmt.Errorf("job %s target %s:%s -> %s: %v",
					result.Target.JobName, result.Target.Provider, result.Target.SourcePath, result.Target.Destination.Path, result.Error))
			}
		}
	}
	report.ExitCode = report.calculateExitCode()
	return report
}

func countBasicResults(results []*SyncResult) (successCount int, firstRunCount int) {
	nonNilResults := filterNilResults(results)
	for _, result := range nonNilResults {
		if result.Success {
			successCount++
		}
		if result.FirstRun {
			firstRunCount++
		}
	}
	return
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
	r.formatTargetsConditional(&builder, verbose, r.SuccessCount, "=== Successful Targets ===\n", r.SucceededTargets, nil)
	r.formatTargetsConditional(&builder, verbose, r.FailureCount, "=== Failed Targets ===\n", r.FailedTargets, r.Errors)
	r.formatTargetsConditional(&builder, verbose, r.FirstRunCount, "=== First-Runs ===\n", r.FirstRunTargets, nil)
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

func (r *SyncReport) formatTargetList(builder *strings.Builder, header string, targets []SyncTarget, errors []error) {
	builder.WriteString(header)
	for i, target := range targets {
		fmt.Fprintf(builder, "%d. [%s] %s:%s -> %s\n", i+1, target.JobName, target.Provider, target.SourcePath, target.Destination.Path)

		if errors != nil && i < len(errors) {
			fmt.Fprintf(builder, "   Error: %v\n", errors[i])
		}
	}
	builder.WriteString("\n")
}

func (r *SyncReport) formatTargetsConditional(builder *strings.Builder, verbose bool, count int, header string, targets []SyncTarget, errors []error) {
	if count > 0 && verbose {
		r.formatTargetList(builder, header, targets, errors)
	}
}

func (r *SyncReport) FormatError() string {
	if !r.HasErrors {
		return "sync completed successfully"
	}
	return fmt.Sprintf("sync failed with %d error(s):\n%s", r.FailureCount, joinErrorMessages(r.Errors, "\n---\n"))
}

func NewReport(results []*SyncResult) *SyncReport {
	return collectResults(results)
}

func newEmptyReport() *SyncReport {
	return &SyncReport{
		FirstRunTargets:  []SyncTarget{},
		FailedTargets:    []SyncTarget{},
		SucceededTargets: []SyncTarget{},
		Errors:           []error{},
	}
}

func filterNilResults(results []*SyncResult) []*SyncResult {
	nonNil := make([]*SyncResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			nonNil = append(nonNil, result)
		}
	}
	return nonNil
}
