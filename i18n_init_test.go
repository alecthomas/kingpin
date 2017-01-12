package kingpin

import (
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestI18N_fr(t *testing.T) {
	f, err := i18n.Tfunc("fr")
	require.NoError(t, err)
	require.Equal(t,
		"Afficher l'aide contextuelle (essayez aussi --help-long et --help-man).",
		f("Show context-sensitive help (also try --help-long and --help-man)."),
	)
}
