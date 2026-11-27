package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ousiassllc/linterly/internal/reporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperCheckProject は一時ディレクトリを作成してカレントディレクトリにし、設定ファイルを配置する。
// max_lines_per_file は 5、ディレクトリ上限は大きい値。
func helperCheckProject(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	oldCfg, oldFmt := configFile, format
	t.Cleanup(func() { configFile, format = oldCfg, oldFmt })

	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, "rules:\n  max_lines_per_file: 5\n  max_lines_per_directory: 100000\n  warning_threshold: 0\ndefault_excludes: false\n")
	configFile = cfgPath
	format = reporter.FormatText
}

// helperCheckJSON は format=json で runCheck を実行し、標準出力の JSON と error を返す。
func helperCheckJSON(t *testing.T, args ...string) (map[string]any, error) {
	t.Helper()
	format = reporter.FormatJSON
	var runErr error
	out := helperCaptureStdout(t, func() {
		runErr = runCheck(checkCmd, args)
	})
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &parsed), out)
	return parsed, runErr
}

// helperResultPaths は JSON レポートから種別ごとのパス一覧を返す。
func helperResultPaths(t *testing.T, report map[string]any, typ string) []string {
	t.Helper()
	var paths []string
	for _, r := range report["results"].([]any) {
		m := r.(map[string]any)
		if m["type"] == typ {
			paths = append(paths, m["path"].(string))
		}
	}
	return paths
}

func TestRunCheck_FileArgumentViolation(t *testing.T) {
	helperCheckProject(t)
	helperWriteFile(t, "a.txt", strings.Repeat("line\n", 8))

	err := runCheck(checkCmd, []string{"a.txt"})
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
}

func TestRunCheck_MultipleFileArguments(t *testing.T) {
	helperCheckProject(t)
	helperWriteFile(t, "src/ok.go", "a\nb\n")
	helperWriteFile(t, "src/big.go", strings.Repeat("line\n", 8))
	helperWriteFile(t, "src/other.go", strings.Repeat("line\n", 8)) // 指定しないので検査されない

	report, err := helperCheckJSON(t, "src/ok.go", "src/big.go")

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
	assert.ElementsMatch(t, []string{"src/ok.go", "src/big.go"}, helperResultPaths(t, report, "file"))
}

func TestRunCheck_FileArgumentRespectsExclude(t *testing.T) {
	helperCheckProject(t)
	helperWriteFile(t, ".linterlyignore", "gen/\n")
	helperWriteFile(t, "gen/big.go", strings.Repeat("line\n", 8))
	helperWriteFile(t, "ok.go", "a\n")

	report, err := helperCheckJSON(t, "gen/big.go", "ok.go")
	require.NoError(t, err)
	assert.Equal(t, []string{"ok.go"}, helperResultPaths(t, report, "file"))
}

func TestRunCheck_DirectoryArgumentUnchanged(t *testing.T) {
	helperCheckProject(t)
	helperWriteFile(t, "src/a.go", "a\n")

	report, err := helperCheckJSON(t, "src")
	require.NoError(t, err)
	// 単一ディレクトリ指定ではそのディレクトリからの相対パスで報告される（従来どおり）
	assert.Equal(t, []string{"a.go"}, helperResultPaths(t, report, "file"))
	assert.Equal(t, []string{"./"}, helperResultPaths(t, report, "directory"))
}

func TestRunCheck_NoArgumentChecksCurrentDirectory(t *testing.T) {
	helperCheckProject(t)
	helperWriteFile(t, "a.go", strings.Repeat("line\n", 8))

	err := runCheck(checkCmd, nil)
	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitViolation, exitErr.Code)
}

func TestRunCheck_NonExistentArgumentIsRuntimeError(t *testing.T) {
	helperCheckProject(t)
	helperWriteFile(t, "a.go", "a\n")

	err := runCheck(checkCmd, []string{"a.go", "missing.go"})
	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
}

func TestCheckCmd_AcceptsMultipleArgs(t *testing.T) {
	assert.NoError(t, checkCmd.Args(checkCmd, []string{"a.go", "b.go", "c.go"}))
}

func TestRunCheck_UnreadableFileIsSkippedWithWarning(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root は権限チェックを回避するためスキップ")
	}
	helperCheckProject(t)
	helperWriteFile(t, "sub/ok.txt", "a\n")
	helperWriteFile(t, "sub/denied.txt", "a\n")
	require.NoError(t, os.Chmod("sub/denied.txt", 0000))
	t.Cleanup(func() { _ = os.Chmod("sub/denied.txt", 0644) })

	var stderr bytes.Buffer
	checkCmd.SetErr(&stderr)
	t.Cleanup(func() { checkCmd.SetErr(nil) })

	report, err := helperCheckJSON(t, ".")
	require.NoError(t, err)
	paths := helperResultPaths(t, report, "file")
	assert.Contains(t, paths, "sub/ok.txt")
	assert.NotContains(t, paths, "sub/denied.txt")
	assert.Contains(t, stderr.String(), "warning: skipped")
	assert.Contains(t, stderr.String(), "denied.txt")
	assert.Contains(t, stderr.String(), "permission denied")
}
