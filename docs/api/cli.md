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

Linterly checks source code line counts against configured rules and reports violations.

Usage:
  linterly [command]

Available Commands:
  check       Run code line count checks
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize a .linterly.yml config file
  version     Print version information

Flags:
  -h, --help              help for linterly
      --lang string       language for messages (en, ja)
      --no-update-check   disable update check

Use "linterly [command] --help" for more information about a command.
```

- ヘルプテキスト（`--help` の出力）は cobra の標準出力であり、`--lang` / `LINTERLY_LANG` / `language` の設定に関わらず常に英語で表示される（i18n の対象外）
- `completion` は cobra が自動で提供するシェル補完スクリプト生成コマンド
- `--lang`・`--no-update-check` は全コマンド共通のグローバルフラグ

### 2.2 `linterly check`

コード量チェックを実行する。

#### 構文

```
linterly check [path...] [flags]
```

#### 引数

| 引数 | 必須 | デフォルト | 説明 |
|------|------|-----------|------|
| `path...` | いいえ | `.`（カレントディレクトリ） | チェック対象のパス。ディレクトリまたはファイルを 0 個以上指定できる |

- `path` にファイルを直接指定できる（例: `linterly check src/a.go src/b.go`）。Git フックの `{staged_files}` や lint-staged が末尾に付与するファイル名もそのまま渡せる
- 結果に出力するパスの基準は、単一のディレクトリを指定した場合はそのディレクトリ、それ以外（ファイルの指定・複数パスの指定）の場合はカレントディレクトリ
- 除外パターン（デフォルト除外・`ignore`）とバイナリ判定は、直接指定したファイルにも通常の走査と同様に適用される。同じファイルが複数回見つかった場合は 1 件として扱う
- ディレクトリ行数のチェックは、指定したパスから見つかったファイルのみを対象に合計する（指定しなかったファイルは合計に含まれない）
- 存在しないパスを指定した場合は実行エラー（終了コード 2）
- 読み取れないファイル（権限なし・走査後の削除など）は stderr に `warning: skipped <path>: <理由>` を出力してスキップし、チェックを続行する。終了コードには影響しない

#### フラグ

| フラグ | 短縮 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--config` | `-c` | `.linterly.yml` | 設定ファイルのパス |
| `--format` | `-f` | `text` | 出力形式（`text` / `json` / `junit` / `checkstyle`）。未知の値は使用エラー（終了コード 2、走査前に検出）: `Error: Unsupported format: xml (text, json, junit, checkstyle)` |
| `--lang` | | | メッセージの言語（`en` / `ja`）。設定ファイルの `language` より優先（グローバルフラグ、全コマンド共通） |
| `--max-lines-per-file` | | `300` | 1ファイルあたりの最大行数。設定ファイルの `rules.max_lines_per_file` を上書き |
| `--max-lines-per-directory` | | `2000` | ディレクトリ直下ファイルの合計最大行数。設定ファイルの `rules.max_lines_per_directory` を上書き |
| `--warning-threshold` | | `10` | 警告閾値（%）。設定ファイルの `rules.warning_threshold` を上書き |
| `--count-mode` | | `all` | 行数カウントモード（`all` / `code_only`）。設定ファイルの `count_mode` を上書き |
| `--ignore` | | | 除外パターン（複数回指定可能）。設定ファイルの `ignore` を上書き。パターンは常にプロジェクトルート基準で評価される |
| `--no-default-excludes` | | | デフォルト除外リストを無効化する。設定ファイルの `default_excludes: false` と同等 |
| `--no-update-check` | | | バージョン更新チェックを無効化する（グローバルフラグ、全コマンド共通） |

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
  "warnings": [],
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
- トップレベルのキーは `warnings` / `results` / `summary`。`warnings` は警告がなくても常に出力され、その場合は空配列（`[]`）
- `warnings` の要素は翻訳済みメッセージではなくメッセージキー（ignore 重複警告時は `"ignore.both_defined"`）。キー名は `language` 設定に依存しない

