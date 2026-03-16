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

func TestNewReporter_JUnit(t *testing.T) {
	var buf bytes.Buffer
	r := NewReporter(FormatJUnit, nil, &buf)
	_, ok := r.(*JUnitReporter)
	assert.True(t, ok)
}

func TestJUnitReporter_Output(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatJUnit, nil, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, nil))

	var suites junitTestSuites
	require.NoError(t, xml.Unmarshal(buf.Bytes(), &suites))

	require.Len(t, suites.TestSuites, 2)

	// file testsuite
	fileSuite := suites.TestSuites[0]
	assert.Equal(t, "file", fileSuite.Name)
	assert.Equal(t, 3, fileSuite.Tests)
	assert.Equal(t, 1, fileSuite.Failures)
	require.Len(t, fileSuite.TestCases, 3)

	// warn testcase: system-out あり、failure なし
	warnTC := fileSuite.TestCases[0]
	assert.Equal(t, "src/handler.go", warnTC.Name)
	assert.Equal(t, "file", warnTC.ClassName)
	assert.Nil(t, warnTC.Failure)
	assert.Equal(t, "WARN src/handler.go (325 lines, limit: 300)", warnTC.SystemOut)

	// error testcase: failure あり
	errorTC := fileSuite.TestCases[1]
	assert.Equal(t, "src/service.go", errorTC.Name)
	assert.NotNil(t, errorTC.Failure)
	assert.Equal(t, "450 lines, limit: 300", errorTC.Failure.Message)
	assert.Equal(t, "ERROR src/service.go (450 lines, limit: 300)", errorTC.Failure.Content)

	// pass testcase: failure なし、system-out なし
	passTC := fileSuite.TestCases[2]
	assert.Equal(t, "src/util.go", passTC.Name)
	assert.Nil(t, passTC.Failure)
	assert.Empty(t, passTC.SystemOut)

	// directory testsuite
	dirSuite := suites.TestSuites[1]
	assert.Equal(t, "directory", dirSuite.Name)
	assert.Equal(t, 1, dirSuite.Tests)
	assert.Equal(t, 0, dirSuite.Failures)
}

func TestJUnitReporter_ValidXML(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatJUnit, nil, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, nil))

	var suites junitTestSuites
	assert.NoError(t, xml.Unmarshal(buf.Bytes(), &suites))

	assert.True(t, strings.HasPrefix(buf.String(), `<?xml version="1.0" encoding="UTF-8"?>`))
}

func TestJUnitReporter_IgnoresWarnings(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatJUnit, nil, &buf)

	report := newTestReport()
	require.NoError(t, reporter.Report(report, []string{"ignore.both_defined"}))

	assert.NotContains(t, buf.String(), "ignore.both_defined")
}

func TestJUnitReporter_AllPass(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewReporter(FormatJUnit, nil, &buf)

	report := &analyzer.AnalysisReport{
		Results: []analyzer.Result{
			{Path: "src/main.go", Type: "file", Lines: 100, Limit: 300, Threshold: 330, Severity: analyzer.SeverityPass},
			{Path: "src/", Type: "directory", Lines: 100, Limit: 2000, Threshold: 2200, Severity: analyzer.SeverityPass},
		},
		Passed: 2,
	}
	require.NoError(t, reporter.Report(report, nil))

	var suites junitTestSuites
	require.NoError(t, xml.Unmarshal(buf.Bytes(), &suites))

	for _, suite := range suites.TestSuites {
		assert.Equal(t, 0, suite.Failures, "suite %s should have no failures", suite.Name)
		for _, tc := range suite.TestCases {
			assert.Nil(t, tc.Failure, "testcase %s should have no failure", tc.Name)
			assert.Empty(t, tc.SystemOut, "testcase %s should have no system-out", tc.Name)
		}
	}
}
