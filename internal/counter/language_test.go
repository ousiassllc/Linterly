package counter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectLanguage_Go(t *testing.T) {
	lang := DetectLanguage("main.go")
	assert.NotNil(t, lang)
	assert.Equal(t, "Go", lang.Name)
}

func TestDetectLanguage_Python(t *testing.T) {
	lang := DetectLanguage("script.py")
	assert.NotNil(t, lang)
	assert.Equal(t, "Python", lang.Name)
}

func TestDetectLanguage_TypeScript(t *testing.T) {
	lang := DetectLanguage("app.tsx")
	assert.NotNil(t, lang)
	assert.Equal(t, "TypeScript", lang.Name)
}

func TestDetectLanguage_Shell(t *testing.T) {
	lang := DetectLanguage("run.sh")
	assert.NotNil(t, lang)
	assert.Equal(t, "Shell", lang.Name)
}

func TestDetectLanguage_HTML(t *testing.T) {
	lang := DetectLanguage("index.html")
	assert.NotNil(t, lang)
	assert.Equal(t, "HTML", lang.Name)
}

func TestDetectLanguage_Unknown(t *testing.T) {
	lang := DetectLanguage("data.xyz")
	assert.Nil(t, lang)
}

func TestDetectLanguage_NoExtension(t *testing.T) {
	lang := DetectLanguage("Makefile")
	assert.Nil(t, lang)
}

func TestDetectLanguage_AllExtensions(t *testing.T) {
	tests := []struct {
		ext  string
		name string
	}{
		{".go", "Go"},
		{".rs", "Rust"},
		{".js", "JavaScript"},
		{".jsx", "JavaScript"},
		{".mjs", "JavaScript"},
		{".ts", "TypeScript"},
		{".tsx", "TypeScript"},
		{".mts", "TypeScript"},
		{".py", "Python"},
		{".rb", "Ruby"},
		{".java", "Java"},
		{".kt", "Kotlin"},
		{".kts", "Kotlin"},
		{".c", "C"},
		{".h", "C"},
		{".cpp", "C++"},
		{".cc", "C++"},
		{".hpp", "C++"},
		{".hh", "C++"},
		{".html", "HTML"},
		{".htm", "HTML"},
		{".xml", "HTML"},
		{".svg", "HTML"},
		{".css", "CSS"},
		{".scss", "SCSS"},
		{".sass", "SCSS"},
		{".sh", "Shell"},
		{".bash", "Shell"},
		{".zsh", "Shell"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			lang := DetectLanguage("file" + tt.ext)
			assert.NotNil(t, lang, "expected language for %s", tt.ext)
			assert.Equal(t, tt.name, lang.Name)
		})
	}
}
