package config

// DefaultExcludePatterns はデフォルト除外パターン一覧を返す。
// default_excludes: true の場合に scanner で使用される。
func DefaultExcludePatterns() []string {
	return []string{
		// 共通
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

		// JavaScript/TypeScript
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

		// Go
		"vendor/",
		"ent/",

		// Rust
		"target/",

		// Python
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
}
