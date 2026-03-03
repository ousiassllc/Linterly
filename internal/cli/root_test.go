package cli

import (
	"runtime"
	"runtime/debug"
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

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		buildInfo *debug.BuildInfo
		ok        bool
		want      string
	}{
		{
			name:    "ldflags で設定済みならそのまま",
			version: "1.2.3",
			want:    "1.2.3",
		},
		{
			name:      "dev で BuildInfo にバージョンあり",
			version:   "dev",
			buildInfo: &debug.BuildInfo{Main: debug.Module{Version: "v0.3.2"}},
			ok:        true,
			want:      "v0.3.2",
		},
		{
			name:    "dev で BuildInfo 取得不可",
			version: "dev",
			ok:      false,
			want:    "dev",
		},
		{
			name:      "dev で BuildInfo が (devel)",
			version:   "dev",
			buildInfo: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:        true,
			want:      "dev",
		},
		{
			name:      "dev で BuildInfo が空文字",
			version:   "dev",
			buildInfo: &debug.BuildInfo{Main: debug.Module{Version: ""}},
			ok:        true,
			want:      "dev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readBI := func() (*debug.BuildInfo, bool) {
				return tt.buildInfo, tt.ok
			}
			assert.Equal(t, tt.want, resolveVersion(tt.version, readBI))
		})
	}
}
