# Linterly

A lightweight linter that checks code line counts. Set line limits per file and per directory to detect bloated code early.

> [日本語](./README.ja.md)

## Features

- **File line count check** — Set a maximum line count per file (default: 300)
- **Directory line count check** — Limit total lines of files directly under a directory (default: 2,000)
- **Graduated violation levels** — Distinguish between `warn` (within threshold) and `error` (exceeds threshold)
- **Code-only counting** — Option to exclude comments and blank lines
- **Multi-language comment recognition** — Go, Rust, JavaScript/TypeScript, Python, Ruby, Java, C/C++, and more
- **Flexible exclusion** — `.linterlyignore` (gitignore format) and config file exclusions
- **Rich default exclusions** — Automatically excludes `node_modules/`, `vendor/`, `.git/`, `dist/`, etc.
- **i18n support** — CLI output in English and Japanese

## Installation

```bash
# Go
go install github.com/ousiassllc/linterly/cmd/linterly@latest

# npm (global)
npm install -g @linterly/cli

# npm (project-local, recommended)
npm install -D @linterly/cli
```

Platform-specific binaries are also available from [GitHub Releases](https://github.com/ousiassllc/linterly/releases). See [@linterly/cli](https://www.npmjs.com/package/@linterly/cli) on npm for package details.

## Usage

```bash
# Check the current directory
linterly check

# Check a specific path
linterly check src/

# Output in JSON format
linterly check --format json

# Override config values with CLI flags
linterly check --max-lines-per-file 500 --count-mode code_only

# Run without a config file (uses defaults)
linterly check --max-lines-per-file 500 --warning-threshold 20

# Add ignore patterns via CLI
linterly check --ignore "vendor/**" --ignore "*.pb.go"

# Disable default excludes
linterly check --no-default-excludes

# Specify a config file
linterly check --config .linterly.yml

# Generate a config file
linterly init

# Show version
linterly version
```

### Example Output

```
  WARN  src/handler.go (325 lines, limit: 300)
  ERROR src/service.go (450 lines, limit: 300)
  ERROR src/ (2500 lines, limit: 2000)

Results: 2 error(s), 1 warning, 42 passed
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All passed (including warnings) |
| `1` | Error-level violations found |
| `2` | Runtime error (invalid config, etc.) |

## Configuration

Place a `.linterly.yml` file in your project root.

```yaml
rules:
  max_lines_per_file: 300        # Max lines per file
  max_lines_per_directory: 2000  # Max lines per directory
  warning_threshold: 10          # Warning threshold (%)

count_mode: all                  # all | code_only
language: en                     # en | ja

ignore:
  - "vendor/**"
  - "*.pb.go"
  - "**/*_generated.go"

default_excludes: true           # Enable/disable default exclusions
```

### Violation Logic

```
limit     = max_lines (e.g. 300)
threshold = limit × (1 + warning_threshold / 100) (e.g. 330)

lines ≤ limit         → PASS
limit < lines ≤ threshold → WARN (exit code 0)
lines > threshold     → ERROR (exit code 1)
```

### Exclude Files

Use `.linterlyignore` with the same format as `.gitignore`. If both `.linterlyignore` and the `ignore` field in the config file are defined, `.linterlyignore` takes precedence.

## Git Hooks Integration

### Lefthook

```yaml
# lefthook.yml
pre-commit:
  commands:
    linterly:
      glob: "*.go"
      run: linterly check {staged_files}
```

### Husky + lint-staged

```json
// package.json
{
  "lint-staged": {
    "*.{js,ts,go,py,rb}": ["linterly check"]
  }
}
```

## CI Usage

### GitHub Actions

```yaml
- run: |
    go install github.com/ousiassllc/linterly/cmd/linterly@latest
    linterly check
```

## Documentation

See [`docs/`](./docs/) for detailed specifications.

| Document | Contents |
|----------|----------|
| [Functional Requirements](./docs/requirements/functional.md) | Use cases, violation logic, default exclusions |
| [Non-functional Requirements](./docs/requirements/non-functional.md) | Performance, platforms, distribution |
| [Architecture](./docs/architecture/overview.md) | Layer structure, tech choices, package design |
| [Config Schema](./docs/architecture/config-schema.md) | YAML spec, validation, ignore format |
| [Component Design](./docs/components/overview.md) | Responsibilities and interfaces of 7 components |
| [CLI Specification](./docs/api/cli.md) | Commands, flags, output, exit codes |

## License

MIT