#### JUnit XML 出力例

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="file" tests="43" failures="1">
    <testcase name="src/handler.go" classname="file">
      <system-out>WARN src/handler.go (325 lines, limit: 300)</system-out>
    </testcase>
    <testcase name="src/service.go" classname="file">
      <failure message="450 lines, limit: 300">ERROR src/service.go (450 lines, limit: 300)</failure>
    </testcase>
    <testcase name="src/utils.go" classname="file" />
    <!-- ... pass は空の testcase -->
  </testsuite>
  <testsuite name="directory" tests="2" failures="1">
    <testcase name="src/" classname="directory">
      <failure message="2500 lines, limit: 2000">ERROR src/ (2500 lines, limit: 2000)</failure>
    </testcase>
    <testcase name="lib/" classname="directory" />
  </testsuite>
</testsuites>
```

- `<testsuite>` は `file` と `directory` の2つに分割
- `error` → `<failure>` 要素で出力
- `warn` → pass 扱い（`<failure>` なし）、内容は `<system-out>` に記載
- `pass` → 空の `<testcase>`（自己閉じタグ）
- メッセージは常に英語（`language` 設定に依存しない）
- ignore 重複警告などの `warnings` は出力しない（JUnit XML に出力する要素がないため破棄される）

#### checkstyle XML 出力例

```xml
<?xml version="1.0" encoding="UTF-8"?>
<checkstyle version="4.3">
  <file name="src/handler.go">
    <error line="1" severity="warning" message="325 lines (limit: 300)" source="linterly" />
  </file>
  <file name="src/service.go">
    <error line="1" severity="error" message="450 lines (limit: 300)" source="linterly" />
  </file>
  <file name="src/">
    <error line="1" severity="error" message="2500 lines (limit: 2000)" source="linterly" />
  </file>
</checkstyle>
```

- violation（warn / error）のみ出力（pass は除外）
- ディレクトリの違反も `<file>` として出力
- `severity` は `"warning"` / `"error"` （checkstyle 標準に準拠）
- `line` 属性は `1` 固定（ファイル全体が対象のため）
- `source` 属性は `"linterly"` 固定
- メッセージは常に英語（`language` 設定に依存しない）
- ignore 重複警告などの `warnings` は出力しない（checkstyle XML に出力する要素がないため破棄される）

#### ignore 重複警告

`.linterlyignore` と設定ファイルの `ignore` が両方存在する場合、text 出力では次の警告を先頭に表示する。JSON 出力では `warnings` 配列にメッセージキー（`ignore.both_defined`）が入る。JUnit XML / checkstyle XML では出力されない。

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

## 3. バージョン更新チェック

全コマンド実行時に、バックグラウンドで最新バージョンをチェックする。更新がある場合、コマンド出力の末尾に stderr で通知を表示する。通知メッセージは i18n 対応（`language` 設定・`--lang` フラグに従う）。

- メインコマンドの終了後、更新チェックが未完了の場合は最大 1 秒だけ完了を待つ。これにより、短時間で終わるコマンドでも API 応答がキャッシュに保存され、通知を表示できる
- 1 秒以内に完了しなかった場合は通知をスキップし、メインコマンドの終了を優先する。キャッシュが有効な場合はチェックが即座に完了するため待ちは発生しない
- チェックの失敗（ネットワークエラー等）は何も出力せずスキップする（エラー表示・ログ出力なし）

#### 出力例（go install 経由）

```
$ linterly check

  WARN  src/handler.go (325 lines, limit: 300)

Results: 0 error(s), 1 warning(s), 42 passed

A new version of linterly is available: v0.3.1 → v0.4.0
Run `go install github.com/ousiassllc/linterly/cmd/linterly@latest` to update.
```

#### 出力例（npm 経由）

```
$ linterly check

  WARN  src/handler.go (325 lines, limit: 300)

Results: 0 error(s), 1 warning(s), 42 passed

