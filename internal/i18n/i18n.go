package i18n

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed messages/*.yml
var messagesFS embed.FS

// supportedLanguages は対応する言語の一覧。
var supportedLanguages = map[string]string{
	"en": "messages/en.yml",
	"ja": "messages/ja.yml",
}

// Translator はメッセージ翻訳を行う。
type Translator struct {
	lang     string
	messages map[string]string
}

// New は指定言語の Translator を生成する。
// 不正な言語が指定された場合はエラーを返す。
func New(lang string) (*Translator, error) {
	path, ok := supportedLanguages[lang]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	data, err := messagesFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read message file: %w", err)
	}

	messages := make(map[string]string)
	if err := yaml.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse message file: %w", err)
	}

	return &Translator{
		lang:     lang,
		messages: messages,
	}, nil
}

// T はメッセージキーに対応する翻訳テキストを返す。
// args はプレースホルダーの置換に使用する。
// 未知のキーが指定された場合はキーをそのまま返す。
func (t *Translator) T(key string, args ...any) string {
	msg, ok := t.messages[key]
	if !ok {
		return key
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

// Lang は Translator の言語を返す。
func (t *Translator) Lang() string {
	return t.lang
}
