# CLI インターフェース仕様

## 1. コマンド一覧

| コマンド | 説明 |
|---------|------|
| `linterly` | ヘルプを表示する |
| `linterly check` | コード量チェックを実行する |
| `linterly init` | 設定ファイルを初期化する |
| `linterly version` | バージョン情報を表示する |

## 2. コマンド詳細

### 2.1 `linterly`（ヘルプ表示）

引数なしで実行するとヘルプを表示する。

```
$ linterly

Linterly - A code line count linter

Usage:
  linterly [command]

Available Commands:
  check       Run code line count checks
  init        Initialize a .linterly.yml config file
  version     Print version information
  help        Help about any command

Flags:
  -h, --help   help for linterly

Use "linterly [command] --help" for more information about a command.
```

日本語設定時（`language: ja`）：

```
$ linterly

Linterly - コード行数チェックツール

使い方:
  linterly [コマンド]

利用可能なコマンド:
  check       コード行数チェックを実行
  init        設定ファイル (.linterly.yml) を初期化
  version     バージョン情報を表示
  help        コマンドのヘルプを表示

フラグ:
  -h, --help   ヘルプを表示

詳しくは "linterly [コマンド] --help" を参照してください。
```

### 2.2 `linterly check`

コード量チェックを実行する。

#### 構文

```
linterly check [path] [flags]
```

#### 引数

| 引数 | 必須 | デフォルト | 説明 |
|------|------|-----------|------|
| `path` | いいえ | `.`（カレントディレクトリ） | チェック対象のパス |

#### フラグ

| フラグ | 短縮 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--config` | `-c` | `.linterly.yml` | 設定ファイルのパス |
| `--format` | `-f` | `text` | 出力形式（`text` / `json`） |
| `--lang` | | | メッセージの言語（`en` / `ja`）。設定ファイルの `language` より優先 |
| `--max-lines-per-file` | | `300` | 1ファイルあたりの最大行数。設定ファイルの `rules.max_lines_per_file` を上書き |
| `--max-lines-per-directory` | | `2000` | ディレクトリ直下ファイルの合計最大行数。設定ファイルの `rules.max_lines_per_directory` を上書き |
| `--warning-threshold` | | `10` | 警告閾値（%）。設定ファイルの `rules.warning_threshold` を上書き |
| `--count-mode` | | `all` | 行数カウントモード（`all` / `code_only`）。設定ファイルの `count_mode` を上書き |
| `--ignore` | | | 除外パターン（複数回指定可能）。設定ファイルの `ignore` を上書き |
| `--no-default-excludes` | | | デフォルト除外リストを無効化する。設定ファイルの `default_excludes: false` と同等 |

#### 設定の優先順位

CLI フラグで指定された値は、設定ファイルの値より常に優先される。

```
CLI フラグ > 設定ファイル > デフォルト値
```

- `--lang` フラグは `LINTERLY_LANG` 環境変数より優先される（既存動作）
- `--ignore` が1回以上指定された場合、設定ファイルの `ignore` は完全に置き換えられる（マージではない）
- `--no-default-excludes` は `--default-excludes=false` とは異なり、否定フラグとして機能する

#### 設定ファイルなしでの実行

設定ファイルが見つからない場合でも、エラーにせず全デフォルト値で動作する。CLI フラグの指定は不要。

```bash
# 設定ファイルなし・フラグなしでも動作（全デフォルト値で実行）
$ linterly check

# 設定ファイルなし・フラグで上書き
$ linterly check --max-lines-per-file 500 --count-mode code_only
```

`--config` または `LINTERLY_CONFIG` で明示的にパスを指定した場合、そのファイルが存在しなければ従来通りエラーになる。

#### CI / GitHub Actions での使用例

```bash
# 設定ファイルなしで CLI フラグのみで実行
linterly check \
  --max-lines-per-file 500 \
  --max-lines-per-directory 3000 \
  --warning-threshold 20 \
  --count-mode code_only \
  --format json

# 設定ファイルの値を一部だけ上書き
linterly check --max-lines-per-file 500

# 除外パターンを CLI で指定
linterly check --ignore vendor/ --ignore generated/ --ignore '*.pb.go'
```

#### テキスト出力例

```
$ linterly check

  WARN  src/handler.go (325 lines, limit: 300)
  ERROR src/service.go (450 lines, limit: 300)
  ERROR src/ (2500 lines, limit: 2000)

Results: 2 error(s), 1 warning(s), 42 passed
```

日本語設定時：

```
$ linterly check

  WARN  src/handler.go (325 行, 上限: 300)
  ERROR src/service.go (450 行, 上限: 300)
  ERROR src/ (2500 行, 上限: 2000)

