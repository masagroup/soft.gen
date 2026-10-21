package typescript

import (
	"embed"
	"io/fs"
	"text/template"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/naming"
	"github.com/masagroup/soft.gen/v2/props"
	"github.com/masagroup/soft.gen/v2/writer"
)

//go:embed templates
var templates embed.FS

// AnnotationSource is the source of the EAnnotations the TypeScript generator
// reads.
const AnnotationSource = "http://net.masagroup/soft/2020/GenTS"

// Language is the TypeScript code generator.
type Language struct{}

// New returns the TypeScript language.
func New() *Language { return &Language{} }

func init() { engine.Register(New()) }

// Name is "ts".
func (*Language) Name() string { return "ts" }

// AnnotationSource is the GenTS annotation source.
func (*Language) AnnotationSource() string { return AnnotationSource }

// Properties declares the properties of the TypeScript generator.
func (*Language) Properties() []props.Decl {
	pkg := props.ScopePackage
	return []props.Decl{
		{Name: "packageDir", Doc: "the output directory of the package: a replacement of its name ($1), e.g. $1/src", Scope: pkg},
		{Name: "packages", Doc: "the modules of the packages: <nsURI>=<interface module>[#<implementation module>], or a module keyed by its last segment; several separated by , ; or spaces; merged with the package annotation", Scope: pkg},
		{Name: "paths", Doc: "the import paths of modules: <module>=<path>, several separated by , ; or spaces (the module of the package is ./internal.js)", Scope: pkg},
		{Name: "namespaces", Doc: "the namespaces of modules: <module>=<namespace>, or <module> for named imports; default: the last segment of the module, ecore for @masagroup/ecore", Scope: pkg},
		{Name: "moduleResolution", Default: "nodenext", Doc: "nodenext or node16: the relative import paths end with .js", Scope: pkg},
		{Name: "allowSyntheticDefaultImports", Default: "false", Doc: "true: a namespace is imported as a default import (import ns from)", Scope: pkg},
		{Name: "testFramework", Default: "jest", Doc: "jest or vitest: the tests import vitest", Scope: pkg},
		{Name: "accessors", Default: "method", Doc: "method, property or both (comma separated): the accessors of the features; also a feature key", Scope: pkg},
		{Name: "async", Default: "false", Doc: "true: the proxies and operations get an Async variant returning a Promise; also an element key", Scope: pkg},
		{Name: "featureDelegation", Doc: "Reflective: the implementations delegate their features to the reflective store", Scope: pkg},
		{Name: "packageInstance", Doc: "interfaces: the package instance is generated with the interfaces and shared by the implementations", Scope: pkg},
		{Name: "checkImplementation", Default: "true", Doc: "false: the setters of a shared package do not check that the objects share the implementation", Scope: pkg},
		{Name: "rootExtendsInterface", Doc: "the interface the root interfaces extend, as <module>/<name> (default: EObject of @masagroup/ecore)", Scope: pkg},
		{Name: "rootExtendsClass", Doc: "the class the root implementations extend, as <module>/<name> (default: EObjectImpl of @masagroup/ecore)", Scope: pkg},
		{Name: "printWidth", Default: "120", Doc: "the line width of the generated code, as prettier's printWidth: the lists the generator builds (imports, parameters, exports) are wrapped beyond it", Scope: pkg},
	}
}

// EntryPoints lists the templates of the -t option, the default ones first,
// as in v1.
func (l *Language) EntryPoints() []engine.EntryPoint {
	return []engine.EntryPoint{
		{Name: "generateModel", Default: true, Doc: "the interfaces, enums, constants, package and factory, and the implementations of the classes", Files: l.modelFiles},
		{Name: "generateModule", Default: true, Doc: "the module files: internal.ts (every export, in dependency order) and index.ts (the public names)", Files: l.moduleFiles},
		{Name: "generateTests", Default: true, Doc: "the tests of the implementations and of the factory", Files: l.testFiles},
		{Name: "generateInterfaces", Doc: "the interfaces, enums, constants, package and factory interfaces; with packageInstance=interfaces, also the shared package, a factory of the data types and the module files", Files: l.interfaceFiles},
		{Name: "generateImplementations", Doc: "the implementations of the classes and the factory; with packageInstance=interfaces, the module files of the implementation, else the package implementation", Files: l.implementationFiles},
	}
}

