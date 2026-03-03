package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"

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
	oldResult := updateResult
	defer func() { updateResult = oldResult }()

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
