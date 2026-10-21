package golang

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
)

// packageTestView is the view model of the tests of the package (template
// package_test.go, v1 generatePackageTests).
type packageTestView struct {
	testNames
	// Name, NsURI and NsPrefix are the NAME, NS_URI and NS_PREFIX constants.
	Name, NsURI, NsPrefix string
	// Classes lists the getters of the classes sorted by name, each followed
	// by the getters of its features and operations in model order; empty
	// without any class.
	Classes []string
	// DataTypes lists the getters of the data types (enums included) sorted
	// by name; empty without any.
	DataTypes []string
}

// testNames is the identifiers of the testing and testify packages a test
// file uses, qualified for the file.
type testNames struct {
	// T is testing.T.
	T string
	// Assert is the assert package prefix ("assert.").
	Assert string
	// Mock is the mock package prefix ("mock."), "" when the file does not use it.
	Mock string
}

// testNamesOf qualifies the testing identifiers for the scope; withMock
// records the mock package.
func testNamesOf(scope *fileScope, withMock bool) testNames {
	names := testNames{
		T:      scope.qualify("testing/testing.T"),
		Assert: identifierPackagePrefix(scope.qualify("github.com/stretchr/testify/assert/assert.Equal")),
	}
	if withMock {
		names.Mock = identifierPackagePrefix(scope.qualify("github.com/stretchr/testify/mock/mock.Anything"))
	}
	return names
}

// packageTestFile is the tests of the package: <package instance
// dir>/<package>package_impl_test.go (v1 generatePackageTests).
func (g *gen) packageTestFile(p ecore.EPackage) []engine.File {
	scope := newFileScope(g.packageInstancePath(p))
	constant := func(name string) string { return scope.qualify(g.qualifiedInterfaceIdentifier(p, name)) }
	view := packageTestView{
		testNames: testNamesOf(scope, false),
		Name:      constant("NAME"),
		NsURI:     constant("NS_URI"),
		NsPrefix:  constant("NS_PREFIX"),
	}
	for _, c := range sortedByName(g.m.Package(p).AllClasses) {
		view.Classes = append(view.Classes, g.reflectiveGetterName(c))
		for _, f := range namedElements(c.GetEStructuralFeatures()) {
			view.Classes = append(view.Classes, g.reflectiveGetterName(f))
		}
		for _, op := range namedElements(c.GetEOperations()) {
			view.Classes = append(view.Classes, g.reflectiveGetterName(op))
		}
	}
	for _, dt := range g.allDataTypesByName(p) {
		view.DataTypes = append(view.DataTypes, g.reflectiveGetterName(dt))
	}
	return []engine.File{{
		Path:   filePath(g.packageInstanceOutputDirectory(p), g.packageTestFileName(p)),
		Body:   goSource(scope, "package_test.go", view),
		Header: true,
	}}
}
