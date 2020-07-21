package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type OutputEncoding struct {
	Format string
	Minify bool
}

func (o *OutputEncoding) RegisterFlags(fg kingpin.FlagGroup) {
	g := fg.FlagGroup("encoding.").
		EnvarNamePrefix("ENCODING_")

	g.Flag("format", "Format to encode the result with.").
		Default("json").
		Envar("FORMAT").
		EnumVar(&o.Format, "json", "yaml")
	g.Flag("minify", "If set the output format will be minified, if possible.").
		Envar("MINIFY").
		BoolVar(&o.Minify)
}
