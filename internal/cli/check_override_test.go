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

func TestRunCheck_FlagMaxLinesPerFile_Override(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagMaxLinesPerFile
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagMaxLinesPerFile = oldFlag
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 100000\n  max_lines_per_directory: 100000\n  warning_threshold: 0\ndefault_excludes: false\n")

	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "main.go"), strings.Repeat("line\n", 10))

	configFile = cfgPath
	format = reporter.FormatText
	flagMaxLinesPerFile = 5
	helperSetFlag(t, "max-lines-per-file")

	err := runCheck(checkCmd, []string{targetDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
}

func TestRunCheck_FlagMaxLinesPerDirectory_Override(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagMaxLinesPerDirectory
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagMaxLinesPerDirectory = oldFlag
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 100000\n  max_lines_per_directory: 100000\n  warning_threshold: 0\ndefault_excludes: false\n")

	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "a.go"), strings.Repeat("line\n", 10))
	helperWriteFile(t, filepath.Join(targetDir, "b.go"), strings.Repeat("line\n", 10))

	configFile = cfgPath
	format = reporter.FormatText
	flagMaxLinesPerDirectory = 5
	helperSetFlag(t, "max-lines-per-directory")

	err := runCheck(checkCmd, []string{targetDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
}

func TestRunCheck_FlagWarningThreshold_Override(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagWarningThreshold
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagWarningThreshold = oldFlag
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	// max_lines_per_file: 10, warning_threshold: 0 → 12行のファイルは error
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 10\n  max_lines_per_directory: 100000\n  warning_threshold: 0\ndefault_excludes: false\n")

	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "main.go"), strings.Repeat("line\n", 12))

	configFile = cfgPath
	format = reporter.FormatText
	// threshold=100 → 閾値=10+10*100/100=20、12行は warn 止まり（exit 0）
	flagWarningThreshold = 100
	helperSetFlag(t, "warning-threshold")

	err := runCheck(checkCmd, []string{targetDir})
	assert.NoError(t, err)
}

func TestRunCheck_FlagIgnore_Override(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagIgnore
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagIgnore = oldFlag
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 3\n  max_lines_per_directory: 100000\n  warning_threshold: 0\ndefault_excludes: false\n")

	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "big.go"), strings.Repeat("line\n", 10))

	configFile = cfgPath
	format = reporter.FormatText
	flagIgnore = []string{"*.go"}
	helperSetFlag(t, "ignore")

	err := runCheck(checkCmd, []string{targetDir})
	assert.NoError(t, err)
}

func TestRunCheck_FlagNoDefaultExcludes_Override(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagNoDefaultExcludes
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagNoDefaultExcludes = oldFlag
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	// default_excludes: true (default) - node_modules would be excluded
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 3\n  max_lines_per_directory: 100000\n  warning_threshold: 0\n")

	// node_modules/ 配下にファイルを作成（通常は除外される）
	targetDir := filepath.Join(tmpDir, "project")
	helperWriteFile(t, filepath.Join(targetDir, "node_modules", "pkg", "big.js"), strings.Repeat("line\n", 10))
	// 正常なファイルも必要（scanResult が空だとテストが不安定になる可能性）
	helperWriteFile(t, filepath.Join(targetDir, "ok.js"), "line\n")

	configFile = cfgPath
	format = reporter.FormatText
	flagNoDefaultExcludes = true
	helperSetFlag(t, "no-default-excludes")

	err := runCheck(checkCmd, []string{targetDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
}

func TestRunCheck_FlagOverride_ValidationError(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagMaxLinesPerFile
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagMaxLinesPerFile = oldFlag
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 300\n")

	configFile = cfgPath
	format = reporter.FormatText
	flagMaxLinesPerFile = -1
	helperSetFlag(t, "max-lines-per-file")

	err := runCheck(checkCmd, []string{tmpDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
	assert.Contains(t, exitErr.Message, "max_lines_per_file")
}

func TestRunCheck_FlagOverride_NoConfigFile(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagMaxLinesPerFile
	oldFlagDE := flagNoDefaultExcludes
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagMaxLinesPerFile = oldFlag
		flagNoDefaultExcludes = oldFlagDE
	}()

	tmpDir := t.TempDir()

	// 設定ファイルなしのディレクトリ
	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "main.go"), strings.Repeat("line\n", 10))

	// 設定ファイルなしのディレクトリに chdir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	configFile = ""
	format = reporter.FormatText
	flagMaxLinesPerFile = 5
	helperSetFlag(t, "max-lines-per-file")
	flagNoDefaultExcludes = true
	helperSetFlag(t, "no-default-excludes")

	err = runCheck(checkCmd, []string{targetDir})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
}

func TestRunCheck_FlagNotChanged_DoesNotOverride(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagMaxLinesPerFile
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagMaxLinesPerFile = oldFlag
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 100000\n  max_lines_per_directory: 100000\n  warning_threshold: 0\ndefault_excludes: false\n")

	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "main.go"), strings.Repeat("line\n", 10))

	configFile = cfgPath
	format = reporter.FormatText
	// フラグ変数をセットするが Changed は設定しない → 上書きされないはず
	flagMaxLinesPerFile = 5
	// helperSetFlag を呼ばない → Changed = false

	err := runCheck(checkCmd, []string{targetDir})
	assert.NoError(t, err)
}

func TestRunCheck_FlagCountMode_Override(t *testing.T) {
	oldCfg := configFile
	oldFmt := format
	oldFlag := flagCountMode
	defer func() {
		configFile = oldCfg
		format = oldFmt
		flagCountMode = oldFlag
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	// count_mode: all, max_lines_per_file: 5
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 5\n  max_lines_per_directory: 100000\n  warning_threshold: 0\ndefault_excludes: false\n")

	targetDir := filepath.Join(tmpDir, "src")
	// 8行のうち3行がコード行（5行がコメント/空行）
	content := "package main\n\n// comment\n// comment\n// comment\n// comment\n// comment\nfunc main() {}\n"
	helperWriteFile(t, filepath.Join(targetDir, "main.go"), content)

	configFile = cfgPath
	format = reporter.FormatText
	// code_only モードに変更 → コード行のみカウント（3行 < 5）
	flagCountMode = "code_only"
	helperSetFlag(t, "count-mode")

	err := runCheck(checkCmd, []string{targetDir})
	assert.NoError(t, err)
}
