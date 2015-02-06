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
	err := fg.parse(tokens, false)
	assert.NoError(t, err)
	assert.False(t, *b)
}

func TestNegateNonBool(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("b", "")
	f.Int()
	fg.init()
	tokens := Tokenize([]string{"--no-b"})
	err := fg.parse(tokens, false)
	assert.Error(t, err)
}

func TestInvalidFlagDefaultCanBeOverridden(t *testing.T) {
	fg := newFlagGroup()
	f := fg.Flag("a", "").Default("invalid")
	f.Bool()
	assert.NoError(t, fg.init())
	tokens := Tokenize([]string{})
	err := fg.parse(tokens, false)
	assert.Error(t, err)
}

func TestRequiredFlag(t *testing.T) {
	fg := newFlagGroup()
	fg.Flag("a", "").Required().Bool()
	assert.NoError(t, fg.init())
	tokens := Tokenize([]string{"--a"})
	err := fg.parse(tokens, false)
	assert.NoError(t, err)
	tokens = Tokenize([]string{})
	err = fg.parse(tokens, false)
	assert.Error(t, err)
}

func TestPatternFlag(t *testing.T) {
	fg := newFlagGroup()
	fg.FlagPattern(`a\.\d+\.b`, "").String()
	assert.NoError(t, fg.init())
	tokens := Tokenize([]string{"--a.42.b", "answer"})
	err := fg.parse(tokens, false)
	assert.NoError(t, err)
	tokens = Tokenize([]string{"--a.nan.b", "nan"})
	err = fg.parse(tokens, false)
	assert.Error(t, err)
}

func TestPatternCapture(t *testing.T) {
	fg := newFlagGroup()
	var submatches []string
	fg.FlagPattern(`a\.(\d+)\.b`, "").Capture(&submatches).Strings()
	assert.NoError(t, fg.init())
	tokens := Tokenize([]string{"--a.42.b", "answer", "--a.66.b", "route"})
	err := fg.parse(tokens, false)
	assert.NoError(t, err)
	assert.Equal(t, submatches, []string{"42", "66"}, "Capture should initialize submatches")
}

func TestInvalidPatternFlag(t *testing.T) {
	fg := newFlagGroup()
	fg.FlagPattern(`a[invalid regex`, "").String()
	assert.Error(t, fg.init())
}
