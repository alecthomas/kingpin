package kingpin

import (
	"net"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/alecthomas/units"
)

var (
	envVarValuesSeparator = "\r?\n"
	envVarValuesTrimmer   = regexp.MustCompile(envVarValuesSeparator + "$")
	envVarValuesSplitter  = regexp.MustCompile(envVarValuesSeparator)
)

type Settings interface {
	SetValue(value Value)
}

type parserMixin struct {
	defaultValues []string
	value         Value
	required      bool
	envar         string
	noEnvar       bool
}

func (p *parserMixin) needsValue() bool {
	haveDefault := len(p.defaultValues) > 0
	return p.required && !(haveDefault || p.HasEnvarValue())
}

func (p *parserMixin) reset() {
	if c, ok := p.value.(cumulativeValue); ok {
		c.Reset()
	}
}

func (p *parserMixin) setDefault() error {
	p.reset()
	if p.HasEnvarValue() {
		if v, ok := p.value.(cumulativeValue); !ok || !v.IsCumulative() {
			// Use the value as-is
			return p.value.Set(p.GetEnvarValue())
		} else {
			for _, value := range p.GetSplitEnvarValue() {
				if err := p.value.Set(value); err != nil {
					return err
				}
			}
			return nil
		}
	}

	if len(p.defaultValues) > 0 {
		for _, defaultValue := range p.defaultValues {
			if err := p.value.Set(defaultValue); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

func (p *parserMixin) HasEnvarValue() bool {
	return p.GetEnvarValue() != ""
}

func (p *parserMixin) GetEnvarValue() string {
	if p.noEnvar || p.envar == "" {
		return ""
	}
	return os.Getenv(p.envar)
}

func (p *parserMixin) GetSplitEnvarValue() []string {
	values := make([]string, 0)

	envarValue := p.GetEnvarValue()
	if envarValue == "" {
		return values
	}

	// Split by new line to extract multiple values, if any.
	trimmed := envVarValuesTrimmer.ReplaceAllString(envarValue, "")
	for _, value := range envVarValuesSplitter.Split(trimmed, -1) {
		values = append(values, value)
	}

	return values
}

func (p *parserMixin) SetValue(value Value) {
	p.value = value
	p.setDefault()
}

// StringMap provides key=value parsing into a map.
func (p *parserMixin) StringMap() (target *map[string]string) {
	target = &(map[string]string{})
	p.StringMapVar(target)
	return
}

// Duration sets the parser to a time.Duration parser.
func (p *parserMixin) Duration() (target *time.Duration) {
	target = new(time.Duration)
	p.DurationVar(target)
	return
}

// Bytes parses numeric byte units. eg. 1.5KB
func (p *parserMixin) Bytes() (target *units.Base2Bytes) {
	target = new(units.Base2Bytes)
	p.BytesVar(target)
	return
}

// IP sets the parser to a net.IP parser.
func (p *parserMixin) IP() (target *net.IP) {
	target = new(net.IP)
	p.IPVar(target)
	return
}

// TCP (host:port) address.
func (p *parserMixin) TCP() (target **net.TCPAddr) {
	target = new(*net.TCPAddr)
	p.TCPVar(target)
	return
}

// TCPVar (host:port) address.
func (p *parserMixin) TCPVar(target **net.TCPAddr) {
	p.SetValue(newTCPAddrValue(target))
}

// ExistingFile sets the parser to one that requires and returns an existing file.
func (p *parserMixin) ExistingFile() (target *string) {
	target = new(string)
	p.ExistingFileVar(target)
	return
}

// ExistingDir sets the parser to one that requires and returns an existing directory.
func (p *parserMixin) ExistingDir() (target *string) {
	target = new(string)
	p.ExistingDirVar(target)
	return
}

// ExistingFileOrDir sets the parser to one that requires and returns an existing file OR directory.
func (p *parserMixin) ExistingFileOrDir() (target *string) {
	target = new(string)
	p.ExistingFileOrDirVar(target)
	return
}

// File returns an os.File against an existing file.
func (p *parserMixin) File() (target **os.File) {
	target = new(*os.File)
	p.FileVar(target)
	return
}

// File attempts to open a File with os.OpenFile(flag, perm).
func (p *parserMixin) OpenFile(flag int, perm os.FileMode) (target **os.File) {
	target = new(*os.File)
	p.OpenFileVar(target, flag, perm)
	return
}

// URL provides a valid, parsed url.URL.
func (p *parserMixin) URL() (target **url.URL) {
	target = new(*url.URL)
	p.URLVar(target)
	return
}

// StringMap provides key=value parsing into a map.
func (p *parserMixin) StringMapVar(target *map[string]string) {
	p.SetValue(newStringMapValue(target))
}

// Float sets the parser to a float64 parser.
func (p *parserMixin) Float() (target *float64) {
	return p.Float64()
}

// Float sets the parser to a float64 parser.
func (p *parserMixin) FloatVar(target *float64) {
	p.Float64Var(target)
}

// Duration sets the parser to a time.Duration parser.
func (p *parserMixin) DurationVar(target *time.Duration) {
	p.SetValue(newDurationValue(target))
}

// BytesVar parses numeric byte units. eg. 1.5KB
func (p *parserMixin) BytesVar(target *units.Base2Bytes) {
	p.SetValue(newBytesValue(target))
}

// IP sets the parser to a net.IP parser.
func (p *parserMixin) IPVar(target *net.IP) {
	p.SetValue(newIPValue(target))
}

// ExistingFile sets the parser to one that requires and returns an existing file.
func (p *parserMixin) ExistingFileVar(target *string) {
	p.SetValue(newExistingFileValue(target))
}

// ExistingDir sets the parser to one that requires and returns an existing directory.
func (p *parserMixin) ExistingDirVar(target *string) {
	p.SetValue(newExistingDirValue(target))
}

// ExistingDir sets the parser to one that requires and returns an existing directory.
func (p *parserMixin) ExistingFileOrDirVar(target *string) {
	p.SetValue(newExistingFileOrDirValue(target))
}

// FileVar opens an existing file.
func (p *parserMixin) FileVar(target **os.File) {
	p.SetValue(newFileValue(target, os.O_RDONLY, 0))
}

// OpenFileVar calls os.OpenFile(flag, perm)
func (p *parserMixin) OpenFileVar(target **os.File, flag int, perm os.FileMode) {
	p.SetValue(newFileValue(target, flag, perm))
}

// URL provides a valid, parsed url.URL.
func (p *parserMixin) URLVar(target **url.URL) {
	p.SetValue(newURLValue(target))
}

// URLList provides a parsed list of url.URL values.
func (p *parserMixin) URLList() (target *[]*url.URL) {
	target = new([]*url.URL)
	p.URLListVar(target)
	return
}

// URLListVar provides a parsed list of url.URL values.
func (p *parserMixin) URLListVar(target *[]*url.URL) {
	p.SetValue(newURLListValue(target))
}

// Enum allows a value from a set of options.
func (p *parserMixin) Enum(options ...string) (target *string) {
	target = new(string)
	p.EnumVar(target, options...)
	return
}

// EnumVar allows a value from a set of options.
func (p *parserMixin) EnumVar(target *string, options ...string) {
	p.SetValue(newEnumFlag(target, options...))
}

// Enums allows a set of values from a set of options.
func (p *parserMixin) Enums(options ...string) (target *[]string) {
	target = new([]string)
	p.EnumsVar(target, options...)
	return
}

// EnumVar allows a value from a set of options.
func (p *parserMixin) EnumsVar(target *[]string, options ...string) {
	p.SetValue(newEnumsFlag(target, options...))
}

// A Counter increments a number each time it is encountered.
func (p *parserMixin) Counter() (target *int) {
	target = new(int)
	p.CounterVar(target)
	return
}

func (p *parserMixin) CounterVar(target *int) {
	p.SetValue(newCounterValue(target))
}
