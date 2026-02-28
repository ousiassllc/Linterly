package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyOverrides_AllFields(t *testing.T) {
	cfg := defaultConfig()
	cfg.Ignore = []string{"old/**"}

	maxFile := 500
	maxDir := 3000
	threshold := 20
	mode := CountModeCodeOnly
	o := &Overrides{
		MaxLinesPerFile:      &maxFile,
		MaxLinesPerDirectory: &maxDir,
		WarningThreshold:     &threshold,
		CountMode:            &mode,
		Ignore:               []string{"vendor/**", "*.pb.go"},
		NoDefaultExcludes:    true,
	}

	err := cfg.ApplyOverrides(o)
	require.NoError(t, err)

	assert.Equal(t, 500, cfg.Rules.MaxLinesPerFile)
	assert.Equal(t, 3000, cfg.Rules.MaxLinesPerDirectory)
	assert.Equal(t, 20, cfg.Rules.WarningThreshold)
	assert.Equal(t, "code_only", cfg.CountMode)
	assert.Equal(t, []string{"vendor/**", "*.pb.go"}, cfg.Ignore)
	assert.Equal(t, false, cfg.DefaultExcludes)
}

func TestApplyOverrides_PartialFields(t *testing.T) {
	cfg := defaultConfig()
	cfg.Ignore = []string{"old/**"}

	maxFile := 500
	o := &Overrides{
		MaxLinesPerFile: &maxFile,
	}

	err := cfg.ApplyOverrides(o)
	require.NoError(t, err)

	assert.Equal(t, 500, cfg.Rules.MaxLinesPerFile)
	assert.Equal(t, 2000, cfg.Rules.MaxLinesPerDirectory)
	assert.Equal(t, 10, cfg.Rules.WarningThreshold)
	assert.Equal(t, "all", cfg.CountMode)
	assert.Equal(t, []string{"old/**"}, cfg.Ignore)
	assert.Equal(t, true, cfg.DefaultExcludes)
}

func TestApplyOverrides_NilOverrides(t *testing.T) {
	cfg := defaultConfig()
	err := cfg.ApplyOverrides(nil)
	require.NoError(t, err)
	assert.Equal(t, 300, cfg.Rules.MaxLinesPerFile)
}

func TestApplyOverrides_ValidationError(t *testing.T) {
	cfg := defaultConfig()
	badValue := -1
	o := &Overrides{
		MaxLinesPerFile: &badValue,
	}

	err := cfg.ApplyOverrides(o)
	require.Error(t, err)

	var valErrs *ValidationErrors
	require.True(t, errors.As(err, &valErrs))
	assert.Contains(t, codeList(valErrs), "validation.max_lines_per_file")
}

func TestApplyOverrides_NoDefaultExcludes(t *testing.T) {
	cfg := defaultConfig()
	o := &Overrides{NoDefaultExcludes: true}

	err := cfg.ApplyOverrides(o)
	require.NoError(t, err)
	assert.Equal(t, false, cfg.DefaultExcludes)
}

func TestApplyOverrides_NoDefaultExcludes_False(t *testing.T) {
	cfg := defaultConfig()
	o := &Overrides{NoDefaultExcludes: false}

	err := cfg.ApplyOverrides(o)
	require.NoError(t, err)
	assert.Equal(t, true, cfg.DefaultExcludes)
}

func TestApplyOverrides_MultipleValidationErrors(t *testing.T) {
	cfg := defaultConfig()
	badFile := -1
	badMode := "invalid"
	o := &Overrides{
		MaxLinesPerFile: &badFile,
		CountMode:       &badMode,
	}

	err := cfg.ApplyOverrides(o)
	require.Error(t, err)

	var valErrs *ValidationErrors
	require.True(t, errors.As(err, &valErrs))
	assert.Len(t, valErrs.Errors, 2)
	codes := codeList(valErrs)
	assert.Contains(t, codes, "validation.max_lines_per_file")
	assert.Contains(t, codes, "validation.count_mode")
}

func TestApplyOverrides_IgnoreOverride(t *testing.T) {
	cfg := defaultConfig()
	cfg.Ignore = []string{"original/**"}

	o := &Overrides{Ignore: []string{}}

	err := cfg.ApplyOverrides(o)
	require.NoError(t, err)
	assert.Empty(t, cfg.Ignore)
	assert.NotNil(t, cfg.Ignore)
}
