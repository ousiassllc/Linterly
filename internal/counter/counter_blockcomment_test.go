package counter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// コードの後ろで始まり複数行に続くブロックコメントは、継続行をコード行に数えない（#75, #79）。
func TestCountCodeOnly_BlockCommentStartsAfterCode(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		input string
		code  int
	}{
		{"Go", "a.go", "package a\nx := 1 /* start\n comment\n comment2\n*/\ny := 2\n// c\n\n/* c */ z := 3\nend\n", 5},
		{"HTML", "a.html", "<p>\n<p> <!-- start\n comment\n-->\n<b>x</b>\n", 3},
		{"CSS", "a.css", "a { color: red } /* start\n comment\n*/\nb { color: blue }\n", 2},
		{"Java", "A.java", "int x = 1; /* start\n comment\n*/ int y = 2;\n", 2},
		{"同一行で閉じるブロックの後に開始", "a.go", "x := 1 /* a */ /* b\n c\n*/\ny := 2\n", 2},
		{"文字列リテラル内の記号は開始とみなさない", "a.go", "p := \"src/*\"\nq := `a/*`\nr := '/'\nz := 1\n", 4},
		{"行コメント内の記号は開始とみなさない", "a.go", "x := 1 // see /* here\ny := 2\n", 2},
		{"閉じていない引用符は文字列とみなさない", "a.rs", "fn f<'a>(x: &'a str) { /* start\n comment\n*/\n}\n", 2},
		{"Python は従来どおり", "a.py", "x = 1  # /* c\ny = 2\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, code, err := countCodeOnly(strings.NewReader(tt.input), DetectLanguage(tt.file))
			require.NoError(t, err)
			assert.Equal(t, tt.code, code)
		})
	}
}
