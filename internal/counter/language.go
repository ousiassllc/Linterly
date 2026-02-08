package counter

import "path/filepath"

// Language はプログラミング言語のコメント構文を定義する。
type Language struct {
	Name              string
	Extensions        []string
	LineCommentStart  []string // 例: ["//", "#"]
	BlockCommentStart string   // 例: "/*"
	BlockCommentEnd   string   // 例: "*/"
}

var languages = []Language{
	{
		Name:              "Go",
		Extensions:        []string{".go"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Rust",
		Extensions:        []string{".rs"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "JavaScript",
		Extensions:        []string{".js", ".jsx", ".mjs"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "TypeScript",
		Extensions:        []string{".ts", ".tsx", ".mts"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Python",
		Extensions:        []string{".py"},
		LineCommentStart:  []string{"#"},
		BlockCommentStart: `"""`,
		BlockCommentEnd:   `"""`,
	},
	{
		Name:              "Ruby",
		Extensions:        []string{".rb"},
		LineCommentStart:  []string{"#"},
		BlockCommentStart: "=begin",
		BlockCommentEnd:   "=end",
	},
	{
		Name:              "Java",
		Extensions:        []string{".java"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Kotlin",
		Extensions:        []string{".kt", ".kts"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "C",
		Extensions:        []string{".c", ".h"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "C++",
		Extensions:        []string{".cpp", ".cc", ".hpp", ".hh"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "HTML",
		Extensions:        []string{".html", ".htm", ".xml", ".svg"},
		LineCommentStart:  nil,
		BlockCommentStart: "<!--",
		BlockCommentEnd:   "-->",
	},
	{
		Name:              "CSS",
		Extensions:        []string{".css"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "SCSS",
		Extensions:        []string{".scss", ".sass"},
		LineCommentStart:  []string{"//"},
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:             "Shell",
		Extensions:       []string{".sh", ".bash", ".zsh"},
		LineCommentStart: []string{"#"},
	},
}

// extToLanguage は拡張子から言語へのマッピング。
var extToLanguage map[string]*Language

func init() {
	extToLanguage = make(map[string]*Language)
	for i := range languages {
		for _, ext := range languages[i].Extensions {
			extToLanguage[ext] = &languages[i]
		}
	}
}

// DetectLanguage はファイルパスの拡張子から言語を検出する。
// 対応する言語が見つからない場合は nil を返す。
func DetectLanguage(path string) *Language {
	ext := filepath.Ext(path)
	return extToLanguage[ext]
}
