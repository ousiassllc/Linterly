package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ousiassllc/linterly/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan_SkipsBinaryFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// テキストファイル
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "main.go"),
		[]byte("package main\n\nfunc main() {}\n"),
		0644,
	))

	// バイナリファイル（null バイトを含む）
	binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "image.png"),
		binaryContent,
		0644,
	))

	// 拡張子なしのバイナリファイル（null バイトを含む）
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "binary_no_ext"),
		binaryContent,
		0644,
	))

	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{},
	}

	result, err := Scan(tmpDir, cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	assert.Contains(t, paths, "main.go", "テキストファイルは含まれるべき")
	assert.NotContains(t, paths, "image.png", "既知バイナリ拡張子はスキップされるべき")
	assert.NotContains(t, paths, "binary_no_ext", "null バイトを含むファイルはスキップされるべき")
}

func TestScan_SkipsBinaryByExtension(t *testing.T) {
	tmpDir := t.TempDir()

	binaryExts := []string{".png", ".jpg", ".exe", ".zip", ".pdf", ".woff2"}
	for _, ext := range binaryExts {
		require.NoError(t, os.WriteFile(
			filepath.Join(tmpDir, "file"+ext),
			[]byte("dummy content"), // 中身がテキストでも拡張子で除外
			0644,
		))
	}

	// テキストファイル
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "code.go"),
		[]byte("package main\n"),
		0644,
	))

	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{},
	}

	result, err := Scan(tmpDir, cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	assert.Contains(t, paths, "code.go")
	for _, ext := range binaryExts {
		assert.NotContains(t, paths, "file"+ext, ext+" はスキップされるべき")
	}
}

func TestScan_NullByteDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// 拡張子がテキストだが中身がバイナリ（null バイト含む）
	binaryContent := make([]byte, 100)
	binaryContent[50] = 0x00 // null バイトを埋め込む
	for i := range binaryContent {
		if i != 50 {
			binaryContent[i] = 'a'
		}
	}
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "corrupted.txt"),
		binaryContent,
		0644,
	))

	// 正常なテキストファイル
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "normal.txt"),
		[]byte("hello world\n"),
		0644,
	))

	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{},
	}

	result, err := Scan(tmpDir, cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	assert.Contains(t, paths, "normal.txt", "正常なテキストファイルは含まれるべき")
	assert.NotContains(t, paths, "corrupted.txt", "null バイトを含むファイルはスキップされるべき")
}

func TestScan_ScriptFilesNotSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	// スクリプトファイル（テキストとして含まれるべき）
	scripts := map[string]string{
		"build.sh":   "#!/bin/bash\necho hello\n",
		"setup.bat":  "@echo off\necho hello\n",
		"deploy.ps1": "Write-Host 'hello'\n",
		"Makefile":   "all:\n\techo hello\n",
		"Dockerfile": "FROM alpine:latest\n",
		"config.yml": "key: value\n",
		"style.css":  "body { color: red; }\n",
		"index.html": "<html></html>\n",
		"app.rb":     "puts 'hello'\n",
		"main.rs":    "fn main() {}\n",
	}
	for name, content := range scripts {
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644))
	}

	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{},
	}

	result, err := Scan(tmpDir, cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	for name := range scripts {
		assert.Contains(t, paths, name, name+" はスキップされないべき")
	}
}

func TestScan_ImageFilesSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	// 画像ファイル（実際のバイナリヘッダ付き）
	// PNG ヘッダ
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "logo.png"), pngHeader, 0644))

	// JPEG ヘッダ
	jpgHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "photo.jpg"), jpgHeader, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "photo.jpeg"), jpgHeader, 0644))

	// GIF ヘッダ
	gifHeader := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x00, 0x00}
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "anim.gif"), gifHeader, 0644))

	// WebP ヘッダ
	webpHeader := []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00}
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "icon.webp"), webpHeader, 0644))

	// ICO（中身はダミーテキストだが拡張子で除外される）
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "favicon.ico"), []byte("dummy"), 0644))

	// SVG はテキストなので含まれるべき
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "icon.svg"),
		[]byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle r="10"/></svg>`),
		0644,
	))

	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{},
	}

	result, err := Scan(tmpDir, cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	assert.NotContains(t, paths, "logo.png", "PNG はスキップされるべき")
	assert.NotContains(t, paths, "photo.jpg", "JPG はスキップされるべき")
	assert.NotContains(t, paths, "photo.jpeg", "JPEG はスキップされるべき")
	assert.NotContains(t, paths, "anim.gif", "GIF はスキップされるべき")
	assert.NotContains(t, paths, "icon.webp", "WebP はスキップされるべき")
	assert.NotContains(t, paths, "favicon.ico", "ICO はスキップされるべき")
	assert.Contains(t, paths, "icon.svg", "SVG はテキストなので含まれるべき")
}

func TestIsBinaryExtension(t *testing.T) {
	tests := []struct {
		path   string
		expect bool
	}{
		// バイナリ（画像）
		{"image.png", true},
		{"photo.jpg", true},
		{"photo.JPEG", true},
		{"icon.gif", true},
		{"icon.webp", true},
		{"icon.ico", true},
		{"icon.bmp", true},
		{"photo.tiff", true},
		{"photo.avif", true},
		{"photo.heic", true},
		// バイナリ（その他）
		{"doc.pdf", true},
		{"app.exe", true},
		{"archive.zip", true},
		{"font.woff2", true},
		{"lib.so", true},
		{"lib.dll", true},
		{"data.db", true},
		// テキスト（スクリプト・コード）
		{"code.go", false},
		{"script.py", false},
		{"build.sh", false},
		{"setup.bat", false},
		{"deploy.ps1", false},
		{"app.rb", false},
		{"main.rs", false},
		{"index.html", false},
		{"style.css", false},
		{"icon.svg", false},
		// テキスト（その他）
		{"readme.md", false},
		{"data.json", false},
		{"config.yml", false},
		{"config.yaml", false},
		{"config.toml", false},
		{"config.xml", false},
		{"noext", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expect, isBinaryExtension(tt.path))
		})
	}
}

func TestIsBinaryContent(t *testing.T) {
	tmpDir := t.TempDir()

	// null バイトを含むファイル
	binaryPath := filepath.Join(tmpDir, "binary")
	require.NoError(t, os.WriteFile(binaryPath, []byte{0x00, 0x01, 0x02}, 0644))
	got, err := isBinaryContent(binaryPath)
	require.NoError(t, err)
	assert.True(t, got, "null バイトを含むファイルはバイナリ判定されるべき")

	// テキストファイル
	textPath := filepath.Join(tmpDir, "text")
	require.NoError(t, os.WriteFile(textPath, []byte("hello\nworld\n"), 0644))
	got, err = isBinaryContent(textPath)
	require.NoError(t, err)
	assert.False(t, got, "テキストファイルはバイナリ判定されないべき")

	// 空ファイル
	emptyPath := filepath.Join(tmpDir, "empty")
	require.NoError(t, os.WriteFile(emptyPath, []byte{}, 0644))
	got, err = isBinaryContent(emptyPath)
	require.NoError(t, err)
	assert.False(t, got, "空ファイルはバイナリ判定されないべき")
}
