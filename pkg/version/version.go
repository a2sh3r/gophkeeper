package version

import (
	"fmt"
	"runtime"
	"time"
)

var (
	Version   string = "1.0.0"
	BuildDate string = time.Now().Format("2006-01-02 15:04:05")
	GitCommit string = "unknown"
	GoVersion string = runtime.Version()
)

// SetBuildInfo sets build information (used by ldflags)
func SetBuildInfo(version, buildTime, gitCommit string) {
	if version != "" {
		Version = version
	}
	if buildTime != "" {
		BuildDate = buildTime
	}
	if gitCommit != "" {
		GitCommit = gitCommit
	}
}

// Info returns version information
func Info() string {
	return fmt.Sprintf("Version: %s\nBuild Date: %s\nGit Commit: %s\nGo Version: %s\nOS/Arch: %s/%s",
		Version, BuildDate, GitCommit, GoVersion, runtime.GOOS, runtime.GOARCH)
}

// ShortInfo returns short version information
func ShortInfo() string {
	return fmt.Sprintf("GophKeeper %s (%s)", Version, BuildDate)
}
