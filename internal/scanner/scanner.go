package scanner

import (
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/denormal/go-gitignore"

	"github.com/ousiassllc/linterly/internal/config"
)

// FileEntry は走査で見つかったファイルの情報。
type FileEntry struct {
	Path string // ターゲットパスからの相対パス
	Dir  string // ファイルが属するディレクトリ（相対パス）
}

// ScanResult は走査結果。
type ScanResult struct {
	Files []FileEntry
	Dirs  []string // チェック対象のディレクトリ一覧（重複なし）
}

// Scan は指定パスを走査し、除外パターンを適用した結果を返す。
func Scan(targetPath string, cfg *config.Config) (*ScanResult, error) {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, err
	}

	// プロジェクトルート（設定ファイル・.linterlyignore の基準ディレクトリ）
	projectRoot, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	matcher, err := buildMatcher(projectRoot, cfg)
	if err != nil {
		return nil, err
	}

	result := &ScanResult{}
	dirSet := make(map[string]bool)

	err = filepath.Walk(absTarget, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// ターゲット相対パス（FileEntry 用）
		relFromTarget, err := filepath.Rel(absTarget, path)
		if err != nil {
			return err
		}

		// ルートディレクトリ自体はスキップ
		if relFromTarget == "." {
			return nil
		}

		// プロジェクトルート相対パス（ignore マッチング用）
		relFromRoot, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}

		// スラッシュ区切りに正規化
		relFromTarget = filepath.ToSlash(relFromTarget)
		relFromRoot = filepath.ToSlash(relFromRoot)

		if info.IsDir() {
			if shouldExclude(matcher, relFromRoot, true) {
				return filepath.SkipDir
			}
			return nil
		}

		// ファイル
		if shouldExclude(matcher, relFromRoot, false) {
			return nil
		}

		dir := filepath.ToSlash(filepath.Dir(relFromTarget))
		if dir == "." {
			dir = "."
		}

		result.Files = append(result.Files, FileEntry{
			Path: relFromTarget,
			Dir:  dir,
		})

		if !dirSet[dir] {
			dirSet[dir] = true
			result.Dirs = append(result.Dirs, dir)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// ルートディレクトリも含める（ファイルがある場合）
	if !dirSet["."] {
		for _, f := range result.Files {
			if f.Dir == "." {
				result.Dirs = append([]string{"."}, result.Dirs...)
				break
			}
		}
	}

	return result, nil
}

// buildMatcher は除外パターンから gitignore マッチャーを構築する。
func buildMatcher(basePath string, cfg *config.Config) (gitignore.GitIgnore, error) {
	var patterns []string

	// デフォルト除外パターン
	if cfg.DefaultExcludes {
		patterns = append(patterns, config.DefaultExcludePatterns()...)
	}

	// ユーザー定義の除外パターン
	ignorePatterns, _, err := cfg.IgnorePatterns()
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, ignorePatterns...)

	if len(patterns) == 0 {
		return nil, nil
	}

	// パターンを改行区切りの文字列にして gitignore パーサーに渡す
	content := strings.Join(patterns, "\n")
	reader := strings.NewReader(content)

	return gitignore.New(reader, basePath, nil), nil
}

// shouldExclude はパスが除外パターンにマッチするかを返す。
func shouldExclude(matcher gitignore.GitIgnore, relPath string, isDir bool) bool {
	if matcher == nil {
		return false
	}

	match := matcher.Relative(relPath, isDir)
	if match == nil {
		return false
	}
	return match.Ignore()
}
