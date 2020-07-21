package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	o := Output{}
	kingpin.FlagsOf(&o)
	kingpin.Parse()

	minifiedPrefix := ""
	if o.Encoding.Minify {
		minifiedPrefix = "minified "
	}
	fmt.Printf("Would send data encoded as %s%s to %s:%d.",
		minifiedPrefix, o.Encoding.Format, o.Connection.Host, o.Connection.Port)
}
