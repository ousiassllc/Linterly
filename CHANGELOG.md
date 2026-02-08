# Changelog

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
