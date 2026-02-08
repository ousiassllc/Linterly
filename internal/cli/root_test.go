package cli

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	assert.Equal(t, "linterly", rootCmd.Use)
}

func TestRootHasSubCommands(t *testing.T) {
	names := make([]string, 0)
	for _, cmd := range rootCmd.Commands() {
		names = append(names, cmd.Name())
	}
	assert.Contains(t, names, "check")
	assert.Contains(t, names, "init")
	assert.Contains(t, names, "version")
}

func TestExecute(t *testing.T) {
	err := Execute()
	assert.NoError(t, err)
}

func TestVersionCommand_Output(t *testing.T) {
	// Version 変数を一時的にセット
	oldVersion := Version
	Version = "1.2.3"
	defer func() { Version = oldVersion }()

	output := helperCaptureStdout(t, func() {
		versionCmd.Run(versionCmd, nil)
	})

	expected := "linterly 1.2.3 (" + runtime.Version() + ", " + runtime.GOOS + "/" + runtime.GOARCH + ")\n"
	assert.Equal(t, expected, output)
}
