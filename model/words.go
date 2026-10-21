package model

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// The word splitting of this file is a private copy of GenerateUtils.java
// (splitAndCombineWords, splitWords): the identifiers of the package constants
// (ClassifierIDName, FeatureIDName...) are built from it. The naming package
// has its own copy: a tiny duplication accepted so that model imports nothing
// of the language packages.

// splitAndCombineWords splits s into words separated by underscores and/or
// mixed-casing and recombines them with separator (v1 splitAndCombineWords,
// GenerateUtils.java). A one character word is not followed by the separator:
// "EModelElement" gives "EModel_Element" with "_".
func splitAndCombineWords(s, separator string) string {
	var parsed []string
	if len(s) != 0 {
		parsed = splitWords(s, '_')
	}
	var result strings.Builder
	for i, component := range parsed {
		result.WriteString(component)
		if i < len(parsed)-1 && utf8.RuneCountInString(component) > 1 {
			result.WriteString(separator)
		}
	}
	return result.String()
}

// splitWords breaks s into words delimited by separator and/or mixed-case
// naming (v1 splitWords, GenerateUtils.java).
func splitWords(s string, separator rune) []string {
	var result []string
	var current []rune
	lastIsLower := false
	for _, cur := range s {
		if unicode.IsUpper(cur) || (!lastIsLower && unicode.IsDigit(cur)) || cur == separator {
			if (lastIsLower && len(current) > 1) || (cur == separator && len(current) > 0) {
				result = append(result, string(current))
				current = nil
			}
			lastIsLower = false
		} else {
			if !lastIsLower {
				if n := len(current); n > 1 {
					last := current[n-1]
					result = append(result, string(current[:n-1]))
					current = []rune{last}
				}
			}
			lastIsLower = true
		}
		if cur != separator {
			current = append(current, cur)
		}
	}
	result = append(result, string(current))
	return result
}

// upperFirst returns s with its first character in upper case (Acceleo toUpperFirst).
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

// lowerFirst returns s with its first character in lower case (Acceleo toLowerFirst).
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}
