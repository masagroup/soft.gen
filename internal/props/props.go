package props

import (
	"fmt"
	"sort"
	"strings"
)

// Lookup finds a key in the generator annotation of the root package. It is
// how the engine gives the package the annotation layer without a dependency on
// the model.
type Lookup func(key string) (string, bool)

// Layer is where a value came from.
type Layer int

const (
	// LayerNone means no layer provided a non-empty value.
	LayerNone Layer = iota
	// LayerCLI is a -p key=value entry or a value set by the engine with Set.
	LayerCLI
	// LayerFile is a -P property file.
	LayerFile
	// LayerAnnotation is the generator annotation of the root package.
	LayerAnnotation
	// LayerDefault is the default of the declaration.
	LayerDefault
)

// String returns the layer name.
func (l Layer) String() string {
	switch l {
	case LayerNone:
		return "none"
	case LayerCLI:
		return "cli"
	case LayerFile:
		return "file"
	case LayerAnnotation:
		return "annotation"
	case LayerDefault:
		return "default"
	default:
		return fmt.Sprintf("Layer(%d)", int(l))
	}
}

// Properties is the resolved, layered property store of a run.
//
// A key is resolved from the first layer giving it a non-empty value, in this
// order: the values set by the engine (Set), the command line, the property
// files (the first file defining the key wins), the generator annotation of the
// root package (declared Package-scoped keys only) and the declared default. An
// empty value at a layer never hides the layers below, as in the v1 generators
// where getPackageProperty tested getProperty(key) <> ”.
//
// Keys given by the command line or a property file but not declared are
// resolved like any other key (v1 accepted anything) and reported by Unknown so
// the engine can warn; they are never read from the annotation.
//
// A Properties is safe for concurrent reads once built; Set must not run
// concurrently with the other methods.
type Properties struct {
	decls      map[string]Decl
	set        map[string]string
	cli        map[string]string
	files      map[string]string
	annotation Lookup
}

// New builds the store. decls are the declarations of the language; the
// Reserved declarations are added unless the language declares a property of
// the same name, and a duplicate or empty name is an error. cli holds the -p
// key=value entries (see ParseKeyValues); files are the -P paths, loaded with
// ParseProperties, the first file defining a key winning; annotation reads the
// generator annotation of the root package and may be nil.
func New(decls []Decl, cli map[string]string, files []string, annotation Lookup) (*Properties, error) {
	p := &Properties{
		decls:      make(map[string]Decl, len(decls)+4),
		set:        map[string]string{},
		cli:        make(map[string]string, len(cli)),
		files:      map[string]string{},
		annotation: annotation,
	}
	for _, d := range decls {
		if d.Name == "" {
			return nil, fmt.Errorf("props: declaration with an empty name")
		}
		if _, dup := p.decls[d.Name]; dup {
			return nil, fmt.Errorf("props: property %q declared twice", d.Name)
		}
		p.decls[d.Name] = d
	}
	for _, d := range Reserved() {
		if _, declared := p.decls[d.Name]; !declared {
			p.decls[d.Name] = d
		}
	}
	for k, v := range cli {
		p.cli[k] = v
	}
	for _, path := range files {
		values, err := ParseProperties(path)
		if err != nil {
			return nil, err
		}
		for k, v := range values {
			if _, defined := p.files[k]; !defined {
				p.files[k] = v
			}
		}
	}
	return p, nil
}

// Get returns the resolved value of key, or "" when no layer provides one.
func (p *Properties) Get(key string) string {
	value, _, _ := p.Lookup(key)
	return value
}

// Lookup resolves key and reports the layer the value came from. When no layer
// provides a non-empty value, value is "", layer LayerNone and ok false.
func (p *Properties) Lookup(key string) (value string, layer Layer, ok bool) {
	if v := p.set[key]; v != "" {
		return v, LayerCLI, true
	}
	if v := p.cli[key]; v != "" {
		return v, LayerCLI, true
	}
	if v := p.files[key]; v != "" {
		return v, LayerFile, true
	}
	decl, declared := p.decls[key]
	if declared && decl.Scope == ScopePackage && p.annotation != nil {
		if v, found := p.annotation(key); found && v != "" {
			return v, LayerAnnotation, true
		}
	}
	if declared && decl.Default != "" {
		return decl.Default, LayerDefault, true
	}
	return "", LayerNone, false
}

// Bool reports whether the resolved value of key is "true".
func (p *Properties) Bool(key string) bool {
	return p.Get(key) == "true"
}

// IsSet reports whether some layer provides a non-empty value for key.
func (p *Properties) IsSet(key string) bool {
	_, _, ok := p.Lookup(key)
	return ok
}

// List splits the resolved value of key on any rune of seps, trims each item
// and drops the empty ones. With empty seps the whole trimmed value is the only
// item. An unset key gives nil.
func (p *Properties) List(key, seps string) []string {
	var items []string
	for _, item := range strings.FieldsFunc(p.Get(key), func(r rune) bool { return strings.ContainsRune(seps, r) }) {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// Set gives key a value decided by the engine (nsURI, templates); it has the
// highest precedence. Like the other layers, an empty value falls through.
func (p *Properties) Set(key, value string) {
	p.set[key] = value
}

// Decls returns the declarations, Reserved ones included, sorted by name.
func (p *Properties) Decls() []Decl {
	decls := make([]Decl, 0, len(p.decls))
	for _, d := range p.decls {
		decls = append(decls, d)
	}
	return sortDecls(decls)
}

// Unknown returns the keys given by the command line or a property file but not
// declared, sorted.
func (p *Properties) Unknown() []string {
	var keys []string
	seen := map[string]bool{}
	for _, layer := range []map[string]string{p.cli, p.files} {
		for k := range layer {
			if _, declared := p.decls[k]; !declared && !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	return keys
}

// Annotation returns the raw annotation layer for key, whatever its
// declaration, for languages that merge the command-line and annotation values
// (for example the Go packages property). Without a Lookup nothing is found.
func (p *Properties) Annotation(key string) (string, bool) {
	if p.annotation == nil {
		return "", false
	}
	return p.annotation(key)
}

// CLI returns the raw command-line or file layer for key, for the same purpose
// as Annotation: the -p entry when there is one, else the value of the first
// file defining the key, even when empty. The values given by Set are not part
// of it.
func (p *Properties) CLI(key string) (string, bool) {
	if v, found := p.cli[key]; found {
		return v, true
	}
	if v, found := p.files[key]; found {
		return v, true
	}
	return "", false
}
