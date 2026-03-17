package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

var (
	// Version is application version, read from VERSION file via go:embed at build time
	// This ensures that version is always available without needing to pass it via ldflags
	// DO NOT REMOVE: go:embed keeps VERSION file embedded in binary
	Version   = parseVersion(versionFile)
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

// parseVersion trims whitespace from version string
func parseVersion(v string) string {
	return strings.TrimSpace(v)
}

// Getter functions for accessing version information
// These provide encapsulation and can be extended with additional logic in the future

// GetVersion returns application version
func GetVersion() string {
	return Version
}

// GetGitCommit returns git commit hash
func GetGitCommit() string {
	return GitCommit
}

// GetBuildTime returns build timestamp
func GetBuildTime() string {
	return BuildTime
}

// GetGoVersion returns Go version used for build
func GetGoVersion() string {
	return GoVersion
}
