package reporter

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/ousiassllc/linterly/internal/analyzer"
)

// JUnitReporter は JUnit XML 形式で結果を出力する。
type JUnitReporter struct {
	writer io.Writer
}

type junitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	TestSuites []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	SystemOut string        `xml:"system-out,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Content string `xml:",chardata"`
}

// Report は分析結果を JUnit XML 形式で出力する。
// warnings は JUnit XML に出力する機構がないため無視する。
func (r *JUnitReporter) Report(report *analyzer.AnalysisReport, _ []string) error {
	suiteMap := map[string]*junitTestSuite{
		"file":      {Name: "file"},
		"directory": {Name: "directory"},
	}

	for _, result := range report.Results {
		suite := suiteMap[result.Type]
		if suite == nil {
			continue
		}

		tc := junitTestCase{
			Name:      result.Path,
			ClassName: result.Type,
		}

		msg := fmt.Sprintf("%d lines, limit: %d", result.Lines, result.Limit)

		switch result.Severity {
		case analyzer.SeverityError:
			tc.Failure = &junitFailure{
				Message: msg,
				Content: fmt.Sprintf("ERROR %s (%s)", result.Path, msg),
			}
			suite.Failures++
		case analyzer.SeverityWarn:
			tc.SystemOut = fmt.Sprintf("WARN %s (%s)", result.Path, msg)
		}

		suite.TestCases = append(suite.TestCases, tc)
		suite.Tests++
	}

	output := junitTestSuites{
		TestSuites: []junitTestSuite{*suiteMap["file"], *suiteMap["directory"]},
	}

	data, err := xml.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.writer, "%s%s\n", xml.Header, data)
	return err
}
