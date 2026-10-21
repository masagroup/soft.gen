package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const shopModel = "../../testdata/models/shop.ecore"

func TestHelpAndVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	assert.Equal(t, 0, run(nil, &out, &errOut))
	assert.Contains(t, out.String(), "Languages:")
	assert.Contains(t, out.String(), "go     soft.generator.go")

	out.Reset()
	assert.Equal(t, 0, run([]string{"go", "-v"}, &out, &errOut))
	assert.Equal(t, "soft.generator.go version: "+developerBuild+"\n", out.String())

	out.Reset()
	Version = "v2.1.0"
	defer func() { Version = developerBuild }()
	assert.Equal(t, 0, run([]string{"-v"}, &out, &errOut))
	assert.Equal(t, "soft-gen version: v2.1.0\n", out.String())

	out.Reset()
	assert.Equal(t, 0, run([]string{"go", "-h"}, &out, &errOut))
	assert.Contains(t, errOut.String(), "generateModel")
	assert.Contains(t, errOut.String(), "featureDelegation")
}

func TestErrors(t *testing.T) {
	var out, errOut bytes.Buffer
	assert.Equal(t, 1, run([]string{"cobol"}, &out, &errOut))
	assert.Contains(t, errOut.String(), `unknown language "cobol"`)

	errOut.Reset()
	assert.Equal(t, 1, run([]string{"go", "-m", shopModel}, &out, &errOut))
	assert.Contains(t, errOut.String(), "required")

	errOut.Reset()
	assert.Equal(t, 1, run([]string{"go", "-m", shopModel, "-o", t.TempDir(), "-p", "novalue"}, &out, &errOut))

	errOut.Reset()
	assert.Equal(t, 1, run([]string{"go", "-m", "missing.ecore", "-o", t.TempDir()}, &out, &errOut))
}

func TestDryRunAndDiff(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	code := run([]string{"go", "-m", shopModel, "-o", dir, "-t", "generateModel", "--dry-run"}, &out, &errOut)
	assert.Equal(t, 0, code, errOut.String())
	assert.Contains(t, out.String(), "file generated: ")
	entries, _ := os.ReadDir(dir)
	assert.Empty(t, entries, "dry run writes nothing")

	out.Reset()
	code = run([]string{"go", "-m", shopModel, "-o", dir, "-t", "generateModel", "--diff", "-s"}, &out, &errOut)
	assert.Equal(t, 1, code, "nothing exists yet: everything differs")
	assert.Contains(t, out.String(), "+++ shop/")

	out.Reset()
	code = run([]string{"go", "-m", shopModel, "-o", dir, "-t", "generateModel", "-s"}, &out, &errOut)
	assert.Equal(t, 0, code, errOut.String())
	assert.FileExists(t, filepath.Join(dir, "shop", "shoppackage.go"))

	out.Reset()
	code = run([]string{"go", "-m", shopModel, "-o", dir, "-t", "generateModel", "--diff"}, &out, &errOut)
	assert.Equal(t, 0, code, "the generated files are up to date")
	assert.Empty(t, out.String())
}
