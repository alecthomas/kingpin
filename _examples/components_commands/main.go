package main

import "gopkg.in/alecthomas/kingpin.v2"

func main() {
	kingpin.CmdsOf(&Info{}, &Show{})
	kingpin.Parse()
}
