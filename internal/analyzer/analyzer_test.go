package analyzer

import (
	"testing"

	"github.com/ousiassllc/linterly/internal/config"
	"github.com/ousiassllc/linterly/internal/counter"
	"github.com/ousiassllc/linterly/internal/scanner"
	"github.com/stretchr/testify/assert"
)

func newTestConfig() *config.Config {
	return &config.Config{
		Rules: config.Rules{
			MaxLinesPerFile:      300,
			MaxLinesPerDirectory: 2000,
			WarningThreshold:     10,
		},
		CountMode: "all",
	}
}

func TestAnalyze_FilePass(t *testing.T) {
	cfg := newTestConfig()
	counts := []counter.LineCount{
		{Path: "src/main.go", TotalLines: 100, CodeLines: 80},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{{Path: "src/main.go", Dir: "src"}},
		Dirs:  []string{"src"},
	}

	report := Analyze(counts, scanResult, cfg)

	// ファイル結果
	fileResult := findResult(report, "src/main.go")
	assert.NotNil(t, fileResult)
	assert.Equal(t, SeverityPass, fileResult.Severity)
	assert.Equal(t, 100, fileResult.Lines)
	assert.Equal(t, 300, fileResult.Limit)
	assert.Equal(t, 330, fileResult.Threshold)
}

func TestAnalyze_FileWarn(t *testing.T) {
	cfg := newTestConfig()
	counts := []counter.LineCount{
		{Path: "src/handler.go", TotalLines: 320, CodeLines: 320},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{{Path: "src/handler.go", Dir: "src"}},
		Dirs:  []string{"src"},
	}

	report := Analyze(counts, scanResult, cfg)

	fileResult := findResult(report, "src/handler.go")
	assert.NotNil(t, fileResult)
	assert.Equal(t, SeverityWarn, fileResult.Severity)
}

func TestAnalyze_FileError(t *testing.T) {
	cfg := newTestConfig()
	counts := []counter.LineCount{
		{Path: "src/service.go", TotalLines: 450, CodeLines: 450},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{{Path: "src/service.go", Dir: "src"}},
		Dirs:  []string{"src"},
	}

	report := Analyze(counts, scanResult, cfg)

	fileResult := findResult(report, "src/service.go")
	assert.NotNil(t, fileResult)
	assert.Equal(t, SeverityError, fileResult.Severity)
}

func TestAnalyze_DirectoryCheck(t *testing.T) {
	cfg := newTestConfig()
	counts := []counter.LineCount{
		{Path: "src/a.go", TotalLines: 1000, CodeLines: 900},
		{Path: "src/b.go", TotalLines: 1200, CodeLines: 1100},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{
			{Path: "src/a.go", Dir: "src"},
			{Path: "src/b.go", Dir: "src"},
		},
		Dirs: []string{"src"},
	}

	report := Analyze(counts, scanResult, cfg)

	dirResult := findResult(report, "src/")
	assert.NotNil(t, dirResult)
	assert.Equal(t, "directory", dirResult.Type)
	assert.Equal(t, 2200, dirResult.Lines)            // 1000 + 1200
	assert.Equal(t, SeverityWarn, dirResult.Severity) // 2200 <= 2200 (threshold)
}

func TestAnalyze_DirectoryOnlyDirectFiles(t *testing.T) {
	cfg := newTestConfig()
	counts := []counter.LineCount{
		{Path: "src/main.go", TotalLines: 100, CodeLines: 80},
		{Path: "src/sub/helper.go", TotalLines: 100, CodeLines: 80},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{
			{Path: "src/main.go", Dir: "src"},
			{Path: "src/sub/helper.go", Dir: "src/sub"},
		},
		Dirs: []string{"src", "src/sub"},
	}

	report := Analyze(counts, scanResult, cfg)

	srcResult := findResult(report, "src/")
	assert.NotNil(t, srcResult)
	assert.Equal(t, 100, srcResult.Lines) // src/main.go のみ
}

func TestAnalyze_Summary(t *testing.T) {
	cfg := newTestConfig()
	counts := []counter.LineCount{
		{Path: "a.go", TotalLines: 100, CodeLines: 100},
		{Path: "b.go", TotalLines: 320, CodeLines: 320},
		{Path: "c.go", TotalLines: 450, CodeLines: 450},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{
			{Path: "a.go", Dir: "."},
			{Path: "b.go", Dir: "."},
			{Path: "c.go", Dir: "."},
		},
		Dirs: []string{"."},
	}

	report := Analyze(counts, scanResult, cfg)

	// ファイル: 1 pass, 1 warn, 1 error
	// ディレクトリ (./ = 870): pass
	assert.Equal(t, 2, report.Passed)   // a.go + dir
	assert.Equal(t, 1, report.Warnings) // b.go
	assert.Equal(t, 1, report.Errors)   // c.go
}

func TestAnalyze_CodeOnlyMode(t *testing.T) {
	cfg := newTestConfig()
	cfg.CountMode = "code_only"

	counts := []counter.LineCount{
		{Path: "src/main.go", TotalLines: 400, CodeLines: 250},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{{Path: "src/main.go", Dir: "src"}},
		Dirs:  []string{"src"},
	}

	report := Analyze(counts, scanResult, cfg)

	fileResult := findResult(report, "src/main.go")
	assert.NotNil(t, fileResult)
	assert.Equal(t, 250, fileResult.Lines)
	assert.Equal(t, SeverityPass, fileResult.Severity)
}

func TestAnalyze_ThresholdZero(t *testing.T) {
	cfg := newTestConfig()
	cfg.Rules.WarningThreshold = 0

	counts := []counter.LineCount{
		{Path: "a.go", TotalLines: 301, CodeLines: 301},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{{Path: "a.go", Dir: "."}},
		Dirs:  []string{"."},
	}

	report := Analyze(counts, scanResult, cfg)

	fileResult := findResult(report, "a.go")
	assert.NotNil(t, fileResult)
	assert.Equal(t, SeverityError, fileResult.Severity) // threshold = 300, 301 > 300
}

func TestAnalyze_ExactLimit(t *testing.T) {
	cfg := newTestConfig()
	counts := []counter.LineCount{
		{Path: "a.go", TotalLines: 300, CodeLines: 300},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{{Path: "a.go", Dir: "."}},
		Dirs:  []string{"."},
	}

	report := Analyze(counts, scanResult, cfg)

	fileResult := findResult(report, "a.go")
	assert.Equal(t, SeverityPass, fileResult.Severity) // 300 <= 300
}

func TestAnalyze_ExactThreshold(t *testing.T) {
	cfg := newTestConfig()
	counts := []counter.LineCount{
		{Path: "a.go", TotalLines: 330, CodeLines: 330},
	}
	scanResult := &scanner.ScanResult{
		Files: []scanner.FileEntry{{Path: "a.go", Dir: "."}},
		Dirs:  []string{"."},
	}

	report := Analyze(counts, scanResult, cfg)

	fileResult := findResult(report, "a.go")
	assert.Equal(t, SeverityWarn, fileResult.Severity) // 330 <= 330
}

func findResult(report *AnalysisReport, path string) *Result {
	for i := range report.Results {
		if report.Results[i].Path == path {
			return &report.Results[i]
		}
	}
	return nil
}
