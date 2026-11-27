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

func TestCountFile_LargeLine_AllMode(t *testing.T) {
	// bufio.MaxScanTokenSize(64KB) を超えるが maxLineKeepSize(1MB) 以内の行は正常にカウントできる
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "large_line.txt")

	// 64KB+1 の行 + 改行 + 通常行 = 2行
	data := append(bytes.Repeat([]byte("a"), bufio.MaxScanTokenSize+1), []byte("\nhello\n")...)
	require.NoError(t, os.WriteFile(path, data, 0644))

	lc, err := CountFile(path, config.CountModeAll)
	require.NoError(t, err)
	assert.Equal(t, 2, lc.TotalLines)
}

func TestCountFile_LargeLine_CodeOnlyMode(t *testing.T) {
	// bufio.MaxScanTokenSize(64KB) を超えるが maxLineKeepSize(1MB) 以内の行は正常にカウントできる
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "large_line.go")

	data := append(bytes.Repeat([]byte("a"), bufio.MaxScanTokenSize+1), []byte("\npackage main\n")...)
	require.NoError(t, os.WriteFile(path, data, 0644))

	lc, err := CountFile(path, config.CountModeCodeOnly)
	require.NoError(t, err)
	assert.Equal(t, 2, lc.TotalLines)
}

func TestCountFile_HugeLine_AllMode(t *testing.T) {
	// maxLineKeepSize(1MB) を超える長大行を含んでもエラーにならず、正しい行数を返す
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "huge_line.txt")

	data := bytes.Repeat([]byte("a"), 2*maxLineKeepSize)
	data = append(data, []byte("\nsecond\nthird")...)
	require.NoError(t, os.WriteFile(path, data, 0644))

	lc, err := CountFile(path, config.CountModeAll)
	require.NoError(t, err)
	assert.Equal(t, 3, lc.TotalLines)
	assert.Equal(t, 3, lc.CodeLines)
}

func TestCountFile_HugeLine_CodeOnlyMode(t *testing.T) {
	// maxLineKeepSize(1MB) を超える長大行を含んでもエラーにならず、正しい行数を返す
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "huge_line.go")

	data := bytes.Repeat([]byte("a"), 2*maxLineKeepSize)
	data = append(data, []byte("\n// comment\n\npackage main")...)
	require.NoError(t, os.WriteFile(path, data, 0644))

	lc, err := CountFile(path, config.CountModeCodeOnly)
	require.NoError(t, err)
	assert.Equal(t, 4, lc.TotalLines)
	assert.Equal(t, 2, lc.CodeLines) // 長大行 + package main
}

func TestCountFile_HugeSingleLine_NoNewline(t *testing.T) {
	// 改行なしの 2MB 単一行は両モードとも 1 行
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "big.js")
	require.NoError(t, os.WriteFile(path, bytes.Repeat([]byte("a"), 2*maxLineKeepSize), 0644))

	for _, mode := range []string{config.CountModeAll, config.CountModeCodeOnly} {
		lc, err := CountFile(path, mode)
		require.NoError(t, err, mode)
		assert.Equal(t, 1, lc.TotalLines, mode)
		assert.Equal(t, 1, lc.CodeLines, mode)
	}
}
