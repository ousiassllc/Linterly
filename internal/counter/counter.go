package counter

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

// LineCount はファイルの行数カウント結果。
type LineCount struct {
	Path       string
	TotalLines int // 全行数
	CodeLines  int // コード行数（コメント・空行除外）
}

// CountFile は指定ファイルの行数をカウントする。
func CountFile(path string, mode string) (*LineCount, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := &LineCount{Path: path}

	if mode == "code_only" {
		lang := DetectLanguage(path)
		result.TotalLines, result.CodeLines = countCodeOnly(f, lang)
	} else {
		total := countAll(f)
		result.TotalLines = total
		result.CodeLines = total
	}

	return result, nil
}

// CountFiles は複数ファイルの行数を並行してカウントする。
func CountFiles(files []string, mode string) ([]LineCount, error) {
	type countResult struct {
		lineCount LineCount
		err       error
		index     int
	}

	results := make([]LineCount, len(files))
	ch := make(chan countResult, len(files))
	var wg sync.WaitGroup

	for i, path := range files {
		wg.Add(1)
		go func(idx int, p string) {
			defer wg.Done()
			lc, err := CountFile(p, mode)
			if err != nil {
				ch <- countResult{err: err, index: idx}
				return
			}
			ch <- countResult{lineCount: *lc, index: idx}
		}(i, path)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var firstErr error
	for cr := range ch {
		if cr.err != nil {
			if firstErr == nil {
				firstErr = cr.err
			}
			continue
		}
		results[cr.index] = cr.lineCount
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return results, nil
}

// countAll はファイルの全行数をカウントする。
func countAll(f *os.File) int {
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}

// countCodeOnly はコード行数を計算する（コメント・空行除外）。
func countCodeOnly(f *os.File, lang *Language) (total int, code int) {
	scanner := bufio.NewScanner(f)
	inBlock := false
	// Python docstring のためのトラッカー
	isPython := lang != nil && lang.Name == "Python"

	for scanner.Scan() {
		total++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 空行チェック
		if trimmed == "" {
			continue
		}

		// 対応言語がない場合、すべてコード行として扱う
		if lang == nil {
			code++
			continue
		}

		// ブロックコメント内にいる場合
		if inBlock {
			if containsBlockEnd(trimmed, lang, isPython) {
				inBlock = false
			}
			continue
		}

		// ブロックコメント開始チェック
		if lang.BlockCommentStart != "" && containsBlockStart(trimmed, lang, isPython) {
			// 同じ行で開始・終了する場合
			if sameLineBlockComment(trimmed, lang, isPython) {
				continue
			}
			inBlock = true
			continue
		}

		// 行コメントチェック
		if isLineComment(trimmed, lang) {
			continue
		}

		code++
	}

	return total, code
}

// isLineComment は行が行コメントであるかを判定する。
func isLineComment(trimmed string, lang *Language) bool {
	for _, prefix := range lang.LineCommentStart {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// containsBlockStart は行がブロックコメント開始を含むかを判定する。
func containsBlockStart(trimmed string, lang *Language, isPython bool) bool {
	if isPython {
		return strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`)
	}
	return strings.HasPrefix(trimmed, lang.BlockCommentStart)
}

// containsBlockEnd は行がブロックコメント終了を含むかを判定する。
func containsBlockEnd(trimmed string, lang *Language, isPython bool) bool {
	if isPython {
		// 開始行と同じ行の場合は sameLineBlockComment で処理済み
		// ここでは終了行のみチェック
		return strings.HasSuffix(trimmed, `"""`) || strings.HasSuffix(trimmed, `'''`)
	}
	return strings.Contains(trimmed, lang.BlockCommentEnd)
}

// sameLineBlockComment は同じ行でブロックコメントが開始・終了するかを判定する。
func sameLineBlockComment(trimmed string, lang *Language, isPython bool) bool {
	if isPython {
		// """...""" or '''...'''
		for _, delim := range []string{`"""`, `'''`} {
			if strings.HasPrefix(trimmed, delim) {
				rest := trimmed[len(delim):]
				if strings.Contains(rest, delim) {
					return true
				}
			}
		}
		return false
	}

	if strings.HasPrefix(trimmed, lang.BlockCommentStart) {
		rest := trimmed[len(lang.BlockCommentStart):]
		return strings.Contains(rest, lang.BlockCommentEnd)
	}
	return false
}
