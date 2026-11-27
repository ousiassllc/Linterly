package updatecheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// --- キャッシュ管理 ---

// errCachedFailure は、直近の取得が成功しておらず failureCacheTTL 内のため通信を行わなかったことを表す。
var errCachedFailure = errors.New("recent update check failed")

// cacheEntry はキャッシュファイルの内容。
// Failed が true の場合は取得を試みたが成功していない記録（LatestVersion は空）で、failureCacheTTL の間だけ有効。
type cacheEntry struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
	Failed        bool      `json:"failed,omitempty"`
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
	ttl := cacheTTL
	if entry.Failed {
		ttl = failureCacheTTL
	}
	if c.now().Sub(entry.CheckedAt) > ttl {
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
