package counter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/ousiassllc/linterly/internal/config"
)

// maxLineKeepSize は code_only モードで 1 行あたりに保持する最大バイト数（1MB）。
// これを超える長大行は先頭のみ判定に使い、残りは読み捨てる（行数には 1 行として数える）。
const maxLineKeepSize = 1024 * 1024

// readChunkSize は countAll の読み取りチャンクサイズ。
const readChunkSize = 64 * 1024

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

	if mode == config.CountModeCodeOnly {
		lang := DetectLanguage(path)
		total, code, err := countCodeOnly(f, lang)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		result.TotalLines = total
		result.CodeLines = code
	} else {
		total, err := countAll(f)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		result.TotalLines = total
		result.CodeLines = total
	}

	return result, nil
}

// maxWorkers は CountFiles が同時に処理するファイル数の上限。
// ファイルごとに goroutine を起動すると fd 上限（ulimit -n）を超えるため、CPU 数に比例した固定数に抑える。
var maxWorkers = runtime.NumCPU()

// CountFiles は複数ファイルの行数を並行してカウントする。
// 同時に処理するファイル数は maxWorkers 以下に制限される。
// 返されるスライスは入力の files スライスと同じインデックス順序を保証する。
// つまり results[i] は files[i] のカウント結果に対応する。
func CountFiles(files []string, mode string) ([]LineCount, error) {
	return countFiles(files, mode, CountFile)
}

// countFiles は countFn をワーカープールで並行実行する。CountFiles の実体（テストで countFn を差し替える）。
func countFiles(files []string, mode string, countFn func(string, string) (*LineCount, error)) ([]LineCount, error) {
	results := make([]LineCount, len(files))
	errs := make([]error, len(files))

	workers := min(maxWorkers, len(files))
	indexes := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range indexes {
				lc, err := countFn(files[i], mode)
				if err != nil {
					errs[i] = err
					continue
				}
				results[i] = *lc
			}
		}()
	}
	for i := range files {
		indexes <- i
	}
	close(indexes)
	wg.Wait()

	// 入力順で最初のエラーを返す
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

// countAll はファイルの全行数をカウントする。行の長さには依存しない。
func countAll(r io.Reader) (int, error) {
	buf := make([]byte, readChunkSize)
	count := 0
	var last byte
	read := false
	for {
		n, err := r.Read(buf)
		if n > 0 {
			read = true
			count += bytes.Count(buf[:n], []byte{'\n'})
			last = buf[n-1]
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	// 末尾が改行で終わらない最終行も 1 行として数える
	if read && last != '\n' {
		count++
	}
	return count, nil
}

// readLine は r から 1 行（改行を除く）を読み取る。
// 行が limit バイトを超える場合、先頭 limit バイトのみ buf に保持し残りは読み捨てる。
// 行を 1 つも読めずに EOF に達した場合は ok=false を返す。
func readLine(r *bufio.Reader, buf []byte, limit int) (line []byte, ok bool, err error) {
	buf = buf[:0]
	for {
		chunk, err := r.ReadSlice('\n')
		if len(chunk) > 0 {
			ok = true
			if room := limit - len(buf); room > 0 {
				buf = append(buf, chunk[:min(room, len(chunk))]...)
			}
		}
		switch {
		case err == nil:
			return bytes.TrimSuffix(buf, []byte{'\n'}), true, nil
		case errors.Is(err, bufio.ErrBufferFull):
			continue
		case errors.Is(err, io.EOF):
			return buf, ok, nil
		default:
			return nil, false, err
		}
	}
}

// countCodeOnly はコード行数を計算する（コメント・空行除外）。
func countCodeOnly(r io.Reader, lang *Language) (total int, code int, err error) {
	reader := bufio.NewReader(r)
	lineBuf := make([]byte, 0, bufio.MaxScanTokenSize)
	inBlock := false
	// Python docstring のためのトラッカー
	isPython := lang != nil && lang.Name == "Python"

	for {
		lineBytes, ok, err := readLine(reader, lineBuf, maxLineKeepSize)
		if err != nil {
			return 0, 0, err
		}
		if !ok {
			break
		}
		lineBuf = lineBytes[:0]
		total++
		trimmed := strings.TrimSpace(string(lineBytes))

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

	return total, code, nil
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
