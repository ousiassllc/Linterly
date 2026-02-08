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

#### テキスト出力例

```
$ linterly check

  WARN  src/handler.go (325 lines, limit: 300)
  ERROR src/service.go (450 lines, limit: 300)
  ERROR src/ (2500 lines, limit: 2000)

Results: 2 error(s), 1 warning, 42 passed
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

  WARN  Both .linterlyignore and ignore in config file are defined.
        .linterlyignore takes precedence. ignore in config file is ignored.

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
linterly v1.0.0 (go1.22.0, linux/amd64)
```

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
| `NO_COLOR` | 設定するとカラー出力を無効化する（[no-color.org](https://no-color.org) 準拠） | なし |

- `--config` フラグが指定された場合は環境変数より優先される

## 改訂履歴

| 版 | 日付 | 変更内容 | 変更理由 |
|---|------|---------|---------|
| 1.0 | 2026-02-08 | 初版作成 | — |
| 1.1 | 2026-02-08 | テキスト出力例のエラー件数を 2 に修正 | JSON 出力例・出力例の ERROR 件数との整合性確保 |
