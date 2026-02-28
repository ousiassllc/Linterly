package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ousiassllc/linterly/internal/analyzer"
	"github.com/ousiassllc/linterly/internal/config"
	"github.com/ousiassllc/linterly/internal/counter"
	"github.com/ousiassllc/linterly/internal/i18n"
	"github.com/ousiassllc/linterly/internal/reporter"
	"github.com/ousiassllc/linterly/internal/scanner"
)

var (
	// configFile は --config フラグの値を保持する。
	configFile string
	// format は --format フラグの値を保持する。
	format string

	// 設定上書きフラグ
	flagMaxLinesPerFile      int
	flagMaxLinesPerDirectory int
	flagWarningThreshold     int
	flagCountMode            string
	flagIgnore               []string
	flagNoDefaultExcludes    bool
)

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Run code line count checks",
	Long:  "Check source code line counts against configured rules and report violations.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file (default is .linterly.yml)")
	checkCmd.Flags().StringVarP(&format, "format", "f", reporter.FormatText, "output format (text or json)")

	// 設定上書きフラグ
	checkCmd.Flags().IntVar(&flagMaxLinesPerFile, "max-lines-per-file", 300, "max lines per file")
	checkCmd.Flags().IntVar(&flagMaxLinesPerDirectory, "max-lines-per-directory", 2000, "max lines per directory")
	checkCmd.Flags().IntVar(&flagWarningThreshold, "warning-threshold", 10, "warning threshold (%)")
	checkCmd.Flags().StringVar(&flagCountMode, "count-mode", "all", "count mode (all or code_only)")
	checkCmd.Flags().StringArrayVar(&flagIgnore, "ignore", nil, "ignore pattern (can be specified multiple times)")
	checkCmd.Flags().BoolVar(&flagNoDefaultExcludes, "no-default-excludes", false, "disable default excludes")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// ターゲットパスの決定
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// config 読み込み前に言語を解決して Translator を初期化
	translator, lang, err := initTranslator()
	if err != nil {
		return err
	}

	// 設定ファイルの読み込み
	cfg, err := config.Load(configFile)
	if err != nil {
		return NewRuntimeError("%s", translateConfigError(translator, err))
	}

	// config.Language がフラグ/環境変数と異なる場合、config 側の言語で再初期化
	if langFlag == "" && os.Getenv("LINTERLY_LANG") == "" && cfg.Language != lang {
		translator, err = i18n.New(cfg.Language)
		if err != nil {
			return NewRuntimeError("failed to initialize i18n: %v", err)
		}
	}

	// CLI フラグによる設定上書き
	overrides := buildOverrides(cmd)
	if err := cfg.ApplyOverrides(overrides); err != nil {
		return NewRuntimeError("%s", translateConfigError(translator, err))
	}

	// ignore パターンの取得と警告
	_, warnings, err := cfg.IgnorePatterns()
	if err != nil {
		return NewRuntimeError("failed to load ignore patterns: %v", err)
	}

	// ファイル走査
	scanResult, err := scanner.Scan(targetPath, cfg)
	if err != nil {
		return NewRuntimeError("failed to scan files: %v", err)
	}

	// ファイルパスを絶対パスに変換（カウント用）
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return NewRuntimeError("failed to resolve path: %v", err)
	}

	filePaths := make([]string, len(scanResult.Files))
	for i, f := range scanResult.Files {
		filePaths[i] = filepath.Join(absTarget, f.Path)
	}

	// 行数カウント
	counts, err := counter.CountFiles(filePaths, cfg.CountMode)
	if err != nil {
		return NewRuntimeError("failed to count lines: %v", err)
	}

	// カウント結果のパスを相対パスに戻す
	for i := range counts {
		counts[i].Path = scanResult.Files[i].Path
	}

	// ルール評価
	report := analyzer.Analyze(counts, scanResult, cfg)

	// 結果出力
	rep := reporter.NewReporter(format, translator, os.Stdout)
	if err := rep.Report(report, warnings); err != nil {
		return NewRuntimeError("failed to write report: %v", err)
	}

	// 終了コード
	if report.Errors > 0 {
		return NewViolationError()
	}

	return nil
}

// buildOverrides は cmd のフラグから Overrides を構築する。
// Changed() == true のフラグのみセットし、未指定のフラグは nil（上書きしない）。
func buildOverrides(cmd *cobra.Command) *config.Overrides {
	o := &config.Overrides{}
	flags := cmd.Flags()

	if flags.Changed("max-lines-per-file") {
		o.MaxLinesPerFile = &flagMaxLinesPerFile
	}
	if flags.Changed("max-lines-per-directory") {
		o.MaxLinesPerDirectory = &flagMaxLinesPerDirectory
	}
	if flags.Changed("warning-threshold") {
		o.WarningThreshold = &flagWarningThreshold
	}
	if flags.Changed("count-mode") {
		o.CountMode = &flagCountMode
	}
	if flags.Changed("ignore") {
		o.Ignore = flagIgnore
	}
	if flags.Changed("no-default-excludes") {
		o.NoDefaultExcludes = flagNoDefaultExcludes
	}

	return o
}

// initTranslator は langFlag から言語を解決し、Translator を初期化する。
// 解決された言語コードも返す（config.Language との比較用）。
func initTranslator() (*i18n.Translator, string, error) {
	lang := i18n.ResolveLanguage(langFlag)
	translator, err := i18n.New(lang)
	if err != nil {
		return nil, "", NewRuntimeError("failed to initialize i18n: %v", err)
	}
	return translator, lang, nil
}

// translateConfigError は config パッケージのエラーを i18n メッセージに変換する。
func translateConfigError(tr *i18n.Translator, err error) string {
	var valErrs *config.ValidationErrors
	if errors.As(err, &valErrs) {
		msgs := make([]string, len(valErrs.Errors))
		for i, e := range valErrs.Errors {
			msgs[i] = tr.T(e.Code)
		}
		return strings.Join(msgs, "; ")
	}

	var cfgErr *config.ConfigError
	if errors.As(err, &cfgErr) {
		if cfgErr.Detail != "" {
			return tr.T(cfgErr.Code, cfgErr.Detail)
		}
		return tr.T(cfgErr.Code)
	}

	// ConfigError でない場合はそのまま返す
	return err.Error()
}
