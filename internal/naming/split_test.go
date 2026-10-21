package naming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name  string
		words []string
	}{
		{"", []string{""}},
		{"a", []string{"a"}},
		{"ab", []string{"ab"}},
		{"aB", []string{"aB"}},
		{"AB", []string{"AB"}},
		{"EClass", []string{"E", "Class"}},
		{"EInt", []string{"E", "Int"}},
		{"eType", []string{"e", "Type"}},
		{"eAllStructuralFeatures", []string{"e", "All", "Structural", "Features"}},
		{"EStringToStringMapEntry", []string{"E", "String", "To", "String", "Map", "Entry"}},
		{"EFeatureMapEntry", []string{"E", "Feature", "Map", "Entry"}},
		{"XMLSchema", []string{"XML", "Schema"}},
		{"eIDAttribute", []string{"eID", "Attribute"}},
		{"operationID", []string{"operation", "ID"}},
		{"URI2", []string{"URI2"}},
		{"a1", []string{"a1"}},
		{"A1b", []string{"A", "1b"}},
		{"AB1c", []string{"AB", "1c"}},
		{"name_with_underscore", []string{"name", "with", "underscore"}},
		{"_leading", []string{"leading"}},
		{"trailing_", []string{"trailing", ""}},
		{"double__underscore", []string{"double", "underscore"}},
		{"a_b", []string{"a", "b"}},
		{"snake_Camel", []string{"snake", "Camel"}},
		{"écoleÉlève", []string{"école", "Élève"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.words, SplitWords(test.name))
		})
	}
}

func TestSplitAndCombine(t *testing.T) {
	tests := []struct {
		name     string
		sep      string
		expected string
	}{
		{"", "_", ""},
		{"a", "_", "a"},
		{"ab", "_", "ab"},
		{"aB", "_", "aB"},
		{"EClass", "_", "EClass"},
		{"eAllStructuralFeatures", "_", "eAll_Structural_Features"},
		{"XMLSchema", "_", "XML_Schema"},
		{"eIDAttribute", "_", "eID_Attribute"},
		{"A1b", "_", "A1b"}, // words [A 1b]: the one-character "A" gets no separator
		{"AB1c", "_", "AB_1c"},
		{"name_with_underscore", "_", "name_with_underscore"},
		{"a_b_c", "_", "abc"},
		{"trailing_", "_", "trailing_"},
		{"eAllStructuralFeatures", "-", "eAll-Structural-Features"},
		{"eAllStructuralFeatures", "", "eAllStructuralFeatures"},
	}
	for _, test := range tests {
		t.Run(test.name+"/"+test.sep, func(t *testing.T) {
			assert.Equal(t, test.expected, SplitAndCombine(test.name, test.sep))
		})
	}
}
