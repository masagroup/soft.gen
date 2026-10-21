package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const shopModel = "../../testdata/models/shop.ecore"

// soft runs the command line and returns its exit code, stdout and stderr.
func soft(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := run(context.Background(), args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestHelpAndVersion(t *testing.T) {
	code, out, _ := soft(t)
	assert.Equal(t, 0, code)
	assert.Contains(t, out, "soft-gen <language> [options]")
	assert.Contains(t, out, "generates go code (soft-gen go "+developerLanguageVersion+")")

	code, out, _ = soft(t, "-v")
	assert.Equal(t, 0, code)
	assert.Equal(t, "soft-gen version: "+developerBuild+"\n", out)

	code, out, _ = soft(t, "go", "-v")
	assert.Equal(t, 0, code)
	assert.Equal(t, "soft-gen go version: "+developerLanguageVersion+"\n", out)

	Version, LanguageVersions = "v2.1.0", "go=v1.13.1"
	defer func() { Version, LanguageVersions = developerBuild, "" }()
	_, out, _ = soft(t, "--version")
	assert.Equal(t, "soft-gen version: v2.1.0\n", out, "the root prints the version of the binary")
	_, out, _ = soft(t, "go", "--version")
	assert.Equal(t, "soft-gen go version: v1.13.1\n", out, "a language prints its own version")
	_, out, _ = soft(t)
	assert.Contains(t, out, "generates go code (soft-gen go v1.13.1)")

	code, out, _ = soft(t, "go", "-h")
	assert.Equal(t, 0, code)
	assert.Contains(t, out, "* generateModel")
	assert.Contains(t, out, "featureDelegation")
	assert.Contains(t, out, "--config")
}

func TestErrors(t *testing.T) {
	code, _, errOut := soft(t, "cobol")
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, `unknown language "cobol"`)

	code, _, errOut = soft(t, "go", "-m", shopModel)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, "the model (-m) and the output folder (-o) are required")

	code, _, errOut = soft(t, "go", "-m", shopModel, "-o", t.TempDir(), "-p", "novalue")
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, `property "novalue" is not key=value`)

	code, _, _ = soft(t, "go", "-m", "missing.ecore", "-o", t.TempDir())
	assert.Equal(t, 1, code)

	code, _, errOut = soft(t, "go", "--config", filepath.Join(t.TempDir(), "missing.yaml"))
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, "configuration file")

	config := filepath.Join(t.TempDir(), "config.ini")
	require.NoError(t, os.WriteFile(config, []byte("model=x"), 0o644))
	code, _, errOut = soft(t, "go", "--config", config)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, `unsupported extension ".ini"`)
}

// TestLanguageVersions checks the parsing of LanguageVersions, the versions
// the Makefile sets for each language, and that an invalid value fails any run.
func TestLanguageVersions(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  languageVersions
		err   string
	}{
		{value: "", want: languageVersions{}},
		{value: "go=v1.13.1", want: languageVersions{"go": "v1.13.1"}},
		{value: " go = v1.13.1 ", want: languageVersions{"go": "v1.13.1"}},
		{value: "go", err: `"go" is not language=version`},
		{value: "=v1", err: `"=v1" is not language=version`},
		{value: "go=", err: "language go: no version"},
		{value: "go=v1,go=v2", err: "language go: version set twice"},
		{value: "go=v1.13.1,cobol=v1", err: `unknown language "cobol"`},
	} {
		got, err := parseLanguageVersions(tc.value)
		if tc.err != "" {
			assert.ErrorContains(t, err, tc.err, tc.value)
			continue
		}
		require.NoError(t, err, tc.value)
		assert.Equal(t, tc.want, got, tc.value)
	}

	LanguageVersions = "cobol=v1"
	defer func() { LanguageVersions = "" }()
	code, _, errOut := soft(t, "go", "-v")
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, `invalid LanguageVersions: unknown language "cobol"`)
}

func TestDryRunAndDiff(t *testing.T) {
	dir := t.TempDir()
	code, out, errOut := soft(t, "go", "-m", shopModel, "-o", dir, "-t", "generateModel", "--dry-run")
	assert.Equal(t, 0, code, errOut)
	assert.Contains(t, out, "file generated: ")
	entries, _ := os.ReadDir(dir)
	assert.Empty(t, entries, "dry run writes nothing")

	code, out, _ = soft(t, "go", "-m", shopModel, "-o", dir, "-t", "generateModel", "--diff", "-s")
	assert.Equal(t, 1, code, "nothing exists yet: everything differs")
	assert.Contains(t, out, "+++ shop/")

	code, _, errOut = soft(t, "go", "-m", shopModel, "-o", dir, "-t", "generateModel", "-s")
	assert.Equal(t, 0, code, errOut)
	assert.FileExists(t, filepath.Join(dir, "shop", "shoppackage.go"))

	code, out, _ = soft(t, "go", "-m", shopModel, "-o", dir, "-t", "generateModel", "--diff")
	assert.Equal(t, 0, code, "the generated files are up to date")
	assert.Empty(t, out)
}

