package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"github.com/ousiassllc/linterly/internal/i18n"
)

const (
	githubReleasesURL = "https://api.github.com/repos/ousiassllc/linterly/releases/latest"
	releasesPageURL   = "https://github.com/ousiassllc/linterly/releases/latest"
	npmUpdateCmd      = "npm update -g @linterly/cli"
	goInstallCmd      = "go install github.com/ousiassllc/linterly/cmd/linterly@latest"
	cacheTTL          = 24 * time.Hour
	cacheFileName     = "latest-version.json"
	httpTimeout       = 3 * time.Second
)

// CheckResult は更新チェックの結果。
type CheckResult struct {
	UpdateAvailable bool
	VersionUnknown  bool
	CurrentVersion  string
	LatestVersion   string
	Message         string
}

// InstallMethod はインストール経路。
type InstallMethod int

const (
	InstallMethodUnknown InstallMethod = iota
	InstallMethodNpm
	InstallMethodGoInstall
)

// Checker はバージョン更新チェッカー。
type Checker struct {
	currentVersion string
	translator     *i18n.Translator
	cacheDir       string
	apiURL         string
	httpClient     *http.Client
	now            func() time.Time
	readBuildInfo  func() (*debug.BuildInfo, bool)
	executablePath func() (string, error)
}

// NewChecker は Checker を生成する。
func NewChecker(currentVersion string, translator *i18n.Translator) *Checker {
	cacheDir := ""
	if dir, err := os.UserCacheDir(); err == nil {
		cacheDir = filepath.Join(dir, "linterly")
	}

	return &Checker{
		currentVersion: currentVersion,
		translator:     translator,
		cacheDir:       cacheDir,
		apiURL:         githubReleasesURL,
		httpClient: &http.Client{
			Timeout: httpTimeout,
			Transport: &userAgentTransport{
				base:      http.DefaultTransport,
				userAgent: "linterly/" + currentVersion,
			},
		},
		now:            time.Now,
		readBuildInfo:  debug.ReadBuildInfo,
		executablePath: os.Executable,
	}
}

// Check は最新バージョンをチェックし結果を返す。
func (c *Checker) Check(ctx context.Context) (*CheckResult, error) {
	current := c.resolveCurrentVersion()
	if current == "" {
		return &CheckResult{
			VersionUnknown: true,
			Message:        c.buildUnknownMessage(),
		}, nil
	}

	current = normalizeVersion(current)

	latest, err := c.getLatestVersion(ctx)
	if err != nil {
		return nil, err
	}

	latest = normalizeVersion(latest)

	if !semver.IsValid(current) || !semver.IsValid(latest) {
		return &CheckResult{CurrentVersion: current, LatestVersion: latest}, nil
	}

	if semver.Compare(latest, current) > 0 {
		method := c.detectInstallMethod()
		return &CheckResult{
			UpdateAvailable: true,
			CurrentVersion:  current,
			LatestVersion:   latest,
			Message:         c.buildUpdateMessage(current, latest, method),
		}, nil
	}

	return &CheckResult{CurrentVersion: current, LatestVersion: latest}, nil
}

// DetectInstallMethod は実行バイナリのパスからインストール経路を推定する。
func DetectInstallMethod() InstallMethod {
	exe, err := os.Executable()
	if err != nil {
		return InstallMethodUnknown
	}
	return detectInstallMethodFromPath(exe)
}

func (c *Checker) resolveCurrentVersion() string {
	if c.currentVersion != "dev" && c.currentVersion != "" {
		return c.currentVersion
	}
	info, ok := c.readBuildInfo()
	if !ok || info == nil {
		return ""
	}
	v := info.Main.Version
	if v == "" || v == "(devel)" {
		return ""
	}
	return v
}

func (c *Checker) getLatestVersion(ctx context.Context) (string, error) {
	if entry, err := c.readCache(); err == nil {
		return entry.LatestVersion, nil
	}

	latest, err := c.fetchLatest(ctx)
	if err != nil {
		return "", err
	}

	_ = c.writeCache(&cacheEntry{
		LatestVersion: latest,
		CheckedAt:     c.now(),
	})

	return latest, nil
}

func (c *Checker) detectInstallMethod() InstallMethod {
	exe, err := c.executablePath()
	if err != nil {
		return InstallMethodUnknown
	}
	return detectInstallMethodFromPath(exe)
}

func detectInstallMethodFromPath(path string) InstallMethod {
	normalized := filepath.ToSlash(path)
	if strings.Contains(normalized, "node_modules") {
		return InstallMethodNpm
	}
	if strings.Contains(normalized, "go/bin") {
		return InstallMethodGoInstall
	}
	return InstallMethodUnknown
}

func normalizeVersion(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// --- キャッシュ管理 ---

type cacheEntry struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

func (c *Checker) readCache() (*cacheEntry, error) {
	if c.cacheDir == "" {
		return nil, fmt.Errorf("cache dir not set")
	}
	data, err := os.ReadFile(filepath.Join(c.cacheDir, cacheFileName))
	if err != nil {
		return nil, err
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if c.now().Sub(entry.CheckedAt) > cacheTTL {
		return nil, fmt.Errorf("cache expired")
	}
	return &entry, nil
}

func (c *Checker) writeCache(entry *cacheEntry) error {
	if c.cacheDir == "" {
		return nil
	}
	if err := os.MkdirAll(c.cacheDir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.cacheDir, cacheFileName), data, 0o644)
}

// --- GitHub API ---

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func (c *Checker) fetchLatest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release response: %w", err)
	}
	if release.TagName == "" {
		return "", fmt.Errorf("empty tag_name in release response")
	}

	return release.TagName, nil
}

// --- メッセージ生成 ---

func (c *Checker) buildUpdateMessage(current, latest string, method InstallMethod) string {
	msg := c.translator.T("update.available", current, latest)
	switch method {
	case InstallMethodNpm:
		msg += "\n" + c.translator.T("update.run", npmUpdateCmd)
	case InstallMethodGoInstall:
		msg += "\n" + c.translator.T("update.run", goInstallCmd)
	default:
		msg += "\n" + c.translator.T("update.visit", releasesPageURL)
	}
	return msg
}

func (c *Checker) buildUnknownMessage() string {
	msg := c.translator.T("update.unknown_version")
	msg += "\n" + c.translator.T("update.unknown_visit", releasesPageURL)
	return msg
}

// --- HTTP トランスポート ---

type userAgentTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", t.userAgent)
	return t.base.RoundTrip(req)
}
