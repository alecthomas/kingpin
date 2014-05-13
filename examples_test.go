package kingpin

import (
	"fmt"
	"net/http"
	"strings"
)

// This example ilustrates how to define custom parsers. HTTPHeader
// cumulatively parses each encountered --header flag into a http.Header struct.
func ExampleParser() {
	// HTTPHeader parses flags cumulatively into a http.Header.
	HTTPHeader := func(s Settings) (target *http.Header) {
		target = &http.Header{}
		s.SetParser(func(value string) error {
			parts := strings.SplitN(value, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("expected HEADER:VALUE got '%s'", value)
			}
			target.Add(parts[0], parts[1])
			return nil
		})
		return
	}

	var (
		curl    = New("curl", "transfer a URL")
		headers = HTTPHeader(curl.Flag("headers", "Add HTTP headers to the request.").Short('H').MetaVar("HEADER:VALUE"))
	)

	curl.Parse([]string{"-H Content-Type:application/octet-stream"})
	for key, value := range *headers {
		fmt.Printf("%s = %s\n", key, value)
	}
}
