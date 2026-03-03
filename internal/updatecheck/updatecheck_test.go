package updatecheck

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ousiassllc/linterly/internal/i18n"
)

func newTestChecker(t *testing.T, version string, serverURL string) *Checker {
	t.Helper()
	translator, err := i18n.New("en")
	require.NoError(t, err)

	c := NewChecker(version, translator)
	c.cacheDir = t.TempDir()
	if serverURL != "" {
		c.apiURL = serverURL
	}
	return c
}

// --- normalizeVersion ---

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.0.0", "v1.0.0"},
		{"1.0.0", "v1.0.0"},
		{"v0.3.1", "v0.3.1"},
		{"0.3.1", "v0.3.1"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeVersion(tt.input))
		})
	}
}

// --- detectInstallMethodFromPath ---

func TestDetectInstallMethodFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected InstallMethod
	}{
		{"npm global", "/home/user/.nvm/versions/node/v18/lib/node_modules/@linterly/cli/bin/linterly", InstallMethodNpm},
		{"npm local", "/home/user/project/node_modules/.bin/linterly", InstallMethodNpm},
		{"go install", "/home/user/go/bin/linterly", InstallMethodGoInstall},
		{"go install windows", "C:/Users/user/go/bin/linterly.exe", InstallMethodGoInstall},
		{"unknown", "/usr/local/bin/linterly", InstallMethodUnknown},
		{"homebrew", "/opt/homebrew/bin/linterly", InstallMethodUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectInstallMethodFromPath(tt.path))
		})
	}
}

// --- resolveCurrentVersion ---

func TestResolveCurrentVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		buildInfo   *debug.BuildInfo
		buildInfoOK bool
		expected    string
	}{
		{
			name:     "normal version",
			version:  "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:    "dev with build info",
			version: "dev",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "v0.3.1"},
			},
			buildInfoOK: true,
			expected:    "v0.3.1",
		},
		{
			name:    "dev with devel",
			version: "dev",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
			},
			buildInfoOK: true,
			expected:    "",
		},
		{
			name:        "dev no build info",
			version:     "dev",
			buildInfoOK: false,
			expected:    "",
		},
		{
			name:        "empty version",
			version:     "",
			buildInfoOK: false,
			expected:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Checker{
				currentVersion: tt.version,
				readBuildInfo: func() (*debug.BuildInfo, bool) {
					return tt.buildInfo, tt.buildInfoOK
				},
			}
			assert.Equal(t, tt.expected, c.resolveCurrentVersion())
		})
	}
}

// --- Cache ---

func TestCache_RoundTrip(t *testing.T) {
	c := &Checker{cacheDir: t.TempDir(), now: time.Now}
	entry := &cacheEntry{LatestVersion: "v0.4.0", CheckedAt: time.Now()}
	require.NoError(t, c.writeCache(entry))

	got, err := c.readCache()
	require.NoError(t, err)
	assert.Equal(t, "v0.4.0", got.LatestVersion)
}

func TestCache_Expired(t *testing.T) {
	c := &Checker{cacheDir: t.TempDir(), now: time.Now}
	entry := &cacheEntry{LatestVersion: "v0.4.0", CheckedAt: time.Now().Add(-25 * time.Hour)}
	require.NoError(t, c.writeCache(entry))

	_, err := c.readCache()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache expired")
}

func TestCache_FileNotFound(t *testing.T) {
	c := &Checker{cacheDir: t.TempDir(), now: time.Now}
	_, err := c.readCache()
	assert.Error(t, err)
}

func TestCache_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	c := &Checker{cacheDir: dir, now: time.Now}
	require.NoError(t, os.WriteFile(filepath.Join(dir, cacheFileName), []byte("{bad json}"), 0o644))

	_, err := c.readCache()
	assert.Error(t, err)
}

func TestCache_EmptyCacheDir(t *testing.T) {
	c := &Checker{cacheDir: ""}
	assert.NoError(t, c.writeCache(&cacheEntry{LatestVersion: "v1.0.0", CheckedAt: time.Now()}))

	_, err := c.readCache()
	assert.Error(t, err)
}

func TestCache_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	c := &Checker{cacheDir: dir, now: time.Now}
	entry := &cacheEntry{LatestVersion: "v0.4.0", CheckedAt: time.Now()}
	require.NoError(t, c.writeCache(entry))

	got, err := c.readCache()
	require.NoError(t, err)
	assert.Equal(t, "v0.4.0", got.LatestVersion)
}

// --- fetchLatest ---

func TestFetchLatest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.0", server.URL)
	version, err := c.fetchLatest(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v0.4.0", version)
}

func TestFetchLatest_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.0", server.URL)
	_, err := c.fetchLatest(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
}

func TestFetchLatest_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid"))
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.0", server.URL)
	_, err := c.fetchLatest(context.Background())
	assert.Error(t, err)
}

func TestFetchLatest_EmptyTagName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: ""})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.0", server.URL)
	_, err := c.fetchLatest(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty tag_name")
}

func TestFetchLatest_UserAgent(t *testing.T) {
	var gotUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.0", server.URL)
	_, err := c.fetchLatest(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "linterly/v0.3.0", gotUserAgent)
}
