package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version はビルド時に -ldflags で設定される。
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("linterly %s (%s, %s/%s)\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}
