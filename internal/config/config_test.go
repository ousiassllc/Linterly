package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_FullConfig(t *testing.T) {
	cfg, err := Load("testdata/valid_full.yml")
	require.NoError(t, err)

	assert.Equal(t, 500, cfg.Rules.MaxLinesPerFile)
	assert.Equal(t, 3000, cfg.Rules.MaxLinesPerDirectory)
	assert.Equal(t, 20, cfg.Rules.WarningThreshold)
	assert.Equal(t, "code_only", cfg.CountMode)
	assert.Equal(t, []string{"vendor/**", "*.pb.go"}, cfg.Ignore)
	assert.Equal(t, false, cfg.DefaultExcludes)
	assert.Equal(t, "ja", cfg.Language)
}

func TestLoad_MinimalConfig(t *testing.T) {
	cfg, err := Load("testdata/valid_minimal.yml")
	require.NoError(t, err)

	assert.Equal(t, 300, cfg.Rules.MaxLinesPerFile)
	// デフォルト値が適用される
	assert.Equal(t, 2000, cfg.Rules.MaxLinesPerDirectory)
	assert.Equal(t, 10, cfg.Rules.WarningThreshold)
	assert.Equal(t, "all", cfg.CountMode)
	assert.Empty(t, cfg.Ignore)
	assert.Equal(t, true, cfg.DefaultExcludes)
	assert.Equal(t, "en", cfg.Language)
}

func TestLoad_RulesOnlyConfig(t *testing.T) {
	cfg, err := Load("testdata/valid_rules_only.yml")
	require.NoError(t, err)

	assert.Equal(t, 200, cfg.Rules.MaxLinesPerFile)
	assert.Equal(t, 1500, cfg.Rules.MaxLinesPerDirectory)
	assert.Equal(t, 5, cfg.Rules.WarningThreshold)
	assert.Equal(t, "all", cfg.CountMode)
	assert.Equal(t, true, cfg.DefaultExcludes)
	assert.Equal(t, "en", cfg.Language)
}

