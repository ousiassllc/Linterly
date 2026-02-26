package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	CountModeAll      = "all"
	CountModeCodeOnly = "code_only"
)

// ConfigError は設定ファイルに関するエラーを表す。
// Code は i18n メッセージキーに対応する。
type ConfigError struct {
	Code    string
	Message string
	Detail  string // err.config_parse 用の詳細情報
}

func (e *ConfigError) Error() string {
	return e.Message
}

// ValidationErrors は複数のバリデーションエラーをまとめる。
type ValidationErrors struct {
	Errors []*ConfigError
}

func (e *ValidationErrors) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Message
	}
	return strings.Join(msgs, "; ")
}

// DefaultConfigTemplate は linterly init で生成する設定テンプレート。
const DefaultConfigTemplate = `# Linterly 設定ファイル
# https://github.com/ousiassllc/linterly

rules:
  max_lines_per_file: 300
  max_lines_per_directory: 2000
  warning_threshold: 10

count_mode: all

# default_excludes: true
# language: en
`

// Config は設定ファイルの内容を表す。
type Config struct {
	Rules           Rules    `yaml:"rules" mapstructure:"rules"`
	CountMode       string   `yaml:"count_mode" mapstructure:"count_mode"`
	Ignore          []string `yaml:"ignore" mapstructure:"ignore"`
	DefaultExcludes bool     `yaml:"default_excludes" mapstructure:"default_excludes"`
	Language        string   `yaml:"language" mapstructure:"language"`
}

// Rules はチェックルールの設定。
type Rules struct {
	MaxLinesPerFile      int `yaml:"max_lines_per_file" mapstructure:"max_lines_per_file"`
	MaxLinesPerDirectory int `yaml:"max_lines_per_directory" mapstructure:"max_lines_per_directory"`
	WarningThreshold     int `yaml:"warning_threshold" mapstructure:"warning_threshold"`
}

// Overrides は CLI フラグによる設定上書きを表す。
// nil のフィールドは「未指定」を意味し、上書きしない。
type Overrides struct {
	MaxLinesPerFile      *int
	MaxLinesPerDirectory *int
	WarningThreshold     *int
	CountMode            *string
	Ignore               []string // nil=未指定, non-nil=上書き
	NoDefaultExcludes    bool     // true の場合 DefaultExcludes を false にする
}

// ApplyOverrides は Overrides の非 nil フィールドで Config を上書きし、
// 最終的な Config をバリデーションする。
func (c *Config) ApplyOverrides(o *Overrides) error {
	if o == nil {
		return validate(c)
	}
	if o.MaxLinesPerFile != nil {
		c.Rules.MaxLinesPerFile = *o.MaxLinesPerFile
	}
	if o.MaxLinesPerDirectory != nil {
		c.Rules.MaxLinesPerDirectory = *o.MaxLinesPerDirectory
	}
	if o.WarningThreshold != nil {
		c.Rules.WarningThreshold = *o.WarningThreshold
	}
	if o.CountMode != nil {
		c.CountMode = *o.CountMode
	}
	if o.Ignore != nil {
		c.Ignore = o.Ignore
	}
	if o.NoDefaultExcludes {
		c.DefaultExcludes = false
	}
	return validate(c)
}

// defaultConfig は設定ファイルなしで動作する際のデフォルト Config を返す。
func defaultConfig() *Config {
	return &Config{
		Rules: Rules{
			MaxLinesPerFile:      300,
			MaxLinesPerDirectory: 2000,
			WarningThreshold:     10,
		},
		CountMode:       CountModeAll,
		Ignore:          []string{},
		DefaultExcludes: true,
		Language:        "en",
	}
}

// Load は設定ファイルを読み込み、バリデーション済みの Config を返す。
// configPath が空でない場合はそのパスのみを読み込む。
// 空の場合は探索順序に従って設定ファイルを探す。
func Load(configPath string) (*Config, error) {
	v := viper.New()

	explicit, err := findAndReadConfig(v, configPath)
	if err != nil {
		// 明示指定のパスが見つからない場合はエラー
		if explicit {
			return nil, err
		}
		// 自動探索で見つからない場合はデフォルト Config を返す
		var cfgErr *ConfigError
		if errors.As(err, &cfgErr) && cfgErr.Code == "err.config_not_found" {
			return defaultConfig(), nil
		}
		return nil, err
	}

	// --- 設定ファイルが見つかった場合の既存ロジック ---
	// rules セクションの存在チェック
	if !v.IsSet("rules") {
		return nil, &ConfigError{
			Code:    "validation.rules_required",
			Message: `"rules" section is required`,
		}
	}

	// デフォルト値の設定
	v.SetDefault("rules.max_lines_per_file", 300)
	v.SetDefault("rules.max_lines_per_directory", 2000)
	v.SetDefault("rules.warning_threshold", 10)
	v.SetDefault("count_mode", CountModeAll)
	v.SetDefault("ignore", []string{})
	v.SetDefault("default_excludes", true)
	v.SetDefault("language", "en")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, &ConfigError{
			Code:    "err.config_parse",
			Message: fmt.Sprintf("failed to parse config file: %s", err),
			Detail:  err.Error(),
		}
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// findAndReadConfig は探索順序に従って設定ファイルを見つけて読み込む。
// explicit は、ユーザーが明示的にパスを指定したかどうかを示す。
func findAndReadConfig(v *viper.Viper, configPath string) (explicit bool, err error) {
	if configPath != "" {
		v.SetConfigFile(configPath)
		return true, v.ReadInConfig()
	}

	// LINTERLY_CONFIG 環境変数
	if envPath := os.Getenv("LINTERLY_CONFIG"); envPath != "" {
		v.SetConfigFile(envPath)
		return true, v.ReadInConfig()
	}

	// カレントディレクトリの .linterly.yml / .linterly.yaml
	for _, name := range []string{".linterly.yml", ".linterly.yaml"} {
		if _, err := os.Stat(name); err == nil {
			v.SetConfigFile(name)
			return false, v.ReadInConfig()
		}
	}

	return false, &ConfigError{
		Code:    "err.config_not_found",
		Message: "config file not found",
	}
}

// validate は Config の各フィールドをバリデーションする。
func validate(cfg *Config) error {
	var errs []*ConfigError

	if cfg.Rules.MaxLinesPerFile <= 0 {
		errs = append(errs, &ConfigError{
			Code:    "validation.max_lines_per_file",
			Message: `"max_lines_per_file" must be a positive integer`,
		})
	}
	if cfg.Rules.MaxLinesPerDirectory <= 0 {
		errs = append(errs, &ConfigError{
			Code:    "validation.max_lines_per_directory",
			Message: `"max_lines_per_directory" must be a positive integer`,
		})
	}
	if cfg.Rules.WarningThreshold < 0 || cfg.Rules.WarningThreshold > 100 {
		errs = append(errs, &ConfigError{
			Code:    "validation.warning_threshold",
			Message: `"warning_threshold" must be between 0 and 100`,
		})
	}
	if cfg.CountMode != CountModeAll && cfg.CountMode != CountModeCodeOnly {
		errs = append(errs, &ConfigError{
			Code:    "validation.count_mode",
			Message: `"count_mode" must be "all" or "code_only"`,
		})
	}
	if cfg.Language != "en" && cfg.Language != "ja" {
		errs = append(errs, &ConfigError{
			Code:    "validation.language",
			Message: `"language" must be "en" or "ja"`,
		})
	}

	if len(errs) > 0 {
		return &ValidationErrors{Errors: errs}
	}
	return nil
}
