package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ousiassllc/linterly/internal/analyzer"
	"github.com/ousiassllc/linterly/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestReport() *analyzer.AnalysisReport {
	return &analyzer.AnalysisReport{
		Results: []analyzer.Result{
			{Path: "src/handler.go", Type: "file", Lines: 325, Limit: 300, Threshold: 330, Severity: analyzer.SeverityWarn},
			{Path: "src/service.go", Type: "file", Lines: 450, Limit: 300, Threshold: 330, Severity: analyzer.SeverityError},
			{Path: "src/util.go", Type: "file", Lines: 100, Limit: 300, Threshold: 330, Severity: analyzer.SeverityPass},
			{Path: "src/", Type: "directory", Lines: 875, Limit: 2000, Threshold: 2200, Severity: analyzer.SeverityPass},
		},
		Errors:   1,
		Warnings: 1,
		Passed:   2,
	}
}

func TestTextReporter_Output(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	tr, err := i18n.New("en")
	require.NoError(t, err)

	var buf bytes.Buffer
	reporter := NewReporter("text", tr, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, nil))

	output := buf.String()
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "handler.go")
	assert.Contains(t, output, "ERROR")
	assert.Contains(t, output, "service.go")
	assert.NotContains(t, output, "util.go") // pass は表示しない
	assert.Contains(t, output, "1 error(s)")
	assert.Contains(t, output, "1 warning(s)")
}

func TestTextReporter_Japanese(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	tr, err := i18n.New("ja")
	require.NoError(t, err)

	var buf bytes.Buffer
	reporter := NewReporter("text", tr, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, nil))

	output := buf.String()
	assert.Contains(t, output, "行")
	assert.Contains(t, output, "上限")
}

func TestTextReporter_WithWarnings(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	tr, err := i18n.New("en")
	require.NoError(t, err)

	var buf bytes.Buffer
	reporter := NewReporter("text", tr, &buf)

	report := newTestReport()
	warnings := []string{"Both .linterlyignore and ignore in config file are defined."}
	require.NoError(t, reporter.Report(report, warnings))

	output := buf.String()
	assert.Contains(t, output, "Both .linterlyignore")
}

func TestTextReporter_NoViolations(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	tr, err := i18n.New("en")
	require.NoError(t, err)

	var buf bytes.Buffer
	reporter := NewReporter("text", tr, &buf)

	report := &analyzer.AnalysisReport{
		Results: []analyzer.Result{
			{Path: "src/main.go", Type: "file", Lines: 100, Limit: 300, Threshold: 330, Severity: analyzer.SeverityPass},
		},
		Passed: 1,
	}
	require.NoError(t, reporter.Report(report, nil))

	output := buf.String()
	assert.Contains(t, output, "0 error(s)")
}

func TestJSONReporter_Output(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter("json", nil, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, nil))

	var output jsonOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &output))

	assert.Len(t, output.Results, 4)      // JSON には pass を含む全結果
	assert.Equal(t, 1, output.Summary.Errors)
	assert.Equal(t, 1, output.Summary.Warnings)
	assert.Equal(t, 2, output.Summary.Passed)
	assert.Equal(t, 4, output.Summary.Total)

	// 結果の詳細を確認
	warnResult := output.Results[0]
	assert.Equal(t, "src/handler.go", warnResult.Path)
	assert.Equal(t, "warn", warnResult.Severity)
	assert.Equal(t, 325, warnResult.Lines)
}

func TestJSONReporter_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter("json", nil, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, nil))

	// 有効な JSON であること
	assert.True(t, json.Valid(buf.Bytes()))
}

func TestNewReporter_Text(t *testing.T) {
	tr, _ := i18n.New("en")
	var buf bytes.Buffer
	r := NewReporter("text", tr, &buf)
	_, ok := r.(*TextReporter)
	assert.True(t, ok)
}

func TestNewReporter_JSON(t *testing.T) {
	var buf bytes.Buffer
	r := NewReporter("json", nil, &buf)
	_, ok := r.(*JSONReporter)
	assert.True(t, ok)
}

func TestTextReporter_Color(t *testing.T) {
	// NO_COLOR が設定されていない場合、カラー出力される
	t.Setenv("NO_COLOR", "")

	tr, err := i18n.New("en")
	require.NoError(t, err)

	var buf bytes.Buffer
	reporter := &TextReporter{writer: &buf, translator: tr}

	report := &analyzer.AnalysisReport{
		Results: []analyzer.Result{
			{Path: "a.go", Type: "file", Lines: 320, Limit: 300, Threshold: 330, Severity: analyzer.SeverityWarn},
		},
		Warnings: 1,
	}
	require.NoError(t, reporter.Report(report, nil))

	output := buf.String()
	assert.True(t, strings.Contains(output, "\033[33m")) // yellow
}
