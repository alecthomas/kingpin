package kingpin

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveWithBuiltin(t *testing.T) {
	a := completionsMixin{}
	a.builtinCompletion.Files = false
	a.builtinCompletion.Directories = true
	a.builtinCompletion.addWords("opt1", "opt2")

	args := a.resolveCompletion()
	assert.Equal(t, a.builtinCompletion, args)
}

func TestResolveWithUser(t *testing.T) {
	a := completionsMixin{}
	a.userCompletion.Files = true
	a.userCompletion.Directories = false
	a.userCompletion.addWords("opt3", "opt4")

	args := a.resolveCompletion()
	assert.Equal(t, a.userCompletion, args)
}

func TestResolveWithCombination(t *testing.T) {
	a := completionsMixin{}

	a.builtinCompletion.Files = false
	a.builtinCompletion.Directories = true
	a.builtinCompletion.addWords("opt1", "opt2")

	a.userCompletion.Files = true
	a.userCompletion.Directories = false
	a.userCompletion.addWords("opt3", "opt4")

	args := a.resolveCompletion()
	// User provided args take preference over builtin (enum-defined) args.
	assert.Equal(t, a.userCompletion, args)
}

func TestAddUserCompletionWords(t *testing.T) {
	a := completionsMixin{}
	a.userCompletion.addWords("opt1", "opt2")

	args := a.resolveCompletion()
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())

	a.userCompletion.addWords("opt3", "opt4")
	args = a.resolveCompletion()
	assert.Equal(t, []string{"opt1", "opt2", "opt3", "opt4"}, args.resolveWords())
}

func TestCompletionSetFileOrDirectory(t *testing.T) {
	a := completionsMixin{}
	a.builtinCompletion.Directories = true
	a.builtinCompletion.Files = false

	args := a.resolveCompletion()
	assert.True(t, args.Directories)
	assert.False(t, args.Files)

	a.builtinCompletion.Directories = false
	a.builtinCompletion.Files = true

	args = a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.True(t, args.Files)
}

func TestAddWords(t *testing.T) {
	a := completionsMixin{}
	a.builtinCompletion = Completion{
		WordActions: []HintAction{func() []string { return []string{"opt1", "opt2"} }}}
	a.builtinCompletion.addWords("opt3", "opt4")

	args := a.resolveCompletion()
	assert.Equal(t, []string{"opt1", "opt2", "opt3", "opt4"}, args.resolveWords())
}

func TestMergeCompletions(t *testing.T) {
	a := Completion{
		Directories: false,
		Files:       false,
		WordActions: []HintAction{func() []string { return []string{"opt1"} }}}

	b := Completion{
		Directories: true,
		Files:       false,
		WordActions: []HintAction{func() []string { return []string{"opt2"} }}}

	c := Completion{
		Directories: false,
		Files:       true,
		WordActions: []HintAction{func() []string { return []string{"opt3"} }}}

	d := mergeCompletions(a, b)
	assert.True(t, d.Directories)
	assert.False(t, d.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, d.resolveWords())

	e := mergeCompletions(d, c)
	assert.True(t, e.Directories)
	assert.True(t, e.Files)
	assert.Equal(t, []string{"opt1", "opt2", "opt3"}, e.resolveWords())
}

func TestGenerateBashString(t *testing.T) {
	ifs := os.Getenv("IFS")
	defer os.Setenv("IFS", ifs)

	os.Setenv("IFS", ",")
	a := Completion{
		Directories: true,
		Files:       false,
		WordActions: []HintAction{func() []string { return []string{"opt1", "opt2"} }}}

	assert.Equal(t, "-d -W opt1,opt2", a.generateBashString())
}
