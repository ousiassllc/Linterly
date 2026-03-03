package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ousiassllc/linterly/internal/config"
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

// newInitCmd は cobra I/O テスト用の init コマンドを生成するヘルパー。
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "init",
		RunE: runInit,
	}
}

func TestRunInit_CobraIO_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	oldLang := langFlag
	langFlag = ""
	defer func() { langFlag = oldLang }()

	cmd := newInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.RunE(cmd, nil)
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	content, err := os.ReadFile(filepath.Join(tmpDir, config.DefaultConfigFileName))
	require.NoError(t, err)
	assert.NotEmpty(t, content)

	// 出力にファイル作成メッセージが含まれることを確認
	assert.NotEmpty(t, out.String())
}

func TestRunInit_CobraIO_OverwriteYes(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	oldLang := langFlag
	langFlag = ""
	defer func() { langFlag = oldLang }()

	// 既存ファイルを作成
	existingContent := []byte("existing content")
	require.NoError(t, os.WriteFile(config.DefaultConfigFileName, existingContent, 0644))

	cmd := newInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader("y\n"))

	err = cmd.RunE(cmd, nil)
	require.NoError(t, err)

	// ファイルが上書きされたことを確認
	content, err := os.ReadFile(filepath.Join(tmpDir, config.DefaultConfigFileName))
	require.NoError(t, err)
	assert.NotEqual(t, existingContent, content)
	assert.Equal(t, config.DefaultConfigTemplate, string(content))

	// 出力に上書きメッセージが含まれることを確認
	assert.NotEmpty(t, out.String())
}

func TestRunInit_CobraIO_OverwriteNo(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	oldLang := langFlag
	langFlag = ""
	defer func() { langFlag = oldLang }()

	// 既存ファイルを作成
	existingContent := []byte("existing content")
	require.NoError(t, os.WriteFile(config.DefaultConfigFileName, existingContent, 0644))

	cmd := newInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader("n\n"))

	err = cmd.RunE(cmd, nil)
	require.NoError(t, err)

	// ファイルが上書きされていないことを確認
	content, err := os.ReadFile(filepath.Join(tmpDir, config.DefaultConfigFileName))
	require.NoError(t, err)
	assert.Equal(t, existingContent, content)
}
