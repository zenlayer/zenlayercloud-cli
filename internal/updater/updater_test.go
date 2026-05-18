package updater

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// mockClient implements ReleaseClient for testing.
type mockClient struct {
	latest    *Release
	all       []*Release
	latestErr error
	allErr    error
}

func (m *mockClient) FetchLatest() (*Release, error) { return m.latest, m.latestErr }
func (m *mockClient) FetchAll() ([]*Release, error)  { return m.all, m.allErr }

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.8", "1.0.9", -1},
		{"1.0.9", "1.0.9", 0},
		{"1.0.9", "1.0.8", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0.0", "2.0.0", -1},
		{"v1.0.8", "v1.0.9", -1},
		{"1.1.0", "1.0.9", 1},
	}
	for _, c := range cases {
		got := CompareVersions(c.a, c.b)
		if got != c.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestArchiveName(t *testing.T) {
	name := ArchiveName("v1.0.9")
	ver := "1.0.9"
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos == "darwin" {
		goarch = "all"
	}
	want := "zeno_" + ver + "_" + goos + "_" + goarch + ".tar.gz"
	if name != want {
		t.Errorf("ArchiveName(%q) = %q, want %q", "v1.0.9", name, want)
	}
}

func TestDownloadURL(t *testing.T) {
	got := DownloadURL("v1.0.9", "zeno_1.0.9_linux_amd64.tar.gz")
	want := "https://github.com/zenlayer/zenlayercloud-cli/releases/download/v1.0.9/zeno_1.0.9_linux_amd64.tar.gz"
	if got != want {
		t.Errorf("DownloadURL = %q, want %q", got, want)
	}
}

func TestChecksumURL(t *testing.T) {
	got := ChecksumURL("v1.0.9")
	want := "https://github.com/zenlayer/zenlayercloud-cli/releases/download/v1.0.9/checksums.txt"
	if got != want {
		t.Errorf("ChecksumURL = %q, want %q", got, want)
	}
}

func TestRollback(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "zeno")
	backup := current + ".bak"
	if err := os.WriteFile(current, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backup, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Rollback(current); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	data, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "old" {
		t.Errorf("after rollback, binary = %q, want %q", string(data), "old")
	}
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Error("backup file should be gone after rollback")
	}
}

func TestRollbackNoBackup(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "zeno")
	if err := os.WriteFile(current, []byte("current"), 0755); err != nil {
		t.Fatal(err)
	}
	err := Rollback(current)
	if err == nil {
		t.Fatal("expected error when no backup exists, got nil")
	}
	if !strings.Contains(err.Error(), "no backup found") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "no backup found")
	}
}

func TestInstall(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "zeno")
	newBin := filepath.Join(dir, "zeno-new")
	if err := os.WriteFile(current, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newBin, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := Install(newBin, current); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	data, _ := os.ReadFile(current)
	if string(data) != "new" {
		t.Errorf("after install, binary = %q, want %q", string(data), "new")
	}
	backup := current + ".bak"
	data, _ = os.ReadFile(backup)
	if string(data) != "old" {
		t.Errorf("backup = %q, want %q", string(data), "old")
	}
}

func TestUpdaterFetchLatest(t *testing.T) {
	u := &Updater{Client: &mockClient{
		latest: &Release{TagName: "v1.0.9", PublishedAt: time.Now()},
	}}
	r, err := u.FetchLatest()
	if err != nil {
		t.Fatal(err)
	}
	if r.TagName != "v1.0.9" {
		t.Errorf("FetchLatest().TagName = %q, want %q", r.TagName, "v1.0.9")
	}
}

func TestUpdaterFetchLatestError(t *testing.T) {
	u := &Updater{Client: &mockClient{
		latestErr: errors.New("network error"),
	}}
	_, err := u.FetchLatest()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdaterFetchAll(t *testing.T) {
	u := &Updater{Client: &mockClient{
		all: []*Release{
			{TagName: "v1.0.9"},
			{TagName: "v1.0.8"},
		},
	}}
	releases, err := u.FetchAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) != 2 {
		t.Errorf("FetchAll() returned %d releases, want 2", len(releases))
	}
}
