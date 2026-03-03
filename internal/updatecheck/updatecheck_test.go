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
	c := &Checker{
		cacheDir: t.TempDir(),
		now:      time.Now,
	}

	entry := &cacheEntry{
		LatestVersion: "v0.4.0",
		CheckedAt:     time.Now(),
	}

	require.NoError(t, c.writeCache(entry))

	got, err := c.readCache()
	require.NoError(t, err)
	assert.Equal(t, "v0.4.0", got.LatestVersion)
}

func TestCache_Expired(t *testing.T) {
	c := &Checker{
		cacheDir: t.TempDir(),
		now:      time.Now,
	}

	entry := &cacheEntry{
		LatestVersion: "v0.4.0",
		CheckedAt:     time.Now().Add(-25 * time.Hour),
	}

	require.NoError(t, c.writeCache(entry))

	_, err := c.readCache()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache expired")
}

func TestCache_FileNotFound(t *testing.T) {
	c := &Checker{
		cacheDir: t.TempDir(),
		now:      time.Now,
	}

	_, err := c.readCache()
	assert.Error(t, err)
}

func TestCache_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	c := &Checker{
		cacheDir: dir,
		now:      time.Now,
	}

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
	c := &Checker{
		cacheDir: dir,
		now:      time.Now,
	}

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

// --- Check integration ---

func TestCheck_UpdateAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.1", server.URL)
	c.executablePath = func() (string, error) { return "/usr/local/bin/linterly", nil }

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.UpdateAvailable)
	assert.False(t, result.VersionUnknown)
	assert.Equal(t, "v0.3.1", result.CurrentVersion)
	assert.Equal(t, "v0.4.0", result.LatestVersion)
	assert.Contains(t, result.Message, "v0.3.1")
	assert.Contains(t, result.Message, "v0.4.0")
	assert.Contains(t, result.Message, releasesPageURL)
}

func TestCheck_UpdateAvailable_GoInstall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.1", server.URL)
	c.executablePath = func() (string, error) { return "/home/user/go/bin/linterly", nil }

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.UpdateAvailable)
	assert.Contains(t, result.Message, goInstallCmd)
}

func TestCheck_UpdateAvailable_Npm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.1", server.URL)
	c.executablePath = func() (string, error) { return "/home/user/node_modules/.bin/linterly", nil }

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.UpdateAvailable)
	assert.Contains(t, result.Message, npmUpdateCmd)
}

func TestCheck_NoUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.3.1"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.1", server.URL)

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.False(t, result.UpdateAvailable)
	assert.False(t, result.VersionUnknown)
	assert.Empty(t, result.Message)
}

func TestCheck_CacheHit(t *testing.T) {
	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.1", server.URL)
	c.executablePath = func() (string, error) { return "/usr/local/bin/linterly", nil }

	// Write a valid cache
	require.NoError(t, c.writeCache(&cacheEntry{
		LatestVersion: "v0.4.0",
		CheckedAt:     time.Now(),
	}))

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.UpdateAvailable)
	assert.False(t, apiCalled, "API should not be called when cache is valid")
}

func TestCheck_CacheExpired(t *testing.T) {
	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.1", server.URL)
	c.executablePath = func() (string, error) { return "/usr/local/bin/linterly", nil }

	// Write an expired cache
	require.NoError(t, c.writeCache(&cacheEntry{
		LatestVersion: "v0.3.1",
		CheckedAt:     time.Now().Add(-25 * time.Hour),
	}))

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.UpdateAvailable)
	assert.True(t, apiCalled, "API should be called when cache is expired")
}

func TestCheck_VersionUnknown(t *testing.T) {
	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
	}))
	defer server.Close()

	c := newTestChecker(t, "dev", server.URL)
	c.readBuildInfo = func() (*debug.BuildInfo, bool) { return nil, false }

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.VersionUnknown)
	assert.False(t, result.UpdateAvailable)
	assert.Contains(t, result.Message, "Unable to determine")
	assert.Contains(t, result.Message, releasesPageURL)
	assert.False(t, apiCalled, "API should not be called when version is unknown")
}

func TestCheck_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.1", server.URL)

	_, err := c.Check(context.Background())
	assert.Error(t, err)
}

func TestCheck_VersionWithoutPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	defer server.Close()

	c := newTestChecker(t, "0.3.1", server.URL)
	c.executablePath = func() (string, error) { return "/usr/local/bin/linterly", nil }

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.UpdateAvailable)
}

func TestCheck_InvalidSemver(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "not-a-version"})
	}))
	defer server.Close()

	c := newTestChecker(t, "v0.3.1", server.URL)

	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.False(t, result.UpdateAvailable)
	assert.False(t, result.VersionUnknown)
	assert.Empty(t, result.Message)
}

// --- Message building ---

func TestBuildUpdateMessage_English(t *testing.T) {
	translator, err := i18n.New("en")
	require.NoError(t, err)

	tests := []struct {
		name     string
		method   InstallMethod
		contains string
	}{
		{"go install", InstallMethodGoInstall, goInstallCmd},
		{"npm", InstallMethodNpm, npmUpdateCmd},
		{"unknown", InstallMethodUnknown, releasesPageURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Checker{translator: translator}
			msg := c.buildUpdateMessage("v0.3.1", "v0.4.0", tt.method)
			assert.Contains(t, msg, "v0.3.1")
			assert.Contains(t, msg, "v0.4.0")
			assert.Contains(t, msg, tt.contains)
		})
	}
}

func TestBuildUpdateMessage_Japanese(t *testing.T) {
	translator, err := i18n.New("ja")
	require.NoError(t, err)

	c := &Checker{translator: translator}
	msg := c.buildUpdateMessage("v0.3.1", "v0.4.0", InstallMethodGoInstall)
	assert.Contains(t, msg, "新しいバージョンが利用可能です")
	assert.Contains(t, msg, goInstallCmd)
}

func TestBuildUnknownMessage_English(t *testing.T) {
	translator, err := i18n.New("en")
	require.NoError(t, err)

	c := &Checker{translator: translator}
	msg := c.buildUnknownMessage()
	assert.Contains(t, msg, "Unable to determine")
	assert.Contains(t, msg, releasesPageURL)
}

func TestBuildUnknownMessage_Japanese(t *testing.T) {
	translator, err := i18n.New("ja")
	require.NoError(t, err)

	c := &Checker{translator: translator}
	msg := c.buildUnknownMessage()
	assert.Contains(t, msg, "バージョンを特定できません")
	assert.Contains(t, msg, releasesPageURL)
}
