package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	assert.Equal(t, "linterly", rootCmd.Use)
}

func TestExecute(t *testing.T) {
	// Execute with no subcommand should succeed (prints help).
	err := Execute()
	assert.NoError(t, err)
}