// TestRepeatedOptions checks the v1 syntax: repeated -t, and repeated -p whose
// values keep their commas (the packages property lists two tokens, the
// second one keyed by name).
func TestRepeatedOptions(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := soft(t, "go", "-s", "-m", shopModel, "-o", dir,
		"-t", "generateModel", "-t", "generateMocks",
		"-p", "packages=http://soft.gen/test/shop=github.com/x/store/store,github.com/x/unused/other",
		"-p", "goVersion=1.17")
	require.Equal(t, 0, code, errOut)
	assert.FileExists(t, filepath.Join(dir, "shop", "shop_mock.go"), "both templates run")
	content, err := os.ReadFile(filepath.Join(dir, "shop", "shop_impl.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "package store", "the packages value keeps its comma: the first token applies")
	assert.Contains(t, string(content), "interface{}", "goVersion=1.17 applies")
}

// TestConfigFiles generates from a configuration file in each format: the
// model, the output, the templates, the property files and silent come from it.
func TestConfigFiles(t *testing.T) {
	model, _ := filepath.Abs(shopModel)
	props, _ := filepath.Abs("../../testdata/props/go.properties")
	yaml := "model: %q\noutput: %q\ntemplate: [generateInterfaces, generateMocks]\nproperties: [%q]\nsilent: true\n"
	for _, tc := range []struct{ ext, content string }{
		{".yaml", yaml},
		{".yml", yaml},
		{".json", `{"model": %q, "output": %q, "template": ["generateInterfaces", "generateMocks"], "properties": [%q], "silent": true}`},
		{".toml", "model = %q\noutput = %q\ntemplate = [\"generateInterfaces\", \"generateMocks\"]\nproperties = [%q]\nsilent = true\n"},
	} {
		t.Run(tc.ext, func(t *testing.T) {
			dir := t.TempDir()
			out := filepath.Join(dir, "out")
			config := filepath.Join(dir, "soft-gen"+tc.ext)
			require.NoError(t, os.WriteFile(config, fmt.Appendf(nil, tc.content, model, out, props), 0o644))
			code, stdout, errOut := soft(t, "go", "--config", config)
			require.Equal(t, 0, code, errOut)
			assert.Empty(t, stdout, "silent comes from the configuration file")
			assert.FileExists(t, filepath.Join(out, "shop", "shop.go"))
			assert.FileExists(t, filepath.Join(out, "shop", "shop_mock.go"))
			assert.NoFileExists(t, filepath.Join(out, "shop", "shop_impl.go"), "only the templates of the file run")
			header, err := os.ReadFile(filepath.Join(out, "shop", "shop.go"))
			require.NoError(t, err)
			assert.Contains(t, string(header), "Copyright(c) 2021 MASA Group", "the property file of the configuration applies")
		})
	}
}

// TestPrecedence checks the order of the sources of an option: the command
// line, then the environment variable, then the configuration file.
func TestPrecedence(t *testing.T) {
	dir := t.TempDir()
	model, _ := filepath.Abs(shopModel)
	config := filepath.Join(dir, "soft-gen.yaml")
	content := fmt.Appendf(nil, "model: %q\noutput: %q\ntemplate: [generateModel]\nsilent: true\n", model, filepath.Join(dir, "from-config"))
	require.NoError(t, os.WriteFile(config, content, 0o644))

	t.Setenv("SOFT_GEN_CONFIG", config)
	code, _, errOut := soft(t, "go")
	require.Equal(t, 0, code, errOut)
	assert.FileExists(t, filepath.Join(dir, "from-config", "shop", "shop_impl.go"), "the configuration file comes from SOFT_GEN_CONFIG")

	t.Setenv("SOFT_GEN_OUTPUT", filepath.Join(dir, "from-env"))
	t.Setenv("SOFT_GEN_TEMPLATE", "generateInterfaces")
	code, _, errOut = soft(t, "go")
	require.Equal(t, 0, code, errOut)
	assert.FileExists(t, filepath.Join(dir, "from-env", "shop", "shop.go"), "the environment wins over the file")
	assert.NoFileExists(t, filepath.Join(dir, "from-env", "shop", "shop_impl.go"))

	code, _, errOut = soft(t, "go", "-o", filepath.Join(dir, "from-cli"), "-t", "generateModel")
	require.Equal(t, 0, code, errOut)
	assert.FileExists(t, filepath.Join(dir, "from-cli", "shop", "shop_impl.go"), "the command line wins over the environment")
}
