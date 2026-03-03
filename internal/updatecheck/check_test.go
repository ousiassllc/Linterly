package updatecheck

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ousiassllc/linterly/internal/i18n"
)

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
