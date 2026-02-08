package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInit_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	// langFlag をデフォルト（空）にして英語で動作させる
	oldLang := langFlag
	langFlag = ""
	defer func() { langFlag = oldLang }()

	err = runInit(initCmd, nil)
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	content, err := os.ReadFile(filepath.Join(tmpDir, ".linterly.yml"))
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestRunInit_CreatesFile_Japanese(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	// LINTERLY_LANG=ja で日本語メッセージが使われることを確認
	t.Setenv("LINTERLY_LANG", "ja")
	oldLang := langFlag
	langFlag = ""
	defer func() { langFlag = oldLang }()

	err = runInit(initCmd, nil)
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	content, err := os.ReadFile(filepath.Join(tmpDir, ".linterly.yml"))
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestRunInit_LangFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	// langFlag="ja" で日本語メッセージが使われることを確認
	oldLang := langFlag
	langFlag = "ja"
	defer func() { langFlag = oldLang }()

	err = runInit(initCmd, nil)
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	content, err := os.ReadFile(filepath.Join(tmpDir, ".linterly.yml"))
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestRunInit_LangFlagOverridesEnv(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	// LINTERLY_LANG=ja だが langFlag="en" の場合、フラグが優先される
	t.Setenv("LINTERLY_LANG", "ja")
	oldLang := langFlag
	langFlag = "en"
	defer func() { langFlag = oldLang }()

	err = runInit(initCmd, nil)
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	content, err := os.ReadFile(filepath.Join(tmpDir, ".linterly.yml"))
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestRunInit_UnsupportedLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	oldLang := langFlag
	langFlag = "fr"
	defer func() { langFlag = oldLang }()

	err = runInit(initCmd, nil)
	require.Error(t, err)

	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, ExitRuntimeError, exitErr.Code)
}
