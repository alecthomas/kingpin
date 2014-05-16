package kingpin

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	// CommandLine is the default Kingpin parser.
	CommandLine = New(filepath.Base(os.Args[0]), "")
)

// Command adds a new command to the default parser.
func Command(name, help string) *CmdClause {
	return CommandLine.Command(name, help)
}

// Flag adds a new flag to the default parser.
func Flag(name, help string) *FlagClause {
	return CommandLine.Flag(name, help)
}

// Arg adds a new argument to the top-level of the default parser.
func Arg(name, help string) *ArgClause {
	return CommandLine.Arg(name, help)
}

// Parse and return the selected command. Will exit with a non-zero status if
// an error was encountered.
func Parse() string {
	selected, err := CommandLine.Parse(os.Args[1:])
	if err != nil {
		Fatalf("%s", err)
	}
	if selected == "" && len(CommandLine.commands) > 0 {
		Usage()
		os.Exit(0)
	}
	return selected
}

// Fatalf prints an error message to stderr and exits.
func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

// FatalIfError prints an error and exits, if err is not nil. The error is printed
// with the given prefix.
func FatalIfError(err error, prefix string) {
	if err != nil {
		Fatalf(prefix+": %s", err)
	}
}

// UsageErrorf prints an error message followed by usage information, then
// exits with a non-zero status.
func UsageErrorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	Usage()
	os.Exit(1)
}

// Usage prints usage to stderr.
func Usage() {
	CommandLine.Usage(os.Stderr)
}
