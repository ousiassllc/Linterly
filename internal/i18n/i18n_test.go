package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_English(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)
	assert.Equal(t, "en", tr.Lang())
}

func TestNew_Japanese(t *testing.T) {
	tr, err := New("ja")
	require.NoError(t, err)
	assert.Equal(t, "ja", tr.Lang())
}

func TestNew_UnsupportedLanguage(t *testing.T) {
	tr, err := New("fr")
	assert.Nil(t, tr)
	assert.EqualError(t, err, "unsupported language: fr")
}

func TestNew_EmptyLanguage(t *testing.T) {
	tr, err := New("")
	assert.Nil(t, tr)
	assert.EqualError(t, err, "unsupported language: ")
}

func TestT_SimpleMessage(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	assert.Equal(t, "Created .linterly.yml", tr.T("init.created"))
}

func TestT_SimpleMessage_Japanese(t *testing.T) {
	tr, err := New("ja")
	require.NoError(t, err)

	assert.Equal(t, ".linterly.yml を作成しました", tr.T("init.created"))
}

func TestT_WithPlaceholders(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	result := tr.T("check.warn", "main.go", 500, 300)
	assert.Equal(t, "WARN  main.go (500 lines, limit: 300)", result)
}

func TestT_WithPlaceholders_Japanese(t *testing.T) {
	tr, err := New("ja")
	require.NoError(t, err)

	result := tr.T("check.warn", "main.go", 500, 300)
	assert.Equal(t, "WARN  main.go (500 行, 上限: 300)", result)
}

func TestT_Summary(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	result := tr.T("check.summary", 2, 3, 10)
	assert.Equal(t, "Results: 2 error(s), 3 warning(s), 10 passed", result)
}

func TestT_Summary_Japanese(t *testing.T) {
	tr, err := New("ja")
	require.NoError(t, err)

	result := tr.T("check.summary", 2, 3, 10)
	assert.Equal(t, "結果: 2 エラー, 3 警告, 10 パス", result)
}

func TestT_UnknownKey(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	assert.Equal(t, "unknown.key", tr.T("unknown.key"))
}

func TestT_UnknownKeyWithArgs(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	// 未知のキーの場合、引数があってもキーをそのまま返す
	assert.Equal(t, "unknown.key", tr.T("unknown.key", "arg1"))
}

func TestT_ErrorMessage(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	result := tr.T("check.error", "controller.go", 800, 300)
	assert.Equal(t, "ERROR controller.go (800 lines, limit: 300)", result)
}

func TestT_ErrorMessage_Japanese(t *testing.T) {
	tr, err := New("ja")
	require.NoError(t, err)

	result := tr.T("check.error", "controller.go", 800, 300)
	assert.Equal(t, "ERROR controller.go (800 行, 上限: 300)", result)
}

func TestT_MultilineMessage(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	result := tr.T("ignore.both_defined")
	expected := "Both .linterlyignore and ignore in config file are defined. .linterlyignore takes precedence. ignore in config file is ignored."
	assert.Equal(t, expected, result)
}

func TestT_MultilineMessage_Japanese(t *testing.T) {
	tr, err := New("ja")
	require.NoError(t, err)

	result := tr.T("ignore.both_defined")
	expected := ".linterlyignore と設定ファイルの ignore が両方定義されています。 .linterlyignore が優先されます。設定ファイルの ignore は無視されます。"
	assert.Equal(t, expected, result)
}

func TestT_VersionInfo(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	result := tr.T("version.info", "1.0.0", "go1.25.6", "linux", "amd64")
	assert.Equal(t, "linterly 1.0.0 (go1.25.6, linux/amd64)", result)
}

func TestT_ValidationMessages(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	tests := []struct {
		key      string
		expected string
	}{
		{"validation.rules_required", `"rules" section is required`},
		{"validation.max_lines_per_file", `"max_lines_per_file" must be a positive integer`},
		{"validation.max_lines_per_directory", `"max_lines_per_directory" must be a positive integer`},
		{"validation.warning_threshold", `"warning_threshold" must be between 0 and 100`},
		{"validation.count_mode", `"count_mode" must be "all" or "code_only"`},
		{"validation.language", `"language" must be "en" or "ja"`},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.expected, tr.T(tt.key))
		})
	}
}

func TestT_ValidationMessages_Japanese(t *testing.T) {
	tr, err := New("ja")
	require.NoError(t, err)

	tests := []struct {
		key      string
		expected string
	}{
		{"validation.rules_required", `"rules" セクションが必要です`},
		{"validation.max_lines_per_file", `"max_lines_per_file" は正の整数である必要があります`},
		{"validation.max_lines_per_directory", `"max_lines_per_directory" は正の整数である必要があります`},
		{"validation.warning_threshold", `"warning_threshold" は 0 から 100 の範囲である必要があります`},
		{"validation.count_mode", `"count_mode" は "all" または "code_only" である必要があります`},
		{"validation.language", `"language" は "en" または "ja" である必要があります`},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.expected, tr.T(tt.key))
		})
	}
}

func TestT_ErrorMessages(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)

	assert.Equal(t, "Config file not found. Run 'linterly init' to create one.", tr.T("err.config_not_found"))
	assert.Equal(t, "Failed to parse config file: invalid yaml", tr.T("err.config_parse", "invalid yaml"))
}

func TestT_ErrorMessages_Japanese(t *testing.T) {
	tr, err := New("ja")
	require.NoError(t, err)

	assert.Equal(t, "設定ファイルが見つかりません。'linterly init' を実行して作成してください。", tr.T("err.config_not_found"))
	assert.Equal(t, "設定ファイルの解析に失敗しました: invalid yaml", tr.T("err.config_parse", "invalid yaml"))
}

func TestT_NoViolations(t *testing.T) {
	tr, err := New("en")
	require.NoError(t, err)
	assert.Equal(t, "No violations found. All checks passed.", tr.T("check.no_violations"))

	trJa, err := New("ja")
	require.NoError(t, err)
	assert.Equal(t, "違反なし。すべてのチェックに合格しました。", trJa.T("check.no_violations"))
}

func TestResolveLanguage_FlagTakesPrecedence(t *testing.T) {
	t.Setenv("LINTERLY_LANG", "ja")
	assert.Equal(t, "en", ResolveLanguage("en"))
}

func TestResolveLanguage_EnvVar(t *testing.T) {
	t.Setenv("LINTERLY_LANG", "ja")
	assert.Equal(t, "ja", ResolveLanguage(""))
}

func TestResolveLanguage_Default(t *testing.T) {
	t.Setenv("LINTERLY_LANG", "")
	assert.Equal(t, "en", ResolveLanguage(""))
}

func TestAllKeysExistInBothLanguages(t *testing.T) {
	en, err := New("en")
	require.NoError(t, err)

	ja, err := New("ja")
	require.NoError(t, err)

	// 英語のキーがすべて日本語にも存在することを確認
	for key := range en.messages {
		_, ok := ja.messages[key]
		assert.True(t, ok, "key %q exists in en but not in ja", key)
	}

	// 日本語のキーがすべて英語にも存在することを確認
	for key := range ja.messages {
		_, ok := en.messages[key]
		assert.True(t, ok, "key %q exists in ja but not in en", key)
	}
}
