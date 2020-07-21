package main

import "gopkg.in/alecthomas/kingpin.v2"

type Output struct {
	Encoding   OutputEncoding
	Connection OutputConnection
}

func (s *Output) RegisterFlags(fg kingpin.FlagGroup) {
	fg.FlagGroup("output.").
		EnvarNamePrefix("OUTPUT_").
		FlagsOf(&s.Encoding, &s.Connection)
}
