package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	if info.Version == "" {
		t.Error("Get().Version should not be empty")
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("Get().GoVersion = %q, want %q", info.GoVersion, runtime.Version())
	}
	if info.OS != runtime.GOOS {
		t.Errorf("Get().OS = %q, want %q", info.OS, runtime.GOOS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("Get().Arch = %q, want %q", info.Arch, runtime.GOARCH)
	}
}

func TestGet_MatchesPackageVars(t *testing.T) {
	info := Get()

	if info.Version != Version {
		t.Errorf("Get().Version = %q, want package var Version %q", info.Version, Version)
	}
	if info.GitCommit != GitCommit {
		t.Errorf("Get().GitCommit = %q, want package var GitCommit %q", info.GitCommit, GitCommit)
	}
	if info.BuildTime != BuildTime {
		t.Errorf("Get().BuildTime = %q, want package var BuildTime %q", info.BuildTime, BuildTime)
	}
}

func TestString(t *testing.T) {
	s := String()

	if s != Version {
		t.Errorf("String() = %q, want %q", s, Version)
	}
	if s == "" {
		t.Error("String() should not return an empty string")
	}
}

func TestFull(t *testing.T) {
	full := Full()

	requiredSubstrings := []string{
		"zeno version",
		"Git Commit:",
		"Build Time:",
		"Go Version:",
		"OS/Arch:",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	}

	for _, sub := range requiredSubstrings {
		if !strings.Contains(full, sub) {
			t.Errorf("Full() output does not contain %q\nFull output:\n%s", sub, full)
		}
	}
}

func TestFull_ContainsVersionString(t *testing.T) {
	full := Full()

	if !strings.Contains(full, Version) {
		t.Errorf("Full() = %q, should contain version %q", full, Version)
	}
}

func TestInfo_DefaultValues(t *testing.T) {
	originalVersion := Version
	originalCommit := GitCommit
	originalBuildTime := BuildTime
	defer func() {
		Version = originalVersion
		GitCommit = originalCommit
		BuildTime = originalBuildTime
	}()

	// Test with default/unset ldflags values
	Version = "dev"
	GitCommit = "none"
	BuildTime = "unknown"

	info := Get()
	if info.Version != "dev" {
		t.Errorf("Get().Version = %q, want %q", info.Version, "dev")
	}
	if info.GitCommit != "none" {
		t.Errorf("Get().GitCommit = %q, want %q", info.GitCommit, "none")
	}
	if info.BuildTime != "unknown" {
		t.Errorf("Get().BuildTime = %q, want %q", info.BuildTime, "unknown")
	}
}

func TestInfo_CustomValues(t *testing.T) {
	originalVersion := Version
	originalCommit := GitCommit
	originalBuildTime := BuildTime
	defer func() {
		Version = originalVersion
		GitCommit = originalCommit
		BuildTime = originalBuildTime
	}()

	Version = "1.2.3"
	GitCommit = "abc123"
	BuildTime = "2026-03-07T00:00:00Z"

	info := Get()
	if info.Version != "1.2.3" {
		t.Errorf("Get().Version = %q, want %q", info.Version, "1.2.3")
	}
	if info.GitCommit != "abc123" {
		t.Errorf("Get().GitCommit = %q, want %q", info.GitCommit, "abc123")
	}
	if info.BuildTime != "2026-03-07T00:00:00Z" {
		t.Errorf("Get().BuildTime = %q, want %q", info.BuildTime, "2026-03-07T00:00:00Z")
	}

	if s := String(); s != "1.2.3" {
		t.Errorf("String() = %q after setting Version, want %q", s, "1.2.3")
	}
}
