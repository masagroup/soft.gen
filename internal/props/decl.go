package props

import (
	"fmt"
	"sort"
	"strings"
)

// Scope says where a property may come from.
type Scope int

const (
	// ScopeRun restricts a property to the command line and the property files.
	ScopeRun Scope = iota
	// ScopePackage also reads the property from the generator annotation of the
	// root package (v1 getPackageProperty).
	ScopePackage
)

// String returns the scope name ("run" or "package").
func (s Scope) String() string {
	switch s {
	case ScopeRun:
		return "run"
	case ScopePackage:
		return "package"
	default:
		return fmt.Sprintf("Scope(%d)", int(s))
	}
}

// Decl declares a property a language understands.
type Decl struct {
	// Name is the property key, as given on the command line, in the property
	// files and in the generator annotation.
	Name string
	// Default is the value used when no layer provides a non-empty value.
	Default string
	// Doc is the documentation shown by FormatHelp; it may span several lines.
	Doc string
	// Scope says whether the generator annotation of the root package is read.
	Scope Scope
}

// Names of the reserved properties, declared for every language.
const (
	// KeyNsURI is the namespace URI of the generated package, set by the engine.
	KeyNsURI = "nsURI"
	// KeyTemplates is the comma-separated list of the templates to run, set by
	// the engine.
	KeyTemplates = "templates"
	// KeyFileNotes is the banner written at the top of every generated file.
	KeyFileNotes = "fileNotes"
	// KeyDebug enables the debug output of the generator.
	KeyDebug = "debug"
)

// Reserved returns the declarations every language has: nsURI, templates,
// fileNotes, debug. New adds them to the declarations of a language unless the
// language declares a property of the same name.
func Reserved() []Decl {
	return []Decl{
		{Name: KeyNsURI, Doc: "namespace URI of the generated package (set by the generator)"},
		{Name: KeyTemplates, Doc: "comma-separated list of the templates to run (set by the generator)"},
		{Name: KeyFileNotes, Doc: "notes written at the top of each generated file", Scope: ScopePackage},
		{Name: KeyDebug, Default: "false", Doc: "print debug information while generating"},
	}
}

// FormatHelp renders the declarations as the help text of a language: one line
// per declaration with its name, default value and documentation, aligned in
// columns, in the given order. An empty default is shown as "-", a
// Package-scoped declaration is marked "[package]" and the lines of a
// multi-line documentation are indented under the first one.
func FormatHelp(decls []Decl) string {
	if len(decls) == 0 {
		return ""
	}
	nameWidth, defaultWidth := 0, 0
	defaults := make([]string, len(decls))
	for i, d := range decls {
		def := d.Default
		if def == "" {
			def = "-"
		}
		defaults[i] = def
		nameWidth = max(nameWidth, len(d.Name))
		defaultWidth = max(defaultWidth, len(def))
	}
	var sb strings.Builder
	for i, d := range decls {
		doc := strings.Split(strings.TrimRight(d.Doc, "\n"), "\n")
		if d.Scope == ScopePackage {
			doc[0] = strings.TrimSpace(doc[0] + " [package]")
		}
		fmt.Fprintf(&sb, "  %-*s  %-*s  %s\n", nameWidth, d.Name, defaultWidth, defaults[i], doc[0])
		for _, line := range doc[1:] {
			fmt.Fprintf(&sb, "  %-*s  %-*s  %s\n", nameWidth, "", defaultWidth, "", line)
		}
	}
	return sb.String()
}

// sortDecls returns a copy of decls sorted by name.
func sortDecls(decls []Decl) []Decl {
	sorted := make([]Decl, len(decls))
	copy(sorted, decls)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	return sorted
}
