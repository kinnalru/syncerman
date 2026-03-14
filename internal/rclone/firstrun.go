package rclone

import (
	"regexp"
)

// FirstRunErrorPatternListings matches the first part of the first-run error pattern.
//
// This regex pattern detects the "cannot find prior Path1 or Path2 listings" error
// that occurs when rclone bisync is run for the first time and no state files exist.
//
// Pattern: "cannot find prior Path1 or Path2 listings"
// Matching: Case-insensitive (?i flag)
// Reference: guides/OVERALL.md:317-326
//
// Use case: Combined with FirstRunErrorPatternFilenames to detect first-run errors
// and trigger automatic --resync mode.
var FirstRunErrorPatternListings = regexp.MustCompile(`(?i)cannot find prior Path1 or Path2 listings`)

// FirstRunErrorPatternFilenames matches the second part of the first-run error pattern.
//
// This regex pattern detects variations of the filename listing message that appears
// alongside the "cannot find prior Path1 or Path2 listings" error.
//
// Pattern: "here are the filenames" or "here are filenames"
// Matching: Case-insensitive (?i flag), optional "the" word
// Reference: guides/OVERALL.md:317-326
//
// Use case: Combined with FirstRunErrorPatternListings to detect first-run errors
// and trigger automatic --resync mode.
var FirstRunErrorPatternFilenames = regexp.MustCompile(`(?i)here are(?:\s+the)?\s+filenames`)

// IsFirstRunError checks if stderr contains the first-run error pattern.
//
// This function detects when rclone bisync has no state files and requires
// the --resync flag to be used for the initial synchronization.
//
// Parameters:
//   - stderr: standard error output from rclone bisync command
//
// Returns:
//   - bool: true if stderr matches the first-run error pattern, false otherwise
//
// Implementation details:
//   - Checks that both pattern parts are present: listing error AND filename message
//   - Performs case-insensitive matching
//   - First-run errors occur when rclone bisync has no existing state files
//
// Use case:
//   - Detect when to automatically re-run sync with --resync flag
//   - Prevent unnecessary sync failures by handling first-run scenarios
//
// Reference: guides/OVERALL.md:311-337
func IsFirstRunError(stderr string) bool {
	return FirstRunErrorPatternListings.MatchString(stderr) && FirstRunErrorPatternFilenames.MatchString(stderr)
}

// ExtractFirstRunErrorPaths extracts file paths from the first-run error message.
//
// This function parses the rclone bisync error message to identify which state files
// (Path1 and Path2 listings) were missing during the sync attempt.
//
// Parameters:
//   - stderr: standard error output from rclone bisync command
//
// Returns:
//   - []string: slice of file paths (Path1 and Path2) if found, empty if not a first-run error
//
// Implementation details:
//   - Parses "Path1: <path>" and "Path2: <path>" patterns from the error message
//   - Returns the paths rclone was looking for but couldn't find
//   - Returns empty slice if stderr doesn't match the first-run error pattern
//
// Use case:
//   - Troubleshooting first-run errors by showing which files are missing
//   - Logging the missing state files for user awareness
//   - Verifying that auto-resync was triggered for the correct reason
func ExtractFirstRunErrorPaths(stderr string) []string {
	if !IsFirstRunError(stderr) {
		return []string{}
	}

	pathPattern := regexp.MustCompile(`Path[12]:\s*(\S+)`)
	matches := pathPattern.FindAllStringSubmatch(stderr, -1)

	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			paths = append(paths, match[1])
		}
	}

	return paths
}

// FirstRunError represents details about a first-run error.
//
// This struct provides a structured representation of the rclone bisync
// first-run error, including the full error message and the extracted paths
// of missing state files.
//
// Fields:
//   - Message: the full error message from rclone bisync stderr output
//   - Paths: extracted file paths (Path1 and Path2) that rclone was looking for
//
// Use case:
//   - Structured error handling for first-run scenarios in the sync workflow
//   - Logging detailed error information for troubleshooting
//   - Providing clear error messages to users about missing state files
type FirstRunError struct {
	Message string
	Paths   []string
}

// ParseFirstRunError parses the first-run error from stderr into a structured format.
//
// This function combines pattern detection and path extraction to create a
// structured error object that contains both the error message and the missing
// state file paths.
//
// Parameters:
//   - stderr: standard error output from rclone bisync command
//
// Returns:
//   - *FirstRunError: structured error containing message and paths, or nil if no first-run error
//
// Implementation details:
//   - Calls IsFirstRunError to verify the stderr matches the first-run error pattern
//   - Extracts paths using ExtractFirstRunErrorPaths function
//   - Returns nil if stderr doesn't match the pattern
//   - Populates both Message and Paths fields on successful parsing
//
// Use case:
//   - Structured error handling in the sync workflow
//   - Providing detailed error information to users and logs
//   - Enabling automated retry logic with --resync flag for first-run scenarios
//
// Reference: guides/OVERALL.md:311-337
func ParseFirstRunError(stderr string) *FirstRunError {
	if !IsFirstRunError(stderr) {
		return nil
	}

	return &FirstRunError{
		Message: stderr,
		Paths:   ExtractFirstRunErrorPaths(stderr),
	}
}
