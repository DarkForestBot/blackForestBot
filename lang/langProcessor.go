package lang

import "git.wetofu.top/tonychee7000/blackForestBot/basis"

var lang string

func SetLang(language string) {
	lang = language
}

func T(key string, args ...interface{}) string {
	lang, ok := basis.GlobalLanguageList[lang]
	if !ok {
		return ""
	}
}
