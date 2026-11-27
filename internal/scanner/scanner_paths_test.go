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

func TestScan_FileTarget(t *testing.T) {
	// Issue #49: ファイルを直接指定した場合、そのファイル 1 件が登録される
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a\nb\n"), 0644))

	t.Chdir(tmpDir)
	cfg := &config.Config{DefaultExcludes: false}

	result, err := Scan("a.txt", cfg)
	require.NoError(t, err)
	require.Len(t, result.Files, 1)
	assert.Equal(t, "a.txt", result.Files[0].Path)
	assert.Equal(t, ".", result.Files[0].Dir)
	assert.Equal(t, filepath.Join(tmpDir, "a.txt"), result.Files[0].AbsPath)
	assert.Equal(t, []string{"."}, result.Dirs)
}

func TestScan_FileTargetInSubdir_PathIsRelativeToProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "src"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "src", "a.go"), []byte("package a\n"), 0644))

	t.Chdir(tmpDir)
	cfg := &config.Config{DefaultExcludes: false}

	result, err := Scan("src/a.go", cfg)
	require.NoError(t, err)
	require.Len(t, result.Files, 1)
	assert.Equal(t, "src/a.go", result.Files[0].Path)
	assert.Equal(t, "src", result.Files[0].Dir)
	assert.Equal(t, []string{"src"}, result.Dirs)
}

func TestScan_FileTargetRespectsExcludeAndBinary(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "vendor"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "vendor", "lib.go"), []byte("package v\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "ignored.txt"), []byte("x\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "bin.dat"), []byte{'a', 0x00, 'b'}, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "img.png"), []byte("not really png"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "keep.txt"), []byte("k\n"), 0644))

	t.Chdir(tmpDir)
	cfg := &config.Config{DefaultExcludes: true, Ignore: []string{"ignored.txt"}}

	result, err := ScanPaths([]string{"vendor/lib.go", "ignored.txt", "bin.dat", "img.png", "keep.txt"}, cfg)
	require.NoError(t, err)
	assert.Equal(t, []string{"keep.txt"}, filePaths(result), "除外・バイナリは直接指定でも適用される")
}

func TestScanPaths_MultipleTargetsAreRelativeToProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	for _, f := range []string{"a/x.go", "a/y.go", "b/z.go", "top.go"} {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(tmpDir, f)), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, f), []byte("package p\n"), 0644))
	}

	t.Chdir(tmpDir)
	cfg := &config.Config{DefaultExcludes: false}

	// ディレクトリとファイルの混在、重複指定（a と a/x.go）
	result, err := ScanPaths([]string{"a", "b/z.go", "a/x.go", "top.go"}, cfg)
	require.NoError(t, err)

	paths := filePaths(result)
	sort.Strings(paths)
	assert.Equal(t, []string{"a/x.go", "a/y.go", "b/z.go", "top.go"}, paths)
	assert.ElementsMatch(t, []string{"a", "b", "."}, result.Dirs)
}

func TestScanPaths_NonExistentTarget(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("package a\n"), 0644))

	t.Chdir(tmpDir)
	_, err := ScanPaths([]string{"a.go", "missing.go"}, &config.Config{})
	assert.Error(t, err)
}

func TestScan_UnreadableEntriesAreSkipped(t *testing.T) {
	// Issue #52: 読み取り不能なファイル・ディレクトリがあっても走査は継続し、Skipped に記録される
	if os.Geteuid() == 0 {
		t.Skip("root は権限チェックを回避するためスキップ")
	}
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "locked"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "ok.txt"), []byte("ok\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "sub", "ok.txt"), []byte("ok\n"), 0644))
	denied := filepath.Join(tmpDir, "sub", "denied.txt")
	require.NoError(t, os.WriteFile(denied, []byte("x\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "locked", "f.txt"), []byte("x\n"), 0644))
	require.NoError(t, os.Chmod(denied, 0000))
	require.NoError(t, os.Chmod(filepath.Join(tmpDir, "locked"), 0000))
	t.Cleanup(func() {
		_ = os.Chmod(denied, 0644)
		_ = os.Chmod(filepath.Join(tmpDir, "locked"), 0755)
	})

	result, err := Scan(tmpDir, &config.Config{DefaultExcludes: false})
	require.NoError(t, err)

	paths := filePaths(result)
	sort.Strings(paths)
	assert.Equal(t, []string{"ok.txt", "sub/ok.txt"}, paths)

	var skipped []string
	for _, s := range result.Skipped {
		assert.ErrorIs(t, s.Err, os.ErrPermission)
		skipped = append(skipped, s.Path)
	}
	assert.ElementsMatch(t, []string{denied, filepath.Join(tmpDir, "locked")}, skipped)
}
