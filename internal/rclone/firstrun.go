package rclone

import (
	"regexp"
)

// FirstRunErrorPatternListings matches first part of first-run error pattern.
var FirstRunErrorPatternListings = regexp.MustCompile(`(?i)cannot find prior Path1 or Path2 listings`)

// FirstRunErrorPatternFilenames matches second part of first-run error pattern.
// Matches "here are the filenames" or "here are filenames".
var FirstRunErrorPatternFilenames = regexp.MustCompile(`(?i)here are(?:\s+the)?\s+filenames`)

// IsFirstRunError checks if stderr contains first-run error pattern.
// First-run errors occur when rclone bisync is run for the first time
// and no state files exist yet.
func IsFirstRunError(stderr string) bool {
	return FirstRunErrorPatternListings.MatchString(stderr) && FirstRunErrorPatternFilenames.MatchString(stderr)
}

// ExtractFirstRunErrorPaths extracts file paths from first-run error message.
// Returns slice of file paths (Path1 and Path2) if found, empty if not.
func ExtractFirstRunErrorPaths(stderr string) []string {
	if !IsFirstRunError(stderr) {
		return []string{}
	}

	pathPattern := regexp.MustCompile(`Path[:12]:\s*(\S+)`)
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
type FirstRunError struct {
	Message string
	Paths   []string
}

// ParseFirstRunError parses a first-run error from stderr.
// Returns nil if stderr doesn't contain first-run error pattern.
func ParseFirstRunError(stderr string) *FirstRunError {
	if !IsFirstRunError(stderr) {
		return nil
	}

	return &FirstRunError{
		Message: stderr,
		Paths:   ExtractFirstRunErrorPaths(stderr),
	}
}
