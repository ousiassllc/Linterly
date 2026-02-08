package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ousiassllc/linterly/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		var exitErr *cli.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Code == cli.ExitRuntimeError {
				// ExitViolation の場合は check コマンド内でレポート出力済みのため追加メッセージ不要
				fmt.Fprintln(os.Stderr, "Error: "+exitErr.Message)
			}
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		os.Exit(2)
	}
}
