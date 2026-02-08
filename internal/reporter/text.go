package reporter

import (
	"fmt"
	"io"
	"os"

	"github.com/ousiassllc/linterly/internal/analyzer"
	"github.com/ousiassllc/linterly/internal/i18n"
)

// TextReporter はテキスト形式で結果を出力する。
type TextReporter struct {
	writer     io.Writer
	translator *i18n.Translator
}

// Report は分析結果をテキスト形式で出力する。
// テキスト出力では violation（warn/error）のみ表示し、pass は表示しない。
func (r *TextReporter) Report(report *analyzer.AnalysisReport, warnings []string) error {
	noColor := os.Getenv("NO_COLOR") != ""

	// ignore 重複警告を先に出力
	for _, w := range warnings {
		line := fmt.Sprintf("  WARN  %s", w)
		if !noColor {
			line = colorYellow(line)
		}
		fmt.Fprintln(r.writer, line)
	}
	if len(warnings) > 0 {
		fmt.Fprintln(r.writer)
	}

	// 違反のみ出力
	hasViolation := false
	for _, result := range report.Results {
		switch result.Severity {
		case analyzer.SeverityWarn:
			line := r.translator.T("check.warn", result.Path, result.Lines, result.Limit)
			if !noColor {
				line = colorYellow("  " + line)
			} else {
				line = "  " + line
			}
			fmt.Fprintln(r.writer, line)
			hasViolation = true
		case analyzer.SeverityError:
			line := r.translator.T("check.error", result.Path, result.Lines, result.Limit)
			if !noColor {
				line = colorRed("  " + line)
			} else {
				line = "  " + line
			}
			fmt.Fprintln(r.writer, line)
			hasViolation = true
		}
	}

	if hasViolation {
		fmt.Fprintln(r.writer)
	}

	// サマリー
	summary := r.translator.T("check.summary", report.Errors, report.Warnings, report.Passed)
	fmt.Fprintln(r.writer, summary)

	return nil
}

// ANSI カラーコード
func colorRed(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func colorYellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}
