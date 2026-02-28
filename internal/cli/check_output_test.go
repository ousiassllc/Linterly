package cli

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ousiassllc/linterly/internal/reporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCheck_ValidationError_Japanese(t *testing.T) {
	old := configFile
	oldLang := langFlag
	oldFormat := format
	defer func() {
		configFile = old
		langFlag = oldLang
		format = oldFormat
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "invalid.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: -1\n")
	configFile = cfgPath
	langFlag = "ja"

	err := runCheck(checkCmd, []string{tmpDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
	// 日本語メッセージが含まれることを確認
	assert.Contains(t, exitErr.Message, "正の整数")
}

func TestRunCheck_ConfigNotFound_Japanese(t *testing.T) {
	old := configFile
	oldLang := langFlag
	defer func() {
		configFile = old
		langFlag = oldLang
	}()

	tmpDir := t.TempDir()
	configFile = filepath.Join(tmpDir, "nonexistent.yml")
	langFlag = "ja"

	err := runCheck(checkCmd, []string{tmpDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
}

func TestRunCheck_FormatJSON(t *testing.T) {
	old := configFile
	oldFormat := format
	defer func() {
		configFile = old
		format = oldFormat
	}()

	tmpDir := t.TempDir()

	// 違反なしの設定ファイル
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	cfgContent := `rules:
  max_lines_per_file: 100000
  max_lines_per_directory: 100000
  warning_threshold: 0
default_excludes: false
`
	helperWriteFile(t, cfgPath, cfgContent)

	// 小さいファイル
	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "small.go"), "line1\nline2\n")

	configFile = cfgPath
	format = reporter.FormatJSON

	var err error
	output := helperCaptureStdout(t, func() {
		err = runCheck(checkCmd, []string{targetDir})
	})

	assert.NoError(t, err)
	// JSON 出力が空でないことを確認
	assert.NotEmpty(t, output)
	assert.True(t, json.Valid([]byte(output)), "output should be valid JSON")
}

func TestRunCheck_FormatJSON_WithViolation(t *testing.T) {
	old := configFile
	oldFormat := format
	defer func() {
		configFile = old
		format = oldFormat
	}()

	tmpDir := t.TempDir()

	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	cfgContent := `rules:
  max_lines_per_file: 3
  max_lines_per_directory: 100000
  warning_threshold: 0
default_excludes: false
`
	helperWriteFile(t, cfgPath, cfgContent)

	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "big.go"), strings.Repeat("line\n", 10))

	configFile = cfgPath
	format = reporter.FormatJSON

	var err error
	output := helperCaptureStdout(t, func() {
		err = runCheck(checkCmd, []string{targetDir})
	})

	// 違反ありの場合は ExitError が返る
	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
	// JSON 出力にファイル情報が含まれる
	assert.Contains(t, output, "big.go")
}

func TestRunCheck_ConfigLanguageSwitchesTranslation(t *testing.T) {
	old := configFile
	oldFormat := format
	oldLang := langFlag
	defer func() {
		configFile = old
		format = oldFormat
		langFlag = oldLang
	}()

	tmpDir := t.TempDir()

	// config に language: ja を指定
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	cfgContent := `rules:
  max_lines_per_file: 3
  max_lines_per_directory: 100000
  warning_threshold: 0
default_excludes: false
language: ja
`
	helperWriteFile(t, cfgPath, cfgContent)

	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "big.go"), strings.Repeat("line\n", 10))

	configFile = cfgPath
	format = reporter.FormatText
	// langFlag 空・LINTERLY_LANG 未設定 → config.Language で再初期化される
	langFlag = ""
	t.Setenv("LINTERLY_LANG", "")

	var err error
	output := helperCaptureStdout(t, func() {
		err = runCheck(checkCmd, []string{targetDir})
	})

	// 違反があるので ExitError が返る
	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
	// 日本語メッセージが含まれることを確認（config.Language: ja による再初期化が有効）
	assert.Contains(t, output, "エラー")
}

func TestRunCheck_MissingRulesSection_Japanese(t *testing.T) {
	old := configFile
	oldLang := langFlag
	defer func() {
		configFile = old
		langFlag = oldLang
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "bad.yml")
	helperWriteFile(t, cfgPath, "not_rules: true\n")
	configFile = cfgPath
	langFlag = "ja"

	err := runCheck(checkCmd, []string{"."})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
	// 日本語メッセージが含まれることを確認
	assert.Contains(t, exitErr.Message, "セクションが必要です")
}

func TestRunCheck_FlagOverride_ValidationError_Japanese(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagMaxLinesPerFile
	oldLang := langFlag
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagMaxLinesPerFile = oldFlag
		langFlag = oldLang
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 300\n")

	configFile = cfgPath
	format = reporter.FormatText
	langFlag = "ja"
	flagMaxLinesPerFile = -1
	helperSetFlag(t, "max-lines-per-file")

	err := runCheck(checkCmd, []string{tmpDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
	assert.Contains(t, exitErr.Message, "正の整数")
}
