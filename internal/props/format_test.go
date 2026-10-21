package props

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testdata = "../../testdata/props"

func fixture(name string) string {
	return filepath.Join(testdata, name)
}

const mplBanner = "// This Source Code Form is subject to the terms of the Mozilla Public\n" +
	"// License, v. 2.0. If a copy of the MPL was not distributed with this\n" +
	"// file, You can obtain one at https://mozilla.org/MPL/2.0/.\n"

func TestFormats(t *testing.T) {
	assert.Equal(t, []string{".json", ".properties", ".yaml", ".yml"}, Formats())
}

func TestParseProperties_UnknownExtension(t *testing.T) {
	_, err := ParseProperties(fixture("notes.toml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `".toml"`)
	for _, ext := range Formats() {
		assert.Contains(t, err.Error(), ext)
	}
}

func TestParseProperties_MissingFile(t *testing.T) {
	_, err := ParseProperties(fixture("missing.properties"))
	require.Error(t, err)
}

func TestParsePropertiesReader_UnknownFormat(t *testing.T) {
	_, err := ParsePropertiesReader(strings.NewReader("a=b"), "toml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "json, properties, yaml")
}

func TestParseProperties_NotesSameInEveryFormat(t *testing.T) {
	want := map[string]string{
		"fileNotes":  mplBanner,
		"packageDir": "$1/src",
	}
	for _, name := range []string{"notes.properties", "notes.yaml", "notes.json"} {
		t.Run(name, func(t *testing.T) {
			got, err := ParseProperties(fixture(name))
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func TestParseProperties_GeneratorFiles(t *testing.T) {
	const goNotes = "// *****************************************************************************\n" +
		"// Copyright(c) 2021 MASA Group                                                 \n" +
		"//                                                                              \n" +
		"// This Source Code Form is subject to the terms of the Mozilla Public          \n" +
		"// License, v. 2.0. If a copy of the MPL was not distributed with this          \n" +
		"// file, You can obtain one at https://mozilla.org/MPL/2.0/.                    \n" +
		"//                                                                              \n" +
		"// *****************************************************************************\n"
	goValues, err := ParseProperties(fixture("go.properties"))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"fileNotes": goNotes}, goValues)

	tsValues, err := ParseProperties(fixture("ts.properties"))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"fileNotes":     strings.Replace(goNotes, "2021", "2024", 1),
		"packageDir":    "$1/src",
		"testFramework": "vitest",
	}, tsValues)
	lines := strings.SplitAfter(tsValues["fileNotes"], "\n")
	require.Equal(t, "", lines[len(lines)-1], "the value must end with a newline")
	for _, line := range lines[:len(lines)-1] {
		assert.True(t, strings.HasSuffix(line, "\n"), "line %q should end with a newline", line)
	}
	assert.True(t, strings.HasPrefix(tsValues["fileNotes"], "// *****"))
}

func TestJava_Syntax(t *testing.T) {
	src := strings.Join([]string{
		"# a comment",
		"! another comment",
		"   # indented comment",
		"",
		"equals=1",
		"colon:2",
		"space 3",
		"  indented = 4  ",
		"tab\t=\t5",
		"empty=",
		"bare",
		"key\\=with\\:seps\\ and\\#hash = value\\!bang",
		"escapes=a\\tb\\nc\\rd\\fe\\\\f\\u00e9\\z",
		"continued=one \\",
		"    two \\",
		"  three",
		"even=ends with two backslashes\\\\",
		"",
		"trailing=at eof\\",
	}, "\n")
	got, err := ParsePropertiesReader(strings.NewReader(src), "properties")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"equals":                 "1",
		"colon":                  "2",
		"space":                  "3",
		"indented":               "4  ",
		"tab":                    "5",
		"empty":                  "",
		"bare":                   "",
		"key=with:seps and#hash": "value!bang",
		"escapes":                "a\tb\nc\rd\fe\\fé" + "z",
		"continued":              "one two three",
		"even":                   "ends with two backslashes\\",
		"trailing":               "at eof",
	}, got)
}

func TestJava_CRLF(t *testing.T) {
	got, err := ParsePropertiesReader(strings.NewReader("a=1\r\nb=x \\\r\n  y\r\nc=3\r"), "properties")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "1", "b": "x y", "c": "3"}, got)
}

