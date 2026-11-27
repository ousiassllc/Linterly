# @linterly/cli

A lightweight linter that checks code line counts. Set line limits per file and per directory to detect bloated code early.

## Installation

```bash
# Global
npm install -g @linterly/cli

# Project-local (recommended)
npm install -D @linterly/cli
```

Also available via Go:

```bash
go install github.com/ousiassllc/linterly/cmd/linterly@latest
```

## Quick Start

```bash
# Generate a config file
npx linterly init

# Check the current directory
npx linterly check

# Check a specific path
npx linterly check src/

# Check specific files (multiple paths allowed; e.g. staged files from Git hooks)
npx linterly check src/handler.go src/service.go

# Output in JSON format
npx linterly check --format json

# Output in JUnit XML / checkstyle XML format
npx linterly check --format junit
npx linterly check --format checkstyle

# Override config values with CLI flags
npx linterly check --max-lines-per-file 500 --count-mode code_only

# Run without a config file (uses defaults)
npx linterly check --max-lines-per-file 500 --warning-threshold 20

# Add ignore patterns via CLI
npx linterly check --ignore "vendor/**" --ignore "*.pb.go"

# Disable default excludes
npx linterly check --no-default-excludes
```

## Configuration

A config file is **optional**. Without one, Linterly runs with sensible defaults. You can override any setting via CLI flags.

Place a `.linterly.yml` in your project root:

```yaml
rules:
  max_lines_per_file: 300
  max_lines_per_directory: 2000
  warning_threshold: 10

count_mode: all       # all | code_only
language: en          # en | ja

ignore:
  - "vendor/**"
  - "*.pb.go"

default_excludes: true
update_check: true    # set to false to disable the update check
```

CLI flags always take precedence over config file values (`CLI flags > config file > defaults`).

You can also create a `.linterlyignore` file (gitignore format) for exclusion patterns. Patterns are evaluated relative to the current directory (the project root).

Linterly checks for a newer release in the background and prints a notice to stderr. Disable it with `--no-update-check`, the `LINTERLY_NO_UPDATE_CHECK` environment variable, or `update_check: false` in the config file.

## Git Hooks Integration

### Husky + lint-staged

```json
{
  "lint-staged": {
    "*.{js,ts,go,py,rb}": ["linterly check"]
  }
}
```

lint-staged appends the staged file names to the command. Directory line limits are evaluated only against those files.

### Lefthook

```yaml
pre-commit:
  commands:
    linterly:
      glob: "*.go"
      run: linterly check {staged_files}
```

`check` accepts file paths, so the staged files are passed as-is. Directory line limits are evaluated only against the files you pass.

## Supported Platforms

| Platform | Package |
|----------|---------|
| Linux x64 | `@linterly/linux-x64` |
| Linux arm64 | `@linterly/linux-arm64` |
| macOS x64 | `@linterly/darwin-x64` |
| macOS arm64 | `@linterly/darwin-arm64` |
| Windows x64 | `@linterly/win32-x64` |

Platform-specific packages are installed automatically via `optionalDependencies`.

## Links

- [GitHub](https://github.com/ousiassllc/linterly)
- [GitHub Releases](https://github.com/ousiassllc/linterly/releases)
- [Documentation](https://github.com/ousiassllc/linterly/tree/main/docs)

## License

MIT
