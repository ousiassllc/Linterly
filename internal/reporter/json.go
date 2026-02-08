package reporter

import (
	"encoding/json"
	"io"

	"github.com/ousiassllc/linterly/internal/analyzer"
)

// JSONReporter は JSON 形式で結果を出力する。
type JSONReporter struct {
	writer io.Writer
}

// jsonOutput は JSON 出力の構造体。
type jsonOutput struct {
	Results []jsonResult `json:"results"`
	Summary jsonSummary  `json:"summary"`
}

type jsonResult struct {
	Path      string `json:"path"`
	Type      string `json:"type"`
	Lines     int    `json:"lines"`
	Limit     int    `json:"limit"`
	Threshold int    `json:"threshold"`
	Severity  string `json:"severity"`
}

type jsonSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Passed   int `json:"passed"`
	Total    int `json:"total"`
}

// Report は分析結果を JSON 形式で出力する。
// JSON 出力には pass を含む全結果を出力する。
func (r *JSONReporter) Report(report *analyzer.AnalysisReport, warnings []string) error {
	output := jsonOutput{
		Results: make([]jsonResult, 0, len(report.Results)),
		Summary: jsonSummary{
			Errors:   report.Errors,
			Warnings: report.Warnings,
			Passed:   report.Passed,
			Total:    report.Errors + report.Warnings + report.Passed,
		},
	}

	for _, result := range report.Results {
		output.Results = append(output.Results, jsonResult{
			Path:      result.Path,
			Type:      result.Type,
			Lines:     result.Lines,
			Limit:     result.Limit,
			Threshold: result.Threshold,
			Severity:  string(result.Severity),
		})
	}

	encoder := json.NewEncoder(r.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
