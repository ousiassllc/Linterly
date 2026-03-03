package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/ousiassllc/linterly/internal/config"
	"github.com/ousiassllc/linterly/internal/i18n"
	"github.com/ousiassllc/linterly/internal/updatecheck"
)

// updateResult は非同期チェック結果を受け取るチャネル。
var updateResult chan *updatecheck.CheckResult

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

	go func() {
		translator, err := i18n.New(lang)
		if err != nil {
			ch <- nil
			return
		}
		checker := updatecheck.NewChecker(Version, translator)
		result, err := checker.Check(context.Background())
		if err != nil {
			ch <- nil
			return
		}
		ch <- result
	}()
}

// printUpdateNotice は更新チェック結果を stderr に表示する（非ブロッキング）。
func printUpdateNotice() {
	if updateResult == nil {
		return
	}
	select {
	case result := <-updateResult:
		if result != nil && (result.UpdateAvailable || result.VersionUnknown) {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, result.Message)
		}
	default:
		// バックグラウンドチェック未完了の場合はブロックしない
	}
}

// readUpdateCheckConfig は設定ファイルから update_check と language を軽量に読み取る。
func readUpdateCheckConfig() (disabled bool, lang string) {
	v := viper.New()

	cfgPath := configFile
	if cfgPath == "" {
		cfgPath = os.Getenv("LINTERLY_CONFIG")
	}

	if cfgPath != "" {
		v.SetConfigFile(cfgPath)
	} else {
		for _, name := range config.DefaultConfigFileNames {
			if _, err := os.Stat(name); err == nil {
				v.SetConfigFile(name)
				break
			}
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return false, ""
	}

	if v.IsSet("update_check") && !v.GetBool("update_check") {
		disabled = true
	}
	if v.IsSet("language") {
		lang = v.GetString("language")
	}
	return
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
