package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Info struct {
	SomeFlag string
}

func (i *Info) RegisterCommands(c kingpin.Cmd) {
	cmd := c.Command("info", "information about me").
		EnvarNamePrefix("INFO_").
		Action(i.action)

	cmd.Flag("some-flag", "just some flag").
		Default("foobar").
		Envar("SOME_FLAG").
		StringVar(&i.SomeFlag)
}

func (i *Info) action(*kingpin.ParseContext) error {
	fmt.Println("This is me.", i.SomeFlag)
	return nil
}
