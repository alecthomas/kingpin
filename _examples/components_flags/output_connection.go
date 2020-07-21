package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type OutputConnection struct {
	Host string
	Port uint16
}

func (o *OutputConnection) RegisterFlags(fg kingpin.FlagGroup) {
	g := fg.FlagGroup("connection.").
		EnvarNamePrefix("CONNECTION_")

	g.Flag("host", "Host to send the output data to.").
		Required().
		Envar("HOST").
		StringVar(&o.Host)
	g.Flag("port", "Port of the host to send the output data to.").
		Default("12345").
		Envar("PORT").
		Uint16Var(&o.Port)
}
