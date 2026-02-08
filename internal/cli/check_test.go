package cli

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ousiassllc/linterly/internal/reporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperWriteFile はテスト用ファイルを作成するヘルパー。
func helperWriteFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

// helperCaptureStdout は fn 実行中の stdout 出力をキャプチャして返すヘルパー。
func helperCaptureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	fn()
	w.Close()
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(data)
}

func TestRunCheck_ConfigNotFound(t *testing.T) {
	old := configFile
	defer func() { configFile = old }()

	configFile = "/nonexistent/path/config.yml"

	err := runCheck(checkCmd, []string{"."})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
	assert.Contains(t, exitErr.Message, "no such file or directory")
}

func TestRunCheck_MissingRulesSection(t *testing.T) {
	old := configFile
	defer func() { configFile = old }()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "bad.yml")
	// rules セクションがない不正な YAML
	helperWriteFile(t, cfgPath, "not_rules: true\n")

	configFile = cfgPath

	err := runCheck(checkCmd, []string{"."})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
}

func TestRunCheck_ValidationError(t *testing.T) {
	old := configFile
	defer func() { configFile = old }()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "invalid.yml")
	// max_lines_per_file が負数のバリデーションエラー
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: -1\n")

	configFile = cfgPath

	err := runCheck(checkCmd, []string{"."})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
	assert.Contains(t, exitErr.Message, "max_lines_per_file")
}

func TestRunCheck_ViolationExitCode1(t *testing.T) {
	old := configFile
	oldFormat := format
	defer func() {
		configFile = old
		format = oldFormat
	}()

	tmpDir := t.TempDir()

	// 小さい上限の設定ファイル
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	cfgContent := `rules:
  max_lines_per_file: 3
  max_lines_per_directory: 100000
  warning_threshold: 0
default_excludes: false
`
	helperWriteFile(t, cfgPath, cfgContent)

	// 上限を超えるファイル（4行）
	targetDir := filepath.Join(tmpDir, "src")
	bigFile := filepath.Join(targetDir, "big.go")
	helperWriteFile(t, bigFile, strings.Repeat("line\n", 4))

	configFile = cfgPath
	format = reporter.FormatText

	err := runCheck(checkCmd, []string{targetDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
}

func TestRunCheck_NoViolation(t *testing.T) {
	old := configFile
	oldFormat := format
	defer func() {
		configFile = old
		format = oldFormat
	}()

	tmpDir := t.TempDir()

	// 大きい上限の設定ファイル
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	cfgContent := `rules:
  max_lines_per_file: 100000
  max_lines_per_directory: 100000
  warning_threshold: 0
default_excludes: false
`
	helperWriteFile(t, cfgPath, cfgContent)

	// 上限内のファイル（2行）
	targetDir := filepath.Join(tmpDir, "src")
	smallFile := filepath.Join(targetDir, "small.go")
	helperWriteFile(t, smallFile, "line1\nline2\n")

	configFile = cfgPath
	format = reporter.FormatText

	err := runCheck(checkCmd, []string{targetDir})
	assert.NoError(t, err)
}

func TestRunCheck_WarningsOnly(t *testing.T) {
	old := configFile
	oldFormat := format
	defer func() {
		configFile = old
		format = oldFormat
	}()

	tmpDir := t.TempDir()

	// warning_threshold を大きくし、上限超えが warn 止まりになる設定
	// threshold = 3 + 3*100/100 = 6 なので、4行のファイルは warn（error にならない）
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	cfgContent := `rules:
  max_lines_per_file: 3
  max_lines_per_directory: 100000
  warning_threshold: 100
default_excludes: false
`
	helperWriteFile(t, cfgPath, cfgContent)

	// 上限を超えるが threshold 内のファイル（4行: 3 < 4 <= 6）
	targetDir := filepath.Join(tmpDir, "src")
	warnFile := filepath.Join(targetDir, "warn.go")
	helperWriteFile(t, warnFile, strings.Repeat("line\n", 4))

	configFile = cfgPath
	format = reporter.FormatText

	err := runCheck(checkCmd, []string{targetDir})
	assert.NoError(t, err)
}

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
