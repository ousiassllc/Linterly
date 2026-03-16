package reporter

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/ousiassllc/linterly/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReporter_Checkstyle(t *testing.T) {
	var buf bytes.Buffer
	r := NewReporter(FormatCheckstyle, nil, &buf)
	_, ok := r.(*CheckstyleReporter)
	assert.True(t, ok)
}

func TestCheckstyleReporter_Output(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatCheckstyle, nil, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, nil))

	var output checkstyleOutput
	require.NoError(t, xml.Unmarshal(buf.Bytes(), &output))

	assert.Equal(t, "4.3", output.Version)

	// pass は除外されるため、warn + error の 2 件のみ
	require.Len(t, output.Files, 2)

	// warn testcase
	warnFile := output.Files[0]
	assert.Equal(t, "src/handler.go", warnFile.Name)
	require.Len(t, warnFile.Errors, 1)
	assert.Equal(t, 1, warnFile.Errors[0].Line)
	assert.Equal(t, "warning", warnFile.Errors[0].Severity)
	assert.Equal(t, "325 lines (limit: 300)", warnFile.Errors[0].Message)
	assert.Equal(t, "linterly", warnFile.Errors[0].Source)

	// error testcase
	errorFile := output.Files[1]
	assert.Equal(t, "src/service.go", errorFile.Name)
	require.Len(t, errorFile.Errors, 1)
	assert.Equal(t, 1, errorFile.Errors[0].Line)
	assert.Equal(t, "error", errorFile.Errors[0].Severity)
	assert.Equal(t, "450 lines (limit: 300)", errorFile.Errors[0].Message)
	assert.Equal(t, "linterly", errorFile.Errors[0].Source)
}

func TestCheckstyleReporter_ValidXML(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatCheckstyle, nil, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, nil))

	var output checkstyleOutput
	assert.NoError(t, xml.Unmarshal(buf.Bytes(), &output))

	assert.True(t, strings.HasPrefix(buf.String(), `<?xml version="1.0" encoding="UTF-8"?>`))
}

func TestCheckstyleReporter_AllPass(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatCheckstyle, nil, &buf)

	report := &analyzer.AnalysisReport{
		Results: []analyzer.Result{
			{Path: "src/main.go", Type: "file", Lines: 100, Limit: 300, Threshold: 330, Severity: analyzer.SeverityPass},
			{Path: "src/", Type: "directory", Lines: 100, Limit: 2000, Threshold: 2200, Severity: analyzer.SeverityPass},
		},
		Passed: 2,
	}
	require.NoError(t, reporter.Report(report, nil))

	var output checkstyleOutput
	require.NoError(t, xml.Unmarshal(buf.Bytes(), &output))

	assert.Equal(t, "4.3", output.Version)
	assert.Empty(t, output.Files)
}

func TestCheckstyleReporter_IgnoresWarnings(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatCheckstyle, nil, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, []string{"ignore.both_defined"}))

	assert.NotContains(t, buf.String(), "ignore.both_defined")
}

func TestCheckstyleReporter_DirectoryViolation(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatCheckstyle, nil, &buf)

	report := &analyzer.AnalysisReport{
		Results: []analyzer.Result{
			{Path: "src/", Type: "directory", Lines: 2500, Limit: 2000, Threshold: 2200, Severity: analyzer.SeverityError},
		},
		Errors: 1,
	}
	require.NoError(t, reporter.Report(report, nil))

	var output checkstyleOutput
	require.NoError(t, xml.Unmarshal(buf.Bytes(), &output))

	require.Len(t, output.Files, 1)
	assert.Equal(t, "src/", output.Files[0].Name)
	assert.Equal(t, "error", output.Files[0].Errors[0].Severity)
	assert.Equal(t, "2500 lines (limit: 2000)", output.Files[0].Errors[0].Message)
}
