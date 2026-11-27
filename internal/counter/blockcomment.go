package counter

import "strings"

// hasCode は 1 行（TrimSpace 済み）にコードが含まれるかを判定し、ブロックコメントの継続状態を inBlock に反映する。
// ブロックコメントの終了後・開始と終了が同じ行にある場合は、残りの文字列を行頭として評価し直す
// （例: "/* c */ x := 1" や "end */ y := 2" はコード行）。
func hasCode(trimmed string, lang *Language, isPython bool, inBlock *bool) bool {
	rest := trimmed
	for rest != "" {
		// ブロックコメント内: 終了まで読み飛ばし、残りを評価し直す
		if *inBlock {
			after, closed := cutBlockEnd(rest, lang, isPython)
			if !closed {
				return false
			}
			*inBlock = false
			rest = strings.TrimSpace(after)
			continue
		}

		// 行頭で始まるブロックコメント（コードの後ろで始まるものは opensBlockAfterCode で扱う）
		if delim := blockStart(rest, lang, isPython); delim != "" {
			afterStart := rest[len(delim):]
			end := lang.BlockCommentEnd
			if isPython {
				end = delim
			}
			_, after, found := strings.Cut(afterStart, end)
			if !found {
				*inBlock = true
				return false
			}
			rest = strings.TrimSpace(after)
			continue
		}

		// 行コメントは行末までコメント
		if isLineComment(rest, lang) {
			return false
		}
		// コードの後ろで始まり、行内で閉じないブロックコメントは次行以降に持ち越す
		if opensBlockAfterCode(rest, lang) {
			*inBlock = true
		}
		return true
	}
	return false
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

// blockStart は行頭にあるブロックコメント開始記号を返す。なければ空文字列を返す。
func blockStart(trimmed string, lang *Language, isPython bool) string {
	if isPython {
		for _, delim := range []string{`"""`, `'''`} {
			if strings.HasPrefix(trimmed, delim) {
				return delim
			}
		}
		return ""
	}
	if lang.BlockCommentStart != "" && strings.HasPrefix(trimmed, lang.BlockCommentStart) {
		return lang.BlockCommentStart
	}
	return ""
}

// cutBlockEnd はブロックコメント内の行から終了記号を探し、終了記号より後ろの文字列を返す。
// Python の docstring は従来どおり、行末が終了記号で終わる場合のみ終了とみなす。
func cutBlockEnd(trimmed string, lang *Language, isPython bool) (after string, closed bool) {
	if isPython {
		return "", strings.HasSuffix(trimmed, `"""`) || strings.HasSuffix(trimmed, `'''`)
	}
	_, after, closed = strings.Cut(trimmed, lang.BlockCommentEnd)
	return after, closed
}

// opensBlockAfterCode は、コードを含む行の途中で始まり、同じ行で閉じないブロックコメントがあるかを返す。
// 行コメントの開始以降は評価しない。文字列リテラル（"..." / '...' / `...`）内の記号は無視する。
// 閉じていない引用符は文字列とみなさない（Rust のライフタイムや英文のアポストロフィ対策）。
// 行頭でしか成立しない記法（Python の docstring、Ruby の =begin）は対象外。
func opensBlockAfterCode(code string, lang *Language) bool {
	start, end := lang.BlockCommentStart, lang.BlockCommentEnd
	if start != "/*" && start != "<!--" {
		return false
	}
	for i := 0; i < len(code); {
		rest := code[i:]
		switch {
		case isLineComment(rest, lang):
			return false
		case strings.HasPrefix(rest, start):
			_, after, closed := strings.Cut(rest[len(start):], end)
			if !closed {
				return true
			}
			i = len(code) - len(after)
		case rest[0] == '"' || rest[0] == '\'' || rest[0] == '`':
			if n := quotedLen(rest); n > 0 {
				i += n
			} else {
				i++
			}
		default:
			i++
		}
	}
	return false
}

// quotedLen は s の先頭の引用符で始まる文字列リテラルの長さ（引用符を含む）を返す。
// 同じ行で閉じていない場合は 0 を返す。バッククォートはエスケープを解釈しない。
func quotedLen(s string) int {
	quote := s[0]
	for i := 1; i < len(s); i++ {
		switch {
		case s[i] == '\\' && quote != '`':
			i++
		case s[i] == quote:
			return i + 1
		}
	}
	return 0
}
