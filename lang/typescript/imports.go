package typescript

import (
	"sort"
	"strings"

	"github.com/masagroup/soft.go/ecore"
)

// A qualified identifier names a type, a function or a member of another
// file: "<module>/<name>", where the module may hold "/" ("@masagroup/ecore/EList",
// "tournament/Team"), the name may be followed by members
// ("tournament/TournamentConstants.TEAM__GROUP") and type arguments
// ("@masagroup/ecore/EMap<string,number>"); an identifier without "/" is
// written as is ("string", "number"). The builders pass every identifier of
// another file through the scope of their file, which records the import and
// returns how the file writes it: the bare name for a module imported by name
// (the module of the package), <namespace>.<name> for a module imported as a
// namespace (ecore.EList). The imports of a file are thus exactly the ones it
// uses.

// scope collects the imports of a file.
type scope struct {
	g *gen
	// named holds the names imported from each module imported by name.
	named map[string]map[string]bool
	// declared holds the qualified identifiers the file declares: never imported.
	declared map[string]bool
	// scanned are the modules whose names are found in the rendered body.
	scanned []scannedModule
	// namespaces holds the modules imported as a namespace.
	namespaces map[string]bool
}

// newScope starts the scope of a file.
func (g *gen) newScope() *scope {
	return &scope{g: g, named: map[string]map[string]bool{}, namespaces: map[string]bool{}, declared: map[string]bool{}}
}

// qualify returns how the file writes a qualified identifier, and records its
// import (v1 getImportedIdentifier).
func (s *scope) qualify(qid string) string {
	base, args := qid, ""
	if open := strings.IndexByte(qid, '<'); open >= 0 {
		end := strings.LastIndexByte(qid, '>')
		var qualified []string
		for _, arg := range splitTypeArguments(qid[open+1 : end]) {
			qualified = append(qualified, s.qualify(strings.TrimSpace(arg)))
		}
		base, args = qid[:open], "<"+strings.Join(qualified, ", ")+">"+qid[end+1:]
	}
	slash := strings.LastIndexByte(base, '/')
	if slash < 0 {
		return base + args
	}
	module, name := base[:slash], base[slash+1:]
	imported, _, _ := strings.Cut(name, ".")
	if s.declared[qualified(module, imported)] {
		return name + args
	}
	if namespace := s.g.moduleNamespace(module); namespace != "" {
		s.namespaces[module] = true
		return namespace + "." + name + args
	}
	if s.named[module] == nil {
		s.named[module] = map[string]bool{}
	}
	s.named[module][imported] = true
	return name + args
}

// declare records a qualified identifier the file declares: the file writes
// it by its name and does not import it.
func (s *scope) declare(qid string) *scope {
	s.declared[qid] = true
	return s
}

