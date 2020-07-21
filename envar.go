package kingpin

import (
	"os"
	"regexp"
)

var (
	envVarValuesSeparator = "\r?\n"
	envVarValuesTrimmer   = regexp.MustCompile(envVarValuesSeparator + "$")
	envVarValuesSplitter  = regexp.MustCompile(envVarValuesSeparator)
)

type envarMixin struct {
	nameResolver envarNameResolver
	envar        string
	noEnvar      bool
}

type envarNameResolver func(string) string

func (e *envarMixin) HasEnvarValue() bool {
	return e.GetEnvarValue() != ""
}

func (e *envarMixin) GetEnvarValue() string {
	if n := e.getEnvar(); n != "" {
		return os.Getenv(n)
	}
	return ""
}

func (e *envarMixin) getEnvar() string {
	if e.noEnvar || e.envar == "" {
		return ""
	}
	if r := e.nameResolver; r != nil {
		return r(e.envar)
	}
	return e.envar
}

func (e *envarMixin) GetSplitEnvarValue() []string {
	envarValue := e.GetEnvarValue()
	if envarValue == "" {
		return []string{}
	}

	// Split by new line to extract multiple values, if any.
	trimmed := envVarValuesTrimmer.ReplaceAllString(envarValue, "")

	return envVarValuesSplitter.Split(trimmed, -1)
}
