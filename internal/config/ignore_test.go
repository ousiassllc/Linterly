package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIgnorePatterns_FromConfig(t *testing.T) {
	// .linterlyignore が存在しないディレクトリで実行
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg := &Config{
		Ignore: []string{"vendor/**", "*.pb.go"},
	}

	patterns, warnings, err := cfg.IgnorePatterns()
	require.NoError(t, err)
	assert.Equal(t, []string{"vendor/**", "*.pb.go"}, patterns)
	assert.Empty(t, warnings)
}

func TestIgnorePatterns_FromLinterlyIgnore(t *testing.T) {
	tmpDir := t.TempDir()
	ignoreContent := "vendor/\n*.pb.go\ngenerated/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterlyignore"), []byte(ignoreContent), 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg := &Config{
		Ignore: nil,
	}

	patterns, warnings, err := cfg.IgnorePatterns()
	require.NoError(t, err)
	assert.Equal(t, []string{"vendor/", "*.pb.go", "generated/"}, patterns)
	assert.Empty(t, warnings)
}

func TestIgnorePatterns_LinterlyIgnoreTakesPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	ignoreContent := "vendor/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterlyignore"), []byte(ignoreContent), 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg := &Config{
		Ignore: []string{"*.pb.go", "generated/"},
	}

	patterns, warnings, err := cfg.IgnorePatterns()
	require.NoError(t, err)
	// .linterlyignore の内容のみが返される
	assert.Equal(t, []string{"vendor/"}, patterns)
	// 両方定義されている場合は警告
	require.Len(t, warnings, 1)
	assert.Equal(t, "ignore.both_defined", warnings[0])
}

func TestIgnorePatterns_NeitherExists(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg := &Config{
		Ignore: nil,
	}

	patterns, warnings, err := cfg.IgnorePatterns()
	require.NoError(t, err)
	assert.Empty(t, patterns)
	assert.Empty(t, warnings)
}

func TestIgnorePatterns_EmptyConfigIgnore(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg := &Config{
		Ignore: []string{},
	}

	patterns, warnings, err := cfg.IgnorePatterns()
	require.NoError(t, err)
	assert.Empty(t, patterns)
	assert.Empty(t, warnings)
}

func TestReadLinterlyIgnore_CommentsAndEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	content := "# Comment\n\nvendor/\n  # Indented comment\n\n*.pb.go\n"
	path := filepath.Join(tmpDir, ".linterlyignore")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	patterns, err := readLinterlyIgnore(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"vendor/", "*.pb.go"}, patterns)
}

func TestReadLinterlyIgnore_AllCommentsAndEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	content := "# Only comments\n\n# Another comment\n\n"
	path := filepath.Join(tmpDir, ".linterlyignore")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	patterns, err := readLinterlyIgnore(path)
	require.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestReadLinterlyIgnore_FileNotFound(t *testing.T) {
	_, err := readLinterlyIgnore("/nonexistent/.linterlyignore")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestReadLinterlyIgnore_WithNegationPattern(t *testing.T) {
	tmpDir := t.TempDir()
	content := "vendor/\n!vendor/important.go\n"
	path := filepath.Join(tmpDir, ".linterlyignore")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	patterns, err := readLinterlyIgnore(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"vendor/", "!vendor/important.go"}, patterns)
}

func TestIgnorePatterns_CacheReturnsSameResult(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg := &Config{
		Ignore: []string{"vendor/**", "*.pb.go"},
	}

	// 1回目の呼び出し
	patterns1, warnings1, err1 := cfg.IgnorePatterns()
	require.NoError(t, err1)
	assert.Equal(t, []string{"vendor/**", "*.pb.go"}, patterns1)
	assert.Empty(t, warnings1)

	// 2回目の呼び出し — キャッシュから同じ結果が返る
	patterns2, warnings2, err2 := cfg.IgnorePatterns()
	require.NoError(t, err2)
	assert.Equal(t, patterns1, patterns2)
	assert.Equal(t, warnings1, warnings2)
}

func TestIgnorePatterns_CacheClearedByApplyOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg := &Config{
		Rules: Rules{
			MaxLinesPerFile:      300,
			MaxLinesPerDirectory: 2000,
			WarningThreshold:     10,
		},
		CountMode: CountModeAll,
		Ignore:    []string{"vendor/**"},
		Language:  "en",
	}

	// 1回目の呼び出しでキャッシュを生成
	patterns1, _, err := cfg.IgnorePatterns()
	require.NoError(t, err)
	assert.Equal(t, []string{"vendor/**"}, patterns1)

	// ApplyOverrides で Ignore を変更 → キャッシュクリア
	newIgnore := []string{"dist/**"}
	require.NoError(t, cfg.ApplyOverrides(&Overrides{Ignore: newIgnore}))

	// 新しいパターンが返る
	patterns2, _, err := cfg.IgnorePatterns()
	require.NoError(t, err)
	assert.Equal(t, []string{"dist/**"}, patterns2)
}

func TestIgnorePatterns_LinterlyIgnoreNoWarningWhenConfigIgnoreEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	ignoreContent := "vendor/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterlyignore"), []byte(ignoreContent), 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg := &Config{
		Ignore: []string{},
	}

	patterns, warnings, err := cfg.IgnorePatterns()
	require.NoError(t, err)
	assert.Equal(t, []string{"vendor/"}, patterns)
	// ignore が空の場合は警告なし
	assert.Empty(t, warnings)
}