// splitTypeArguments splits type arguments on their top-level commas.
func splitTypeArguments(args string) []string {
	var parts []string
	depth, start := 0, 0
	for i, r := range args {
		switch r {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, args[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, args[start:])
}

// imports returns the import declarations of the file, one per module in the
// order of the modules (v1 generateImports): a named import, its names sorted
// and wrapped beyond the print width, or a namespace import (a default import
// with allowSyntheticDefaultImports). Empty when the file imports nothing.
func (s *scope) imports() string {
	modules := make([]string, 0, len(s.named)+len(s.namespaces))
	for module := range s.named {
		modules = append(modules, module)
	}
	for module := range s.namespaces {
		modules = append(modules, module)
	}
	sort.Strings(modules)
	var b strings.Builder
	for _, module := range modules {
		path := quote(s.g.modulePath(module))
		if s.namespaces[module] {
			namespace := s.g.moduleNamespace(module)
			if s.g.propertyIsTrue("allowSyntheticDefaultImports") {
				b.WriteString("import " + namespace + " from " + path + "\n")
			} else {
				b.WriteString("import * as " + namespace + " from " + path + "\n")
			}
			continue
		}
		names := make([]string, 0, len(s.named[module]))
		for name := range s.named[module] {
			names = append(names, name)
		}
		sort.Strings(names)
		b.WriteString(wrapList("import {", names, "} from "+path, 0, s.g.printWidth()) + "\n")
	}
	return b.String()
}

// The qualified identifiers of the elements of the model.

// qualified joins a module and an identifier.
func qualified(module, identifier string) string { return module + "/" + identifier }

// interfaceIdentifier is an identifier of the interface module of a package.
func (g *gen) interfaceIdentifier(p ecore.EPackage, identifier string) string {
	return qualified(g.interfaceModule(p), identifier)
}

// implementationIdentifier is an identifier of the implementation module of a package.
func (g *gen) implementationIdentifier(p ecore.EPackage, identifier string) string {
	return qualified(g.implementationModule(p), identifier)
}

// ecoreIdentifier is an identifier of the ecore module.
func (g *gen) ecoreIdentifier(identifier string) string {
	return g.interfaceIdentifier(g.m.Ecore(), identifier)
}

// classInterfaceIdentifier is the interface of a class.
func (g *gen) classInterfaceIdentifier(c ecore.EClass) string {
	return g.interfaceIdentifier(c.GetEPackage(), g.interfaceName(c))
}

// classImplementationIdentifier is the implementation of a class.
func (g *gen) classImplementationIdentifier(c ecore.EClass) string {
	return g.implementationIdentifier(c.GetEPackage(), g.implementationName(c))
}

// classExtensionIdentifier is the hand-written extension of a class.
func (g *gen) classExtensionIdentifier(c ecore.EClass) string {
	return g.implementationIdentifier(c.GetEPackage(), g.extensionName(c))
}

// enumIdentifier is an enum.
func (g *gen) enumIdentifier(e ecore.EEnum) string {
	return g.interfaceIdentifier(e.GetEPackage(), e.GetName())
}

// constantsIdentifier is a member of the constants of a package.
func (g *gen) constantsIdentifier(p ecore.EPackage, member string) string {
	return g.interfaceIdentifier(p, g.constantsName(p)+"."+member)
}

// packageGetterIdentifier is the function returning the package instance
// (getLibraryPackage).
func (g *gen) packageGetterIdentifier(p ecore.EPackage) string {
	return g.interfaceIdentifier(p, "get"+g.packageName(p))
}

// scannedModule is a module of functions the templates write by their name
// (vitest, ts-mockito): its names are imported when the body calls them.
type scannedModule struct {
	module string
	// defaultName is the name of a default import ("deepEqual"), else the
	// names are imported by name.
	defaultName string
	names       []string
}

// scan adds a module whose names are found in the rendered body; the
// scanned modules come first, in the order they are added (v1 wrote them
// before the imports of the model).
func (s *scope) scan(module, defaultName string, names ...string) {
	s.scanned = append(s.scanned, scannedModule{module: module, defaultName: defaultName, names: names})
}

// scannedImports are the imports of the scanned modules whose names the body
// calls (name( or name<).
func (s *scope) scannedImports(body string) string {
	var b strings.Builder
	for _, m := range s.scanned {
		if m.defaultName != "" {
			if calls(body, m.defaultName) {
				b.WriteString("import " + m.defaultName + " from " + quote(m.module) + "\n")
			}
			continue
		}
		var used []string
		for _, name := range m.names {
			if calls(body, name) {
				used = append(used, name)
			}
		}
		if len(used) > 0 {
			sort.Strings(used)
			b.WriteString(wrapList("import {", used, "} from "+quote(m.module), 0, s.g.printWidth()) + "\n")
		}
	}
	return b.String()
}

// calls reports whether body calls a function: name followed by ( or <, not
// preceded by a letter, a digit, _, $ or a dot.
func calls(body, name string) bool {
	for i := 0; ; {
		j := strings.Index(body[i:], name)
		if j < 0 {
			return false
		}
		start, end := i+j, i+j+len(name)
		before := byte(' ')
		if start > 0 {
			before = body[start-1]
		}
		if end < len(body) && (body[end] == '(' || body[end] == '<') && !(before == '.' || before == '_' || before == '$' || isAlphanumeric(before)) {
			return true
		}
		i = end
	}
}
