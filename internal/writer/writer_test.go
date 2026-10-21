package writer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goRegions = RegionSyntax{Start: "// Start of user code %s", End: "// End of user code"}

func TestWriteHeaderAndNewlines(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	w := New(dir, Output(&out), Header(func(string) string { return "// Code generated. DO NOT EDIT.\n\n" }))
	status, err := w.Write("pkg/a.go", []byte("package a\r\n\r\nvar x = 1"), WithHeader)
	require.NoError(t, err)
	assert.Equal(t, Written, status)
	content, err := os.ReadFile(filepath.Join(dir, "pkg", "a.go"))
	require.NoError(t, err)
	assert.Equal(t, "// Code generated. DO NOT EDIT.\n\npackage a\n\nvar x = 1\n", string(content))
	assert.Contains(t, out.String(), "file generated: ")
	assert.Contains(t, out.String(), filepath.Join("pkg", "a.go"))
}

func TestWriteSilentAndDryRun(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	w := New(dir, Output(&out), Silent(true), DryRun(true))
	status, err := w.Write("a.txt", []byte("x\n"), 0)
	require.NoError(t, err)
	assert.Equal(t, Skipped, status)
	assert.Empty(t, out.String())
	assert.NoFileExists(t, filepath.Join(dir, "a.txt"))
}

func TestWriteDiff(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644))
	var out bytes.Buffer
	w := New(dir, Output(&out), Diff(true))
	status, err := w.Write("a.txt", []byte("one\ntwo\n"), 0)
	require.NoError(t, err)
	assert.Equal(t, Same, status)
	assert.Empty(t, out.String())
	status, err = w.Write("a.txt", []byte("one\nthree\n"), 0)
	require.NoError(t, err)
	assert.Equal(t, Different, status)
	assert.Contains(t, out.String(), "-two\n+three\n")
	content, _ := os.ReadFile(filepath.Join(dir, "a.txt"))
	assert.Equal(t, "one\ntwo\n", string(content), "diff mode writes nothing")
}

func TestWriteFormatsGo(t *testing.T) {
	dir := t.TempDir()
	w := New(dir, Silent(true), Format(GoFormatter{}))
	_, err := w.Write("a.go", []byte("package a\nvar   x   =   1\n"), 0)
	require.NoError(t, err)
	content, _ := os.ReadFile(filepath.Join(dir, "a.go"))
	assert.Equal(t, "package a\n\nvar x = 1\n", string(content))
	_, err = w.Write("a.go", []byte("package a\nvar x = \n"), 0)
	assert.Error(t, err, "unparsable code is an error")
}

func TestWritePreservesRegions(t *testing.T) {
	dir := t.TempDir()
	existing := "// header\n\ntype A interface {\n\t// Start of user code A\n\tCustom()\n\t// End of user code\n}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte(existing), 0o644))
	w := New(dir, Silent(true), Regions(&goRegions))
	generated := "// header\n\ntype A interface {\n\tGetName() string\n\t// Start of user code A\n\t// End of user code\n}\n"
	_, err := w.Write("a.go", []byte(generated), WithRegions)
	require.NoError(t, err)
	content, _ := os.ReadFile(filepath.Join(dir, "a.go"))
	assert.Equal(t, "// header\n\ntype A interface {\n\tGetName() string\n\t// Start of user code A\n\tCustom()\n\t// End of user code\n}\n", string(content))
}

func TestPreserveRegions(t *testing.T) {
	existing := []byte("a\n// Start of user code one\nkept 1\n// End of user code\nb\n// Start of user code gone\nold\n// End of user code\n")
	generated := []byte("a\n// Start of user code one\n// End of user code\nc\n// Start of user code two\ndefault 2\n// End of user code\n")
	merged, lost, err := PreserveRegions(goRegions, existing, generated)
	require.NoError(t, err)
	assert.Equal(t, "a\n// Start of user code one\nkept 1\n// End of user code\nc\n// Start of user code two\ndefault 2\n// End of user code\n", string(merged))
	assert.Equal(t, []string{"gone"}, lost)

	same, lost, err := PreserveRegions(goRegions, []byte("no regions\n"), generated)
	require.NoError(t, err)
	assert.Equal(t, generated, same)
	assert.Empty(t, lost)

	_, _, err = PreserveRegions(goRegions, existing, []byte("// Start of user code x\nunterminated\n"))
	assert.Error(t, err)
}

func TestRegionSyntax(t *testing.T) {
	assert.Equal(t, "// Start of user code EClass", goRegions.Open("EClass"))
	assert.Equal(t, "// End of user code", goRegions.Close())
}

func TestUnifiedDiff(t *testing.T) {
	assert.Empty(t, UnifiedDiff("f", []byte("a\nb\n"), []byte("a\nb\n")))
	diff := UnifiedDiff("f", []byte("1\n2\n3\n4\n5\n6\n7\n8\n9\n"), []byte("1\n2\n3\n4\nfive\n6\n7\n8\n9\n"))
	assert.Equal(t, "--- f\n+++ f\n@@ -2,7 +2,7 @@\n 2\n 3\n 4\n-5\n+five\n 6\n 7\n 8\n", diff)
	diff = UnifiedDiff("f", nil, []byte("new\n"))
	assert.Equal(t, "--- f\n+++ f\n@@ -1,0 +1,1 @@\n+new\n", diff)
}

func TestExecFormatterUnavailable(t *testing.T) {
	f := ExecFormatter{Command: "soft-gen-no-such-formatter-xyz"}
	assert.False(t, f.Available())
	_, err := f.Format("a.ts", []byte("x"))
	assert.Error(t, err)
	skipped := ExecFormatter{Command: "whatever", Extensions: []string{".ts"}}
	out, err := skipped.Format("a.go", []byte("x"))
	require.NoError(t, err)
	assert.Equal(t, "x", string(out))
}
