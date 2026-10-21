package typescript

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
)

// The module files of a package (v1 lib/generateModule.mtl): internal.ts
// exports every file of the package in an order where a class follows the
// classes it extends, index.ts the public names, through internal.ts.

// moduleView is the view model of a module file: sections of lines separated
// by a blank line.
type moduleView struct {
	Sections [][]string
}

// exportFrom is the export of every name of a file of the package.
func (g *gen) exportFrom(fileName string) string {
	return `export * from "./` + g.moduleName(fileName) + `"`
}

// interfaceExports are the sections exporting the interfaces, the enums, the
// package, factory and constants files.
func (g *gen) interfaceExports(p ecore.EPackage) [][]string {
	info := g.m.Package(p)
	interfaces := []string{"// interfaces"}
	for _, c := range info.OrderedClasses {
		interfaces = append(interfaces, g.exportFrom(g.interfaceFileName(c, c.GetName())))
	}
	enums := []string{"// enums"}
	for _, e := range info.AllEnums {
		enums = append(enums, g.exportFrom(g.interfaceFileName(e, e.GetName())))
	}
	pkg := []string{"// package - factory- constants",
		g.exportFrom(g.interfaceFileName(p, g.packageName(p))),
		g.exportFrom(g.interfaceFileName(p, g.factoryName(p))),
		g.exportFrom(g.constantsFileName(p)),
	}
	return [][]string{interfaces, enums, pkg}
}

// implementationExports are the exports of the implementations of the
// classes, each followed by its extension.
func (g *gen) implementationExports(p ecore.EPackage) []string {
	var exports []string
	for _, c := range g.m.Package(p).OrderedClasses {
		if c.IsInterface() {
			continue
		}
		exports = append(exports, g.exportFrom(g.implementationFileName(c, c.GetName())))
		if g.HasExtension(c) {
			exports = append(exports, g.exportFrom(g.ElementName(c, c.GetName(), "extensionFileName", "$1Ext.ts")))
		}
	}
	return exports
}

// packageImplementationExports are the exports of the package and factory
// implementations, and of the factory extension.
func (g *gen) packageImplementationExports(p ecore.EPackage) []string {
	exports := []string{
		g.exportFrom(g.implementationFileName(p, g.packageName(p))),
		g.exportFrom(g.implementationFileName(p, g.factoryName(p))),
	}
	if g.KeyIs(p, "factory", "extension") {
		exports = append(exports, g.exportFrom(g.ElementName(p, g.factoryName(p), "extensionFileName", "$1Ext.ts")))
	}
	return exports
}

// internalFile is internal.ts (v1 generateInternal).
func (g *gen) internalFile(p ecore.EPackage) engine.File {
	implementations := append([]string{"// implementations"}, g.implementationExports(p)...)
	implementations = append(implementations, g.packageImplementationExports(p)...)
	view := moduleView{Sections: append(g.interfaceExports(p), implementations)}
	return g.file(g.internalFileName(p), g.newScope(), "module.ts", view)
}

// interfacesInternalFile is internal.ts of the interfaces of a shared
// package (v1 generateInterfacesInternal): the interfaces, the package and the
// factory of its data types.
func (g *gen) interfacesInternalFile(p ecore.EPackage) engine.File {
	dataTypes := append([]string{"// package - factory of the data types"}, g.packageImplementationExports(p)...)
	view := moduleView{Sections: append(g.interfaceExports(p), dataTypes)}
	return g.file(g.internalFileName(p), g.newScope(), "module.ts", view)
}

// implementationsInternalFile is internal.ts of an implementation of a shared
// package (v1 generateImplementationsInternal): the implementations of the
// classes and the factory.
func (g *gen) implementationsInternalFile(p ecore.EPackage) engine.File {
	view := moduleView{Sections: [][]string{
		append([]string{"// implementations"}, g.implementationExports(p)...),
		{"// factory", g.exportFrom(g.implementationFileName(p, g.factoryName(p)))},
	}}
	return g.file(g.internalFileName(p), g.newScope(), "module.ts", view)
}

// publicNames are the names index.ts exports: the interfaces, the enums, the
// package, factory and constants, the getter of the package.
func (g *gen) publicNames(p ecore.EPackage) []string {
	info := g.m.Package(p)
	var names []string
	for _, c := range info.OrderedClasses {
		names = append(names, g.interfaceName(c))
	}
	for _, e := range info.AllEnums {
		names = append(names, e.GetName())
	}
	return append(names, g.packageInterfaceName(p), g.factoryName(p), g.constantsName(p), "get"+g.packageInterfaceName(p))
}

// indexFile is index.ts (v1 generateIndex).
func (g *gen) indexFile(p ecore.EPackage) engine.File {
	return g.exportsFile(p, append(g.publicNames(p), "get"+g.factoryName(p)))
}

// interfacesIndexFile is index.ts of the interfaces of a shared package.
func (g *gen) interfacesIndexFile(p ecore.EPackage) engine.File {
	return g.exportsFile(p, g.publicNames(p))
}

// implementationsIndexFile is index.ts of an implementation of a shared
// package: its factory, getFactory and install.
func (g *gen) implementationsIndexFile(p ecore.EPackage) engine.File {
	return g.exportsFile(p, []string{g.factoryImplementationName(p), "getFactory", "install"})
}

// exportsFile is an index.ts exporting names from internal.ts.
func (g *gen) exportsFile(p ecore.EPackage, names []string) engine.File {
	view := moduleView{Sections: [][]string{{wrapList("export {", names, "} from "+quote(g.internalPath()), 0, g.printWidth())}}}
	return g.file(g.indexFileName(p), g.newScope(), "module.ts", view)
}
