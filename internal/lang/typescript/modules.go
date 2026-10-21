package typescript

import (
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/lang/common"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// ecoreModule is the module of the ecore package.
const ecoreModule = "@masagroup/ecore"

// ecoreNamespace is the namespace of the ecore module in the other packages.
const ecoreNamespace = "ecore"

// Indexes of the modules of a package in the packages property
// ("uri=interface#implementation").
const (
	interfaceModuleIndex      = 1
	implementationModuleIndex = 2
)

// packageModules returns the modules of the packages (v1
// getPackageToModuleMap): the ecore package, then the tokens of the packages
// property, merged from the command line and the package annotation.
func (g *gen) packageModules() common.PackageSpecs {
	if g.modules == nil {
		defaults := map[string]string{g.m.Ecore().GetNsURI(): ecoreModule}
		g.modules = common.ParsePackageSpecs(g.ListProperty("packages"), defaults)
	}
	return g.modules
}

// indexModule returns the module of the package at the index (v1
// getIndexModule): from its spec split on "#" (empty parts dropped, as the
// OCL tokenize), else the name of the package.
func (g *gen) indexModule(p ecore.EPackage, index int) string {
	spec, ok := g.packageModules().Spec(p)
	if !ok {
		return p.GetName()
	}
	parts := strings.FieldsFunc(spec, func(r rune) bool { return r == '#' })
	if len(parts) == 0 {
		return p.GetName()
	}
	return common.IndexPart(parts, index)
}

// interfaceModule is the module of the interfaces of a package.
func (g *gen) interfaceModule(p ecore.EPackage) string { return g.indexModule(p, interfaceModuleIndex) }

// implementationModule is the module of the implementations of a package.
func (g *gen) implementationModule(p ecore.EPackage) string {
	return g.indexModule(p, implementationModuleIndex)
}

// currentModule is the module of the generated package: its interface module.
func (g *gen) currentModule() string { return g.interfaceModule(g.pkg) }

// isSharedPackage reports whether the package is shared by several
// implementations (v1 isSharedPackage): packageInstance=interfaces and an
// implementation module distinct from the interface module. Its package and a
// factory of its data types are generated with the interfaces, each
// implementation holds its classes and a factory of its own.
func (g *gen) isSharedPackage(p ecore.EPackage) bool {
	return !g.m.IsEcore(p) && g.PackagePropertyIs("packageInstance", "interfaces") &&
		g.interfaceModule(p) != g.implementationModule(p)
}

// isCheckImplementation reports whether the setters of a shared package check
// that the objects share the implementation (property checkImplementation,
// default true).
func (g *gen) isCheckImplementation(p ecore.EPackage) bool {
	return g.isSharedPackage(p) && !g.PackagePropertyIs("checkImplementation", "false")
}

// isExplicitFileExtension reports whether the relative import paths end with
// .js: moduleResolution nodenext (the default) or node16.
func (g *gen) isExplicitFileExtension() bool {
	resolution, _ := g.PackageProperty("moduleResolution")
	resolution = strings.ToLower(resolution)
	return resolution == "" || resolution == "nodenext" || resolution == "node16"
}

// internalPath is the import path of the internal module of the package:
// "./internal.js" (or "./internal").
func (g *gen) internalPath() string {
	if g.isExplicitFileExtension() {
		return "./internal.js"
	}
	return "./internal"
}

// keyValues parses the tokens "key=value" of a list property; a token
// without "=" gets the value of bare, or is dropped when bare is nil.
func keyValues(list string, bare *string) map[string]string {
	values := map[string]string{}
	for _, token := range common.Tokens(list) {
		if key, value, ok := strings.Cut(token, "="); ok {
			values[key] = value
		} else if bare != nil {
			values[token] = *bare
		}
	}
	return values
}

// modulePaths returns the import path of the modules (v1 getModuleToPathMap):
// the paths property, and the current module mapped to the internal module.
// For a shared package the implementation module is the internal module too,
// and the interface module is unless the paths property gives its path (a run
// generating an implementation gives the path of the interfaces).
func (g *gen) modulePaths() map[string]string {
	if g.paths != nil {
		return g.paths
	}
	paths := keyValues(g.ListProperty("paths"), nil)
	current := g.currentModule()
	if g.isSharedPackage(g.pkg) {
		if _, ok := paths[current]; !ok {
			paths[current] = g.internalPath()
		}
		paths[g.implementationModule(g.pkg)] = g.internalPath()
	} else {
		paths[current] = g.internalPath()
	}
	g.paths = paths
	return paths
}

// modulePath is the import path of a module: its path, else the module.
func (g *gen) modulePath(module string) string {
	if path, ok := g.modulePaths()[module]; ok {
		return path
	}
	return module
}

// moduleNamespaces returns the namespaces of the modules (v1
// getModuleToNamespaceMap): ecore for @masagroup/ecore, none for the current
// module (and the implementation module of a shared package), then the
// namespaces property, where a bare module has none.
func (g *gen) moduleNamespaces() map[string]string {
	if g.namespaces != nil {
		return g.namespaces
	}
	namespaces := map[string]string{ecoreModule: ecoreNamespace, g.currentModule(): ""}
	if g.isSharedPackage(g.pkg) {
		namespaces[g.implementationModule(g.pkg)] = ""
	}
	none := ""
	for module, namespace := range keyValues(g.ListProperty("namespaces"), &none) {
		namespaces[module] = namespace
	}
	g.namespaces = namespaces
	return namespaces
}

// moduleNamespace is the namespace of a module (v1 getModuleNamespace): its
// namespace, else the last segment of a module with a "/", else none: its
// names are imported by name.
func (g *gen) moduleNamespace(module string) string {
	if namespace, ok := g.moduleNamespaces()[module]; ok {
		return namespace
	}
	if !strings.Contains(module, "/") {
		return ""
	}
	return common.LastSegment(module)
}

// packageDirectory is the output directory of a package (v1
// getPackageDirectory): its name replaced by the packageDir property ($1 is
// the name), default the name.
func (g *gen) packageDirectory(p ecore.EPackage) string {
	replacement, ok := g.PackageProperty("packageDir")
	if !ok || replacement == "" {
		return p.GetName()
	}
	rule, err := naming.ParseRule("(.*):"+replacement, naming.Regex)
	if err != nil {
		g.Context.Log.Warn("packageDir", "value", replacement, "error", err)
		return p.GetName()
	}
	return rule.Apply(p.GetName())
}

// filePath is the path of a file of the package: its directory and name.
func (g *gen) filePath(p ecore.EPackage, name string) string {
	return g.packageDirectory(p) + "/" + name
}
