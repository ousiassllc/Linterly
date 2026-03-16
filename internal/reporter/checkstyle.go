package reporter

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/ousiassllc/linterly/internal/analyzer"
)

// CheckstyleReporter は checkstyle XML 形式で結果を出力する。
type CheckstyleReporter struct {
	writer io.Writer
}

type checkstyleOutput struct {
	XMLName xml.Name         `xml:"checkstyle"`
	Version string           `xml:"version,attr"`
	Files   []checkstyleFile `xml:"file"`
}

type checkstyleFile struct {
	Name   string            `xml:"name,attr"`
	Errors []checkstyleError `xml:"error"`
}

type checkstyleError struct {
	Line     int    `xml:"line,attr"`
	Severity string `xml:"severity,attr"`
	Message  string `xml:"message,attr"`
	Source   string `xml:"source,attr"`
}

// Report は分析結果を checkstyle XML 形式で出力する。
// warnings は checkstyle XML に出力する機構がないため無視する。
func (r *CheckstyleReporter) Report(report *analyzer.AnalysisReport, _ []string) error {
	output := checkstyleOutput{
		Version: "4.3",
	}

	for _, result := range report.Results {
		if result.Severity == analyzer.SeverityPass {
			continue
		}

		severity := "warning"
		if result.Severity == analyzer.SeverityError {
			severity = "error"
		}

		f := checkstyleFile{
			Name: result.Path,
			Errors: []checkstyleError{
				{
					Line:     1,
					Severity: severity,
					Message:  fmt.Sprintf("%d lines (limit: %d)", result.Lines, result.Limit),
					Source:   "linterly",
				},
			},
		}

		output.Files = append(output.Files, f)
	}

	data, err := xml.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.writer, "%s%s\n", xml.Header, data)
	return err
}