func TestJava_BadUnicode(t *testing.T) {
	_, err := ParsePropertiesReader(strings.NewReader("a=\\u12"), "properties")
	require.Error(t, err)
	_, err = ParsePropertiesReader(strings.NewReader("a=\\uzzzz"), "properties")
	require.Error(t, err)
}

func TestYAML_Scalars(t *testing.T) {
	src := `
str: hello
quoted: "true"
bool: true
int: 42
float: 1.50
hex: 0x1F
nothing: null
tilde: ~
empty:
anchored: &a value
alias: *a
block: |
  line 1
  line 2
folded: >-
  a
  b
`
	got, err := ParsePropertiesReader(strings.NewReader(src), "yaml")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"str":      "hello",
		"quoted":   "true",
		"bool":     "true",
		"int":      "42",
		"float":    "1.50",
		"hex":      "0x1F",
		"nothing":  "",
		"tilde":    "",
		"empty":    "",
		"anchored": "value",
		"alias":    "value",
		"block":    "line 1\nline 2\n",
		"folded":   "a b",
	}, got)
}

func TestYAML_Empty(t *testing.T) {
	for _, src := range []string{"", "# only a comment\n", "---\n"} {
		got, err := ParsePropertiesReader(strings.NewReader(src), "yaml")
		require.NoError(t, err, "%q", src)
		assert.Empty(t, got)
	}
}

func TestYAML_Errors(t *testing.T) {
	_, err := ParsePropertiesReader(strings.NewReader("a:\n  b: 1\n"), "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"a"`)

	_, err = ParsePropertiesReader(strings.NewReader("a: [1, 2]\n"), "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"a"`)

	_, err = ParsePropertiesReader(strings.NewReader("- a\n- b\n"), "yaml")
	require.Error(t, err)

	_, err = ParsePropertiesReader(strings.NewReader("a: [\n"), "yaml")
	require.Error(t, err)
}

func TestJSON_Scalars(t *testing.T) {
	src := `{"str": "hello", "bool": true, "no": false, "int": 42, "float": 1.50, "big": 12345678901234567890, "exp": 1e3, "nothing": null, "multi": "a\nb\n"}`
	got, err := ParsePropertiesReader(strings.NewReader(src), "json")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"str":     "hello",
		"bool":    "true",
		"no":      "false",
		"int":     "42",
		"float":   "1.50",
		"big":     "12345678901234567890",
		"exp":     "1e3",
		"nothing": "",
		"multi":   "a\nb\n",
	}, got)
}

func TestJSON_Empty(t *testing.T) {
	got, err := ParsePropertiesReader(strings.NewReader(""), "json")
	require.NoError(t, err)
	assert.Empty(t, got)
	got, err = ParsePropertiesReader(strings.NewReader("{}"), "json")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestJSON_Errors(t *testing.T) {
	_, err := ParsePropertiesReader(strings.NewReader(`{"a": {"b": 1}}`), "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"a"`)

	_, err = ParsePropertiesReader(strings.NewReader(`{"a": [1]}`), "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"a"`)

	_, err = ParsePropertiesReader(strings.NewReader(`[1]`), "json")
	require.Error(t, err)

	_, err = ParsePropertiesReader(strings.NewReader(`{"a": 1} {"b": 2}`), "json")
	require.Error(t, err)

	_, err = ParsePropertiesReader(strings.NewReader(`{"a": `), "json")
	require.Error(t, err)
}

func TestParseKeyValues(t *testing.T) {
	got, err := ParseKeyValues([]string{"a=1", "b=x=y", " c =", "a=2"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "2", "b": "x=y", "c": ""}, got)

	got, err = ParseKeyValues(nil)
	require.NoError(t, err)
	assert.Empty(t, got)

	_, err = ParseKeyValues([]string{"a=1", "novalue"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"novalue"`)

	_, err = ParseKeyValues([]string{"=1"})
	require.Error(t, err)
}
