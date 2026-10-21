// Package goldentest compares the files a language generates for a model with
// the expected files committed in this repository (testdata/golden).
//
// Run with -update to write the generated files as the new expected files of
// every case: the expected directory then holds exactly what the case
// generates. Review the result with git before committing it.
package goldentest

import (
	"bytes"
	"context"
	"flag"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/writer"
)

// update is the -update flag of the test binaries that import this package.
var update = flag.Bool("update", false, "write the generated files as the expected files of the golden cases")

// Updating reports whether the test runs with -update.
func Updating() bool { return *update }

// Case is one golden comparison: a generation request and the directory of
// its expected files.
type Case struct {
	// Name names the sub-test.
	Name string
	// Lang is the language to run.
	Lang engine.Language
	// Model is the .ecore file.
	Model string
	// Props, PropFiles and EntryPoints are the -p, -P and -t options.
	Props       map[string]string
	PropFiles   []string
	EntryPoints []string
	// Expected is the directory of the expected files, laid out as the
	// output directory: it holds exactly the files the case generates.
	Expected string
	// MaxDiffLines bounds the diff printed per file (0: 80 lines).
	MaxDiffLines int
}

// Run generates the case into a temporary directory and compares it with its
// expected files: a unified diff per differing file, the generated files that
// are not expected and the expected files that were not generated. With
// -update it writes the generated files as the expected files instead.
func Run(t *testing.T, c Case) {
	t.Helper()
	t.Run(c.Name, func(t *testing.T) {
		t.Helper()
		out := t.TempDir()
		req := engine.Request{
			Lang: c.Lang, ModelPath: c.Model, OutDir: out,
			Props: c.Props, PropFiles: c.PropFiles, EntryPoints: c.EntryPoints,
			Silent: true, Output: io.Discard, ErrOutput: os.Stderr,
		}
		if _, err := engine.Run(context.Background(), req); err != nil {
			t.Fatalf("generation failed: %v", err)
		}
		if *update {
			if err := ReplaceDir(c.Expected, out); err != nil {
				t.Fatalf("update %s: %v", c.Expected, err)
			}
			return
		}
		if _, err := os.Stat(c.Expected); err != nil {
			t.Fatalf("no expected files at %s: run the test with -update, then review them", c.Expected)
		}
		generated, err := ListFiles(out)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := ListFiles(c.Expected)
		if err != nil {
			t.Fatal(err)
		}
		expectedSet := map[string]bool{}
		for _, file := range expected {
			expectedSet[file] = true
		}
		generatedSet := map[string]bool{}
		for _, file := range generated {
			generatedSet[file] = true
			if !expectedSet[file] {
				t.Errorf("%s: generated but not expected", file)
				continue
			}
			want, _ := os.ReadFile(filepath.Join(c.Expected, filepath.FromSlash(file)))
			got, _ := os.ReadFile(filepath.Join(out, filepath.FromSlash(file)))
			want = bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n"))
			if !bytes.Equal(got, want) {
				t.Errorf("%s differs:\n%s", file, truncate(writer.UnifiedDiff(file, want, got), c.MaxDiffLines))
			}
		}
		for _, file := range expected {
			if !generatedSet[file] {
				t.Errorf("%s: expected but not generated", file)
			}
		}
	})
}

// ReplaceDir makes dst hold exactly the files of src: dst is removed, then
// every file of src is copied into it.
func ReplaceDir(dst, src string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o644)
	})
}

// ListFiles lists the files under root, slash-separated and sorted (nil when
// root does not exist).
func ListFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, p)
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	sort.Strings(files)
	return files, err
}

// truncate keeps the first lines of a diff.
func truncate(diff string, max int) string {
	if max == 0 {
		max = 80
	}
	lines := strings.SplitAfter(diff, "\n")
	if len(lines) <= max {
		return diff
	}
	return strings.Join(lines[:max], "") + "... (" + strconv.Itoa(len(lines)-max) + " more lines)\n"
}
