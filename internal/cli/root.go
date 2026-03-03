package cli

import (
	"github.com/spf13/cobra"
)

// langFlag は --lang フラグの値を保持する。
var langFlag string

// flagNoUpdateCheck は --no-update-check フラグの値を保持する。
var flagNoUpdateCheck bool

// rootCmd はルートコマンド。SilenceErrors/SilenceUsage により main.go でエラー出力を一元管理する。
var rootCmd = &cobra.Command{
	Use:           "linterly",
	Short:         "Linterly - A code line count linter",
	Long:          "Linterly checks source code line counts against configured rules and reports violations.",
	SilenceErrors: true,
	SilenceUsage:  true,
	// PersistentPreRun は全サブコマンドに継承される。
	// 注意: サブコマンドが PersistentPreRun を再定義すると、
	// このフックは呼ばれなくなる（Cobra の仕様）。
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		startUpdateCheck()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", "language for messages (en, ja)")
	rootCmd.PersistentFlags().BoolVar(&flagNoUpdateCheck, "no-update-check", false, "disable update check")
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command.
func Execute() error {
	err := rootCmd.Execute()
	printUpdateNotice()
	return err
}
