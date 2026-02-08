package reporter

import (
	"io"

	"github.com/ousiassllc/linterly/internal/analyzer"
	"github.com/ousiassllc/linterly/internal/i18n"
)

// Reporter は結果出力のインターフェース。
type Reporter interface {
	Report(report *analyzer.AnalysisReport, warnings []string) error
}

// NewReporter はフォーマット指定に応じた Reporter を返す。
func NewReporter(format string, translator *i18n.Translator, writer io.Writer) Reporter {
	if format == "json" {
		return &JSONReporter{writer: writer}
	}
	return &TextReporter{
		writer:     writer,
		translator: translator,
	}
}
