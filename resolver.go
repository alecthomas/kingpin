package kingpin

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	envarTransformRegexp = regexp.MustCompile(`[^a-zA-Z_]+`)
)

// A Resolver retrieves flag values from an external source, such as a configuration file or environment variables.
type Resolver interface {
	// Resolve key in the given parse context.
	//
	// A nil slice should be returned if the key can not be resolved.
	Resolve(key string, context *ParseContext) ([]string, error)
}

// ResolverFunc is a function that is also a Resolver.
type ResolverFunc func(key string, context *ParseContext) ([]string, error)

func (r ResolverFunc) Resolve(key string, context *ParseContext) ([]string, error) {
	return r(key, context)
}

// A resolver that pulls values from the flag defaults. This resolver is always installed in the ParseContext.
func defaultsResolver() Resolver {
	return ResolverFunc(func(key string, context *ParseContext) ([]string, error) {
		for _, clause := range context.CombinedFlagsAndArgs() {
			if clause.name == key {
				return clause.defaultValues, nil
			}
		}
		return nil, nil
	})
}

func parseEnvar(envar, sep string) []string {
	value, ok := os.LookupEnv(envar)
	if !ok {
		return nil
	}
	if sep == "" {
		return []string{value}
	}
	return strings.Split(value, sep)
}

// Resolves a clause value from the envar configured on that clause, if any.
func envarResolver(sep string) Resolver {
	return ResolverFunc(func(key string, context *ParseContext) ([]string, error) {
		for _, clause := range context.CombinedFlagsAndArgs() {
			if key == clause.name {
				if clause.noEnvar || clause.envar == "" {
					return nil, nil
				}
				return parseEnvar(clause.envar, sep), nil
			}
		}
		return nil, nil
	})
}

// MapResolver resolves values from a static map.
func MapResolver(values map[string][]string) Resolver {
	return ResolverFunc(func(key string, context *ParseContext) ([]string, error) {
		return values[key], nil
	})
}

// JSONResolver returns a Resolver that retrieves values from a JSON source.
func JSONResolver(data []byte) (Resolver, error) {
	values := map[string]interface{}{}
	err := json.Unmarshal(data, &values)
	if err != nil {
		return nil, err
	}
	mapping := map[string][]string{}
	for key, value := range values {
		sub, err := jsonDecodeValue(value)
		if err != nil {
			return nil, err
		}
		mapping[key] = sub
	}
	return MapResolver(mapping), nil
}

func jsonDecodeValue(value interface{}) ([]string, error) {
	switch v := value.(type) {
	case []interface{}:
		out := []string{}
		for _, sv := range v {
			next, err := jsonDecodeValue(sv)
			if err != nil {
				return nil, err
			}
			out = append(out, next...)
		}
		return out, nil
	case string:
		return []string{v}, nil
	case float64:
		return []string{fmt.Sprintf("%v", v)}, nil
	case bool:
		if v {
			return []string{"true"}, nil
		}
		return []string{"false"}, nil
	}
	return nil, fmt.Errorf("unsupported JSON value %v (of type %T)", value, value)
}

// RenamingResolver creates a resolver for remapping names for a child resolver.
//
// This is useful if your configuration file uses a naming convention that does not map directly to
// flag names.
func RenamingResolver(resolver Resolver, rename func(string) string) Resolver {
	return ResolverFunc(func(key string, context *ParseContext) ([]string, error) {
		return resolver.Resolve(rename(key), context)
	})
}

// PrefixedEnvarResolver resolves any flag/argument via environment variables.
//
// "prefix" is the common-prefix for the environment variables. "separator", is the character used to separate
// multiple values within a single envar (eg. ";")
//
// With a prefix of APP_, flags in the form --some-flag will be transformed to APP_SOME_FLAG.
func PrefixedEnvarResolver(prefix, separator string) Resolver {
	return ResolverFunc(func(key string, context *ParseContext) ([]string, error) {
		key = envarTransform(prefix + key)
		return parseEnvar(key, separator), nil
	})
}

// DontResolve returns a Resolver that will never return values for the given keys, even if provided.
func DontResolve(resolver Resolver, keys ...string) Resolver {
	disabled := map[string]bool{}
	for _, key := range keys {
		disabled[key] = true
	}
	return ResolverFunc(func(key string, context *ParseContext) ([]string, error) {
		if disabled[key] {
			return nil, nil
		}
		return resolver.Resolve(key, context)
	})
}

func envarTransform(name string) string {
	return strings.ToUpper(envarTransformRegexp.ReplaceAllString(name, "_"))
}
