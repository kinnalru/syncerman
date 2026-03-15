package version

var (
	Version   = "dev"
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
