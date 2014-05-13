package kingpin

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Parser func(string) error

type parserMixin struct {
	parser   Parser
	required bool
}

// String sets the parser to a string parser.
func (p *parserMixin) String() (target *string) {
	return StringParser(p)
}

// Strings appends multiple occurrences to a string slice.
func (p *parserMixin) Strings() (target []string) {
	return StringsParser(p)
}

// StringMap provides key=value parsing into a map.
func (p *parserMixin) StringMap() (target map[string]string) {
	return StringMapParser(p)
}

// Bool sets the parser to a boolean parser.
func (p *parserMixin) Bool() (target *bool) {
	return BoolParser(p)
}

// Int sets the parser to an int parser.
func (p *parserMixin) Int() (target *int) {
	return IntParser(p)
}

// Float sets the parser to a float64 parser.
func (p *parserMixin) Float() (target *float64) {
	return FloatParser(p)
}

// Duration sets the parser to a time.Duration parser.
func (p *parserMixin) Duration() (target *time.Duration) {
	return DurationParser(p)
}

// IP sets the parser to a net.IP parser.
func (p *parserMixin) IP() (target *net.IP) {
	return IPParser(p)
}

// ExistingFile sets the parser to one that requires and returns an existing file.
func (p *parserMixin) ExistingFile() (target *string) {
	return ExistingFileParser(p)
}

// ExistingDir sets the parser to one that requires and returns an existing directory.
func (p *parserMixin) ExistingDir() (target *string) {
	return ExistingDirParser(p)
}

// File sets the parser to one that requires and opens a valid os.File.
func (p *parserMixin) File() (target **os.File) {
	return FileParser(p)
}

// URL provides a valid, parsed url.URL.
func (p *parserMixin) URL() (target **url.URL) {
	return URLParser(p)
}

// SetParser sets the parser used to convert strings to values. Typically used
// by a Parser function implementation.
func (p *parserMixin) SetParser(parser Parser) {
	p.parser = parser
}

type Settings interface {
	SetParser(parser Parser)
}

type FlagSettings interface {
	SetIsBoolean()
}

func StringParser(s Settings) (target *string) {
	target = new(string)
	s.SetParser(func(value string) error {
		*target = value
		return nil
	})
	return
}

func StringsParser(s Settings) (target []string) {
	s.SetParser(func(value string) error {
		target = append(target, value)
		return nil
	})
	return
}

func StringMapParser(s Settings) (target map[string]string) {
	target = make(map[string]string)
	s.SetParser(func(value string) error {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected KEY=VALUE got '%s'", value)
		}
		target[parts[0]] = parts[1]
		return nil
	})
	return
}

func BoolParser(s Settings) (target *bool) {
	target = new(bool)
	if s, ok := s.(FlagSettings); ok {
		s.SetIsBoolean()
	}
	s.SetParser(func(value string) error {
		if value == "0" || value == "false" {
			*target = false
		} else {
			*target = true
		}
		return nil
	})
	return
}

func IntParser(s Settings) (target *int) {
	target = new(int)
	s.SetParser(func(value string) error {
		if out, err := strconv.ParseInt(value, 10, 64); err != nil {
			return err
		} else {
			*target = int(out)
			return nil
		}
	})
	return
}

func FloatParser(s Settings) (target *float64) {
	target = new(float64)
	s.SetParser(func(value string) error {
		if out, err := strconv.ParseFloat(value, 64); err != nil {
			return err
		} else {
			*target = out
			return nil
		}
	})
	return
}

func DurationParser(s Settings) (target *time.Duration) {
	target = new(time.Duration)
	s.SetParser(func(value string) error {
		d, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*target = d
		return nil
	})
	return
}

func IPParser(s Settings) (target *net.IP) {
	target = new(net.IP)
	s.SetParser(func(value string) error {
		if ip := net.ParseIP(value); ip == nil {
			return fmt.Errorf("'%s' is not an IP address", value)
		} else {
			*target = ip
			return nil
		}
	})
	return
}

func ExistingFileParser(s Settings) (target *string) {
	target = new(string)
	s.SetParser(func(value string) error {
		if s, err := os.Stat(value); err != nil || s.IsDir() {
			return fmt.Errorf("'%s' is a directory", value)
		}
		*target = value
		return nil
	})
	return
}

func ExistingDirParser(s Settings) (target *string) {
	target = new(string)
	s.SetParser(func(value string) error {
		if s, err := os.Stat(value); err != nil || !s.IsDir() {
			return fmt.Errorf("'%s' is not a valid directory", value)
		}
		*target = value
		return nil
	})
	return
}

func FileParser(s Settings) (target **os.File) {
	target = new(*os.File)
	s.SetParser(func(value string) error {
		f, err := os.Open(value)
		if err != nil {
			return err
		}
		*target = f
		return nil
	})
	return
}

func URLParser(s Settings) (target **url.URL) {
	target = new(*url.URL)
	s.SetParser(func(value string) error {
		if u, err := url.Parse(value); err != nil {
			return fmt.Errorf("invalid URL: %s", err)
		} else {
			*target = u
			return nil
		}
	})
	return
}
