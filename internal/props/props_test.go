package props

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile writes a temporary property file and returns its path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func mapLookup(m map[string]string) Lookup {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

var testDecls = []Decl{
	{Name: "module", Doc: "go module path", Scope: ScopePackage},
	{Name: "exported", Default: "false", Doc: "export everything", Scope: ScopePackage},
	{Name: "outputDir", Default: "out", Doc: "output directory"},
	{Name: "packages", Doc: "package mapping", Scope: ScopePackage},
}

func TestNew_DeclErrors(t *testing.T) {
	_, err := New([]Decl{{Name: ""}}, nil, nil, nil)
	require.Error(t, err)
	_, err = New([]Decl{{Name: "a"}, {Name: "a"}}, nil, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"a"`)
}

func TestNew_FileError(t *testing.T) {
	_, err := New(nil, nil, []string{filepath.Join(t.TempDir(), "missing.properties")}, nil)
	require.Error(t, err)
	_, err = New(nil, nil, []string{writeFile(t, "bad.yaml", "a:\n  b: 1\n")}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad.yaml")
}

func TestReservedAdded(t *testing.T) {
	p, err := New(testDecls, nil, nil, nil)
	require.NoError(t, err)
	names := make([]string, 0, len(p.Decls()))
	for _, d := range p.Decls() {
		names = append(names, d.Name)
	}
	assert.Equal(t, []string{"debug", "exported", "fileNotes", "module", "nsURI", "outputDir", "packages", "templates"}, names)
	assert.Equal(t, "false", p.Get("debug"))
	assert.False(t, p.Bool("debug"))
	assert.Empty(t, p.Unknown())
}

func TestReservedOverriddenByLanguage(t *testing.T) {
	p, err := New([]Decl{{Name: "debug", Default: "true", Doc: "mine"}}, nil, nil, nil)
	require.NoError(t, err)
	assert.True(t, p.Bool("debug"))
	assert.Len(t, p.Decls(), 4)
	assert.Equal(t, "mine", p.Decls()[0].Doc)
}

func TestPrecedence(t *testing.T) {
	first := writeFile(t, "first.properties", "module=from-first\nexported=\noutputDir=first-out\nonlyFirst=1\n")
	second := writeFile(t, "second.yaml", "module: from-second\nexported: true\noutputDir: second-out\nonlySecond: 2\n")
	annotation := mapLookup(map[string]string{
		"module":    "from-annotation",
		"exported":  "from-annotation",
		"outputDir": "from-annotation",
		"undecl":    "from-annotation",
	})
	cli := map[string]string{"module": "from-cli", "outputDir": "", "cliOnly": "3"}
	p, err := New(testDecls, cli, []string{first, second}, annotation)
	require.NoError(t, err)

	// CLI wins
	v, layer, ok := p.Lookup("module")
	assert.True(t, ok)
	assert.Equal(t, "from-cli", v)
	assert.Equal(t, LayerCLI, layer)

	// an empty CLI value falls through to the files; the first file wins
	v, layer, ok = p.Lookup("outputDir")
	assert.True(t, ok)
	assert.Equal(t, "first-out", v)
	assert.Equal(t, LayerFile, layer)

	// the first file defines exported as empty: falls through to the annotation
	v, layer, ok = p.Lookup("exported")
	assert.True(t, ok)
	assert.Equal(t, "from-annotation", v)
	assert.Equal(t, LayerAnnotation, layer)

	// keys given by one file only
	assert.Equal(t, "1", p.Get("onlyFirst"))
	assert.Equal(t, "2", p.Get("onlySecond"))
	assert.Equal(t, "3", p.Get("cliOnly"))

	// an undeclared key is never read from the annotation
	assert.Equal(t, "", p.Get("undecl"))
	assert.False(t, p.IsSet("undecl"))
	raw, found := p.Annotation("undecl")
	assert.True(t, found)
	assert.Equal(t, "from-annotation", raw)

	// Set beats everything
	p.Set("module", "from-engine")
	v, layer, ok = p.Lookup("module")
	assert.True(t, ok)
	assert.Equal(t, "from-engine", v)
	assert.Equal(t, LayerCLI, layer)
	p.Set("module", "")
	assert.Equal(t, "from-cli", p.Get("module"))

	assert.Equal(t, []string{"cliOnly", "onlyFirst", "onlySecond"}, p.Unknown())
}

func TestDefaults(t *testing.T) {
	p, err := New(testDecls, nil, nil, nil)
	require.NoError(t, err)

	v, layer, ok := p.Lookup("outputDir")
	assert.True(t, ok)
	assert.Equal(t, "out", v)
	assert.Equal(t, LayerDefault, layer)
	assert.True(t, p.IsSet("outputDir"))

	v, layer, ok = p.Lookup("module")
	assert.False(t, ok)
	assert.Equal(t, "", v)
	assert.Equal(t, LayerNone, layer)
	assert.False(t, p.IsSet("module"))

	v, layer, ok = p.Lookup("nope")
	assert.False(t, ok)
	assert.Equal(t, "", v)
	assert.Equal(t, LayerNone, layer)
}

func TestAnnotationScope(t *testing.T) {
	annotation := mapLookup(map[string]string{"outputDir": "ann", "module": "ann", "empty": ""})
	p, err := New(append(testDecls, Decl{Name: "empty", Scope: ScopePackage}), nil, nil, annotation)
	require.NoError(t, err)
	// ScopeRun declaration: the annotation is ignored, the default is used
	assert.Equal(t, "out", p.Get("outputDir"))
	// ScopePackage declaration
	assert.Equal(t, "ann", p.Get("module"))
	// an empty annotation value is not a value
	assert.False(t, p.IsSet("empty"))

	// nil annotation
	p, err = New(testDecls, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "", p.Get("module"))
	_, found := p.Annotation("module")
	assert.False(t, found)
}

func TestBoolAndList(t *testing.T) {
	p, err := New(testDecls, map[string]string{
		"exported":  "true",
		"packages":  " a=b ; c=d,, e=f ;",
		"templates": "model, tests",
		"single":    "  x  ",
	}, nil, nil)
	require.NoError(t, err)
	assert.True(t, p.Bool("exported"))
	assert.False(t, p.Bool("module"))
	assert.False(t, p.Bool("debug"))
	assert.Equal(t, []string{"a=b", "c=d", "e=f"}, p.List("packages", ";,"))
	assert.Equal(t, []string{"model", "tests"}, p.List("templates", ","))
	assert.Equal(t, []string{"x"}, p.List("single", ""))
	assert.Nil(t, p.List("module", ","))
}

func TestRawLayers(t *testing.T) {
	file := writeFile(t, "f.json", `{"packages": "from-file", "fileEmpty": "", "fileOnly": "fo"}`)
	p, err := New(testDecls, map[string]string{"packages": "from-cli", "cliEmpty": ""}, []string{file}, mapLookup(map[string]string{"packages": "from-ann"}))
	require.NoError(t, err)
	p.Set("packages", "from-engine")

	v, ok := p.CLI("packages")
	assert.True(t, ok)
	assert.Equal(t, "from-cli", v)
	v, ok = p.CLI("cliEmpty")
	assert.True(t, ok)
	assert.Equal(t, "", v)
	v, ok = p.CLI("fileEmpty")
	assert.True(t, ok)
	assert.Equal(t, "", v)
	v, ok = p.CLI("fileOnly")
	assert.True(t, ok)
	assert.Equal(t, "fo", v)
	_, ok = p.CLI("module")
	assert.False(t, ok)

	v, ok = p.Annotation("packages")
	assert.True(t, ok)
	assert.Equal(t, "from-ann", v)
	_, ok = p.Annotation("module")
	assert.False(t, ok)

	assert.Equal(t, "from-engine", p.Get("packages"))
	assert.Equal(t, []string{"cliEmpty", "fileEmpty", "fileOnly"}, p.Unknown())
}

func TestGeneratorPropertiesFile(t *testing.T) {
	p, err := New(nil, nil, []string{filepath.Join(testdata, "ts.properties")}, nil)
	require.NoError(t, err)
	v, layer, ok := p.Lookup("fileNotes")
	assert.True(t, ok)
	assert.Equal(t, LayerFile, layer)
	assert.Contains(t, v, "Copyright(c) 2024 MASA Group")
	assert.Equal(t, "$1/src", p.Get("packageDir"))
	assert.Equal(t, []string{"packageDir", "testFramework"}, p.Unknown())
}

func TestFormatHelp(t *testing.T) {
	assert.Equal(t, "", FormatHelp(nil))
	got := FormatHelp([]Decl{
		{Name: "module", Doc: "go module path", Scope: ScopePackage},
		{Name: "outputDir", Default: "out", Doc: "output directory\nrelative to the model"},
		{Name: "debug", Default: "false", Doc: "print debug information"},
	})
	want := "" +
		"  module     -      go module path [package]\n" +
		"  outputDir  out    output directory\n" +
		"                    relative to the model\n" +
		"  debug      false  print debug information\n"
	assert.Equal(t, want, got)
}

func TestStringers(t *testing.T) {
	assert.Equal(t, "run", ScopeRun.String())
	assert.Equal(t, "package", ScopePackage.String())
	assert.Equal(t, "Scope(7)", Scope(7).String())
	for layer, name := range map[Layer]string{LayerNone: "none", LayerCLI: "cli", LayerFile: "file", LayerAnnotation: "annotation", LayerDefault: "default", Layer(9): "Layer(9)"} {
		assert.Equal(t, name, layer.String())
	}
}
