package updatecheck

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 失敗しているサーバーを立て、受けたリクエスト数を返せるようにする。
func newFailingServer(t *testing.T, hits *atomic.Int32, healthy *atomic.Bool) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if !healthy.Load() {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.4.0"})
	}))
	t.Cleanup(server.Close)
	return server
}

// Issue #78: 取得に失敗した後は、TTL 内の 2 回目以降は通信せず即座に失敗を返す。
func TestCheck_FailureIsNegativelyCached(t *testing.T) {
	var hits atomic.Int32
	var healthy atomic.Bool
	server := newFailingServer(t, &hits, &healthy)

	c := newTestChecker(t, "v0.3.1", server.URL)
	now := time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return now }

	_, err := c.Check(context.Background())
	require.Error(t, err)
	assert.Equal(t, int32(1), hits.Load())

	// サーバーが復旧しても TTL 内は通信しない
	healthy.Store(true)
	now = now.Add(failureCacheTTL - time.Minute)
	_, err = c.Check(context.Background())
	require.ErrorIs(t, err, errCachedFailure)
	assert.Equal(t, int32(1), hits.Load(), "TTL 内は API を呼ばない")
}

// 応答を待たずに打ち切られた（プロセス終了を想定）場合も、TTL 内は再試行しない。
func TestCheck_AbandonedAttemptIsNotRetried(t *testing.T) {
	var hits atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		<-release
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })

	c := newTestChecker(t, "v0.3.1", server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := c.Check(ctx)
	require.Error(t, err)

	_, err = c.Check(context.Background())
	require.ErrorIs(t, err, errCachedFailure)
	assert.Equal(t, int32(1), hits.Load())
}

func TestCheck_FailureCacheExpires(t *testing.T) {
	var hits atomic.Int32
	var healthy atomic.Bool
	server := newFailingServer(t, &hits, &healthy)

	c := newTestChecker(t, "v0.3.1", server.URL)
	c.executablePath = func() (string, error) { return "/usr/local/bin/linterly", nil }
	now := time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return now }

	_, err := c.Check(context.Background())
	require.Error(t, err)

	// TTL 経過後は再び通信し、成功すれば通常のキャッシュ（24 時間）になる
	healthy.Store(true)
	now = now.Add(failureCacheTTL + time.Minute)
	result, err := c.Check(context.Background())
	require.NoError(t, err)
	assert.True(t, result.UpdateAvailable)
	assert.Equal(t, int32(2), hits.Load())

	now = now.Add(2 * failureCacheTTL)
	_, err = c.Check(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int32(2), hits.Load(), "成功結果は 24 時間キャッシュされる")
}

// 失敗記録のない従来形式のキャッシュは、従来どおり 24 時間有効。
func TestCache_LegacyEntryWithoutFailedField(t *testing.T) {
	c := newTestChecker(t, "v0.3.1", "")
	require.NoError(t, c.writeCache(&cacheEntry{LatestVersion: "v0.4.0", CheckedAt: time.Now().Add(-2 * time.Hour)}))

	entry, err := c.readCache()
	require.NoError(t, err)
	assert.False(t, entry.Failed)
	assert.Equal(t, "v0.4.0", entry.LatestVersion)
}
