package analyzer

import (
	"path/filepath"

	"github.com/ousiassllc/linterly/internal/config"
	"github.com/ousiassllc/linterly/internal/counter"
	"github.com/ousiassllc/linterly/internal/scanner"
)

// Severity は違反レベル。
type Severity string

const (
	SeverityPass  Severity = "pass"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

// Result は1つのチェック結果。
type Result struct {
	Path      string   `json:"path"`
	Type      string   `json:"type"`      // "file" または "directory"
	Lines     int      `json:"lines"`     // 実際の行数
	Limit     int      `json:"limit"`     // 設定上限
	Threshold int      `json:"threshold"` // warn/error 境界値
	Severity  Severity `json:"severity"`
}

// AnalysisReport は全体のチェック結果。
type AnalysisReport struct {
	Results  []Result
	Errors   int
	Warnings int
	Passed   int
}

// Analyze はカウント結果をルール設定と比較し、レポートを返す。
func Analyze(counts []counter.LineCount, scanResult *scanner.ScanResult, cfg *config.Config) *AnalysisReport {
	report := &AnalysisReport{}

	maxFile := cfg.Rules.MaxLinesPerFile
	maxDir := cfg.Rules.MaxLinesPerDirectory
	thresholdPct := cfg.Rules.WarningThreshold
	codeOnly := cfg.CountMode == "code_only"

	fileThreshold := calcThreshold(maxFile, thresholdPct)
	dirThreshold := calcThreshold(maxDir, thresholdPct)

	// ファイルごとのチェック
	for _, lc := range counts {
		lines := lc.TotalLines
		if codeOnly {
			lines = lc.CodeLines
		}

		severity := judgeSeverity(lines, maxFile, fileThreshold)
		result := Result{
			Path:      filepath.ToSlash(lc.Path),
			Type:      "file",
			Lines:     lines,
			Limit:     maxFile,
			Threshold: fileThreshold,
			Severity:  severity,
		}
		report.Results = append(report.Results, result)
		countSeverity(report, severity)
	}

	// ディレクトリごとのチェック（直下ファイルのみ集計）
	dirLines := calcDirectoryLines(counts, codeOnly)

	for _, dir := range scanResult.Dirs {
		lines := dirLines[dir]
		severity := judgeSeverity(lines, maxDir, dirThreshold)
		dirPath := dir
		if dirPath != "." {
			dirPath = dirPath + "/"
		} else {
			dirPath = "./"
		}
		result := Result{
			Path:      dirPath,
			Type:      "directory",
			Lines:     lines,
			Limit:     maxDir,
			Threshold: dirThreshold,
			Severity:  severity,
		}
		report.Results = append(report.Results, result)
		countSeverity(report, severity)
	}

	return report
}

// calcThreshold は warn/error 境界値を計算する。
func calcThreshold(limit int, thresholdPct int) int {
	return limit + limit*thresholdPct/100
}

// judgeSeverity は行数と上限・閾値から severity を判定する。
func judgeSeverity(lines, limit, threshold int) Severity {
	if lines <= limit {
		return SeverityPass
	}
	if lines <= threshold {
		return SeverityWarn
	}
	return SeverityError
}

// countSeverity はレポートの集計値を更新する。
func countSeverity(report *AnalysisReport, severity Severity) {
	switch severity {
	case SeverityPass:
		report.Passed++
	case SeverityWarn:
		report.Warnings++
	case SeverityError:
		report.Errors++
	}
}

// calcDirectoryLines はディレクトリ直下のファイルの行数を集計する。
func calcDirectoryLines(counts []counter.LineCount, codeOnly bool) map[string]int {
	dirLines := make(map[string]int)
	for _, lc := range counts {
		dir := filepath.ToSlash(filepath.Dir(lc.Path))
		if dir == "." {
			dir = "."
		}
		lines := lc.TotalLines
		if codeOnly {
			lines = lc.CodeLines
		}
		dirLines[dir] += lines
	}
	return dirLines
}
