package naming

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// UpperSnake returns the EMF constant form of a name, the words of
// SplitAndCombine joined with '_' and upper-cased: "eAllStructuralFeatures"
// gives "EALL_STRUCTURAL_FEATURES" and "EStringToStringMapEntry" gives
// "ESTRING_TO_STRING_MAP_ENTRY".
func UpperSnake(s string) string {
	return strings.ToUpper(SplitAndCombine(s, "_"))
}

// UpperFirst returns s with its first rune upper-cased and the rest unchanged,
// like the Acceleo toUpperFirst: "book" gives "Book" and "" gives "".
func UpperFirst(s string) string {
	return mapFirst(s, unicode.ToUpper)
}

// LowerFirst returns s with its first rune lower-cased and the rest unchanged,
// like the Acceleo toLowerFirst: "EClass" gives "eClass" and "" gives "".
func LowerFirst(s string) string {
	return mapFirst(s, unicode.ToLower)
}

// mapFirst applies f to the first rune of s.
func mapFirst(s string, f func(rune) rune) string {
	if s == "" {
		return s
	}
	first, size := utf8.DecodeRuneInString(s)
	mapped := f(first)
	if mapped == first {
		return s
	}
	return string(mapped) + s[size:]
}
