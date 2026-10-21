package naming

import (
	"strings"
	"unicode"
)

// wordSeparator is the explicit word delimiter of the EMF word splitting.
const wordSeparator = '_'

// SplitWords breaks s into words delimited by '_' and/or mixed-case naming,
// exactly as the EMF code generator does: "eAllStructuralFeatures" gives
// [e All Structural Features], "XMLSchema" gives [XML Schema], and "a_b"
// gives [a b]; the underscores are dropped and an empty s gives [""].
func SplitWords(s string) []string {
	// Verbatim port of GenerateUtils.splitWords(s, '_') of the v1 generators,
	// on runes instead of Java chars.
	result := make([]string, 0, 4)
	currentWord := make([]rune, 0, len(s))
	lastIsLower := false
	for _, curChar := range s {
		if unicode.IsUpper(curChar) || (!lastIsLower && unicode.IsDigit(curChar)) || curChar == wordSeparator {
			if (lastIsLower && len(currentWord) > 1) || (curChar == wordSeparator && len(currentWord) > 0) {
				result = append(result, string(currentWord))
				currentWord = currentWord[:0]
			}
			lastIsLower = false
		} else {
			if !lastIsLower {
				if n := len(currentWord); n > 1 {
					// The last upper-case rune starts the new word: "XMLSchema" -> "XML" + "S".
					lastChar := currentWord[n-1]
					result = append(result, string(currentWord[:n-1]))
					currentWord = append(currentWord[:0], lastChar)
				}
			}
			lastIsLower = true
		}
		if curChar != wordSeparator {
			currentWord = append(currentWord, curChar)
		}
	}
	result = append(result, string(currentWord))
	return result
}

// SplitAndCombine splits s into words with SplitWords and joins them with
// sep, except that a word of one character gets no separator after it:
// "eAllStructuralFeatures" with "_" gives "eAll_Structural_Features".
func SplitAndCombine(s, sep string) string {
	// Verbatim port of GenerateUtils.splitAndCombineWords of the v1 generators.
	if s == "" {
		return ""
	}
	words := SplitWords(s)
	var result strings.Builder
	for i, word := range words {
		result.WriteString(word)
		if i < len(words)-1 && len([]rune(word)) > 1 {
			result.WriteString(sep)
		}
	}
	return result.String()
}
