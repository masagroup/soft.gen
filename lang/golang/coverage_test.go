package golang

import (
	"context"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/tmpl"
)

// unreachableTemplates lists the defines no model can render, with the reason.
// The coverage test fails when one of them becomes reachable, so the list
// stays exact.
var unreachableTemplates = map[string]string{
	"feature/basicget.container": "a basic getter exists for a single proxy, and a proxy is never a container",
}

// TestTemplateCoverage checks that the fixtures, together, render every
// template define of the Go language but the unreachable ones.
func TestTemplateCoverage(t *testing.T) {
	if len(fixtures) == 0 {
		t.Skip("no fixture registered")
	}
	lang := New()
	set, err := tmpl.Load(lang.Templates(), lang.Funcs())
	if err != nil {
		t.Fatal(err)
	}
	rendered := map[string]bool{}
	for _, f := range sortedFixtures() {
		req := engine.Request{
			Lang: lang, ModelPath: f.modelPath(), OutDir: t.TempDir(),
			Props: f.Props, PropFiles: []string{fixtureProps}, EntryPoints: f.EntryPoints,
			Silent: true, DryRun: true, Output: io.Discard, ErrOutput: io.Discard,
			OnRender: func(name string) { rendered[name] = true },
		}
		if _, err := engine.Run(context.Background(), req); err != nil {
			t.Fatalf("%s: %v", f.Name, err)
		}
	}
	var missing, stale []string
	for _, name := range set.Names() {
		_, unreachable := unreachableTemplates[name]
		switch {
		case !rendered[name] && !unreachable:
			missing = append(missing, name)
		case rendered[name] && unreachable:
			stale = append(stale, name)
		}
	}
	for name := range unreachableTemplates {
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
		t.Errorf("unreachableTemplates entries that are rendered or undefined:\n  %s", strings.Join(stale, "\n  "))
	}
}
