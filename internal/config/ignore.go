package config

import (
	"bufio"
	"os"
	"strings"
)

const linterlyIgnoreFile = ".linterlyignore"

// IgnorePatterns は有効な除外パターン一覧を返す。
// .linterlyignore が存在すればそちらを優先し、設定ファイルにも ignore が定義されている場合は warnings に警告を追加する。
// .linterlyignore が存在しない場合は設定ファイルの ignore フィールドを返す。
func (c *Config) IgnorePatterns() (patterns []string, warnings []string, err error) {
	ignorePatterns, fileErr := readLinterlyIgnore(linterlyIgnoreFile)
	if fileErr != nil {
		if !os.IsNotExist(fileErr) {
			return nil, nil, fileErr
		}
		// .linterlyignore が存在しない場合は設定ファイルの ignore を使用
		return c.Ignore, nil, nil
	}

	// .linterlyignore が存在する場合
	if len(c.Ignore) > 0 {
		warnings = append(warnings, "Both .linterlyignore and ignore in config file are defined. .linterlyignore takes precedence. ignore in config file is ignored.")
	}
	return ignorePatterns, warnings, nil
}

// readLinterlyIgnore は .linterlyignore ファイルを読み込み、パターンのリストを返す。
// コメント行（# で始まる）と空行は除外する。
func readLinterlyIgnore(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}
