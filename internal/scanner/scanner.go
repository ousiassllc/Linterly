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
	Path    string // 基準ディレクトリからの相対パス（基準は ScanPaths を参照）
	Dir     string // ファイルが属するディレクトリ（Path と同じ基準の相対パス）
	AbsPath string // ファイルの絶対パス（行数カウント用）
}

// SkippedEntry は読み取れずにスキップされたエントリ。
type SkippedEntry struct {
	Path string // 絶対パス
	Err  error
}

// ScanResult は走査結果。
type ScanResult struct {
	Files   []FileEntry
	Dirs    []string       // チェック対象のディレクトリ一覧（重複なし）
	Skipped []SkippedEntry // 読み取れずにスキップしたエントリ
}

// Scan は指定パスを走査し、除外パターンを適用した結果を返す。
// ScanPaths に単一のパスを渡す場合と同じ。
func Scan(targetPath string, cfg *config.Config) (*ScanResult, error) {
	return ScanPaths([]string{targetPath}, cfg)
}

// ScanPaths は指定された複数のパス（ディレクトリまたはファイル）を走査し、除外パターンを適用した結果を返す。
//
// FileEntry.Path / Dir の基準ディレクトリは次のとおり。
//   - 単一のディレクトリを指定した場合: そのディレクトリ
//   - それ以外（ファイルの指定、複数パスの指定）: カレントディレクトリ（プロジェクトルート）
//
// ファイルを直接指定した場合も、除外パターン・バイナリ判定は通常どおり適用する。
// 除外対象のディレクトリを直接指定した場合は、ファイル指定と同様に何も走査しない。
// 指定パスがシンボリックリンクの場合はリンク先を走査する（走査中に出会うリンクは辿らない）。
// 同じファイルが複数回見つかった場合は 1 件として扱う。
// 読み取れないエントリは走査を中断せず、ScanResult.Skipped に記録して続行する。
// 指定パス自体が存在しない場合はエラーを返す。
func ScanPaths(targetPaths []string, cfg *config.Config) (*ScanResult, error) {
	// プロジェクトルート（設定ファイル・.linterlyignore の基準ディレクトリ）
	projectRoot, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	matcher, err := buildMatcher(projectRoot, cfg)
	if err != nil {
		return nil, err
	}

	absTargets := make([]string, len(targetPaths))
	for i, p := range targetPaths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		absTargets[i], err = resolveTarget(abs, projectRoot)
		if err != nil {
			return nil, err
		}
	}

	// 基準ディレクトリ（FileEntry の相対パスの基準）
	base := projectRoot
	if len(absTargets) == 1 {
		if info, err := os.Lstat(absTargets[0]); err == nil && info.IsDir() {
			base = absTargets[0]
		}
	}

	s := &scan{
		base:        base,
		projectRoot: projectRoot,
		matcher:     matcher,
		result:      &ScanResult{},
		dirSet:      make(map[string]bool),
		seen:        make(map[string]bool),
	}
	for _, absTarget := range absTargets {
		if err := s.walk(absTarget); err != nil {
			return nil, err
		}
	}
	return s.result, nil
}

// scan は ScanPaths の走査状態。
type scan struct {
	base        string // FileEntry の相対パスの基準ディレクトリ
	projectRoot string // ignore マッチングの基準ディレクトリ
	matcher     gitignore.GitIgnore
	result      *ScanResult
	dirSet      map[string]bool
	seen        map[string]bool // 登録済みファイルの絶対パス
}

