package main

import "gopkg.in/alecthomas/kingpin.v2"

type Show struct {
	ShowFiles
	ShowVariables
}

func (s *Show) RegisterCommands(c kingpin.Cmd) {
	c.Command("show", "show some stuff").
		EnvarNamePrefix("SHOW_").
		CmdsOf(&s.ShowFiles, &s.ShowVariables)
}
