package reporter

import (
	"io"
	"os"
	"slices"

	"github.com/ousiassllc/linterly/internal/analyzer"
	"github.com/ousiassllc/linterly/internal/i18n"
)

const (
	FormatText       = "text"
	FormatJSON       = "json"
	FormatJUnit      = "junit"
	FormatCheckstyle = "checkstyle"
)

// SupportedFormats は --format に指定できる値の一覧。
var SupportedFormats = []string{FormatText, FormatJSON, FormatJUnit, FormatCheckstyle}

// IsSupportedFormat は format が対応しているフォーマット名かを返す。
func IsSupportedFormat(format string) bool {
	return slices.Contains(SupportedFormats, format)
}

// Reporter は結果出力のインターフェース。
type Reporter interface {
	Report(report *analyzer.AnalysisReport, warnings []string) error
}

// NewReporter はフォーマット指定に応じた Reporter を返す。
// 未知のフォーマットはテキスト Reporter になるため、呼び出し側で IsSupportedFormat により検証すること。
func NewReporter(format string, translator *i18n.Translator, writer io.Writer) Reporter {
	switch format {
	case FormatJSON:
		return &JSONReporter{writer: writer}
	case FormatJUnit:
		return &JUnitReporter{writer: writer}
	case FormatCheckstyle:
		return &CheckstyleReporter{writer: writer}
	default:
		return &TextReporter{
			writer:     writer,
			translator: translator,
			noColor:    os.Getenv("NO_COLOR") != "",
		}
	}
}