// walk は absTarget（ディレクトリまたはファイル）を走査して結果に追加する。
func (s *scan) walk(absTarget string) error {
	return filepath.Walk(absTarget, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// 指定パス自体にアクセスできない場合は走査できないためエラーとする
			if path == absTarget && info == nil {
				return walkErr
			}
			s.skip(path, walkErr)
			return nil
		}

		// プロジェクトルート相対パス（ignore マッチング用）
		relFromRoot, err := filepath.Rel(s.projectRoot, path)
		if err != nil {
			return err
		}
		relFromRoot = filepath.ToSlash(relFromRoot)

		// 指定ディレクトリ自体は登録しない。除外対象（自身または親が除外）の場合は配下も走査しない
		if path == absTarget && info.IsDir() {
			if s.targetDirExcluded(relFromRoot) {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			if shouldExclude(s.matcher, relFromRoot, true) {
				return filepath.SkipDir
			}
			return nil
		}

		// 正規ファイル以外（シンボリックリンク等）はスキップ
		if !info.Mode().IsRegular() {
			return nil
		}

		// ファイル
		if shouldExclude(s.matcher, relFromRoot, false) {
			return nil
		}

		// 直接指定されたファイルは親ディレクトリを辿らないため、親ディレクトリの除外を明示的に確認する
		if path == absTarget && s.parentExcluded(relFromRoot) {
			return nil
		}

		// 重複登録を避ける
		if s.seen[path] {
			return nil
		}

		// バイナリファイルはスキップ
		binary, err := isBinary(path)
		if err != nil {
			s.skip(path, err)
			return nil
		}
		if binary {
			return nil
		}

		relFromBase, err := filepath.Rel(s.base, path)
		if err != nil {
			return err
		}
		relFromBase = filepath.ToSlash(relFromBase)
		dir := filepath.ToSlash(filepath.Dir(relFromBase))

		s.seen[path] = true
		s.result.Files = append(s.result.Files, FileEntry{
			Path:    relFromBase,
			Dir:     dir,
			AbsPath: path,
		})

		if !s.dirSet[dir] {
			s.dirSet[dir] = true
			s.result.Dirs = append(s.result.Dirs, dir)
		}

		return nil
	})
}

// targetDirExcluded は直接指定されたディレクトリ（プロジェクトルート相対パス）が除外対象かを返す。
// ファイルを直接指定した場合と同様に、ディレクトリ自身と親ディレクトリの除外パターンを評価する。
// プロジェクトルート自身とルート外のディレクトリには除外パターンを適用しない。
func (s *scan) targetDirExcluded(relPath string) bool {
	if relPath == "." || relPath == ".." || strings.HasPrefix(relPath, "../") {
		return false
	}
	return shouldExclude(s.matcher, relPath, true) || s.parentExcluded(relPath)
}

// resolveTarget は指定パスがシンボリックリンクの場合に、リンク先の実体パスへ解決する。
// 実体がプロジェクトルート配下にある場合は、ルートがシンボリックリンク経由でも
// ルート相対パスが変わらないよう projectRoot からのパスで返す。
// リンクでないパスはそのまま返す。リンク先が存在しない場合はエラーを返す。
func resolveTarget(absTarget, projectRoot string) (string, error) {
	info, err := os.Lstat(absTarget)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return absTarget, nil // 存在しない場合のエラーは走査時に報告する
	}
	resolved, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		return "", err
	}
	if realRoot, err := filepath.EvalSymlinks(projectRoot); err == nil {
		if rel, err := filepath.Rel(realRoot, resolved); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return filepath.Join(projectRoot, rel), nil
		}
	}
	return resolved, nil
}

// parentExcluded は relPath（プロジェクトルート相対）の親ディレクトリのいずれかが除外対象かを返す。
func (s *scan) parentExcluded(relPath string) bool {
	parts := strings.Split(relPath, "/")
	for i := 1; i < len(parts); i++ {
		if parts[i-1] == ".." {
			continue // プロジェクトルート外の親は評価しない
		}
		if shouldExclude(s.matcher, strings.Join(parts[:i], "/"), true) {
			return true
		}
	}
	return false
}

// skip は読み取れなかったエントリを記録する。
func (s *scan) skip(path string, err error) {
	s.result.Skipped = append(s.result.Skipped, SkippedEntry{Path: path, Err: err})
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
