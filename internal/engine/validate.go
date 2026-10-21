package engine

import (
	"errors"
	"fmt"

	"github.com/masagroup/soft.gen/v2/internal/tmpl"
)

// Validate checks a language before it runs: its names are set, its entry
// points and properties are unique and complete, its templates parse and
// follow the template conventions. Every language runs it in its own tests.
func Validate(lang Language) error {
	var errs []error
	fail := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }
	if lang.Name() == "" {
		fail("the language has no name")
	}
	if lang.AnnotationSource() == "" {
		fail("language %s: no annotation source", lang.Name())
	}
	if len(lang.EntryPoints()) == 0 {
		fail("language %s: no entry point", lang.Name())
	}
	entries := map[string]bool{}
	for _, ep := range lang.EntryPoints() {
		switch {
		case ep.Name == "":
			fail("language %s: an entry point has no name", lang.Name())
		case entries[ep.Name]:
			fail("language %s: entry point %q declared twice", lang.Name(), ep.Name)
		case ep.Files == nil:
			fail("language %s: entry point %q lists no files", lang.Name(), ep.Name)
		}
		entries[ep.Name] = true
	}
	decls := map[string]bool{}
	for _, decl := range lang.Properties() {
		switch {
		case decl.Name == "":
			fail("language %s: a property has no name", lang.Name())
		case decls[decl.Name]:
			fail("language %s: property %q declared twice", lang.Name(), decl.Name)
		case decl.Doc == "":
			fail("language %s: property %q has no documentation", lang.Name(), decl.Name)
		}
		decls[decl.Name] = true
	}
	if lang.Templates() == nil {
		fail("language %s: no templates", lang.Name())
	} else {
		if _, err := tmpl.Load(lang.Templates(), lang.Funcs()); err != nil {
			fail("language %s: %v", lang.Name(), err)
		}
		for _, problem := range tmpl.Lint(lang.Templates(), lang.Funcs()) {
			fail("language %s: template %s", lang.Name(), problem)
		}
	}
	return errors.Join(errs...)
}
