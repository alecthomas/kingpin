package kingpin

import (
	"github.com/stretchrcom/testify/assert"

	"testing"
)

func TestBool(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("b", "")
	b := f.Bool()
	fg.init()
	tokens := Tokenize([]string{"--b"})
	fg.parse(tokens, true)
	assert.True(t, *b)
}

// func TestNoBool(t *testing.T) {
// 	fg := newFlagGroup()
// 	f := fg.Flag("b", "").Default("true")
// 	b := f.Bool()
// 	fg.init()
// 	tokens := Tokenize([]string{"--no-b"})
// 	fg.parse(tokens, false)
// 	assert.False(t, *b)
// }
