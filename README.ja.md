# Linterly

コードの行数をチェックする軽量リンターツール。ファイル単位・ディレクトリ単位で行数制限を設定し、肥大化したコードを早期に検出します。

> [English](./README.md)

## 特徴

- **ファイル行数チェック** — ファイルごとの行数上限を設定（デフォルト: 300行）
- **ディレクトリ行数チェック** — ディレクトリ直下ファイルの合計行数を制限（デフォルト: 2,000行）
- **段階的な違反レベル** — `warn`（閾値内）と `error`（閾値超過）を区別
- **コード行のみカウント** — コメント・空行を除外するモードに対応
- **多言語コメント認識** — Go, Rust, JavaScript/TypeScript, Python, Ruby, Java, C/C++ など
- **柔軟な除外設定** — `.linterlyignore`（gitignore形式）と設定ファイルによる除外
- **豊富なデフォルト除外** — `node_modules/`, `vendor/`, `.git/`, `dist/` など自動除外
- **日本語対応** — CLI出力の日英切り替え

## インストール

```bash
# Go
go install github.com/ousiassllc/linterly/cmd/linterly@latest

# npm（グローバル）
npm install -g @linterly/cli

# npm（プロジェクトローカル・推奨）
npm install -D @linterly/cli
```

[GitHub Releases](https://github.com/ousiassllc/linterly/releases) からプラットフォーム別のバイナリも入手できます。npm パッケージの詳細は [@linterly/cli](https://www.npmjs.com/package/@linterly/cli) を参照してください。

## 使い方

```bash
# カレントディレクトリをチェック
linterly check

# 特定のパスをチェック
linterly check src/

# JSON形式で出力
linterly check --format json

# 設定ファイルを指定
linterly check --config .linterly.yml

# 設定ファイルを生成
linterly init

# バージョン表示
linterly version
```

### 出力例

```
  WARN  src/handler.go (325 lines, limit: 300)
  ERROR src/service.go (450 lines, limit: 300)
  ERROR src/ (2500 lines, limit: 2000)

Results: 2 error(s), 1 warning, 42 passed
```

### 終了コード

| コード | 意味 |
|--------|------|
| `0` | すべてパス（warn含む） |
| `1` | error レベルの違反あり |
| `2` | 実行エラー（設定不正など） |

## 設定

`.linterly.yml` をプロジェクトルートに配置します。

```yaml
rules:
  max_lines_per_file: 300        # ファイル行数上限
  max_lines_per_directory: 2000  # ディレクトリ行数上限
  warning_threshold: 10          # 警告閾値 (%)

count_mode: all                  # all | code_only
language: en                     # en | ja

ignore:
  - "vendor/**"
  - "*.pb.go"
  - "**/*_generated.go"

default_excludes: true           # デフォルト除外の有効/無効
```

### 違反判定ロジック

```
上限 = max_lines（例: 300）
閾値 = 上限 × (1 + warning_threshold / 100)（例: 330）

行数 ≤ 上限        → PASS
上限 < 行数 ≤ 閾値 → WARN（終了コード 0）
行数 > 閾値        → ERROR（終了コード 1）
```

### 除外ファイル

`.linterlyignore` を gitignore と同じ形式で記述できます。`.linterlyignore` の設定は設定ファイルの `ignore` より優先されます。

## Git Hooks との連携

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

## CI での利用

### GitHub Actions

```yaml
- run: |
    go install github.com/ousiassllc/linterly/cmd/linterly@latest
    linterly check
```

## ドキュメント

詳細な仕様は [`docs/`](./docs/) を参照してください。

| ドキュメント | 内容 |
|-------------|------|
| [機能要件](./docs/requirements/functional.md) | ユースケース・違反ロジック・デフォルト除外 |
| [非機能要件](./docs/requirements/non-functional.md) | パフォーマンス・プラットフォーム・配布 |
| [アーキテクチャ](./docs/architecture/overview.md) | レイヤー構成・技術選定・パッケージ設計 |
| [設定スキーマ](./docs/architecture/config-schema.md) | YAML仕様・バリデーション・ignore形式 |
| [コンポーネント設計](./docs/components/overview.md) | 7コンポーネントの責務・インターフェース |
| [CLI仕様](./docs/api/cli.md) | コマンド・フラグ・出力・終了コード |

## License

MIT
