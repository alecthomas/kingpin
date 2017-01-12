package kingpin

import (
	"errors"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

// Struct allows applications to define flags with struct tags.
//
// Supported struct tags are: help, default, placeholder, required, hidden, long and short.
//
// A field MUST have at least the "help" tag present to be converted to a flag.
//
// The name of the flag will default to the CamelCase name transformed to camel-case. This can
// be overridden with the "long" tag.
//
// All basic Go types are supported including floats, ints, strings, time.Duration,
// and slices of same.
//
// For compatibility, also supports the tags used by https://github.com/jessevdk/go-flags
func (c *cmdMixin) Struct(v interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return errors.New(T("expected a struct but received ") + reflect.TypeOf(v).String())
	}
	for i := 0; i < rv.NumField(); i++ {
		// Parse out tags
		field := rv.Field(i)
		ft := rv.Type().Field(i)
		tag := ft.Tag
		help := tag.Get("help")
		if help == "" {
			help = tag.Get("description")
		}
		if help == "" {
			continue
		}
		placeholder := tag.Get("placeholder")
		if placeholder == "" {
			placeholder = tag.Get("value-name")
		}
		dflt := tag.Get("default")
		short := tag.Get("short")
		required := tag.Get("required")
		hidden := tag.Get("hidden")
		name := strings.ToLower(strings.Join(camelCase(ft.Name), "-"))
		if tag.Get("long") != "" {
			name = tag.Get("long")
		}
		arg := tag.Get("arg")

		// Define flag using extracted tags
		var clause *Clause
		if arg != "" {
			clause = c.Arg(name, help)
		} else {
			clause = c.Flag(name, help)
		}
		if dflt != "" {
			clause = clause.Default(dflt)
		}
		if short != "" {
			r, _ := utf8.DecodeRuneInString(short)
			if r == utf8.RuneError {
				return errors.New(T("invalid short flag {{.Arg0}}", map[string]interface{}{"Arg0": short}))
			}
			clause = clause.Short(r)
		}
		if required != "" {
			clause = clause.Required()
		}
		if hidden != "" {
			clause = clause.Hidden()
		}
		if placeholder != "" {
			clause = clause.PlaceHolder(placeholder)
		}
		ptr := field.Addr().Interface()
		if ft.Type == reflect.TypeOf(time.Duration(0)) {
			clause.DurationVar(ptr.(*time.Duration))
		} else {
			switch ft.Type.Kind() {
			case reflect.String:
				clause.StringVar(ptr.(*string))

			case reflect.Bool:
				clause.BoolVar(ptr.(*bool))

			case reflect.Float32:
				clause.Float32Var(ptr.(*float32))
			case reflect.Float64:
				clause.Float64Var(ptr.(*float64))

			case reflect.Int:
				clause.IntVar(ptr.(*int))
			case reflect.Int8:
				clause.Int8Var(ptr.(*int8))
			case reflect.Int16:
				clause.Int16Var(ptr.(*int16))
			case reflect.Int32:
				clause.Int32Var(ptr.(*int32))
			case reflect.Int64:
				clause.Int64Var(ptr.(*int64))

			case reflect.Uint:
				clause.UintVar(ptr.(*uint))
			case reflect.Uint8:
				clause.Uint8Var(ptr.(*uint8))
			case reflect.Uint16:
				clause.Uint16Var(ptr.(*uint16))
			case reflect.Uint32:
				clause.Uint32Var(ptr.(*uint32))
			case reflect.Uint64:
				clause.Uint64Var(ptr.(*uint64))

			case reflect.Slice:
				if ft.Type == reflect.TypeOf(time.Duration(0)) {
					clause.DurationListVar(ptr.(*[]time.Duration))
				} else {
					switch ft.Type.Elem().Kind() {
					case reflect.String:
						clause.StringsVar(field.Addr().Interface().(*[]string))

					case reflect.Bool:
						clause.BoolListVar(field.Addr().Interface().(*[]bool))

					case reflect.Float32:
						clause.Float32ListVar(ptr.(*[]float32))
					case reflect.Float64:
						clause.Float64ListVar(ptr.(*[]float64))

					case reflect.Int:
						clause.IntsVar(field.Addr().Interface().(*[]int))
					case reflect.Int8:
						clause.Int8ListVar(ptr.(*[]int8))
					case reflect.Int16:
						clause.Int16ListVar(ptr.(*[]int16))
					case reflect.Int32:
						clause.Int32ListVar(ptr.(*[]int32))
					case reflect.Int64:
						clause.Int64ListVar(ptr.(*[]int64))

					case reflect.Uint:
						clause.UintsVar(ptr.(*[]uint))
					case reflect.Uint8:
						clause.Uint8ListVar(ptr.(*[]uint8))
					case reflect.Uint16:
						clause.Uint16ListVar(ptr.(*[]uint16))
					case reflect.Uint32:
						clause.Uint32ListVar(ptr.(*[]uint32))
					case reflect.Uint64:
						clause.Uint64ListVar(ptr.(*[]uint64))

					default:
						return errors.New(T("unsupported field type {{.Arg0}} for field {{.Arg1}}", map[string]interface{}{"Arg0": ft.Type.String(), "Arg1": ft.Name}))
					}
				}

			default:
				return errors.New(T("unsupported field type {{.Arg0}} for field {{.Arg1}}", map[string]interface{}{"Arg0": ft.Type.String(), "Arg1": ft.Name}))
			}
		}
	}
	return nil
}
