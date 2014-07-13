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
	fg.parse(tokens, false)
	assert.True(t, *b)
}

func TestNoBool(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("b", "").Default("true")
	b := f.Bool()
	fg.init()
	tokens := Tokenize([]string{"--no-b"})
	_, err := fg.parse(tokens, false)
	assert.NoError(t, err)
	assert.False(t, *b)
}

func TestNegateNonBool(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("b", "")
	f.Int()
	fg.init()
	tokens := Tokenize([]string{"--no-b"})
	_, err := fg.parse(tokens, false)
	assert.Error(t, err)
}

func TestInvalidFlagDefaultCanBeOverridden(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("a", "").Default("invalid")
	f.Bool()
	assert.NoError(t, fg.init())
	tokens := Tokenize([]string{})
	_, err := fg.parse(tokens, false)
	assert.Error(t, err)
}

func TestRequiredFlag(t *testing.T) {
	fg := newFlagGroup()
	fg.Flag("a", "").Required().Bool()
	assert.NoError(t, fg.init())
	tokens := Tokenize([]string{"--a"})
	tokens, err := fg.parse(tokens, false)
	assert.NoError(t, err)
	tokens = Tokenize([]string{})
	tokens, err = fg.parse(tokens, false)
	assert.Error(t, err)
}
