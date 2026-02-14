package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func isToken(s []byte) bool {
	if len(s) == 0 {
		return false
	}

	for _, b := range s {
		// visible ASCII only
		if b < 0x21 || b > 0x7E {
			return false
		}

		switch b {
		case '"', '(', ')', ',', '/', ':', ';', '<', '=', '>', '?',
			'@', '[', '\\', ']', '{', '}':
			return false
		}
	}

	return true
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if i := bytes.Index(data, []byte("\r\n")); i != -1 {
		if i == 0 {
			return 2, true, nil
		}
		line, afterLine, _ := bytes.Cut(data, []byte("\r\n"))
		before, after, ok := bytes.Cut(line, []byte(":"))

		if !ok {
			return 0, false, fmt.Errorf("Error: no colon delimiter")
		}

		before = bytes.TrimLeft(before, " \t")
		after = bytes.Trim(after, " \t")

		if len(after) == 0 || len(before) == 0 {
			return 0, false, fmt.Errorf("Error: nothing after/before the colon")
		}
		if !isToken(before) {
			return 0, false, fmt.Errorf("Error: field-name must be a token")
		}
		h.Set(string(before), string(after))
		return len(data) - len(afterLine), false, nil
	}
	return 0, false, nil
}

func (h Headers) Get(name string) string {
	return h[strings.ToLower(name)]
}
func (h Headers) Set(name, value string) {
	lowerKey := strings.ToLower(name)
	if v, ok := h[lowerKey]; ok {
		h[lowerKey] = v + ", " + value
		return
	}
	h[lowerKey] = value
}
