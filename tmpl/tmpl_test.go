package tmpl

import (
	"testing"
	"testing/fstest"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var conventional = fstest.MapFS{
	"class/impl.go.tmpl": &fstest.MapFile{Data: []byte(`{{ define "class/impl.go" -}}
type {{ .Name }} struct {
{{- range .Fields }}
	{{ include (printf "field/decl.%s" .Kind) . }}
{{- end }}
}
{{ blank }}
{{- if .Doc }}
// {{ .Doc }}
{{- end }}
{{- end }}
`)},
	"field/decl.tmpl": &fstest.MapFile{Data: []byte(`{{/* a field declaration, one variant per kind */}}
{{ define "field/decl.plain" -}}
{{ .Name }} {{ .Type }}
{{- end }}

{{ define "field/decl.pointer" -}}
{{ .Name }} *{{ .Type }}
{{- end }}
`)},
}

type field struct{ Name, Type, Kind string }

func TestLoadRenderInclude(t *testing.T) {
	set, err := Load(conventional, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"class/impl.go", "field/decl.plain", "field/decl.pointer"}, set.Names())
	assert.True(t, set.Has("field/decl.plain"))
	assert.False(t, set.Has("field/decl.other"))

	out, err := set.Render("class/impl.go", map[string]any{
		"Name":   "bookImpl",
		"Doc":    "a book",
		"Fields": []field{{"title", "string", "plain"}, {"author", "Writer", "pointer"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "type bookImpl struct {\n\ttitle string\n\tauthor *Writer\n}\n\n// a book", string(out))

	_, err = set.Render("missing", nil)
	assert.Error(t, err)
	_, err = set.Render("class/impl.go", map[string]any{"Name": "x", "Fields": []field{{"a", "b", "unknown"}}})
	assert.Error(t, err, "an unknown variant is an error")
}

func TestLoadLanguageFuncsOverride(t *testing.T) {
	fsys := fstest.MapFS{"x.tmpl": &fstest.MapFile{Data: []byte(`{{ define "x" -}}
{{ upper .Name }} {{ shout .Name }}
{{- end }}
`)}}
	set, err := Load(fsys, template.FuncMap{"shout": func(s string) string { return s + "!" }})
	require.NoError(t, err)
	out, err := set.Render("x", map[string]string{"Name": "a"})
	require.NoError(t, err)
	assert.Equal(t, "A a!", string(out))
}

func TestBaseFuncs(t *testing.T) {
	fsys := fstest.MapFS{"f.tmpl": &fstest.MapFile{Data: []byte(`{{ define "f" -}}
{{ join ", " (list "a" "b") }}|{{ upperFirst "book" }}|{{ lowerFirst "Book" }}|{{ upperSnake "eAllStructuralFeatures" }}|{{ quote "a\"b" }}|{{ ternary true "y" "n" }}|{{ default "d" "" }}|{{ default "d" "v" }}|{{ add 1 2 }}|{{ (dict "k" "v").k }}
{{ indent 1 "x\ny" }}
{{- end }}
`)}}
	set, err := Load(fsys, nil)
	require.NoError(t, err)
	out, err := set.Render("f", nil)
	require.NoError(t, err)
	assert.Equal(t, "a, b|Book|book|EALL_STRUCTURAL_FEATURES|\"a\\\"b\"|y|d|v|3|v\n\tx\n\ty", string(out))
}

func TestLintAcceptsConventionalTemplates(t *testing.T) {
	assert.Empty(t, Lint(conventional))
}

func TestLintReportsViolations(t *testing.T) {
	fsys := fstest.MapFS{
		"feature/getter.tmpl": &fstest.MapFile{Data: []byte(`stray text
{{ define "other/name" }}
{{ if .X }}
a
{{- end -}}


{{ "\n" }}
{{ end }}
`)},
	}
	problems := Lint(fsys)
	messages := make([]string, len(problems))
	for i, p := range problems {
		messages[i] = p.String()
	}
	assert.Contains(t, messages, `feature/getter.tmpl:2: define "other/name" is not named after its file ("feature/getter" or "feature/getter.<variant>")`)
	assert.Contains(t, messages, "feature/getter.tmpl:9: control line must be left-trimmed: {{- end ... }}")
	assert.Contains(t, messages, `feature/getter.tmpl:2: define "other/name" must open with {{ define "other/name" -}}`)
	assert.Contains(t, messages, "feature/getter.tmpl:3: control line must be left-trimmed: {{- if ... }}")
	assert.Contains(t, messages, "feature/getter.tmpl:5: control line must not be right-trimmed: {{- end ... }}")
	assert.Contains(t, messages, "feature/getter.tmpl:7: two consecutive blank lines")
	assert.Contains(t, messages, `feature/getter.tmpl:8: newline literal {{ "\n" }}: use a blank line in the template`)
	assert.Contains(t, messages, "feature/getter.tmpl:1: text outside a define block")
}
