package kingpin

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestBool(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("b", "")
	b := f.Bool()
	fg.init()
	tokens := tokenize([]string{"--b"})
	fg.parse(tokens, false)
	assert.True(t, *b)
}

func TestNoBool(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("b", "").Default("true")
	b := f.Bool()
	fg.init()
	tokens := tokenize([]string{"--no-b"})
	err := fg.parse(tokens, false)
	assert.NoError(t, err)
	assert.False(t, *b)
}

func TestNegateNonBool(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("b", "")
	f.Int()
	fg.init()
	tokens := tokenize([]string{"--no-b"})
	err := fg.parse(tokens, false)
	assert.Error(t, err)
}

func TestInvalidFlagDefaultCanBeOverridden(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("a", "").Default("invalid")
	f.Bool()
	assert.NoError(t, fg.init())
	tokens := tokenize([]string{})
	err := fg.parse(tokens, false)
	assert.Error(t, err)
}

func TestRequiredFlag(t *testing.T) {
	fg := newFlagGroup()
	fg.Flag("a", "").Required().Bool()
	assert.NoError(t, fg.init())
	tokens := tokenize([]string{"--a"})
	err := fg.parse(tokens, false)
	assert.NoError(t, err)
	tokens = tokenize([]string{})
	err = fg.parse(tokens, false)
	assert.Error(t, err)
}
