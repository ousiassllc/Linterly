package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const formatTestConfig = `rules:
  max_lines_per_file: 10
  max_lines_per_directory: 100000
  warning_threshold: 50
default_excludes: false
`

// runCheckWithFormat は content を持つ 1 ファイルを対象に指定フォーマットで check を実行する。
func runCheckWithFormat(t *testing.T, fmtName string, lines int) (string, error) {
	t.Helper()
	oldConfig, oldFormat, oldLang := configFile, format, langFlag
	t.Cleanup(func() {
		configFile, format, langFlag = oldConfig, oldFormat, oldLang
	})

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".linterly.yml")
	helperWriteFile(t, cfgPath, formatTestConfig)
	targetDir := filepath.Join(tmpDir, "src")
	helperWriteFile(t, filepath.Join(targetDir, "a.go"), strings.Repeat("line\n", lines))

	configFile = cfgPath
	format = fmtName
	langFlag = "en"

	var err error
	output := helperCaptureStdout(t, func() {
		err = runCheck(checkCmd, []string{targetDir})
	})
	return output, err
}

func TestRunCheck_UnsupportedFormat(t *testing.T) {
	for _, f := range []string{"foo", "JSON", "sarif", ""} {
		t.Run("format="+f, func(t *testing.T) {
			output, err := runCheckWithFormat(t, f, 1)

			var exitErr *ExitError
			require.True(t, errors.As(err, &exitErr))
			assert.Equal(t, ExitRuntimeError, exitErr.Code)
			assert.Contains(t, exitErr.Message, "Unsupported format: "+f)
			assert.Contains(t, exitErr.Message, "text, json, junit, checkstyle")
			assert.Empty(t, output)
		})
	}
}
