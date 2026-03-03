# Changelog

## [v0.3.3] - 2026-03-03

### 🐛 Bug Fixes / バグ修正
- `go install` 時に `debug.ReadBuildInfo` でバージョンをフォールバック取得するよう修正 / Fix version fallback to use `debug.ReadBuildInfo` for `go install` builds (#38)

## [v0.3.2] - 2026-03-03

### ✨ New Features / 新機能
- バージョン更新チェック機能を追加（非同期チェック・キャッシュ・--no-update-check フラグ） / Add version update check with async checking, cache, and --no-update-check flag (#30)
- バイナリファイルをスキャン対象から自動除外 / Automatically skip binary files from scan targets (#29)

### 🐛 Bug Fixes / バグ修正
- サブディレクトリ指定時の ignore パターン不一致を修正 / Fix ignore pattern mismatch when targeting subdirectories (#27)
- ignore 重複警告メッセージを i18n 対応 / Add i18n support for ignore duplicate warning messages (#33)
- version コマンド出力に v プレフィックスを付与 / Add v prefix to version command output (#33)
- CSS の行コメント `//` を非標準として除外 / Remove non-standard `//` line comment from CSS definition (#35)
- findAndReadConfig のエラーを ConfigError にラップして i18n 対応 / Wrap findAndReadConfig errors in ConfigError for i18n support (#35)

### 🔧 Improvements / 改善
- デフォルト値と設定ファイル名を定数化して重複除去 / Extract default values and config file names into constants (#35)
- サポート言語判定を i18n.IsSupportedLanguage に統一 / Unify supported language validation via i18n.IsSupportedLanguage (#35)
- init コマンドで cobra の I/O メソッドを使用しテスタビリティ向上 / Use cobra I/O methods in init command for better testability (#35)
- IgnorePatterns にキャッシュを導入し二重呼び出しを解消 / Add cache to IgnorePatterns to eliminate duplicate file I/O (#35)
- countAll/countCodeOnly の引数を io.Reader に変更 / Change countAll/countCodeOnly parameter from *os.File to io.Reader (#35)
- TextReporter の NO_COLOR 判定を初期化時に移動 / Move NO_COLOR check to TextReporter initialization (#35)
- JSONReporter に warnings フィールドを追加 / Add warnings field to JSONReporter output (#35)
- scanner のデッドコード（ルートディレクトリ補完ブロック）を削除 / Remove dead code for root directory补completion in scanner (#35)

### 📝 Documentation / ドキュメント
- ドキュメント乖離を修正（Config/Checker 定義・配布形態・設定ファイル） / Fix documentation drift for Config/Checker definitions, distribution, and config files (#34)
- バージョン更新チェック機能の設計ドキュメントを追加 / Add design documentation for version update check feature (#30)
- バイナリファイル自動スキップ機能のドキュメントを追加 / Add documentation for binary file auto-skip feature (#29)

### 🏗️ Infrastructure / インフラ
- .gitignore に .env / .husky/ を追加 / Add .env and .husky/ to .gitignore (#35)
- lefthook の pre-commit にテスト実行を追加 / Add test execution to lefthook pre-commit (#35)
- CI の go test に -count=1 を追加しキャッシュ無効化 / Add -count=1 to CI go test to disable caching (#35)

## [v0.3.1] - 2026-02-28

### 📝 ドキュメント / Documentation
- README に CLI フラグ上書き・設定ファイル不要の使用例を追加 / Add CLI flag override and config-free usage examples to README

## [v0.3.0] - 2026-02-28

CLI フラグによる設定値の上書きと設定ファイルなしでの実行に対応。
Add config override via CLI flags and support running without a config file.

### ✨ 新機能 / New Features
- check コマンドに設定上書きフラグ 6 種を追加（--max-lines-per-file, --max-lines-per-func, --count-mode, --threshold, --exclude, --include） / Add 6 config override flags to check command (#22)
- Overrides 型と ApplyOverrides メソッドを追加し、設定ファイルなしでの実行に対応 / Add Overrides type and ApplyOverrides method to support running without config file (#22)

### 🐛 バグ修正 / Bug Fixes
- スキャナーバッファを 1MB に拡張し、エラーメッセージにファイルパスを付与 / Expand scanner buffer to 1MB and include file path in error messages (#24)

### 🔧 改善 / Improvements
- テストファイルを 300 行以内に分割し、.linterlyignore に除外パターンを追加 / Split test files under 300 lines and add ignore patterns (#25)
- lefthook の pre-commit/pre-push フック構成を整理 / Reorganize lefthook pre-commit/pre-push hook configuration

### 📝 ドキュメント / Documentation
- CLI フラグによる設定値上書きと設定ファイルなし実行の仕様を追加 / Add spec for CLI flag overrides and config-free execution (#22)

## [v0.2.2] - 2026-02-23

### 📝 ドキュメント / Documentation
- README を英語化し日本語版を README.ja.md に分離 / Split README into English (README.md) and Japanese (README.ja.md)
- npm パッケージリンクを追加し未実装の Action 参照を削除 / Add npm package links and remove unimplemented Action references

## [v0.2.0] - 2026-02-23

GoReleaser による自動リリースと npm パッケージ配布に対応。
Add automated releases via GoReleaser and npm package distribution.

### ✨ 新機能 / New Features
- GoReleaser を導入しクロスコンパイル・リリースを自動化 / Introduce GoReleaser for cross-compilation and automated releases
- npm パッケージ配布用の構成を追加 / Add npm package distribution setup
- リリースワークフローを追加し GoReleaser 設定を調整 / Add release workflow and adjust GoReleaser config

### 🐛 バグ修正 / Bug Fixes
- GoReleaser の出力先を build/ に変更し dist/npm/ との競合を解消 / Change GoReleaser output to build/ to avoid conflict with dist/npm/
- Makefile の goreleaser パス解決と release-check ターゲット追加 / Fix goreleaser path resolution in Makefile and add release-check target

### 🔧 改善 / Improvements
- Go バージョンを 1.25.6 から 1.26 に更新 / Update Go version from 1.25.6 to 1.26

### 📝 ドキュメント / Documentation
- インストール手順の修正と Git Hooks 連携セクションを追加 / Fix install instructions and add Git Hooks integration section
- release ターゲットに GITHUB_TOKEN が必要な旨を明記 / Document GITHUB_TOKEN requirement for release target

## [v0.1.0] - 2026-02-08

初回リリース。コード行数チェック CLI ツール「Linterly」の基本機能を実装。

### ✨ 新機能
- Go プロジェクトを初期化（go.mod, main.go, CLI 骨格）
- i18n コンポーネントを実装（英語・日本語メッセージ管理）
- config コンポーネントを実装（設定読み込み・バリデーション・ignore 優先ルール）
- scanner コンポーネントを実装（ファイル走査・除外フィルタ・gitignore パターンマッチ）
- counter コンポーネントを実装（行数カウント・言語検出・コメント認識）
- analyzer コンポーネントを実装（ルール評価・閾値判定・ディレクトリ集計）
- reporter コンポーネントを実装（テキスト/JSON 出力・カラー対応・i18n 連携）
- CLI コンポーネントを実装（check・init・version コマンド）(#2)
- 終了コード 2（実行エラー）を実装
- init コマンドの i18n 対応と --lang フラグの追加
- config.Load のエラーメッセージを i18n 対応 (#4)
- Makefile を追加（build・test・cover・lint・fmt・clean）
- プロジェクト自身の .linterly.yml と .linterlyignore を追加
- GitHub Actions CI ワークフローを追加（lint, test, build）
- golangci-lint を導入し Makefile と CI に組み込み
- lefthook で pre-commit/pre-push フックを導入 (#6)

### 🐛 バグ修正
- .gitignore に bin/ を追加
- countAll/countCodeOnly に bufio.Scanner.Err() チェックを追加
- init.go の ReadString エラーを処理し EOF 時に安全に終了 (#4)

### 🔧 改善
- max_lines_per_file のデフォルト値を 400 から 300 に変更
- マジックストリングを定数化（CountMode, Format）
- i18n 初期化パターンを initTranslator ヘルパーに共通化
- analyzer.go の冗長な no-op 条件分岐を削除
- tools.go を削除（go-gitignore は scanner で直接使用）
- husky を lefthook に置換し Node.js 依存を排除 (#6)

### 📝 ドキュメント
- README を作成
- 機能要件・非機能要件ドキュメントを作成
- アーキテクチャ設計・設定ファイルスキーマドキュメントを作成
- CLI インターフェース仕様・コンポーネント設計ドキュメントを作成
- 仕様書を実装に合わせて更新 (#4)

### 🧪 テスト
- CLI パッケージのテストカバレッジを向上（version, JSON 出力, 言語切替）(#6)
