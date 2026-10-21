package golang

import (
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/lang/internal/common"
)

// ecorePackagePath is the path of the soft.go ecore package: "<import path>/<package name>".
const ecorePackagePath = "github.com/masagroup/soft.go/ecore/ecore"

// Indexes of the paths of a package in the packages property ("uri=interface#implementation#mock").
const (
	interfacePathIndex      = 1
	implementationPathIndex = 2
	mockPathIndex           = 3
)

// packagePaths returns the paths of the packages (v1 getAllPackagePaths): the
// ecore package, then the tokens of the packages property, merged from the
// command line and the package annotation. A token is "uri=paths" or "paths",
// keyed by the last segment of its first path.
func (g *gen) packagePaths() common.PackageSpecs {
	if g.paths == nil {
		defaults := map[string]string{g.m.Ecore().GetNsURI(): ecorePackagePath}
		g.paths = common.ParsePackageSpecs(g.ListProperty("packages"), defaults)
	}
	return g.paths
}

// indexPath returns the path of the package at the index (v1 getIndexPath):
// the paths found by nsURI, else by name, else "<name>/<name>"; a missing index
// gives the last path.
func (g *gen) indexPath(p ecore.EPackage, index int) string {
	spec, ok := g.packagePaths().Spec(p)
	if !ok {
		return p.GetName() + "/" + p.GetName()
	}
	return common.IndexPart(strings.Split(spec, "#"), index)
}

// interfacePackagePath is the path of the package holding the interfaces.
func (g *gen) interfacePackagePath(p ecore.EPackage) string {
	return g.indexPath(p, interfacePathIndex)
}

// implementationPackagePath is the path of the package holding the implementations.
func (g *gen) implementationPackagePath(p ecore.EPackage) string {
	return g.indexPath(p, implementationPathIndex)
}

// mockPackagePath is the path of the package holding the mocks.
func (g *gen) mockPackagePath(p ecore.EPackage) string {
	return g.indexPath(p, mockPathIndex)
}

// isSharedPackage reports whether the package instance (GetPackage) is shared
// by several implementations: the property packageInstance is "interfaces" and
// the implementation path differs from the interface path (v1 isSharedPackage).
func (g *gen) isSharedPackage(p ecore.EPackage) bool {
	return !g.m.IsEcore(p) &&
		g.PackagePropertyIs("packageInstance", "interfaces") &&
		g.interfacePackagePath(p) != g.implementationPackagePath(p)
}

// isCheckImplementation reports whether the setters of a shared package check
// that the objects do not mix implementations (v1 isCheckImplementation).
func (g *gen) isCheckImplementation(p ecore.EPackage) bool {
	return g.isSharedPackage(p) && !g.PackagePropertyIs("checkImplementation", "false")
}

// packageInstancePath is the path of the package holding GetPackage.
func (g *gen) packageInstancePath(p ecore.EPackage) string {
	if g.isSharedPackage(p) {
		return g.interfacePackagePath(p)
	}
	return g.implementationPackagePath(p)
}

// packageInstanceOutputDirectory is the output directory of the package holding GetPackage.
func (g *gen) packageInstanceOutputDirectory(p ecore.EPackage) string {
	if g.isSharedPackage(p) {
		return g.interfaceOutputDirectory(p)
	}
	return g.implementationOutputDirectory(p)
}

// outputDirectory is the directory, relative to the output folder, of the
// package at the index (v1 getOutputDirectory): with the module property,
// the directory of the path minus the module prefix ("" when equal); without
// it, or outside the module, the package name.
func (g *gen) outputDirectory(p ecore.EPackage, index int) string {
	module, ok := g.PackageProperty("module")
	if !ok {
		return p.GetName()
	}
	dir := packageDirectoryFromPath(g.indexPath(p, index))
	switch {
	case dir == module:
		return ""
	case len(dir) > len(module) && strings.HasPrefix(dir, module):
		return dir[len(module)+1:]
	}
	return p.GetName()
}

// interfaceOutputDirectory is the output directory of the interfaces.
func (g *gen) interfaceOutputDirectory(p ecore.EPackage) string {
	return g.outputDirectory(p, interfacePathIndex)
}

// implementationOutputDirectory is the output directory of the implementations.
func (g *gen) implementationOutputDirectory(p ecore.EPackage) string {
	return g.outputDirectory(p, implementationPathIndex)
}

// mockOutputDirectory is the output directory of the mocks.
func (g *gen) mockOutputDirectory(p ecore.EPackage) string {
	return g.outputDirectory(p, mockPathIndex)
}

// packageDirectoryFromPath returns the import path of a package path
// ("github.com/x/y/y" -> "github.com/x/y"); "" without a slash.
func packageDirectoryFromPath(path string) string {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return ""
	}
	return path[:i]
}

// packageNameFromPath returns the package name of a package path (its last segment).
func packageNameFromPath(path string) string {
	return common.LastSegment(path)
}

// filePath joins an output directory and a file name with a slash.
func filePath(dir, name string) string {
	if dir == "" {
		return name
	}
	return dir + "/" + name
}
