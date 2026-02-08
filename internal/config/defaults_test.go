package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultExcludePatterns_NotEmpty(t *testing.T) {
	patterns := DefaultExcludePatterns()
	assert.NotEmpty(t, patterns)
}

func TestDefaultExcludePatterns_ContainsCommonPatterns(t *testing.T) {
	patterns := DefaultExcludePatterns()

	common := []string{
		".git/",
		"dist/",
		"build/",
		"out/",
		"*.min.js",
		"*.min.css",
		"*.lock",
		"*-lock.*",
		"*.gen.*",
		"*.generated.*",
		".idea/",
		".vscode/",
		".claude/",
		".cursor/",
		".gemini/",
		".cache/",
	}
	for _, p := range common {
		assert.Contains(t, patterns, p)
	}
}

func TestDefaultExcludePatterns_ContainsJavaScriptPatterns(t *testing.T) {
	patterns := DefaultExcludePatterns()

	jsPatterns := []string{
		"node_modules/",
		"bower_components/",
		".next/",
		".nuxt/",
		".svelte-kit/",
		".angular/",
		".turbo/",
		".parcel-cache/",
		".vite/",
		"coverage/",
	}
	for _, p := range jsPatterns {
		assert.Contains(t, patterns, p)
	}
}

func TestDefaultExcludePatterns_ContainsGoPatterns(t *testing.T) {
	patterns := DefaultExcludePatterns()

	goPatterns := []string{
		"vendor/",
		"ent/",
	}
	for _, p := range goPatterns {
		assert.Contains(t, patterns, p)
	}
}

func TestDefaultExcludePatterns_ContainsRustPatterns(t *testing.T) {
	patterns := DefaultExcludePatterns()
	assert.Contains(t, patterns, "target/")
}

func TestDefaultExcludePatterns_ContainsPythonPatterns(t *testing.T) {
	patterns := DefaultExcludePatterns()

	pyPatterns := []string{
		"__pycache__/",
		"*.pyc",
		"*.pyo",
		".venv/",
		"venv/",
		"env/",
		"*.egg-info/",
		".eggs/",
		".mypy_cache/",
		".pytest_cache/",
		".tox/",
	}
	for _, p := range pyPatterns {
		assert.Contains(t, patterns, p)
	}
}

func TestDefaultExcludePatterns_ReturnsNewSlice(t *testing.T) {
	p1 := DefaultExcludePatterns()
	p2 := DefaultExcludePatterns()

	// 別のスライスが返されることを確認（変更が互いに影響しない）
	assert.Equal(t, p1, p2)
	p1[0] = "modified"
	assert.NotEqual(t, p1[0], p2[0])
}
