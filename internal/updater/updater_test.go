package updater

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestExtractBinary(t *testing.T) {
	// Build a minimal tar.gz containing a fake "zeno" binary
	archivePath := buildTestArchive(t, "zeno", []byte("fake-binary"))

	destDir := t.TempDir()
	got, err := ExtractBinary(archivePath, destDir)
	if err != nil {
		t.Fatalf("ExtractBinary failed: %v", err)
	}
	data, _ := os.ReadFile(got)
	if string(data) != "fake-binary" {
		t.Errorf("extracted content = %q, want %q", data, "fake-binary")
	}
}

func TestExtractBinaryNotFound(t *testing.T) {
	// Archive contains a file with wrong name
	archivePath := buildTestArchive(t, "other-binary", []byte("data"))

	destDir := t.TempDir()
	_, err := ExtractBinary(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error when binary not in archive, got nil")
	}
	if !strings.Contains(err.Error(), "not found in archive") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "not found in archive")
	}
}

// buildTestArchive creates a .tar.gz archive with a single file named fileName
// containing content, and returns the archive path.
func buildTestArchive(t *testing.T, fileName string, content []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "archive-*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name:     fileName,
		Mode:     0755,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gw.Close()
	return f.Name()
}

func TestExtractBinarySymlinkSkipped(t *testing.T) {
	// Create an archive with a symlink entry named "zeno" — should be skipped
	archivePath := buildTestArchiveWithSymlink(t, "zeno", "/etc/passwd")

	destDir := t.TempDir()
	_, err := ExtractBinary(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error when only a symlink named 'zeno' is present, got nil")
	}
	if !strings.Contains(err.Error(), "not found in archive") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "not found in archive")
	}
}

func TestExtractBinaryPathTraversal(t *testing.T) {
	// Create an archive with a path-traversal entry
	archivePath := buildTestArchiveWithPath(t, "../../etc/zeno", []byte("payload"))

	destDir := t.TempDir()
	_, err := ExtractBinary(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error for path-traversal entry, got nil")
	}
	if !strings.Contains(err.Error(), "unsafe path") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "unsafe path")
	}
}

// buildTestArchiveWithSymlink creates a tar.gz with a single symlink entry.
func buildTestArchiveWithSymlink(t *testing.T, linkName, linkTarget string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "archive-*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{
		Name:     linkName,
		Typeflag: tar.TypeSymlink,
		Linkname: linkTarget,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gw.Close()
	return f.Name()
}

// buildTestArchiveWithPath creates a tar.gz with a single regular file at the given path.
func buildTestArchiveWithPath(t *testing.T, path string, content []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "archive-*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{
		Name:     path,
		Mode:     0755,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gw.Close()
	return f.Name()
}

func TestVerifyChecksumMatch(t *testing.T) {
	// Create a temp file with known content
	content := []byte("test content")
	f, err := os.CreateTemp(t.TempDir(), "archive-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Compute expected SHA256
	h := sha256.New()
	h.Write(content)
	expected := fmt.Sprintf("%x", h.Sum(nil))

	archiveName := "zeno_1.0.9_linux_amd64.tar.gz"
	checksumContent := expected + "  " + archiveName + "\n"

	// Start a local HTTP server that returns the checksums.txt
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(checksumContent))
	}))
	defer srv.Close()

	if err := VerifyChecksum(f.Name(), srv.URL+"/checksums.txt", archiveName); err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}
}

func TestVerifyChecksumMismatch(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "archive-*")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("real content"))
	f.Close()

	archiveName := "zeno_1.0.9_linux_amd64.tar.gz"
	checksumContent := "0000000000000000000000000000000000000000000000000000000000000000  " + archiveName + "\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(checksumContent))
	}))
	defer srv.Close()

	err = VerifyChecksum(f.Name(), srv.URL+"/checksums.txt", archiveName)
	if err == nil {
		t.Fatal("expected checksum mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "checksum mismatch")
	}
}

func TestVerifyChecksumMissingEntry(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "archive-*")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("content"))
	f.Close()

	archiveName := "zeno_1.0.9_linux_amd64.tar.gz"
	// checksums.txt has an entry for a different file
	checksumContent := "abc123  other_file.tar.gz\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(checksumContent))
	}))
	defer srv.Close()

	err = VerifyChecksum(f.Name(), srv.URL+"/checksums.txt", archiveName)
	if err == nil {
		t.Fatal("expected 'no checksum entry' error, got nil")
	}
	if !strings.Contains(err.Error(), "no checksum entry") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "no checksum entry")
	}
}
