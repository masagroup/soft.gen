package naming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvoidKeywords(t *testing.T) {
	// The Go reserved words of the v1 getReservedKeywords.
	goKeywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	tests := []struct {
		name     string
		keywords map[string]bool
		suffix   string
		expected string
	}{
		{"type", goKeywords, "_", "type_"},
		{"range", goKeywords, "_", "range_"},
		{"name", goKeywords, "_", "name"},
		{"Type", goKeywords, "_", "Type"},
		{"", goKeywords, "_", ""},
		{"type", goKeywords, "", "type"},
		{"type", nil, "_", "type"},
		{"type", map[string]bool{"type": false}, "_", "type"},
		{"class", map[string]bool{"class": true}, "Value", "classValue"},
	}
	for _, test := range tests {
		t.Run(test.name+test.suffix, func(t *testing.T) {
			assert.Equal(t, test.expected, AvoidKeywords(test.name, test.keywords, test.suffix))
		})
	}
}
