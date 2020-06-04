package kingpin

import (
	"strings"
)

func Fuzz(data []byte) int {
	app := New("test", "")
	_, err := app.Parse(strings.Split(string(data), " "))
	if err != nil {
		return 0
	}
	return 1
}
