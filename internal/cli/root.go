package cli

import (
	"github.com/spf13/cobra"
)

var langFlag string

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
