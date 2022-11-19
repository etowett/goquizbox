package project

import (
	"strings"
	"unicode"
)

// TrimSpace trims space and "zero-width no-break space".
func TrimSpace(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == '\uFEFF'
	})
}

// TrimSpaceAndNonPrintable trims spaces and non-printable chars from the beginning and end of a string.
func TrimSpaceAndNonPrintable(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || !unicode.IsPrint(r)
	})
}
