package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "linterly",
	Short: "Linterly - A code line count linter",
	Long:  "Linterly checks source code line counts against configured rules and reports violations.",
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .linterly.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".linterly")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
