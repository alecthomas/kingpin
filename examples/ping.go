package main

import (
	"fmt"

	"github.com/alecthomas/kingpin"
)

var (
	debug   = kingpin.Flag("debug", "Enable debug mode.").Bool()
	timeout = kingpin.Flag("timeout", "Timeout waiting for ping.").Required().Short('t').Duration()
	ip      = kingpin.Arg("ip", "IP address to ping.").Required().IP()
	moo     = kingpin.Arg("moo", "moo").String()
)

func main() {
	kingpin.Parse()
	fmt.Printf("Would ping: %s with timeout %s", *ip, *timeout)
}
