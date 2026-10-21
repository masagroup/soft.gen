package golang

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

// AnnotationSource is the source of the EAnnotations the Go generator reads.
const AnnotationSource = "http://net.masagroup/soft/2019/GenGo"

// Language is the Go code generator.
type Language struct{}

// New returns the Go language.
func New() *Language { return &Language{} }

func init() { engine.Register(New()) }

// Name is "go".
func (*Language) Name() string { return "go" }

// AnnotationSource is the GenGo annotation source.
func (*Language) AnnotationSource() string { return AnnotationSource }

// Properties declares the properties of the Go generator.
func (*Language) Properties() []props.Decl {
	pkg := props.ScopePackage
	return []props.Decl{
		{Name: "packages", Doc: "the paths of the packages: <nsURI>=<interface path>[#<implementation path>[#<mock path>]], or a path keyed by its last segment; several separated by , ; or spaces; merged with the package annotation", Scope: pkg},
		{Name: "module", Doc: "the Go module path: the output directories are the package paths minus the module", Scope: pkg},
		{Name: "featureDelegation", Doc: "Reflective: the implementations delegate their features to the reflective store", Scope: pkg},
		{Name: "packageInstance", Doc: "interfaces: the package instance (GetPackage) is generated with the interfaces and shared by the implementations", Scope: pkg},
		{Name: "checkImplementation", Default: "true", Doc: "false: the setters of a shared package do not check that the objects share the implementation", Scope: pkg},
		{Name: "exported", Default: "false", Doc: "true: the implementation names are exported", Scope: pkg},
		{Name: "goVersion", Default: "1.19", Doc: "the Go version of the generated code: any from 1.18, else interface{}", Scope: pkg},
		{Name: "rootExtendsInterface", Default: "ecore.EObject", Doc: "the interface the root interfaces embed", Scope: pkg},
		{Name: "rootExtendsClass", Default: "ecore.EObjectImpl", Doc: "the struct the root implementations embed", Scope: pkg},
		{Name: "rootExtendsMock", Default: "ecore.MockEObjectInternal", Doc: "the struct the root mocks embed", Scope: pkg},
	}
}

// EntryPoints lists the templates of the Go generator.
func (l *Language) EntryPoints() []engine.EntryPoint {
	return []engine.EntryPoint{
		{Name: "generateModel", Default: true, Doc: "the interfaces, enums, package, factory and implementations of the model", Files: l.modelFiles},
		{Name: "generateInterfaces", Doc: "the interfaces, enums, package and factory interfaces (and the shared package instance)", Files: l.interfaceFiles},
		{Name: "generateImplementations", Doc: "the implementations of the classes, the package (unless shared) and the factory", Files: l.implementationFiles},
		{Name: "generateTests", Default: true, Doc: "the tests of the implementations, the factory and the package", Files: l.testFiles},
		{Name: "generateMocks", Default: true, Doc: "the mocks of the classes", Files: l.mockFiles},
		{Name: "generateMocksTests", Default: true, Doc: "the tests of the mocks", Files: l.mockTestFiles},
	}
}

// Templates is the embedded template directory.
func (*Language) Templates() fs.FS {
	sub, err := fs.Sub(templates, "templates")
	if err != nil {
		panic(err)
	}
	return sub
}

// Funcs adds no function.
func (*Language) Funcs() template.FuncMap { return nil }

// Regions is the syntax of the user code region of the interfaces.
func (*Language) Regions() *writer.RegionSyntax {
	return &writer.RegionSyntax{Start: "// Start of user code %s", End: "// End of user code"}
}

// Formatter formats the files with gofmt.
func (*Language) Formatter() writer.Formatter { return writer.GoFormatter{} }

// NameStyle is the regex style (pattern:replacement with $1).
func (*Language) NameStyle() naming.Style { return naming.Regex }

// modelFiles is generateModel: the interfaces of every class, the enums, the
// package and factory interfaces, the implementations of the concrete
// classes, the package implementation, the data type factory of a shared
// package and the factory implementation.
func (*Language) modelFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	info := g.m.Package(pkg)
	var files []engine.File
	for _, class := range info.AllClasses {
		files = append(files, g.interfaceFile(class)...)
	}
	for _, enum := range info.AllEnums {
		files = append(files, g.enumFile(enum))
	}
	files = append(files, g.packageInterfaceFile(pkg))
	files = append(files, g.factoryInterfaceFile(pkg)...)
	for _, class := range info.AllClasses {
		if !class.IsInterface() {
			files = append(files, g.implementationFile(class)...)
		}
	}
	files = append(files, g.packageImplementationFile(pkg)...)
	if g.isSharedPackage(pkg) {
		files = append(files, g.dataTypeFactoryFile(pkg)...)
	}
	files = append(files, g.factoryImplementationFile(pkg)...)
	return files, nil
}

// interfaceFiles is generateInterfaces: the interfaces, the enums, the package
// and factory interfaces; for a shared package also the package
// implementation and the data type factory, with the interfaces.
func (*Language) interfaceFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	info := g.m.Package(pkg)
	var files []engine.File
	for _, class := range info.AllClasses {
		files = append(files, g.interfaceFile(class)...)
	}
	for _, enum := range info.AllEnums {
		files = append(files, g.enumFile(enum))
	}
	files = append(files, g.packageInterfaceFile(pkg))
	files = append(files, g.factoryInterfaceFile(pkg)...)
	if g.isSharedPackage(pkg) {
		files = append(files, g.packageImplementationFile(pkg)...)
		files = append(files, g.dataTypeFactoryFile(pkg)...)
	}
	return files, nil
}

// implementationFiles is generateImplementations: the implementations of the
// concrete classes, the package implementation unless shared, the factory
// implementation.
func (*Language) implementationFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	var files []engine.File
	for _, class := range g.m.Package(pkg).AllClasses {
		if !class.IsInterface() {
			files = append(files, g.implementationFile(class)...)
		}
	}
	if !g.isSharedPackage(pkg) {
		files = append(files, g.packageImplementationFile(pkg)...)
	}
	files = append(files, g.factoryImplementationFile(pkg)...)
	return files, nil
}

// testFiles is generateTests: the tests of the concrete classes, the factory and the package.
func (*Language) testFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	var files []engine.File
	for _, class := range g.m.Package(pkg).AllClasses {
		if !class.IsInterface() {
			files = append(files, g.implementationTestFile(class)...)
		}
	}
	files = append(files, g.factoryTestFile(pkg)...)
	files = append(files, g.packageTestFile(pkg)...)
	return files, nil
}

// mockFiles is generateMocks: a mock per class, interfaces included.
func (*Language) mockFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	var files []engine.File
	for _, class := range g.m.Package(pkg).AllClasses {
		files = append(files, g.mockFile(class)...)
	}
	return files, nil
}

// mockTestFiles is generateMocksTests: the tests of the mock of every class.
func (*Language) mockTestFiles(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
	g := newGen(ctx)
	var files []engine.File
	for _, class := range g.m.Package(pkg).AllClasses {
		files = append(files, g.mockTestFile(class)...)
	}
	return files, nil
}
