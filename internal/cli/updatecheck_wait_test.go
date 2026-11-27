package cli

import (
	"context"
	"testing"
	"time"

	"github.com/ousiassllc/linterly/internal/updatecheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- 短時間コマンドでも更新チェックの完了を（上限付きで）待つ ---

// stubUpdateCheck は更新チェックの状態とフック（フラグ・結果チャネル・実行関数・待機上限）を差し替え、
// テスト終了時に元へ戻す。
func stubUpdateCheck(t *testing.T, disabled bool, wait time.Duration, run func(context.Context, string) (*updatecheck.CheckResult, error)) {
	t.Helper()
	oldFlag, oldResult, oldRun, oldWait := flagNoUpdateCheck, updateResult, runUpdateCheck, updateWaitTimeout
	t.Cleanup(func() {
		flagNoUpdateCheck, updateResult, runUpdateCheck, updateWaitTimeout = oldFlag, oldResult, oldRun, oldWait
	})
	t.Setenv("LINTERLY_NO_UPDATE_CHECK", "")
	t.Setenv("LINTERLY_CONFIG", "")
	flagNoUpdateCheck, updateResult, runUpdateCheck, updateWaitTimeout = disabled, nil, run, wait
}

// startUpdateCheck 後にチェックが遅れて完了しても、上限時間内なら通知を表示する。
func TestPrintUpdateNotice_WaitsForSlowCheck(t *testing.T) {
	release := make(chan struct{})
	stubUpdateCheck(t, false, 10*time.Second, func(context.Context, string) (*updatecheck.CheckResult, error) {
		<-release
		return &updatecheck.CheckResult{UpdateAvailable: true, Message: "update available: v9.9.9"}, nil
	})

	startUpdateCheck()
	require.NotNil(t, updateResult)
	// printUpdateNotice が待機に入った後でチェックを完了させる
	time.AfterFunc(20*time.Millisecond, func() { close(release) })

	assert.Contains(t, helperCaptureStderr(t, printUpdateNotice), "update available: v9.9.9")
}

// チェックが上限時間内に終わらなければ、通知せずに戻る（メインコマンドを無制限に遅らせない）。
func TestPrintUpdateNotice_GivesUpAfterTimeout(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	stubUpdateCheck(t, false, 10*time.Millisecond, func(context.Context, string) (*updatecheck.CheckResult, error) {
		<-block
		return &updatecheck.CheckResult{UpdateAvailable: true, Message: "never shown"}, nil
	})

	startUpdateCheck()
	require.NotNil(t, updateResult)

	assert.Empty(t, helperCaptureStderr(t, printUpdateNotice))
}

// 無効化されている場合は runUpdateCheck（通信）も待機も発生しない。
func TestStartUpdateCheck_DisabledDoesNotRun(t *testing.T) {
	called := make(chan struct{}, 1)
	stubUpdateCheck(t, true, time.Second, func(context.Context, string) (*updatecheck.CheckResult, error) {
		called <- struct{}{}
		return nil, nil
	})

	startUpdateCheck()
	assert.Nil(t, updateResult)
	assert.Empty(t, helperCaptureStderr(t, printUpdateNotice))
	assert.Empty(t, called)
}