結果: 2 エラー, 1 警告, 42 パス
```

#### JSON 出力例

```json
{
  "results": [
    {
      "path": "src/handler.go",
      "type": "file",
      "lines": 325,
      "limit": 300,
      "threshold": 330,
      "severity": "warn"
    },
    {
      "path": "src/service.go",
      "type": "file",
      "lines": 450,
      "limit": 300,
      "threshold": 330,
      "severity": "error"
    },
    {
      "path": "src/",
      "type": "directory",
      "lines": 2500,
      "limit": 2000,
      "threshold": 2200,
      "severity": "error"
    }
  ],
  "summary": {
    "errors": 2,
    "warnings": 1,
    "passed": 42,
    "total": 45
  }
}
```

- JSON 出力のキー名は常に英語（`language` 設定に依存しない）
- `threshold` は `limit × (1 + warning_threshold / 100)` の計算値

#### ignore 重複警告

`.linterlyignore` と設定ファイルの `ignore` が両方存在する場合：

```
$ linterly check

  WARN  Both .linterlyignore and ignore in config file are defined. .linterlyignore takes precedence. ignore in config file is ignored.

  WARN  src/handler.go (325 lines, limit: 300)
  ...
```

### 2.3 `linterly init`

設定ファイルを初期化する。

#### 構文

```
linterly init [flags]
```

#### フラグ

なし。

#### 動作

1. カレントディレクトリに `.linterly.yml` を生成する
2. ファイルが既に存在する場合は上書き確認を行う

```
$ linterly init
Created .linterly.yml

$ linterly init
.linterly.yml already exists. Overwrite? [y/N]: y
Overwritten .linterly.yml
```

### 2.4 `linterly version`

バージョン情報を表示する。

```
$ linterly version
linterly v1.0.0 (go1.25.6, linux/amd64)
```

- バージョン文字列はビルド時に `-ldflags` で設定される。開発時は `dev` が表示される
- `v` プレフィックスは `git tag` のタグ名に含めることを前提とする

## 3. 終了コード

| コード | 意味 |
|--------|------|
| `0` | チェック成功（違反なし、または warn のみ） |
| `1` | チェック失敗（error が 1 つ以上存在） |
| `2` | 実行エラー（設定ファイル不正、引数エラー等） |

## 4. 環境変数

| 変数 | 説明 | デフォルト |
|------|------|-----------|
| `LINTERLY_CONFIG` | 設定ファイルのパス（`--config` フラグと同等） | なし |
| `LINTERLY_LANG` | メッセージの言語（`en` / `ja`）。`--lang` フラグと同等 | なし |
| `NO_COLOR` | 設定するとカラー出力を無効化する（[no-color.org](https://no-color.org) 準拠） | なし |

- `--config` フラグが指定された場合は `LINTERLY_CONFIG` より優先される
- `--lang` フラグが指定された場合は `LINTERLY_LANG` より優先される
- 言語の優先順位: `--lang` フラグ > `LINTERLY_LANG` 環境変数 > 設定ファイルの `language` > デフォルト `en`

## 5. 設定の優先順位（全体）

```
CLI フラグ > 環境変数 > 設定ファイル > デフォルト値
```

各設定項目の優先順位:

| 設定項目 | CLI フラグ | 環境変数 | 設定ファイル | デフォルト |
|---------|-----------|----------|------------|-----------|
| 設定ファイルパス | `--config` | `LINTERLY_CONFIG` | — | `.linterly.yml` 探索 |
| 言語 | `--lang` | `LINTERLY_LANG` | `language` | `en` |
| 最大行数/ファイル | `--max-lines-per-file` | — | `rules.max_lines_per_file` | `300` |
| 最大行数/ディレクトリ | `--max-lines-per-directory` | — | `rules.max_lines_per_directory` | `2000` |
| 警告閾値 | `--warning-threshold` | — | `rules.warning_threshold` | `10` |
| カウントモード | `--count-mode` | — | `count_mode` | `all` |
| 除外パターン | `--ignore` | — | `ignore` | `[]` |
| デフォルト除外 | `--no-default-excludes` | — | `default_excludes` | `true` |
| カラー無効化 | — | `NO_COLOR` | — | 未設定（カラー有効） |

## 改訂履歴

| 版 | 日付 | 変更内容 | 変更理由 |
|---|------|---------|---------|
| 1.0 | 2026-02-08 | 初版作成 | — |
| 1.1 | 2026-02-08 | テキスト出力例のエラー件数を 2 に修正 | JSON 出力例・出力例の ERROR 件数との整合性確保 |
| 1.2 | 2026-02-08 | warning(s) 表記修正、ignore 重複警告を 1 行表記に修正、version 出力例更新、--lang フラグと LINTERLY_LANG 環境変数を追加 | ドキュメント乖離レポート (#3) 対応 |
| 1.3 | 2026-02-24 | check コマンドに設定上書きフラグ（--max-lines-per-file 等6種）を追加、設定ファイルなし実行の対応、優先順位の明記 | #22 CLI フラグによる設定値の上書き対応 |
