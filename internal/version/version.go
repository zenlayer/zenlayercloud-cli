// Package version provides version information for zeno.
package version

import "runtime"

// These variables are set via ldflags during build.
var (
	// Version is the semantic version of the CLI.
	Version = "dev"
	// GitCommit is the git commit hash.
	GitCommit = "none"
	// BuildTime is the build timestamp.
	BuildTime = "unknown"
)

// Info contains version information.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the version information.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a human-readable version string.
func String() string {
	return Version
}

// Full returns a detailed version string.
func Full() string {
	info := Get()
	return "zeno version " + info.Version + "\n" +
		"  Git Commit: " + info.GitCommit + "\n" +
		"  Build Time: " + info.BuildTime + "\n" +
		"  Go Version: " + info.GoVersion + "\n" +
		"  OS/Arch:    " + info.OS + "/" + info.Arch
}
