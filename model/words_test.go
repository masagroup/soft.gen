package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitAndCombineWords(t *testing.T) {
	tests := map[string]string{
		"":                           "",
		"EClass":                     "EClass",
		"EModelElement":              "EModel_Element",
		"eSuperTypes":                "eSuper_Types",
		"EStringToStringMapEntry":    "EString_To_String_Map_Entry",
		"eGet":                       "eGet",
		"getEAnnotation":             "get_EAnnotation",
		"iD":                         "iD",
		"ID":                         "ID",
		"xMLNSPrefixMap":             "xMLNS_Prefix_Map",
		"snake_case_name":            "snake_case_name",
		"with2Digits10":              "with2_Digits10",
		"a_b":                        "ab", // a one character word gets no separator (v1)
		"ab_cd":                      "ab_cd",
		"__":                         "",
		"EInvocationTargetException": "EInvocation_Target_Exception",
	}
	for input, expected := range tests {
		assert.Equal(t, expected, splitAndCombineWords(input, "_"), input)
	}
	assert.Equal(t, "EModel-Element", splitAndCombineWords("EModelElement", "-"))
}

func TestSplitWords(t *testing.T) {
	assert.Equal(t, []string{"E", "Model", "Element"}, splitWords("EModelElement", '_'))
	assert.Equal(t, []string{"a", "b"}, splitWords("a_b", '_'))
	assert.Equal(t, []string{""}, splitWords("", '_'))
}

func TestFirstCase(t *testing.T) {
	assert.Equal(t, "Abc", upperFirst("abc"))
	assert.Equal(t, "ABC", upperFirst("ABC"))
	assert.Equal(t, "", upperFirst(""))
	assert.Equal(t, "abc", lowerFirst("Abc"))
	assert.Equal(t, "eClass", lowerFirst("EClass"))
	assert.Equal(t, "", lowerFirst(""))
	assert.Equal(t, "École", upperFirst("école"))
}
