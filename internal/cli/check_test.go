package cli

import (
	"errors"
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
