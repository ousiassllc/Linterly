package cli

import (
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
