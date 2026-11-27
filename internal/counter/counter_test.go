package counter

import (
	"bufio"
	"bytes"
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

func TestCountFile_LargeLine_AllMode(t *testing.T) {
	// bufio.MaxScanTokenSize(64KB) を超えるが maxScanBufSize(1MB) 以内の行は正常にカウントできる
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
	// bufio.MaxScanTokenSize(64KB) を超えるが maxScanBufSize(1MB) 以内の行は正常にカウントできる
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

	results, err := countFiles(files, config.CountModeAll, countFn)
	require.NoError(t, err)
	require.Len(t, results, len(files))
	for i, r := range results {
		assert.Equal(t, files[i], r.Path, "入力順が維持されること")
	}
	assert.LessOrEqual(t, int(peak.Load()), maxWorkers)
	assert.Positive(t, int(peak.Load()))
}

func TestCountFiles_Empty(t *testing.T) {
	results, err := CountFiles(nil, config.CountModeAll)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestCountCodeOnly_BlockCommentWithCodeOnSameLine(t *testing.T) {
	// ブロックコメントの開始行・終了行・同一行に同居するコードは行数に数える
	input := "/* c */ x := 1\n/* start\nend */ y := 2\nz\n"
	total, code, err := countCodeOnly(strings.NewReader(input), DetectLanguage("u.go"))
	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, 3, code)
}

func TestCountCodeOnly_BlockCommentSameLineEdges(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		input string
		code  int
	}{
		{"複数ブロックの後にコード", "a.go", "/* a */ /* b */ x := 1\n", 1},
		{"複数ブロックのみ", "a.go", "/* a */ /* b */\n", 0},
		{"ブロック後に行コメント", "a.go", "/* a */ // note\n", 0},
		{"終了行が閉じ記号のみ", "a.go", "/*\nfoo\n*/\nx := 1\n", 1},
		{"終了行の後にさらにブロック開始", "a.go", "/* a\nb */ /* c\nd */ x := 1\n", 1},
		{"コード後ろのコメントはコード行", "a.go", "x := 1 /* c */\n", 1},
		{"HTML コメント後にコード", "a.html", "<!-- c --> <p>x</p>\n", 1},
		{"CSS ブロック終了後にコード", "a.css", "/* c\n*/ a { color: red }\n", 1},
		{"Python docstring は従来どおり", "a.py", "\"\"\"doc\"\"\"\nx = 1\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, code, err := countCodeOnly(strings.NewReader(tt.input), DetectLanguage(tt.file))
			require.NoError(t, err)
			assert.Equal(t, tt.code, code)
		})
	}
}

// TestCountCodeOnly_LanguageComments は言語ごとのコメント認識（行コメント / ブロックコメント / 空行 / コード）を検証する。
// F-021（Rust）, F-022（JS/TS）, F-024（Ruby）, F-025（Java/Kotlin）, F-026（C/C++）, F-028（CSS/SCSS）。
func TestCountCodeOnly_LanguageComments(t *testing.T) {
	cStyle := strings.Join([]string{
		"// line comment",
		"",
		"/* block start",
		" * middle",
		" */",
		"/* single */",
		"code1",
		"/* a */ code2",
		"code3 // trailing",
	}, "\n") + "\n"
	// 非コード行 = 行コメント(1) + 空行(1) + ブロック(3) + 単一行ブロック(1) = 6、コード = 3
	const cStyleTotal, cStyleCode = 9, 3

	tests := []struct {
		name  string
		file  string
		input string
		total int
		code  int
	}{
		{"Rust", "main.rs", cStyle, cStyleTotal, cStyleCode},
		{"JavaScript", "app.js", cStyle, cStyleTotal, cStyleCode},
		{"JavaScript(jsx)", "app.jsx", cStyle, cStyleTotal, cStyleCode},
		{"JavaScript(mjs)", "app.mjs", cStyle, cStyleTotal, cStyleCode},
		{"TypeScript", "app.ts", cStyle, cStyleTotal, cStyleCode},
		{"TypeScript(tsx)", "app.tsx", cStyle, cStyleTotal, cStyleCode},
		{"Java", "Main.java", cStyle, cStyleTotal, cStyleCode},
		{"Kotlin", "Main.kt", cStyle, cStyleTotal, cStyleCode},
		{"Kotlin(kts)", "build.gradle.kts", cStyle, cStyleTotal, cStyleCode},
		{"C", "main.c", cStyle, cStyleTotal, cStyleCode},
		{"C(header)", "main.h", cStyle, cStyleTotal, cStyleCode},
		{"C++", "main.cpp", cStyle, cStyleTotal, cStyleCode},
		{"C++(hpp)", "main.hpp", cStyle, cStyleTotal, cStyleCode},
		{"SCSS", "style.scss", cStyle, cStyleTotal, cStyleCode},
		{
			// Ruby: # 行コメントと =begin ... =end の複数行ブロック
			name: "Ruby",
			file: "app.rb",
			input: strings.Join([]string{
				"# line comment",
				"",
				"=begin",
				"multi line",
				"comment",
				"=end",
				"puts 'hello'",
				"x = 1 # trailing",
			}, "\n") + "\n",
			total: 8,
			code:  2,
		},
		{
			// CSS は // を行コメントとして認識しない
			name:  "CSS",
			file:  "style.css",
			input: "/* c */\n// not a comment\na { color: red }\n",
			total: 3,
			code:  2,
		},
		{
			// SCSS は // 行コメントと /* */ の両方を認識する
			name:  "SCSS 行コメントとブロック",
			file:  "style.scss",
			input: "// c\n/* c */\n$x: 1;\n",
			total: 3,
			code:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := DetectLanguage(tt.file)
			require.NotNil(t, lang, tt.file)
			total, code, err := countCodeOnly(strings.NewReader(tt.input), lang)
			require.NoError(t, err)
			assert.Equal(t, tt.total, total, "total")
			assert.Equal(t, tt.code, code, "code")
		})
	}
}

func TestCountCodeOnly_Ruby_BlockNotAtLineStart(t *testing.T) {
	// =begin は行頭のみブロックコメントの開始として認識する
	input := "x = 1 =begin\ny = 2\n"
	_, code, err := countCodeOnly(strings.NewReader(input), DetectLanguage("a.rb"))
	require.NoError(t, err)
	assert.Equal(t, 2, code)
}
