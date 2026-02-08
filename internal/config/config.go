package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

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

// Load は設定ファイルを読み込み、バリデーション済みの Config を返す。
// configPath が空でない場合はそのパスのみを読み込む。
// 空の場合は探索順序に従って設定ファイルを探す。
func Load(configPath string) (*Config, error) {
	v := viper.New()

	if err := findAndReadConfig(v, configPath); err != nil {
		return nil, err
	}

	// rules セクションの存在チェック（デフォルト値設定前に行う）
	if !v.IsSet("rules") {
		return nil, errors.New(`"rules" section is required`)
	}

	// デフォルト値の設定（rules 存在チェック後に行う）
	v.SetDefault("rules.max_lines_per_file", 300)
	v.SetDefault("rules.max_lines_per_directory", 2000)
	v.SetDefault("rules.warning_threshold", 10)
	v.SetDefault("count_mode", "all")
	v.SetDefault("ignore", []string{})
	v.SetDefault("default_excludes", true)
	v.SetDefault("language", "en")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// findAndReadConfig は探索順序に従って設定ファイルを見つけて読み込む。
func findAndReadConfig(v *viper.Viper, configPath string) error {
	if configPath != "" {
		v.SetConfigFile(configPath)
		return v.ReadInConfig()
	}

	// LINTERLY_CONFIG 環境変数
	if envPath := os.Getenv("LINTERLY_CONFIG"); envPath != "" {
		v.SetConfigFile(envPath)
		return v.ReadInConfig()
	}

	// カレントディレクトリの .linterly.yml / .linterly.yaml
	for _, name := range []string{".linterly.yml", ".linterly.yaml"} {
		if _, err := os.Stat(name); err == nil {
			v.SetConfigFile(name)
			return v.ReadInConfig()
		}
	}

	return errors.New("config file not found")
}

// validate は Config の各フィールドをバリデーションする。
func validate(cfg *Config) error {
	var errs []string

	if cfg.Rules.MaxLinesPerFile <= 0 {
		errs = append(errs, `"max_lines_per_file" must be a positive integer`)
	}
	if cfg.Rules.MaxLinesPerDirectory <= 0 {
		errs = append(errs, `"max_lines_per_directory" must be a positive integer`)
	}
	if cfg.Rules.WarningThreshold < 0 || cfg.Rules.WarningThreshold > 100 {
		errs = append(errs, `"warning_threshold" must be between 0 and 100`)
	}
	if cfg.CountMode != "all" && cfg.CountMode != "code_only" {
		errs = append(errs, `"count_mode" must be "all" or "code_only"`)
	}
	if cfg.Language != "en" && cfg.Language != "ja" {
		errs = append(errs, `"language" must be "en" or "ja"`)
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
