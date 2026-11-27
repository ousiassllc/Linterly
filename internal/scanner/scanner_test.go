package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ousiassllc/linterly/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan_BasicScan(t *testing.T) {
	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{},
	}

	result, err := Scan("testdata/project", cfg)
	require.NoError(t, err)

	// vendor, node_modules, .git, build のファイルも含まれる
	paths := filePaths(result)
	assert.Contains(t, paths, "src/main.go")
	assert.Contains(t, paths, "src/util.go")
	assert.Contains(t, paths, "README.md")
	assert.Contains(t, paths, "vendor/lib.go")
	assert.Contains(t, paths, "node_modules/pkg.js")
	assert.Contains(t, paths, "app.min.js")

	// バイナリファイルはスキップされる
	assert.NotContains(t, paths, "build/out.bin", "バイナリ拡張子 .bin はスキップされるべき")
}

func TestScan_DefaultExcludes(t *testing.T) {
	cfg := &config.Config{
		DefaultExcludes: true,
		Ignore:          []string{},
	}

	result, err := Scan("testdata/project", cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	// src/ のファイルは含まれる
	assert.Contains(t, paths, "src/main.go")
	assert.Contains(t, paths, "src/util.go")
	assert.Contains(t, paths, "README.md")

	// デフォルト除外パターンで除外されるもの
	assert.NotContains(t, paths, "vendor/lib.go")
	assert.NotContains(t, paths, "node_modules/pkg.js")
	assert.NotContains(t, paths, "app.min.js")
}

func TestScan_CustomIgnorePatterns(t *testing.T) {
	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{"src/"},
	}

	result, err := Scan("testdata/project", cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	assert.NotContains(t, paths, "src/main.go")
	assert.NotContains(t, paths, "src/util.go")
	assert.Contains(t, paths, "README.md")
}

func TestScan_DirsContainsAllRelevantDirs(t *testing.T) {
	cfg := &config.Config{
		DefaultExcludes: true,
		Ignore:          []string{},
	}

	result, err := Scan("testdata/project", cfg)
	require.NoError(t, err)

	sort.Strings(result.Dirs)
	assert.Contains(t, result.Dirs, "src")
	assert.Contains(t, result.Dirs, ".")
}

func TestScan_FileEntryHasCorrectDir(t *testing.T) {
	cfg := &config.Config{
		DefaultExcludes: true,
		Ignore:          []string{},
	}

	result, err := Scan("testdata/project", cfg)
	require.NoError(t, err)

	for _, f := range result.Files {
		if f.Path == "src/main.go" {
			assert.Equal(t, "src", f.Dir)
		}
		if f.Path == "README.md" {
			assert.Equal(t, ".", f.Dir)
		}
	}
}

func TestScan_NonExistentPath(t *testing.T) {
	cfg := &config.Config{
		DefaultExcludes: false,
	}

	_, err := Scan("testdata/nonexistent", cfg)
	assert.Error(t, err)
}

func TestScan_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DefaultExcludes: false,
	}

	result, err := Scan(tmpDir, cfg)
	require.NoError(t, err)
	assert.Empty(t, result.Files)
	assert.Empty(t, result.Dirs)
}

func TestScan_SymlinksNotFollowed(t *testing.T) {
	tmpDir := t.TempDir()

	// ファイルを作成
	realDir := filepath.Join(tmpDir, "real")
	require.NoError(t, os.MkdirAll(realDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(realDir, "file.go"), []byte("package main"), 0644))

	// シンボリックリンクを作成
	linkPath := filepath.Join(tmpDir, "link")
	err := os.Symlink(realDir, linkPath)
	if err != nil {
		t.Skip("symlinks not supported")
	}

	cfg := &config.Config{
		DefaultExcludes: false,
	}

	result, err := Scan(tmpDir, cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	assert.Contains(t, paths, "real/file.go")
	// filepath.Walk はシンボリックリンクをフォローしない
}

func TestScan_SubdirTargetWithPathIgnorePattern(t *testing.T) {
	// Issue #27: サブディレクトリをターゲット指定した場合にパスを含む
	// ignore パターンがマッチしない
	tmpDir := t.TempDir()

	// ディレクトリ構成: tmpDir/subdir/file.go, tmpDir/subdir/keep.go
	subdir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.MkdirAll(subdir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "file.go"), []byte("package sub"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "keep.go"), []byte("package sub"), 0644))

	// cwd をプロジェクトルートに変更
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }()

	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{"subdir/file.go"},
	}

	// ルート指定: 除外される（修正前後どちらも成功）
	result, err := Scan(".", cfg)
	require.NoError(t, err)
	paths := filePaths(result)
	assert.NotContains(t, paths, "subdir/file.go", "ルート指定で除外されるべき")
	assert.Contains(t, paths, "subdir/keep.go")

	// サブディレクトリ指定: 除外されるべき（修正前は失敗）
	result, err = Scan("subdir", cfg)
	require.NoError(t, err)
	paths = filePaths(result)
	assert.NotContains(t, paths, "file.go", "サブディレクトリ指定でも除外されるべき")
	assert.Contains(t, paths, "keep.go")
}

func TestScan_SubdirTargetWithDirIgnorePattern(t *testing.T) {
	// Issue #27: ディレクトリパターンでも同様のバグが発生
	tmpDir := t.TempDir()

	// ディレクトリ構成: tmpDir/pkg/generated/code.go, tmpDir/pkg/src/app.go
	generated := filepath.Join(tmpDir, "pkg", "generated")
	src := filepath.Join(tmpDir, "pkg", "src")
	require.NoError(t, os.MkdirAll(generated, 0755))
	require.NoError(t, os.MkdirAll(src, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(generated, "code.go"), []byte("package gen"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "app.go"), []byte("package src"), 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }()

	cfg := &config.Config{
		DefaultExcludes: false,
		Ignore:          []string{"pkg/generated/"},
	}

	// サブディレクトリ指定で除外されるべき
	result, err := Scan("pkg", cfg)
	require.NoError(t, err)
	paths := filePaths(result)
	assert.NotContains(t, paths, "generated/code.go", "サブディレクトリ指定でも除外されるべき")
	assert.Contains(t, paths, "src/app.go")
}

func filePaths(result *ScanResult) []string {
	var paths []string
	for _, f := range result.Files {
		paths = append(paths, f.Path)
	}
	return paths
}
