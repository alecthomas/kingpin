package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
)

type ShowFiles struct {
	Root string
}

func (s *ShowFiles) RegisterCommands(c kingpin.Cmd) {
	cmd := c.Command("files", "show some files").
		EnvarNamePrefix("FILES_").
		Action(s.action)

	cmd.Flag("root", "where to start to show files of").
		Default(".").
		Envar("ROOT").
		ExistingDirVar(&s.Root)
}

func (s *ShowFiles) action(*kingpin.ParseContext) error {
	files, err := ioutil.ReadDir(s.Root)
	if err != nil {
		return err
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}
	return nil
}
