package kingpin

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Completion struct {
	Directories bool
	Files       bool
	WordActions []HintAction
}

// HintAction is a function type who is expected to return a slice of possible
// wordlist command line arguments.
type HintAction func() []string

type completionsMixin struct {
	userCompletion    Completion
	builtinCompletion Completion
}

func (c *Completion) addWords(words ...string) {
	c.WordActions = append(c.WordActions, func() []string { return words })
}

func (c *Completion) empty() bool {
	return !c.Directories && !c.Files && len(c.WordActions) == 0
}

func (c *Completion) generateBashString() string {
	if c.empty() {
		return ""
	}

	result := &bytes.Buffer{}

	if c.Directories {
		fmt.Fprint(result, "-d ")
	}

	if c.Files {
		fmt.Fprintf(result, "-f ")
	}

	if len(c.WordActions) > 0 {
		fmt.Fprint(result, "-W ")

		ifs := os.Getenv("IFS")
		if ifs == "" || strings.ContainsRune(ifs, '\n') {
			ifs = "\n"
		} else {
			ifs = ifs[:1]
		}

		fmt.Fprint(result, strings.Join(c.resolveWords(), ifs))
	}

	return result.String()
}

func mergeCompletions(c1, c2 Completion) Completion {
	directories := c1.Directories || c2.Directories
	files := c1.Files || c2.Files
	wordActions := append(c1.WordActions, c2.WordActions...)

	return Completion{Directories: directories, Files: files, WordActions: wordActions}
}

func (a *completionsMixin) resolveCompletion() Completion {
	if !a.userCompletion.empty() {
		return a.userCompletion
	}
	return a.builtinCompletion
}

func (c Completion) resolveWords() []string {
	result := []string{}
	for _, wordAction := range c.WordActions {
		result = append(result, wordAction()...)
	}
	sort.Strings(result)
	return result
}
