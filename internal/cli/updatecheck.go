package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ousiassllc/linterly/internal/config"
	"github.com/ousiassllc/linterly/internal/i18n"
	"github.com/ousiassllc/linterly/internal/updatecheck"
)

// updateWaitTimeout は、メインコマンド終了後に更新チェックの完了を待つ上限時間。
// 短時間で終わるコマンドでも API 応答をキャッシュへ保存し通知を出せるよう、
// 有限の時間だけ待つ（httpTimeout の 3 秒を超えない）。
// キャッシュが有効な場合はチェックが即座に完了するため待ちは発生しない。
var updateWaitTimeout = 1 * time.Second

// updateResult は非同期チェック結果を受け取るチャネル。
var updateResult chan *updatecheck.CheckResult

// runUpdateCheck は更新チェックを実行する。テストで差し替え可能にしている。
var runUpdateCheck = func(ctx context.Context, lang string) (*updatecheck.CheckResult, error) {
	translator, err := i18n.New(lang)
	if err != nil {
		return nil, err
	}
	return updatecheck.NewChecker(Version, translator).Check(ctx)
}

// startUpdateCheck はバックグラウンドで更新チェックを開始する。
func startUpdateCheck() {
	if flagNoUpdateCheck || os.Getenv("LINTERLY_NO_UPDATE_CHECK") != "" {
		return
	}

	configDisabled, configLang := readUpdateCheckConfig()
	if configDisabled {
		return
	}

	lang := resolveUpdateCheckLang(configLang)

	ch := make(chan *updatecheck.CheckResult, 1)
	updateResult = ch

	run := runUpdateCheck
	go func() {
		result, err := run(context.Background(), lang)
		if err != nil {
			ch <- nil
			return
		}
		ch <- result
	}()
}

// printUpdateNotice は更新チェック結果を stderr に表示する。
// 未完了の場合は updateWaitTimeout まで待ち、それでも完了しなければスキップする。
func printUpdateNotice() {
	if updateResult == nil {
		return
	}
	timer := time.NewTimer(updateWaitTimeout)
	defer timer.Stop()
	select {
	case result := <-updateResult:
		if result != nil && (result.UpdateAvailable || result.VersionUnknown) {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, result.Message)
		}
	case <-timer.C:
		// 上限時間内に完了しなかった場合はメインコマンドの終了を優先する
	}
}

// readUpdateCheckConfig は設定ファイルから update_check と language を軽量に読み取る。
// 設定ファイルの探索は config.Load と同じ規則（config.LoadUpdateSettings 経由）。
func readUpdateCheckConfig() (disabled bool, lang string) {
	updateCheck, lang := config.LoadUpdateSettings(configFile)
	return !updateCheck, lang
}

// resolveUpdateCheckLang は更新チェック用の言語を決定する。
// configLang は readUpdateCheckConfig から取得した設定ファイルの language 値。
func resolveUpdateCheckLang(configLang string) string {
	lang := i18n.ResolveLanguage(langFlag)
	if langFlag != "" || os.Getenv("LINTERLY_LANG") != "" {
		return lang
	}
	if i18n.IsSupportedLanguage(configLang) {
		return configLang
	}
	return lang
}
