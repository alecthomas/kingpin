package kingpin

//go:generate go run ./cmd/embedi18n/main.go en-AU
//go:generate go run ./cmd/embedi18n/main.go fr

import (
	"os"

	"github.com/nicksnyder/go-i18n/i18n"
)

var T = initI18N()

func initI18N() i18n.TranslateFunc {
	// Initialise translations.
	i18n.ParseTranslationFileBytes("i18n/en-AU.all.json", i18n_en_AU)
	i18n.ParseTranslationFileBytes("i18n/fr.all.json", i18n_fr)

	lang := os.Getenv("LANG")
	t, err := i18n.Tfunc(lang, "en")
	if err != nil {
		panic(err)
	}
	return t
}
