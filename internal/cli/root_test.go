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

	expected := "linterly v1.2.3 (" + runtime.Version() + ", " + runtime.GOOS + "/" + runtime.GOARCH + ")\n"
	assert.Equal(t, expected, output)
}

func TestDisplayVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "dev はそのまま", version: "dev", want: "dev"},
		{name: "v プレフィックスなし", version: "1.0.0", want: "v1.0.0"},
		{name: "v プレフィックスあり", version: "v2.1.0", want: "v2.1.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldVersion := Version
			Version = tt.version
			defer func() { Version = oldVersion }()

			assert.Equal(t, tt.want, displayVersion())
		})
	}
}
