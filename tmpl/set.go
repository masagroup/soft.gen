package tmpl

import (
	"bytes"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"text/template"
)

// Extension is the file extension of a template file.
const Extension = ".tmpl"

// Set is the template set of a language: every template file of its embedded
// file system parsed into one text/template, with the base functions, the
// functions of the language and the include function bound to the set.
type Set struct {
	root     *template.Template
	names    []string
	onRender func(name string)
}

// OnRender sets a function called with the name of every template the set
// renders, file templates and include alike (nil: none). The coverage tests use
// it to know which defines were rendered.
func (s *Set) OnRender(fn func(name string)) {
	s.onRender = fn
}

// Load parses every *.tmpl file found under fsys (recursively) into one set.
// funcs adds or overrides the functions of BaseFuncs; it may be nil.
func Load(fsys fs.FS, funcs template.FuncMap) (*Set, error) {
	set := &Set{}
	all := BaseFuncs()
	all["include"] = set.include
	for name, fn := range funcs {
		all[name] = fn
	}
	root := template.New("").Funcs(all).Option("missingkey=error")
	paths, err := templatePaths(fsys)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("template %s: %w", path, err)
		}
		if _, err := root.New(path).Parse(string(content)); err != nil {
			return nil, fmt.Errorf("template %s: %w", path, err)
		}
	}
	set.root = root
	for _, t := range root.Templates() {
		if t.Tree != nil && !strings.HasSuffix(t.Name(), Extension) {
			set.names = append(set.names, t.Name())
		}
	}
	sort.Strings(set.names)
	return set, nil
}

// templatePaths lists the *.tmpl files of fsys, sorted, with slash separators.
func templatePaths(fsys fs.FS) ([]string, error) {
	var paths []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, Extension) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

// Render executes the named template with data and returns its text.
func (s *Set) Render(name string, data any) ([]byte, error) {
	t := s.root.Lookup(name)
	if t == nil {
		return nil, fmt.Errorf("template %q not found", name)
	}
	if s.onRender != nil {
		s.onRender(name)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Has reports whether the set defines the named template.
func (s *Set) Has(name string) bool {
	return s.root.Lookup(name) != nil
}

// Names lists the defined template names, sorted (the file templates
// themselves, named after their path, are left out).
func (s *Set) Names() []string {
	return append([]string(nil), s.names...)
}

// include executes the named template with data and returns its text, so a
// template can choose a variant by name:
//
//	{{ include (printf "feature/getter.%s" .Strategy) . }}
func (s *Set) include(name string, data any) (string, error) {
	text, err := s.Render(name, data)
	if err != nil {
		return "", err
	}
	return string(text), nil
}
