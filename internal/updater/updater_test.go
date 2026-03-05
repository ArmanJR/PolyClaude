package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"v0.2.0", "0.1.0", true},
		{"v1.0.0", "0.9.9", true},
		{"v0.1.1", "0.1.0", true},
		{"v0.1.0", "0.1.0", false},
		{"v0.1.0", "0.2.0", false},
		{"v0.0.1", "0.1.0", false},
		{"v2.0.0", "1.99.99", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.latest, tt.current), func(t *testing.T) {
			got := isNewer(tt.latest, tt.current)
			if got != tt.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}

func TestCheckCachedDevBuild(t *testing.T) {
	result := CheckCached("dev", t.TempDir())
	if result != "" {
		t.Errorf("expected empty string for dev build, got %q", result)
	}
}

func TestCheckCachedUsesFreshCache(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, cacheFile)

	// Write a fresh cache entry with a version that is "newer" than current
	cacheData := fmt.Sprintf("%d\nv99.0.0\n", time.Now().Unix())
	if err := os.WriteFile(cachePath, []byte(cacheData), 0o644); err != nil {
		t.Fatal(err)
	}

	result := CheckCached("0.1.0", dir)
	if result != "v99.0.0" {
		t.Errorf("expected v99.0.0 from cache, got %q", result)
	}
}

func TestCheckCachedNoUpdateAvailable(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, cacheFile)

	// Cache says latest is same as current
	cacheData := fmt.Sprintf("%d\nv0.1.0\n", time.Now().Unix())
	if err := os.WriteFile(cachePath, []byte(cacheData), 0o644); err != nil {
		t.Fatal(err)
	}

	result := CheckCached("0.1.0", dir)
	if result != "" {
		t.Errorf("expected empty string when up to date, got %q", result)
	}
}
