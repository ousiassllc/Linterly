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

	// ファイルが既に存在するか確認
	if _, err := os.Stat(filename); err == nil {
		fmt.Print(filename + " already exists. Overwrite? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			return nil
		}
		if err := os.WriteFile(filename, []byte(config.DefaultConfigTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		fmt.Println("Overwritten " + filename)
		return nil
	}

	if err := os.WriteFile(filename, []byte(config.DefaultConfigTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	fmt.Println("Created " + filename)
	return nil
}
