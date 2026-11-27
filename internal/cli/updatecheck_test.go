package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ousiassllc/linterly/internal/updatecheck"
)

// helperCaptureStderr は fn 実行中の stderr 出力をキャプチャして返すヘルパー。
func helperCaptureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	defer func() { os.Stderr = old }()
	fn()
	w.Close()
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(data)
}

// --- startUpdateCheck disable conditions ---

func TestStartUpdateCheck_DisabledByFlag(t *testing.T) {
	oldFlag := flagNoUpdateCheck
	oldResult := updateResult
	defer func() {
		flagNoUpdateCheck = oldFlag
		updateResult = oldResult
	}()

	flagNoUpdateCheck = true
	updateResult = nil

	startUpdateCheck()
	assert.Nil(t, updateResult)
}

func TestStartUpdateCheck_DisabledByEnv(t *testing.T) {
	oldResult := updateResult
	defer func() { updateResult = oldResult }()

	t.Setenv("LINTERLY_NO_UPDATE_CHECK", "1")
	updateResult = nil

	startUpdateCheck()
	assert.Nil(t, updateResult)
}

func TestStartUpdateCheck_DisabledByConfig(t *testing.T) {
	oldCfg := configFile
	oldResult := updateResult
	defer func() {
		configFile = oldCfg
		updateResult = oldResult
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 300\nupdate_check: false\n")

	configFile = cfgPath
	updateResult = nil

	startUpdateCheck()
	assert.Nil(t, updateResult)
}

func TestStartUpdateCheck_NotDisabled(t *testing.T) {
	oldFlag := flagNoUpdateCheck
	oldResult := updateResult
	defer func() {
		flagNoUpdateCheck = oldFlag
		updateResult = oldResult
	}()

	flagNoUpdateCheck = false
	t.Setenv("LINTERLY_NO_UPDATE_CHECK", "")
	updateResult = nil

	startUpdateCheck()
	// goroutine が起動され、チャネルが作成される
	assert.NotNil(t, updateResult)
}

// --- printUpdateNotice ---

func TestPrintUpdateNotice_Nil(t *testing.T) {
	oldResult := updateResult
	defer func() { updateResult = oldResult }()

	updateResult = nil

	output := helperCaptureStderr(t, func() {
		printUpdateNotice()
	})
	assert.Empty(t, output)
}

func TestPrintUpdateNotice_UpdateAvailable(t *testing.T) {
	oldResult := updateResult
	defer func() { updateResult = oldResult }()

	ch := make(chan *updatecheck.CheckResult, 1)
	ch <- &updatecheck.CheckResult{
		UpdateAvailable: true,
		Message:         "A new version is available: v0.4.0",
	}
	updateResult = ch

	output := helperCaptureStderr(t, func() {
		printUpdateNotice()
	})
	assert.Contains(t, output, "A new version is available: v0.4.0")
}

func TestPrintUpdateNotice_VersionUnknown(t *testing.T) {
	oldResult := updateResult
	defer func() { updateResult = oldResult }()

	ch := make(chan *updatecheck.CheckResult, 1)
	ch <- &updatecheck.CheckResult{
		VersionUnknown: true,
		Message:        "Unable to determine current version",
	}
	updateResult = ch

	output := helperCaptureStderr(t, func() {
		printUpdateNotice()
	})
	assert.Contains(t, output, "Unable to determine current version")
}

func TestPrintUpdateNotice_NoUpdate(t *testing.T) {
	oldResult := updateResult
	defer func() { updateResult = oldResult }()

	ch := make(chan *updatecheck.CheckResult, 1)
	ch <- &updatecheck.CheckResult{
		UpdateAvailable: false,
		VersionUnknown:  false,
	}
	updateResult = ch

	output := helperCaptureStderr(t, func() {
		printUpdateNotice()
	})
	assert.Empty(t, output)
}

func TestPrintUpdateNotice_NotCompleted(t *testing.T) {
	oldResult, oldWait := updateResult, updateWaitTimeout
	defer func() { updateResult, updateWaitTimeout = oldResult, oldWait }()
	updateWaitTimeout = 10 * time.Millisecond

	// 空のバッファ付きチャネル（何も送信されていない）
	ch := make(chan *updatecheck.CheckResult, 1)
	updateResult = ch

	output := helperCaptureStderr(t, func() {
		printUpdateNotice()
	})
	assert.Empty(t, output)
}

func TestPrintUpdateNotice_NilResult(t *testing.T) {
	oldResult := updateResult
	defer func() { updateResult = oldResult }()

	ch := make(chan *updatecheck.CheckResult, 1)
	ch <- nil
	updateResult = ch

	output := helperCaptureStderr(t, func() {
		printUpdateNotice()
	})
	assert.Empty(t, output)
}

// --- readUpdateCheckConfig ---

func TestReadUpdateCheckConfig_Disabled(t *testing.T) {
	oldCfg := configFile
	defer func() { configFile = oldCfg }()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 300\nupdate_check: false\n")

	configFile = cfgPath

	disabled, _ := readUpdateCheckConfig()
	assert.True(t, disabled)
}

func TestReadUpdateCheckConfig_LanguageFromConfig(t *testing.T) {
	oldCfg := configFile
	defer func() { configFile = oldCfg }()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 300\nlanguage: ja\n")

	configFile = cfgPath

	disabled, lang := readUpdateCheckConfig()
	assert.False(t, disabled)
	assert.Equal(t, "ja", lang)
}

func TestReadUpdateCheckConfig_NoConfigFile(t *testing.T) {
	oldCfg := configFile
	defer func() { configFile = oldCfg }()

	configFile = "/nonexistent/path/config.yml"

	disabled, lang := readUpdateCheckConfig()
	assert.False(t, disabled)
	assert.Empty(t, lang)
}

// --- resolveUpdateCheckLang ---

func TestResolveUpdateCheckLang_FlagOverrides(t *testing.T) {
	oldLang := langFlag
	defer func() { langFlag = oldLang }()

	langFlag = "ja"
	assert.Equal(t, "ja", resolveUpdateCheckLang(""))
}

func TestResolveUpdateCheckLang_EnvOverrides(t *testing.T) {
	oldLang := langFlag
	defer func() { langFlag = oldLang }()

	langFlag = ""
	t.Setenv("LINTERLY_LANG", "ja")
	assert.Equal(t, "ja", resolveUpdateCheckLang(""))
}

func TestResolveUpdateCheckLang_ConfigLanguage(t *testing.T) {
	oldLang := langFlag
	defer func() { langFlag = oldLang }()

	langFlag = ""
	assert.Equal(t, "ja", resolveUpdateCheckLang("ja"))
}

func TestResolveUpdateCheckLang_Default(t *testing.T) {
	oldLang := langFlag
	defer func() { langFlag = oldLang }()

	langFlag = ""
	assert.Equal(t, "en", resolveUpdateCheckLang(""))
}

func TestResolveUpdateCheckLang_InvalidConfigLanguage(t *testing.T) {
	oldLang := langFlag
	defer func() { langFlag = oldLang }()

	langFlag = ""
	// 無効な言語コードの場合はデフォルト "en" にフォールバック
	assert.Equal(t, "en", resolveUpdateCheckLang("fr"))
}

// --- 短時間コマンドでも更新チェックの完了を（上限付きで）待つ ---

// startUpdateCheck 後にチェックが遅れて完了しても、上限時間内なら通知を表示する。
func TestPrintUpdateNotice_WaitsForSlowCheck(t *testing.T) {
	oldFlag, oldResult, oldRun, oldWait := flagNoUpdateCheck, updateResult, runUpdateCheck, updateWaitTimeout
	defer func() {
		flagNoUpdateCheck, updateResult, runUpdateCheck, updateWaitTimeout = oldFlag, oldResult, oldRun, oldWait
	}()
	flagNoUpdateCheck = false
	t.Setenv("LINTERLY_NO_UPDATE_CHECK", "")
	t.Setenv("LINTERLY_CONFIG", "")
	updateResult = nil
	updateWaitTimeout = 10 * time.Second

	release := make(chan struct{})
	runUpdateCheck = func(ctx context.Context, lang string) (*updatecheck.CheckResult, error) {
		<-release
		return &updatecheck.CheckResult{UpdateAvailable: true, Message: "update available: v9.9.9"}, nil
	}

	startUpdateCheck()
	require.NotNil(t, updateResult)
	// printUpdateNotice が待機に入った後でチェックを完了させる
	time.AfterFunc(20*time.Millisecond, func() { close(release) })

	output := helperCaptureStderr(t, printUpdateNotice)
	assert.Contains(t, output, "update available: v9.9.9")
}

// チェックが上限時間内に終わらなければ、通知せずに戻る（メインコマンドを無制限に遅らせない）。
func TestPrintUpdateNotice_GivesUpAfterTimeout(t *testing.T) {
	oldFlag, oldResult, oldRun, oldWait := flagNoUpdateCheck, updateResult, runUpdateCheck, updateWaitTimeout
	defer func() {
		flagNoUpdateCheck, updateResult, runUpdateCheck, updateWaitTimeout = oldFlag, oldResult, oldRun, oldWait
	}()
	flagNoUpdateCheck = false
	t.Setenv("LINTERLY_NO_UPDATE_CHECK", "")
	t.Setenv("LINTERLY_CONFIG", "")
	updateResult = nil
	updateWaitTimeout = 10 * time.Millisecond

	block := make(chan struct{})
	defer close(block)
	runUpdateCheck = func(ctx context.Context, lang string) (*updatecheck.CheckResult, error) {
		<-block
		return &updatecheck.CheckResult{UpdateAvailable: true, Message: "never shown"}, nil
	}

	startUpdateCheck()
	require.NotNil(t, updateResult)

	output := helperCaptureStderr(t, printUpdateNotice)
	assert.Empty(t, output)
}

// 無効化されている場合は runUpdateCheck（通信）も待機も発生しない。
func TestStartUpdateCheck_DisabledDoesNotRun(t *testing.T) {
	oldFlag, oldResult, oldRun := flagNoUpdateCheck, updateResult, runUpdateCheck
	defer func() { flagNoUpdateCheck, updateResult, runUpdateCheck = oldFlag, oldResult, oldRun }()
	flagNoUpdateCheck = true
	updateResult = nil
	called := make(chan struct{}, 1)
	runUpdateCheck = func(ctx context.Context, lang string) (*updatecheck.CheckResult, error) {
		called <- struct{}{}
		return nil, nil
	}

	startUpdateCheck()
	assert.Nil(t, updateResult)
	assert.Empty(t, helperCaptureStderr(t, printUpdateNotice))
	assert.Empty(t, called)
}
