package kingpin

import (
	"fmt"
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
func (f *flagGroup) FlagsStruct(v interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct but received " + reflect.TypeOf(v).String())
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

		// Define flag using extracted tags
		flag := f.Flag(name, help)
		if dflt != "" {
			flag = flag.Default(dflt)
		}
		if short != "" {
			r, _ := utf8.DecodeRuneInString(short)
			if r == utf8.RuneError {
				return fmt.Errorf("invalid short flag %q", short)
			}
			flag = flag.Short(r)
		}
		if required != "" {
			flag = flag.Required()
		}
		if hidden != "" {
			flag = flag.Hidden()
		}
		if placeholder != "" {
			flag = flag.PlaceHolder(placeholder)
		}
		ptr := field.Addr().Interface()
		switch ft.Type.Kind() {
		case reflect.String:
			flag.StringVar(ptr.(*string))

		case reflect.Bool:
			flag.BoolVar(ptr.(*bool))

		case reflect.Float32:
			flag.Float32Var(ptr.(*float32))
		case reflect.Float64:
			flag.Float64Var(ptr.(*float64))

		case reflect.Int:
			flag.IntVar(ptr.(*int))
		case reflect.Int8:
			flag.Int8Var(ptr.(*int8))
		case reflect.Int16:
			flag.Int16Var(ptr.(*int16))
		case reflect.Int32:
			flag.Int32Var(ptr.(*int32))
		case reflect.Int64:
			flag.Int64Var(ptr.(*int64))

		case reflect.Uint:
			flag.UintVar(ptr.(*uint))
		case reflect.Uint8:
			flag.Uint8Var(ptr.(*uint8))
		case reflect.Uint16:
			flag.Uint16Var(ptr.(*uint16))
		case reflect.Uint32:
			flag.Uint32Var(ptr.(*uint32))
		case reflect.Uint64:
			flag.Uint64Var(ptr.(*uint64))

		case reflect.Slice:
			switch ft.Type.Elem().Kind() {
			case reflect.String:
				flag.StringsVar(field.Addr().Interface().(*[]string))

			case reflect.Bool:
				flag.BoolListVar(field.Addr().Interface().(*[]bool))

			case reflect.Float32:
				flag.Float32ListVar(ptr.(*[]float32))
			case reflect.Float64:
				flag.Float64ListVar(ptr.(*[]float64))

			case reflect.Int:
				flag.IntsVar(field.Addr().Interface().(*[]int))
			case reflect.Int8:
				flag.Int8ListVar(ptr.(*[]int8))
			case reflect.Int16:
				flag.Int16ListVar(ptr.(*[]int16))
			case reflect.Int32:
				flag.Int32ListVar(ptr.(*[]int32))
			case reflect.Int64:
				flag.Int64ListVar(ptr.(*[]int64))

			case reflect.Uint:
				flag.UintsVar(ptr.(*[]uint))
			case reflect.Uint8:
				flag.Uint8ListVar(ptr.(*[]uint8))
			case reflect.Uint16:
				flag.Uint16ListVar(ptr.(*[]uint16))
			case reflect.Uint32:
				flag.Uint32ListVar(ptr.(*[]uint32))
			case reflect.Uint64:
				flag.Uint64ListVar(ptr.(*[]uint64))

			default:
				if ft.Type == reflect.TypeOf(time.Duration(0)) {
					flag.DurationListVar(ptr.(*[]time.Duration))
				} else {
					return fmt.Errorf("unsupported field type %s for field %s", ft.Type.String(), ft.Name)
				}
			}

		default:
			if ft.Type == reflect.TypeOf(time.Duration(0)) {
				flag.DurationVar(ptr.(*time.Duration))
			} else {
				return fmt.Errorf("unsupported field type %s for field %s", ft.Type.String(), ft.Name)
			}
		}
	}
	return nil
}
