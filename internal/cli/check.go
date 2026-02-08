package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ousiassllc/linterly/internal/analyzer"
	"github.com/ousiassllc/linterly/internal/config"
	"github.com/ousiassllc/linterly/internal/counter"
	"github.com/ousiassllc/linterly/internal/i18n"
	"github.com/ousiassllc/linterly/internal/reporter"
	"github.com/ousiassllc/linterly/internal/scanner"
)

var (
	configFile string
	format     string
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
	checkCmd.Flags().StringVarP(&format, "format", "f", "text", "output format (text or json)")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// ターゲットパスの決定
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// 設定ファイルの読み込み
	cfg, err := config.Load(configFile)
	if err != nil {
		return NewRuntimeError("%s", err)
	}

	// i18n の初期化
	translator, err := i18n.New(cfg.Language)
	if err != nil {
		return NewRuntimeError("failed to initialize i18n: %v", err)
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
