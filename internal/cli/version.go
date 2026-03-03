package cli

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// Version はビルド時に -ldflags で設定される。
var Version = "dev"

func init() {
	Version = resolveVersion(Version, debug.ReadBuildInfo)
}

// resolveVersion は Version が "dev" の場合に BuildInfo からバージョンを解決する。
func resolveVersion(version string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if version != "dev" {
		return version
	}
	info, ok := readBuildInfo()
	if !ok || info == nil {
		return version
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	return version
}

// displayVersion は表示用のバージョン文字列を返す。
// GoReleaser が v プレフィックスを除去するため、表示時に補完する。
func displayVersion() string {
	if Version == "dev" {
		return "dev"
	}
	if strings.HasPrefix(Version, "v") {
		return Version
	}
	return "v" + Version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("linterly %s (%s, %s/%s)\n", displayVersion(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}
