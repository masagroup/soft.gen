package golang

import (
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/naming"
)

// factoryTestView is the view model of the tests of the factory (template
// factory_test.go, v1 generateFactoryTests).
type factoryTestView struct {
	factoryTestNames
	// AsInternal is whether AsInternal is tested (not for a shared package).
	AsInternal bool
	// Classes lists the concrete classes sorted by name.
	Classes []factoryTestClassView
	// Conversions lists the conversions of the serializable data types
	// (enums included) sorted by name.
	Conversions []factoryTestConversionView
}

// factoryTestNames is the identifiers every part of the factory tests uses.
type factoryTestNames struct {
	testNames
	// Constructor is the constructor of the factory implementation:
	// new<Package>FactoryImpl, from the package name as v1 did.
	Constructor string
	// MockEClass, MockEDataType and MockEList are the constructors of the ecore mocks.
	MockEClass, MockEDataType, MockEList string
}

// factoryTestClassView is the creation tests of a concrete class.
type factoryTestClassView struct {
	// Name is the class name; ID its classifier id constant.
	Name, ID string
	// Containers lists the container references sorted by name.
	Containers []factoryTestContainerView
}

// factoryTestContainerView is the creation tests of a class from one of its
// container references (v1 TestFactoryCreate<X>From<Reference>Container).
type factoryTestContainerView struct {
	factoryTestNames
	// Reference is the reference name, upper-cased first.
	Reference string
	// MockContainer is the mock of the container class.
	MockContainer string
	// Opposite is the opposite name, upper-cased first; OppositeMany whether
	// it is many-valued (the element is added to its list, else set).
	Opposite     string
	OppositeMany bool
	// WithClassID is whether the class has an element id: the
	// FromContainerAndClassID creation is tested too.
	WithClassID bool
}

// factoryTestConversionView is the conversion tests of a data type
// (templates test/createfromstring.<CreateKind>, test/converttostring.<ConvertKind>).
type factoryTestConversionView struct {
	factoryTestNames
	// ID is the classifier id constant.
	ID string
	// CreateKind and ConvertKind are the test variants: enum, string, int,
	// float, byte, bytes, bool, reflect (creation only), time or unsupported.
	CreateKind, ConvertKind string
	// Type is the Go type of the int and float variants.
	Type string
	// Times is the number of GetClassifierID calls of an enum: one per literal, plus one.
	Times string
	// Literals lists the literals of an enum sorted by name.
	Literals []factoryTestLiteralView
	// Reflect is the reflect package prefix of the reflect variant, Time
	// the time package prefix of the time variant.
	Reflect, Time string
}

// factoryTestLiteralView is one literal of an enum: its constant and its name.
type factoryTestLiteralView struct {
	Const, Name string
}

// factoryTestFile is the tests of the factory: <implementation
// dir>/<package>factory_impl_test.go (v1 generateFactoryTests).
func (g *gen) factoryTestFile(p ecore.EPackage) []engine.File {
	scope := newFileScope(g.implementationPackagePath(p))
	ecoreMock := func(name string) string { return scope.qualify(g.qualifiedMockIdentifier(g.m.Ecore(), name)) }
	names := factoryTestNames{
		testNames:     testNamesOf(scope, true),
		Constructor:   "new" + naming.UpperFirst(p.GetName()) + "FactoryImpl",
		MockEClass:    ecoreMock("NewMockEClass"),
		MockEDataType: ecoreMock("NewMockEDataType"),
		MockEList:     ecoreMock("NewMockEList"),
	}
	view := factoryTestView{factoryTestNames: names, AsInternal: !g.isSharedPackage(p)}
	for _, c := range g.concreteClassesByName(p) {
		view.Classes = append(view.Classes, g.factoryTestClass(c, names, scope))
	}
	for _, dt := range g.allDataTypesByName(p) {
		if dt.IsSerializable() {
			view.Conversions = append(view.Conversions, g.factoryTestConversion(dt, names, scope))
		}
	}
	return []engine.File{{
		Path:   filePath(g.implementationOutputDirectory(p), g.factoryTestFileName(p)),
		Body:   goSource(scope, "factory_test.go", view),
		Header: true,
	}}
}

// factoryTestClass builds the creation tests of a concrete class (v1
// generateFactoryTests, for each class): its id and its containers.
func (g *gen) factoryTestClass(c ecore.EClass, names factoryTestNames, scope *fileScope) factoryTestClassView {
	view := factoryTestClassView{
		Name: c.GetName(),
		ID:   scope.qualify(g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.m.ClassifierIDName(c))),
	}
	withID := g.m.ElementIDAccessorName(c) != ""
	for _, ref := range sortedByName(g.m.Class(c).AllContainers) {
		opposite := ref.GetEOpposite()
		view.Containers = append(view.Containers, factoryTestContainerView{
			factoryTestNames: names,
			Reference:        naming.UpperFirst(ref.GetName()),
			MockContainer:    g.mockTestValue(ref.GetEType(), scope),
			Opposite:         naming.UpperFirst(opposite.GetName()),
			OppositeMany:     opposite.GetUpperBound() != 1,
			WithClassID:      withID,
		})
	}
	return view
}

// factoryTestConversion builds the conversion tests of a data type (v1
// generateCreateFrom, generateConvertTo).
func (g *gen) factoryTestConversion(dt ecore.EDataType, names factoryTestNames, scope *fileScope) factoryTestConversionView {
	view := factoryTestConversionView{
		factoryTestNames: names,
		ID:               scope.qualify(g.qualifiedInterfaceIdentifier(dt.GetEPackage(), g.m.ClassifierIDName(dt))),
	}
	if enum, ok := dt.(ecore.EEnum); ok {
		view.CreateKind, view.ConvertKind = "enum", "enum"
		literals := elementsOf[ecore.EEnumLiteral](enum.GetELiterals())
		view.Times = strconv.Itoa(len(literals) + 1)
		for _, l := range sortedByName(literals) {
			view.Literals = append(view.Literals, factoryTestLiteralView{
				Const: scope.qualify(g.qualifiedInterfaceIdentifier(enum.GetEPackage(), g.constantsName(l))),
				Name:  l.GetName(),
			})
		}
		return view
	}
	view.Type = g.instanceTypeName(dt)
	view.CreateKind = factoryTestKind(conversionKind(view.Type))
	view.ConvertKind = view.CreateKind
	switch view.CreateKind {
	case "reflect":
		view.ConvertKind = "unsupported"
		view.Reflect = identifierPackagePrefix(scope.qualify("reflect/reflect.Ptr"))
	case "time":
		view.Time = identifierPackagePrefix(scope.qualify("time/time.Time"))
	}
	return view
}

// factoryTestKind groups the conversion kinds the tests share: the integer
// kinds are int, the float kinds float.
func factoryTestKind(kind string) string {
	switch kind {
	case "int", "int64", "int32", "int16", "int8":
		return "int"
	case "float64", "float32":
		return "float"
	}
	return kind
}