// Templates is the embedded templates directory.
func (*Language) Templates() fs.FS {
	sub, err := fs.Sub(templates, "templates")
	if err != nil {
		panic(err)
	}
	return sub
}

// Funcs adds the TypeScript layout functions: an indentation of 4 spaces and
// double-quoted string literals.
func (*Language) Funcs() template.FuncMap {
	return template.FuncMap{
		"indent": indent,
		"quote":  quote,
		"lines":  lines,
	}
}

// Regions is nil: the TypeScript files are generated whole, a hand-written
// type extends them instead (<Class>Ext.ts).
func (*Language) Regions() *writer.RegionSyntax { return nil }

// Formatter formats the files with prettier, when --format is given and
// prettier is on the PATH; the generated code already has its layout.
func (*Language) Formatter() writer.Formatter {
	return writer.ExecFormatter{Command: "prettier", Args: []string{"--stdin-filepath", "{path}"}, Extensions: []string{".ts"}}
}

// NameStyle is the regex style (pattern:replacement with $1).
func (*Language) NameStyle() naming.Style { return naming.Regex }

// modelFiles is generateModel: the interface of every class and, for a
// concrete class, its implementation; the enums; the factory interface and
// implementation, the constants, the package interface and implementation.
func (*Language) modelFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	var files []engine.File
	for _, class := range g.m.Package(pkg).AllClasses {
		files = append(files, g.interfaceFile(class))
		if !class.IsInterface() {
			files = append(files, g.implementationFile(class))
		}
	}
	for _, enum := range g.m.Package(pkg).AllEnums {
		files = append(files, g.enumFile(enum))
	}
	files = append(files, g.factoryInterfaceFile(pkg), g.factoryImplementationFile(pkg))
	files = append(files, g.constantsFile(pkg), g.packageInterfaceFile(pkg), g.packageImplementationFile(pkg))
	return files, nil
}

// interfaceFiles is generateInterfaces: the interfaces, the enums, the factory
// interface, the constants and the package interface; for a shared package
// also the package implementation, the factory of the data types and the
// module files of the interfaces.
func (*Language) interfaceFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	var files []engine.File
	for _, class := range g.m.Package(pkg).AllClasses {
		files = append(files, g.interfaceFile(class))
	}
	for _, enum := range g.m.Package(pkg).AllEnums {
		files = append(files, g.enumFile(enum))
	}
	files = append(files, g.factoryInterfaceFile(pkg), g.constantsFile(pkg), g.packageInterfaceFile(pkg))
	if g.isSharedPackage(pkg) {
		files = append(files, g.packageImplementationFile(pkg), g.dataTypeFactoryFile(pkg))
		files = append(files, g.interfacesInternalFile(pkg), g.interfacesIndexFile(pkg))
	}
	return files, nil
}

// implementationFiles is generateImplementations: the implementations of the
// concrete classes and the factory; the module files of the implementation
// for a shared package, else the package implementation.
func (*Language) implementationFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	var files []engine.File
	for _, class := range g.m.Package(pkg).AllClasses {
		if !class.IsInterface() {
			files = append(files, g.implementationFile(class))
		}
	}
	files = append(files, g.factoryImplementationFile(pkg))
	if g.isSharedPackage(pkg) {
		files = append(files, g.implementationsInternalFile(pkg), g.implementationsIndexFile(pkg))
	} else {
		files = append(files, g.packageImplementationFile(pkg))
	}
	return files, nil
}

// moduleFiles is generateModule: internal.ts and index.ts.
func (*Language) moduleFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	return []engine.File{g.internalFile(pkg), g.indexFile(pkg)}, nil
}

// testFiles is generateTests: the tests of the concrete classes, but with a
// reflective feature delegation, and of the factory.
func (*Language) testFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	var files []engine.File
	for _, class := range g.m.Package(pkg).AllClasses {
		if !class.IsInterface() && !g.IsReflectiveDelegation() {
			files = append(files, g.implementationTestFile(class))
		}
	}
	files = append(files, g.factoryTestFile(pkg))
	return files, nil
}