A new version of linterly is available: v0.3.1 → v0.4.0
Run `npm update -g @linterly/cli` to update.
```

#### 出力例（その他 / 経路不明）

```
A new version of linterly is available: v0.3.1 → v0.4.0
Visit https://github.com/ousiassllc/linterly/releases/latest to update.
```

#### 出力例（バージョン不明時）

```
Unable to determine the current version of linterly.
Visit https://github.com/ousiassllc/linterly/releases/latest to get the latest version.
```

- バージョン不明時は最新バージョンとの比較を行わず、毎回この通知を表示する
- キャッシュは使用しない（バージョン比較ができないため）

#### インストール経路の検出ロジック

| 経路 | 検出方法 | 案内メッセージ |
|------|---------|--------------|
| npm | 実行バイナリのパスに `node_modules` を含む | `npm update -g @linterly/cli` |
| go install | 実行バイナリのパスに `go/bin` を含む | `go install github.com/ousiassllc/linterly/cmd/linterly@latest` |
| その他 | 上記に該当しない | GitHub Releases URL |

#### 無効化

- `--no-update-check` フラグ（全コマンド共通）
- `LINTERLY_NO_UPDATE_CHECK` 環境変数（値が空でなければ無効化）
- 設定ファイルの `update_check: false`

## 4. 終了コード

| コード | 意味 |
|--------|------|
| `0` | チェック成功（違反なし、または warn のみ） |
| `1` | チェック失敗（error が 1 つ以上存在） |
| `2` | 実行エラー（設定ファイル不正、引数エラー、未知の `--format` 値、存在しないパスの指定等） |

## 5. 環境変数

| 変数 | 説明 | デフォルト |
|------|------|-----------|
| `LINTERLY_CONFIG` | 設定ファイルのパス（`--config` フラグと同等） | なし |
| `LINTERLY_LANG` | メッセージの言語（`en` / `ja`）。`--lang` フラグと同等 | なし |
| `NO_COLOR` | 設定するとカラー出力を無効化する（[no-color.org](https://no-color.org) 準拠） | なし |
| `LINTERLY_NO_UPDATE_CHECK` | 設定するとバージョン更新チェックを無効化する（`--no-update-check` フラグと同等） | なし |

- `--config` フラグが指定された場合は `LINTERLY_CONFIG` より優先される
- `--lang` フラグが指定された場合は `LINTERLY_LANG` より優先される
- 言語の優先順位: `--lang` フラグ > `LINTERLY_LANG` 環境変数 > 設定ファイルの `language` > デフォルト `en`

## 6. 設定の優先順位（全体）

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
| 更新チェック無効化 | `--no-update-check` | `LINTERLY_NO_UPDATE_CHECK` | `update_check` | 未設定（チェック有効） |

## 改訂履歴

| 版 | 日付 | 変更内容 | 変更理由 |
|---|------|---------|---------|
| 1.0 | 2026-02-08 | 初版作成 | — |
| 1.1 | 2026-02-08 | テキスト出力例のエラー件数を 2 に修正 | JSON 出力例・出力例の ERROR 件数との整合性確保 |
| 1.2 | 2026-02-08 | warning(s) 表記修正、ignore 重複警告を 1 行表記に修正、version 出力例更新、--lang フラグと LINTERLY_LANG 環境変数を追加 | ドキュメント乖離レポート (#3) 対応 |
| 1.3 | 2026-02-24 | check コマンドに設定上書きフラグ（--max-lines-per-file 等6種）を追加、設定ファイルなし実行の対応、優先順位の明記 | #22 CLI フラグによる設定値の上書き対応 |
| 1.4 | 2026-03-03 | 3. バージョン更新チェックセクション追加（出力例・経路検出・無効化）、--no-update-check フラグと LINTERLY_NO_UPDATE_CHECK 環境変数を追加、優先順位表に更新チェック行を追加 | #30 バージョン更新チェック機能 |
| 1.5 | 2026-03-03 | 通知メッセージを i18n 対応に変更、バージョン不明時の出力例を追加、無効化条件からバージョン不明時スキップを削除 | #30 フィードバック反映 |
| 1.6 | 2026-03-03 | 無効化に設定ファイルの `update_check: false` を追加、優先順位表に update_check 列を追加 | #30 設定ファイル対応 |
| 1.7 | 2026-03-07 | `--format` に `junit` / `checkstyle` を追加、JUnit XML / checkstyle XML の出力例・仕様を追加 | #43 JUnit XML / checkstyle XML 出力フォーマット追加 |
| 1.8 | 2026-11-28 | ヘルプ出力例を実出力に合わせて更新し日本語ヘルプ例を削除（ヘルプは英語のみ）、`check` を `[path...]`（ファイル直接指定・複数パス・結果パスの基準位置・読み取り不能ファイルのスキップ警告）に更新、未知の `--format` を使用エラーとして明記、JSON 出力に `warnings` を追加、junit / checkstyle が warnings を出力しない旨を追記、更新チェックのメインコマンド終了後の最大 1 秒待機を追記 | #49, #50, #52, #57, #58, #59, #66 |
