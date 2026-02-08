BINARY_NAME := linterly
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/ousiassllc/linterly/internal/cli.Version=$(VERSION)"
GORELEASER := $(shell command -v goreleaser 2>/dev/null || echo $(shell go env GOPATH)/bin/goreleaser)

.DEFAULT_GOAL := help

.PHONY: help build run test test-v cover lint fmt clean setup-hooks release release-check release-dry-run

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## バイナリをビルド
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/linterly

run: build ## ビルドして実行（引数: ARGS="check ."）
	./bin/$(BINARY_NAME) $(ARGS)

test: ## テストを実行
	go test ./...

test-v: ## テストを詳細表示で実行
	go test ./... -v

cover: ## カバレッジレポートを表示
	go test ./... -cover

lint: ## golangci-lint を実行
	golangci-lint run

fmt: ## コードをフォーマット
	gofmt -w .

clean: ## ビルド成果物を削除
	rm -rf bin/ dist/

release: ## GoReleaser でリリース
	$(GORELEASER) release --clean

release-check: ## GoReleaser の設定を検証
	$(GORELEASER) check

release-dry-run: ## GoReleaser でスナップショットビルド（ドライラン）
	$(GORELEASER) release --snapshot --clean

setup-hooks: ## lefthook で Git Hooks をインストール
	lefthook install
