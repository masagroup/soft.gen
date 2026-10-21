package goldentest

import (
	"context"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/tmpl"
)

// Fixture is a golden case of a language: a model, the entry points and the
// properties of a generation, and how it compares with the v1 generator.
type Fixture struct {
	// Name is the case, and the directory of its expected files.
	Name string
	// Model is the path of the model, relative to the models directory of the
	// suite.
	Model string
	// EntryPoints are the -t options; empty means the default ones.
	EntryPoints []string
	// Props are the -p options.
	Props map[string]string
	// V1Differences lists the expected files where v2 differs from v1 on
	// purpose (slash-separated path -> reason): the v1 cross-check leaves them
	// out and -v1update keeps them.
	V1Differences map[string]string
	// V1Skip, when set, is why the v1 cross-check leaves the case out: v1
	// cannot generate it (its output does not parse or does not compile, or
	// it fails); the expected files are v2's output, reviewed.
	V1Skip string
}

// Suite holds the fixtures of a language and runs the golden tests on them:
// the comparison with the expected files, the template coverage and the v1
// cross-check. A language declares one suite in its tests; each topic file
// registers its fixtures from init.
type Suite struct {
	// Lang is the language of the fixtures.
	Lang engine.Language
	// ModelsDir is the directory of the models (testdata/models).
	ModelsDir string
	// GoldenDir is the directory of the expected files, one sub-directory per
	// fixture (testdata/golden/<language>).
	GoldenDir string
	// PropFiles are the -P files of every fixture.
	PropFiles []string
	// Unreachable lists the template defines no model can render, with the
	// reason: the coverage test fails when one of them becomes reachable, so
	// the list stays exact.
	Unreachable map[string]string

	fixtures []Fixture
}

// Register adds fixtures to the suite.
func (s *Suite) Register(fixtures ...Fixture) {
	s.fixtures = append(s.fixtures, fixtures...)
}

// Fixtures returns the fixtures sorted by name.
func (s *Suite) Fixtures() []Fixture {
	sorted := append([]Fixture(nil), s.fixtures...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	return sorted
}

// ModelPath is the path of the model of a fixture.
func (s *Suite) ModelPath(f Fixture) string {
	return filepath.Join(s.ModelsDir, filepath.FromSlash(f.Model))
}

// ExpectedDir is the directory of the expected files of a fixture.
func (s *Suite) ExpectedDir(f Fixture) string { return filepath.Join(s.GoldenDir, f.Name) }

// TestFixtures compares the generation of every fixture with its expected
// files (Run; -update writes them).
func (s *Suite) TestFixtures(t *testing.T) {
	t.Helper()
	if len(s.fixtures) == 0 {
		t.Skip("no fixture registered")
	}
	names := map[string]bool{}
	for _, f := range s.Fixtures() {
		if names[f.Name] {
			t.Fatalf("fixture %q registered twice", f.Name)
		}
		names[f.Name] = true
		Run(t, Case{
			Name: f.Name, Lang: s.Lang, Model: s.ModelPath(f),
			Props: f.Props, PropFiles: s.PropFiles, EntryPoints: f.EntryPoints,
			Expected: s.ExpectedDir(f),
		})
	}
}

// TestTemplateCoverage checks that the fixtures, together, render every
// template define of the language but the unreachable ones, and that every
// unreachable one is defined and rendered by no fixture.
func (s *Suite) TestTemplateCoverage(t *testing.T) {
	t.Helper()
	if len(s.fixtures) == 0 {
		t.Skip("no fixture registered")
	}
	set, err := tmpl.Load(s.Lang.Templates(), s.Lang.Funcs())
	if err != nil {
		t.Fatal(err)
	}
	rendered := map[string]bool{}
	for _, f := range s.Fixtures() {
		req := engine.Request{
			Lang: s.Lang, ModelPath: s.ModelPath(f), OutDir: t.TempDir(),
			Props: f.Props, PropFiles: s.PropFiles, EntryPoints: f.EntryPoints,
			Silent: true, DryRun: true, Output: io.Discard, ErrOutput: io.Discard,
			OnRender: func(name string) { rendered[name] = true },
		}
		if _, err := engine.Run(context.Background(), req); err != nil {
			t.Fatalf("%s: %v", f.Name, err)
		}
	}
	var missing, stale []string
	for _, name := range set.Names() {
		_, unreachable := s.Unreachable[name]
		switch {
		case !rendered[name] && !unreachable:
			missing = append(missing, name)
		case rendered[name] && unreachable:
			stale = append(stale, name)
		}
	}
	for name := range s.Unreachable {
		if !set.Has(name) {
			stale = append(stale, name+" (not defined)")
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	if len(missing) > 0 {
		t.Errorf("%d of %d templates rendered by no fixture:\n  %s", len(missing), len(set.Names()), strings.Join(missing, "\n  "))
	}
	if len(stale) > 0 {
		t.Errorf("unreachable entries that are rendered or undefined:\n  %s", strings.Join(stale, "\n  "))
	}
}
