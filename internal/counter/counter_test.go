package counter

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ousiassllc/linterly/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountFile_AllMode_Go(t *testing.T) {
	lc, err := CountFile("testdata/sample.go", config.CountModeAll)
	require.NoError(t, err)
	assert.Equal(t, 13, lc.TotalLines)
	assert.Equal(t, 13, lc.CodeLines) // all モードでは同じ
}

func TestCountFile_CodeOnly_Go(t *testing.T) {
	lc, err := CountFile("testdata/sample.go", config.CountModeCodeOnly)
	require.NoError(t, err)
	assert.Equal(t, 13, lc.TotalLines)
	// 空行(3) + 行コメント(1) + ブロックコメント(4) = 8 非コード行
	// 13 - 8 = 5 コード行
	assert.Equal(t, 5, lc.CodeLines)
}

func TestCountFile_CodeOnly_Python(t *testing.T) {
	lc, err := CountFile("testdata/sample.py", config.CountModeCodeOnly)
	require.NoError(t, err)
	assert.Equal(t, 15, lc.TotalLines)
	// 空行(4) + 行コメント(2) + ブロックコメント(4行) + docstring(1行) = 11 非コード行
	// 15 - 11 = 4 コード行 (import os, def hello():, print("hello"), x = 1)
	assert.Equal(t, 4, lc.CodeLines)
}

func TestCountFile_CodeOnly_HTML(t *testing.T) {
	lc, err := CountFile("testdata/sample.html", config.CountModeCodeOnly)
	require.NoError(t, err)
	assert.Equal(t, 10, lc.TotalLines)
	// ブロックコメント(5行) + 空行(0) = 5 非コード行
	// 10 - 5 = 5 コード行
	assert.Equal(t, 5, lc.CodeLines)
}

func TestCountFile_CodeOnly_Shell(t *testing.T) {
	lc, err := CountFile("testdata/sample.sh", config.CountModeCodeOnly)
	require.NoError(t, err)
	assert.Equal(t, 7, lc.TotalLines)
	// 空行(2) + 行コメント(3: shebang含む) = 5 非コード行
	// 7 - 5 = 2 コード行 (echo "hello", exit 0)
	assert.Equal(t, 2, lc.CodeLines)
}

func TestCountFile_EmptyFile(t *testing.T) {
	lc, err := CountFile("testdata/empty.txt", config.CountModeAll)
	require.NoError(t, err)
	assert.Equal(t, 0, lc.TotalLines)
	assert.Equal(t, 0, lc.CodeLines)
}

func TestCountFile_UnknownLanguage_CodeOnly(t *testing.T) {
	lc, err := CountFile("testdata/unknown.xyz", config.CountModeCodeOnly)
	require.NoError(t, err)
	assert.Equal(t, 3, lc.TotalLines)
	assert.Equal(t, 3, lc.CodeLines) // 対応言語なし → 全行コード行
}

func TestCountFile_NonExistent(t *testing.T) {
	_, err := CountFile("testdata/nonexistent.go", config.CountModeAll)
	assert.Error(t, err)
}

func TestCountFiles_Parallel(t *testing.T) {
	files := []string{
		"testdata/sample.go",
		"testdata/sample.py",
		"testdata/sample.sh",
	}

	results, err := CountFiles(files, config.CountModeAll)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// 順序が維持されることを確認
	assert.Equal(t, "testdata/sample.go", results[0].Path)
	assert.Equal(t, "testdata/sample.py", results[1].Path)
	assert.Equal(t, "testdata/sample.sh", results[2].Path)
}

func TestCountFiles_WithError(t *testing.T) {
	files := []string{
		"testdata/sample.go",
		"testdata/nonexistent.go",
	}

	_, err := CountFiles(files, config.CountModeAll)
	assert.Error(t, err)
}

func TestCountFile_NoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "no_newline.go")
	require.NoError(t, os.WriteFile(path, []byte("package main\nfunc main() {}"), 0644))

	lc, err := CountFile(path, config.CountModeAll)
	require.NoError(t, err)
	assert.Equal(t, 2, lc.TotalLines)
}

func TestCountFile_ScannerError_AllMode(t *testing.T) {
	// bufio.MaxScanTokenSize を超える改行なしデータで scanner.Err() がエラーを返すことを確認
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "huge_line.txt")

	data := bytes.Repeat([]byte("a"), bufio.MaxScanTokenSize+1)
	require.NoError(t, os.WriteFile(path, data, 0644))

	_, err := CountFile(path, config.CountModeAll)
	assert.Error(t, err)
}

func TestCountFile_ScannerError_CodeOnlyMode(t *testing.T) {
	// bufio.MaxScanTokenSize を超える改行なしデータで scanner.Err() がエラーを返すことを確認
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "huge_line.go")

	data := bytes.Repeat([]byte("a"), bufio.MaxScanTokenSize+1)
	require.NoError(t, os.WriteFile(path, data, 0644))

	_, err := CountFile(path, config.CountModeCodeOnly)
	assert.Error(t, err)
}