func TestLoad_MissingRulesSection(t *testing.T) {
	_, err := Load("testdata/missing_rules.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"rules" section is required`)

	var cfgErr *ConfigError
	require.True(t, errors.As(err, &cfgErr))
	assert.Equal(t, "validation.rules_required", cfgErr.Code)
}

func TestLoad_InvalidMaxLinesPerFile(t *testing.T) {
	_, err := Load("testdata/invalid_max_lines_per_file.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"max_lines_per_file" must be a positive integer`)

	var valErrs *ValidationErrors
	require.True(t, errors.As(err, &valErrs))
	assert.Contains(t, codeList(valErrs), "validation.max_lines_per_file")
}

func TestLoad_InvalidMaxLinesPerDirectory(t *testing.T) {
	_, err := Load("testdata/invalid_max_lines_per_directory.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"max_lines_per_directory" must be a positive integer`)
}

func TestLoad_InvalidWarningThresholdOver100(t *testing.T) {
	_, err := Load("testdata/invalid_warning_threshold.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"warning_threshold" must be between 0 and 100`)
}

func TestLoad_InvalidWarningThresholdNegative(t *testing.T) {
	_, err := Load("testdata/invalid_warning_threshold_negative.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"warning_threshold" must be between 0 and 100`)
}

func TestLoad_InvalidCountMode(t *testing.T) {
	_, err := Load("testdata/invalid_count_mode.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"count_mode" must be "all" or "code_only"`)
}

func TestLoad_InvalidLanguage(t *testing.T) {
	_, err := Load("testdata/invalid_language.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"language" must be "en" or "ja"`)
}

func TestLoad_MultipleValidationErrors(t *testing.T) {
	_, err := Load("testdata/invalid_multiple.yml")
	require.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, `"max_lines_per_file" must be a positive integer`)
	assert.Contains(t, errMsg, `"max_lines_per_directory" must be a positive integer`)
	assert.Contains(t, errMsg, `"warning_threshold" must be between 0 and 100`)
	assert.Contains(t, errMsg, `"count_mode" must be "all" or "code_only"`)
	assert.Contains(t, errMsg, `"language" must be "en" or "ja"`)

	var valErrs *ValidationErrors
	require.True(t, errors.As(err, &valErrs))
	codes := codeList(valErrs)
	assert.Contains(t, codes, "validation.max_lines_per_file")
	assert.Contains(t, codes, "validation.max_lines_per_directory")
	assert.Contains(t, codes, "validation.warning_threshold")
	assert.Contains(t, codes, "validation.count_mode")
	assert.Contains(t, codes, "validation.language")
	assert.Len(t, valErrs.Errors, 5)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("testdata/nonexistent.yml")
	require.Error(t, err)
}

func TestLoad_ConfigPathEmpty_FallbackToCurrentDir(t *testing.T) {
	// 一時ディレクトリに .linterly.yml を作成してそこに chdir する
	tmpDir := t.TempDir()
	content := []byte("rules:\n  max_lines_per_file: 250\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterly.yml"), content, 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 250, cfg.Rules.MaxLinesPerFile)
}

func TestLoad_ConfigPathEmpty_FallbackToYamlExtension(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("rules:\n  max_lines_per_file: 350\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterly.yaml"), content, 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 350, cfg.Rules.MaxLinesPerFile)
}

func TestLoad_ConfigPathEmpty_YmlTakesPrecedenceOverYaml(t *testing.T) {
	tmpDir := t.TempDir()
	ymlContent := []byte("rules:\n  max_lines_per_file: 100\n")
	yamlContent := []byte("rules:\n  max_lines_per_file: 200\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterly.yml"), ymlContent, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterly.yaml"), yamlContent, 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 100, cfg.Rules.MaxLinesPerFile)
}

func TestLoad_EnvVariable(t *testing.T) {
	t.Setenv("LINTERLY_CONFIG", "testdata/valid_full.yml")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 500, cfg.Rules.MaxLinesPerFile)
	assert.Equal(t, "ja", cfg.Language)
}

func TestLoad_NoConfigFile_ReturnsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, 300, cfg.Rules.MaxLinesPerFile)
	assert.Equal(t, 2000, cfg.Rules.MaxLinesPerDirectory)
	assert.Equal(t, 10, cfg.Rules.WarningThreshold)
	assert.Equal(t, "all", cfg.CountMode)
	assert.Empty(t, cfg.Ignore)
	assert.Equal(t, true, cfg.DefaultExcludes)
	assert.Equal(t, "en", cfg.Language)
}

func TestLoad_ExplicitConfigPath_NotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yml")
	require.Error(t, err)
}

func TestLoad_EnvVariable_NotFound(t *testing.T) {
	t.Setenv("LINTERLY_CONFIG", "/nonexistent/path.yml")
	_, err := Load("")
	require.Error(t, err)
}

func TestLoad_WarningThresholdZeroIsValid(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("rules:\n  max_lines_per_file: 300\n  max_lines_per_directory: 2000\n  warning_threshold: 0\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterly.yml"), content, 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.Rules.WarningThreshold)
}

func TestLoad_WarningThreshold100IsValid(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("rules:\n  max_lines_per_file: 300\n  max_lines_per_directory: 2000\n  warning_threshold: 100\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".linterly.yml"), content, 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 100, cfg.Rules.WarningThreshold)
}

func TestDefaultConfigTemplate(t *testing.T) {
	assert.Contains(t, DefaultConfigTemplate, "max_lines_per_file: 300")
	assert.Contains(t, DefaultConfigTemplate, "max_lines_per_directory: 2000")
	assert.Contains(t, DefaultConfigTemplate, "warning_threshold: 10")
	assert.Contains(t, DefaultConfigTemplate, "count_mode: all")
	assert.Contains(t, DefaultConfigTemplate, "# default_excludes: true")
	assert.Contains(t, DefaultConfigTemplate, "# language: en")
}

// codeList は ValidationErrors から Code の一覧を返すヘルパー。
func codeList(ve *ValidationErrors) []string {
	codes := make([]string, len(ve.Errors))
	for i, e := range ve.Errors {
		codes[i] = e.Code
	}
	return codes
}
