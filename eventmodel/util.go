package eventmodel

import (
	"strings"
	"unicode"
)

// Capitalizes the first letter of a string
func CapitalizeFirstCharacter(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// Simple slugify: remove non-alphanumeric, normalize casing
func Slugify(s string) string {
	// Lowercase and remove non-alphanumeric characters (except space and dash)
	var builder strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == ' ' {
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}
