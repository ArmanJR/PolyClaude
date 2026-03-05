package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	repoOwner     = "ArmanJR"
	repoName      = "PolyClaude"
	releaseAPIURL = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
	checkInterval = 24 * time.Hour
	httpTimeout   = 5 * time.Second
	cacheFile     = ".update-check"
)

type releaseResponse struct {
	TagName string `json:"tag_name"`
}

// CheckCached checks for a newer version using a 24h file cache.
// Returns the latest version string if an update is available, empty string otherwise.
// Errors are silently logged — this should never block the user.
func CheckCached(currentVersion, cacheDir string) string {
	if currentVersion == "dev" {
		return ""
	}

	cachePath := filepath.Join(cacheDir, cacheFile)

	// Read cache
	if data, err := os.ReadFile(cachePath); err == nil {
		lines := strings.SplitN(string(data), "\n", 2)
		if len(lines) == 2 {
			if ts, err := strconv.ParseInt(lines[0], 10, 64); err == nil {
				if time.Since(time.Unix(ts, 0)) < checkInterval {
					latest := strings.TrimSpace(lines[1])
					if isNewer(latest, currentVersion) {
						return latest
					}
					return ""
				}
			}
		}
	}

	// Cache miss or stale — fetch from API
	latest, err := fetchLatestVersion()
	if err != nil {
		slog.Debug("update check failed", "error", err)
		return ""
	}

	// Write cache
	_ = os.MkdirAll(cacheDir, 0o755)
	cacheData := fmt.Sprintf("%d\n%s\n", time.Now().Unix(), latest)
	if err := os.WriteFile(cachePath, []byte(cacheData), 0o644); err != nil {
		slog.Debug("failed to write update cache", "error", err)
	}

	if isNewer(latest, currentVersion) {
		return latest
	}
	return ""
}

// SelfUpdate downloads the latest release and replaces the current binary.
func SelfUpdate(currentVersion string) error {
	if currentVersion == "dev" {
		return fmt.Errorf("cannot update a dev build — install from a release first")
	}

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("checking latest version: %w", err)
	}

	if !isNewer(latest, currentVersion) {
		fmt.Printf("Already up to date (v%s).\n", currentVersion)
		return nil
	}

	fmt.Printf("Updating v%s -> %s...\n", currentVersion, latest)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	downloadURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/%s/%s_%s_%s_%s.tar.gz",
		repoOwner, repoName, latest, repoName, strings.TrimPrefix(latest, "v"), runtime.GOOS, runtime.GOARCH,
	)

	slog.Info("downloading release", "url", downloadURL)

	tmpDir, err := os.MkdirTemp("", "polyclaude-update-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("downloading release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d (is %s/%s supported?)", resp.StatusCode, runtime.GOOS, runtime.GOARCH)
	}

	// Extract binary from tarball
	newBin, err := extractBinary(resp.Body, tmpDir)
	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	// Replace current binary: rename old, move new, remove old
	oldPath := execPath + ".old"
	if err := os.Rename(execPath, oldPath); err != nil {
		return fmt.Errorf("backing up current binary: %w", err)
	}

	if err := copyFile(newBin, execPath, 0o755); err != nil {
		// Attempt rollback
		_ = os.Rename(oldPath, execPath)
		return fmt.Errorf("installing new binary: %w", err)
	}

	_ = os.Remove(oldPath)

	fmt.Printf("Updated to %s.\n", latest)
	return nil
}

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(releaseAPIURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no tag_name in release response")
	}

	return release.TagName, nil
}

func extractBinary(r io.Reader, destDir string) (string, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if filepath.Base(hdr.Name) == "polyclaude" && hdr.Typeflag == tar.TypeReg {
			outPath := filepath.Join(destDir, "polyclaude")
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0o755)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return outPath, nil
		}
	}

	return "", fmt.Errorf("polyclaude binary not found in archive")
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// isNewer returns true if latest (e.g. "v0.2.0") is newer than current (e.g. "0.1.0").
func isNewer(latest, current string) bool {
	parse := func(s string) [3]int {
		s = strings.TrimPrefix(s, "v")
		parts := strings.SplitN(s, ".", 3)
		var v [3]int
		for i := 0; i < 3 && i < len(parts); i++ {
			v[i], _ = strconv.Atoi(parts[i])
		}
		return v
	}

	l := parse(latest)
	c := parse(current)

	for i := 0; i < 3; i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}
