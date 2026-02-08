package main

import (
	"os"

	"github.com/ousiassllc/linterly/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
