package updater

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubRepo = "zenlayer/zenlayercloud-cli"
	binaryName = "zeno"
)

var httpCli = &http.Client{Timeout: 60 * time.Second}

// Release represents a single GitHub release entry.
type Release struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
}

// ReleaseClient fetches release information from a remote source.
type ReleaseClient interface {
	FetchLatest() (*Release, error)
	FetchAll() ([]*Release, error)
}

// Updater delegates release fetching to a ReleaseClient.
type Updater struct {
	Client ReleaseClient
}

// New returns an Updater backed by the real GitHub API.
func New() *Updater {
	return &Updater{Client: &httpClient{}}
}

// FetchLatest returns the latest release.
func (u *Updater) FetchLatest() (*Release, error) {
	return u.Client.FetchLatest()
}

// FetchAll returns all releases.
func (u *Updater) FetchAll() ([]*Release, error) {
	return u.Client.FetchAll()
}

// CompareVersions compares two semver strings (with or without leading "v").
// Returns -1 if a < b, 0 if equal, 1 if a > b.
func CompareVersions(a, b string) int {
	pa := parseVersion(a)
	pb := parseVersion(b)
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	var nums [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		fmt.Sscanf(parts[i], "%d", &nums[i])
	}
	return nums
}

// ArchiveName returns the release archive filename for the current platform.
// Darwin always uses arch "all" (universal binary from goreleaser).
func ArchiveName(tag string) string {
	ver := strings.TrimPrefix(tag, "v")
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos == "darwin" {
		goarch = "all"
	}
	return fmt.Sprintf("%s_%s_%s_%s.tar.gz", binaryName, ver, goos, goarch)
}

// DownloadURL returns the full URL for a release archive asset.
func DownloadURL(tag, archive string) string {
	return fmt.Sprintf(
		"https://github.com/%s/releases/download/%s/%s",
		githubRepo, tag, archive,
	)
}

// ChecksumURL returns the checksums.txt URL for a release.
func ChecksumURL(tag string) string {
	return fmt.Sprintf(
		"https://github.com/%s/releases/download/%s/checksums.txt",
		githubRepo, tag,
	)
}

// Download fetches url into a temp file and returns its path.
// The caller must remove the file when done.
func Download(url string) (string, error) {
	resp, err := httpCli.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d for %s", resp.StatusCode, url)
	}
	f, err := os.CreateTemp("", "zeno-download-*")
	if err != nil {
		return "", err
	}
	defer f.Close()
	const maxDownloadSize int64 = 100 << 20 // 100 MB
	lr := io.LimitReader(resp.Body, maxDownloadSize+1)
	n, copyErr := io.Copy(f, lr)
	if copyErr != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("download failed: %w", copyErr)
	}
	if n > maxDownloadSize {
		os.Remove(f.Name())
		return "", fmt.Errorf("download exceeded 100 MB limit for %s", url)
	}
	return f.Name(), nil
}

// VerifyChecksum checks the SHA256 of archivePath against the entry in checksumURL for archiveName.
func VerifyChecksum(archivePath, checksumURL, archiveName string) error {
	resp, err := httpCli.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksums not found: HTTP %d", resp.StatusCode)
	}

	var expected string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[1] == archiveName {
			expected = fields[0]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read checksums: %w", err)
	}
	if expected == "" {
		return fmt.Errorf("no checksum entry for %s in checksums.txt", archiveName)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s (expected %s, got %s)", archiveName, expected, actual)
	}
	return nil
}

// ExtractBinary extracts the zeno binary from a .tar.gz archive into destDir.
// Returns the path to the extracted binary.
func ExtractBinary(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("invalid gzip archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("corrupt archive: %w", err)
		}
		// 跳过 symlink/hardlink 防止攻击
		if hdr.Typeflag == tar.TypeSymlink || hdr.Typeflag == tar.TypeLink {
			continue
		}
		// 拒绝含 ".." 的路径
		if strings.Contains(filepath.Clean(hdr.Name), "..") {
			return "", fmt.Errorf("archive contains unsafe path: %s", hdr.Name)
		}
		if filepath.Base(hdr.Name) == binaryName && hdr.Typeflag == tar.TypeReg {
			dest := filepath.Join(destDir, binaryName)
			out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return "", err
			}
			const maxExtractSize int64 = 100 << 20 // 100 MB
			lr := io.LimitReader(tr, maxExtractSize+1)
			n, copyErr := io.Copy(out, lr)
			out.Close()
			if copyErr != nil {
				os.Remove(dest)
				return "", fmt.Errorf("failed to extract binary: %w", copyErr)
			}
			if n > maxExtractSize {
				os.Remove(dest)
				return "", fmt.Errorf("extracted binary exceeds 100 MB limit")
			}
			return dest, nil
		}
	}
	return "", fmt.Errorf("binary '%s' not found in archive", binaryName)
}

// Install replaces currentBinary with newBinary, backing up current as <path>.bak.
// Falls back to copy+delete if os.Rename fails across device boundaries.
func Install(newBinary, currentBinary string) error {
	backup := currentBinary + ".bak"
	if err := copyFile(currentBinary, backup); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	if err := os.Rename(newBinary, currentBinary); err != nil {
		// Fallback: copy then delete (handles cross-device links)
		if err2 := copyFile(newBinary, currentBinary); err2 != nil {
			return fmt.Errorf("failed to install new binary: %w (rename: %v)", err2, err)
		}
		os.Remove(newBinary)
	}
	return nil
}

// Rollback restores <binaryPath>.bak to binaryPath.
func Rollback(binaryPath string) error {
	backup := binaryPath + ".bak"
	if _, err := os.Stat(backup); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backup)
	}
	if err := os.Rename(backup, binaryPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// httpClient implements ReleaseClient against the GitHub REST API.
type httpClient struct{}

func (c *httpClient) FetchLatest() (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	resp, err := httpCli.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("GitHub API rate limit exceeded, please try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release: HTTP %d", resp.StatusCode)
	}
	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}
	return &r, nil
}

func (c *httpClient) FetchAll() ([]*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100", githubRepo)
	resp, err := httpCli.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("GitHub API rate limit exceeded, please try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases: HTTP %d", resp.StatusCode)
	}
	var releases []*Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}
	return releases, nil
}
