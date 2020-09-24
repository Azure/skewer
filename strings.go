package skewer

import (
	"strings"
	"unicode"
)

func stringToLower(input string) string {
	var output string
	for _, c := range input {
		if !unicode.IsSpace(c) {
			output += string(c)
		}
	}
	return strings.ToLower(output)
}

func stringEquals(a, b string) bool {
	return stringToLower(a) == stringToLower(b)
}
