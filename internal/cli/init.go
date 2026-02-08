package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ousiassllc/linterly/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a .linterly.yml config file",
	Long:  "Create a .linterly.yml configuration file with default settings in the current directory.",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	const filename = ".linterly.yml"

	// config 読み込み前に言語を解決して Translator を初期化
	translator, _, err := initTranslator()
	if err != nil {
		return err
	}

	// ファイルが既に存在するか確認
	if _, err := os.Stat(filename); err == nil {
		fmt.Print(translator.T("init.overwrite") + " ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return nil // 入力がない場合は上書きしない
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			return nil
		}
		if err := os.WriteFile(filename, []byte(config.DefaultConfigTemplate), 0644); err != nil {
			return NewRuntimeError("failed to write config file: %v", err)
		}
		fmt.Println(translator.T("init.overwritten"))
		return nil
	}

	if err := os.WriteFile(filename, []byte(config.DefaultConfigTemplate), 0644); err != nil {
		return NewRuntimeError("failed to write config file: %v", err)
	}
	fmt.Println(translator.T("init.created"))
	return nil
}
