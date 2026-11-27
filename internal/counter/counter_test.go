package counter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
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

func TestCountFile_CodeOnly_CSS(t *testing.T) {
	lc, err := CountFile("testdata/sample.css", config.CountModeCodeOnly)
	require.NoError(t, err)
	assert.Equal(t, 5, lc.TotalLines)
	// ブロックコメント(1行) = 1 非コード行
	// 5 - 1 = 4 コード行（"// This is NOT a CSS comment" はコード行として数える）
	assert.Equal(t, 4, lc.CodeLines)
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

	results, skipped := CountFiles(files, config.CountModeAll)
	assert.Empty(t, skipped)
	assert.Len(t, results, 3)

	// 順序が維持されることを確認
	assert.Equal(t, "testdata/sample.go", results[0].Path)
	assert.Equal(t, "testdata/sample.py", results[1].Path)
	assert.Equal(t, "testdata/sample.sh", results[2].Path)
}

func TestCountFiles_SkipsUnreadableFiles(t *testing.T) {
	// 読み取れないファイルがあっても他のファイルは集計され、スキップ情報が返る
	files := []string{
		"testdata/sample.go",
		"testdata/nonexistent.go",
		"testdata/sample.sh",
	}

	results, skipped := CountFiles(files, config.CountModeAll)
	require.Len(t, results, 2)
	assert.Equal(t, "testdata/sample.go", results[0].Path)
	assert.Equal(t, "testdata/sample.sh", results[1].Path)
	require.Len(t, skipped, 1)
	assert.Equal(t, "testdata/nonexistent.go", skipped[0].Path)
	assert.ErrorIs(t, skipped[0].Err, os.ErrNotExist)
}

func TestCountFiles_SkipsPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root は権限チェックを回避するためスキップ")
	}
	tmpDir := t.TempDir()
	ok := filepath.Join(tmpDir, "ok.txt")
	denied := filepath.Join(tmpDir, "denied.txt")
	require.NoError(t, os.WriteFile(ok, []byte("a\nb\n"), 0644))
	require.NoError(t, os.WriteFile(denied, []byte("a\n"), 0000))

	results, skipped := CountFiles([]string{denied, ok}, config.CountModeAll)
	require.Len(t, results, 1)
	assert.Equal(t, ok, results[0].Path)
	assert.Equal(t, 2, results[0].TotalLines)
	require.Len(t, skipped, 1)
	assert.Equal(t, denied, skipped[0].Path)
	assert.ErrorIs(t, skipped[0].Err, os.ErrPermission)
}

func TestCountFile_NoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "no_newline.go")
	require.NoError(t, os.WriteFile(path, []byte("package main\nfunc main() {}"), 0644))

	lc, err := CountFile(path, config.CountModeAll)
	require.NoError(t, err)
	assert.Equal(t, 2, lc.TotalLines)
}

func TestCountAll_Edges(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"only newline", "\n", 1},
		{"no trailing newline", "a\nb", 2},
		{"trailing newline", "a\nb\n", 2},
		{"blank lines", "\n\n\n", 3},
		{"crlf", "a\r\nb\r\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := countAll(strings.NewReader(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountAll_FromReader(t *testing.T) {
	r := strings.NewReader("line1\nline2\nline3\n")
	count, err := countAll(r)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestCountCodeOnly_FromReader(t *testing.T) {
	r := strings.NewReader("package main\n\n// comment\nfunc main() {}\n")
	lang := DetectLanguage("example.go")
	total, code, err := countCodeOnly(r, lang)
	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, 2, code)
}

func TestCountFiles_ConcurrencyIsBounded(t *testing.T) {
	// 同時に処理されるファイル数が maxWorkers を超えないこと（fd 枯渇の回帰テスト）
	files := make([]string, 200)
	for i := range files {
		files[i] = fmt.Sprintf("f%d.txt", i)
	}

	var current, peak atomic.Int32
	countFn := func(path, _ string) (*LineCount, error) {
		n := current.Add(1)
		defer current.Add(-1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		runtime.Gosched()
		return &LineCount{Path: path}, nil
	}

	results, skipped := countFiles(files, config.CountModeAll, countFn)
	assert.Empty(t, skipped)
	require.Len(t, results, len(files))
	for i, r := range results {
		assert.Equal(t, files[i], r.Path, "入力順が維持されること")
	}
	assert.LessOrEqual(t, int(peak.Load()), maxWorkers)
	assert.Positive(t, int(peak.Load()))
}

func TestCountFiles_Empty(t *testing.T) {
	results, skipped := CountFiles(nil, config.CountModeAll)
	assert.Empty(t, results)
	assert.Empty(t, skipped)
}
