package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

func parseVersion(v string) string {
	return strings.TrimSpace(v)
}

var (
	Version   = parseVersion(versionFile)
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

func GetVersion() string {
	return Version
}

func GetGitCommit() string {
	return GitCommit
}

func GetBuildTime() string {
	return BuildTime
}

func GetGoVersion() string {
	return GoVersion
}

func GetFullVersion() string {
	return Version
}
