package cli

import (
	"github.com/spf13/cobra"
)

// langFlag は --lang フラグの値を保持する。
var langFlag string

// rootCmd はルートコマンド。SilenceErrors/SilenceUsage により main.go でエラー出力を一元管理する。
var rootCmd = &cobra.Command{
	Use:           "linterly",
	Short:         "Linterly - A code line count linter",
	Long:          "Linterly checks source code line counts against configured rules and reports violations.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", "language for messages (en, ja)")
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
