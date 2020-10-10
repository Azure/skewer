package skewer

import (
	"strings"
	"unicode"
)

func stringNormalize(input string) string {
	var output string
	for _, c := range input {
		if !unicode.IsSpace(c) {
			output += string(c)
		}
	}
	return strings.ToLower(output)
}

func stringEqualsWithNormalization(a, b string) bool {
	return stringNormalize(a) == stringNormalize(b)
}
