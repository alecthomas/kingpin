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

// NOTE: Most of the base type values were lifted from:
// http://golang.org/src/pkg/flag/flag.go?s=20146:20222#L79

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
//
// If a Value has an IsBoolFlag() bool method returning true,
// the command-line parser makes -name equivalent to -name=true
// rather than using the next command-line argument.
type Value interface {
	String() string
	Set(string) error
}

// Getter is an interface that allows the contents of a Value to be retrieved.
// It wraps the Value interface, rather than being part of it, because it
// appeared after Go 1 and its compatibility rules. All Value types provided
// by this package satisfy the Getter interface.
type Getter interface {
	Value
	Get() interface{}
}

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type boolFlag interface {
	Value
	IsBoolFlag() bool
}

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() interface{} { return bool(*b) }

func (b *boolValue) String() string { return fmt.Sprintf("%v", *b) }

func (b *boolValue) IsBoolFlag() bool { return true }

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = intValue(v)
	return err
}

func (i *intValue) Get() interface{} { return int(*i) }

func (i *intValue) String() string { return fmt.Sprintf("%v", *i) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() interface{} { return int64(*i) }

func (i *int64Value) String() string { return fmt.Sprintf("%v", *i) }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() interface{} { return uint(*i) }

func (i *uintValue) String() string { return fmt.Sprintf("%v", *i) }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() interface{} { return uint64(*i) }

func (i *uint64Value) String() string { return fmt.Sprintf("%v", *i) }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return fmt.Sprintf("%s", *s) }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() interface{} { return float64(*f) }

func (f *float64Value) String() string { return fmt.Sprintf("%v", *f) }

// -- time.Duration Value
type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() interface{} { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

// -- []string Value
type stringsValue []string

func newStringsValue(p *[]string) *stringsValue {
	return (*stringsValue)(p)
}

func (s *stringsValue) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *stringsValue) String() string {
	return strings.Join(*s, ",")
}

// -- map[string]string Value
type stringMapValue map[string]string

func newStringMapValue(p *map[string]string) *stringMapValue {
	return (*stringMapValue)(p)
}

func (s *stringMapValue) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected KEY=VALUE got '%s'", value)
	}
	(*s)[parts[0]] = parts[1]
	return nil
}
func (s *stringMapValue) String() string {
	return fmt.Sprintf("%s", map[string]string(*s))
}

// -- net.IP Value
type ipValue net.IP

func newIPValue(p *net.IP) *ipValue {
	return (*ipValue)(p)
}

func (i *ipValue) Set(value string) error {
	if ip := net.ParseIP(value); ip == nil {
		return fmt.Errorf("'%s' is not an IP address", value)
	} else {
		*i = *(*ipValue)(&ip)
		return nil
	}
}

func (i *ipValue) String() string {
	return (*net.IP)(i).String()
}

// -- existingFile Value

type fileStatValue string

func newFileStatValue(p *string, predicate func(os.FileInfo) bool) *fileStatValue {
	return (*fileStatValue)(p)
}

func (e *fileStatValue) Set(value string) error {
	if s, err := os.Stat(value); err != nil || s.IsDir() {
		return fmt.Errorf("'%s' is a directory", value)
	}
	*e = fileStatValue(value)
	return nil
}

func (e *fileStatValue) String() string {
	return (string)(*e)
}

// -- os.File value

type fileValue struct {
	f **os.File
}

func newFileValue(p **os.File) *fileValue {
	return &fileValue{p}
}

func (f *fileValue) Set(value string) error {
	if fd, err := os.Open(value); err != nil {
		return err
	} else {
		*f.f = fd
		return nil
	}
}

func (f *fileValue) String() string {
	if *f.f == nil {
		return "<nil>"
	}
	return (*f.f).Name()
}

// -- url.URL Value
type urlValue struct {
	u **url.URL
}

func newURLValue(p **url.URL) *urlValue {
	return &urlValue{p}
}

func (u *urlValue) Set(value string) error {
	if url, err := url.Parse(value); err != nil {
		return fmt.Errorf("invalid URL: %s", err)
	} else {
		*u.u = url
		return nil
	}
}

func (u *urlValue) String() string {
	if *u.u == nil {
		return "<nil>"
	}
	return (*u.u).String()
}
