package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	name := os.Args[1]
	r, err := os.Open("i18n/" + name + ".all.json")
	if err != nil {
		panic(err)
	}
	defer r.Close()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	id := strings.Replace(name, "-", "_", -1)
	w, err := os.Create("i18n_" + id + ".go")
	if err != nil {
		panic(err)
	}
	defer w.Close()
	fmt.Fprintf(w, `package kingpin

var i18n_%s = []byte(%q)
`, id, data)
}
