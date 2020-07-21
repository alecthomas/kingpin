package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"regexp"
	"strings"
)

type ShowVariables struct {
	Filter *regexp.Regexp
}

func (s *ShowVariables) RegisterCommands(c kingpin.Cmd) {
	cmd := c.Command("variables", "show some environment variables").
		EnvarNamePrefix("VARIABLES_").
		Action(s.action)

	cmd.Flag("filter", "name which should match this pattern").
		Default(".*").
		Envar("FILTER").
		RegexpVar(&s.Filter)
}

func (s *ShowVariables) action(*kingpin.ParseContext) error {
	for _, keyToValue := range os.Environ() {
		keyAndValue := strings.SplitN(keyToValue, "=", 2)
		if s.Filter.MatchString(keyAndValue[0]) {
			fmt.Println(keyToValue)
		}
	}
	return nil
}
