package typescript

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
)

// factoryTestView is the data of a test of the factory implementation (v1
// tests/generateFactoryTests.mtl).
type factoryTestView struct {
	names
	// Factory is the factory implementation tested.
	Factory string
	// Classes are the ids of the classes created by create (createFromEClass).
	Classes []string
	// Class is the name of a class; Container the type of a mock container,
	// ContainerList the type of the list of its opposite ("" for a single
	// opposite) and ContainerGet the read of that list; WithClassID whether
	// the creator takes a class id.
	Class, Container, ContainerList, ContainerGet string
	WithClassID                                   bool
	// Test is the name of a test.
	Test string
	// DataType is a data type: its name, id and variant.
	DataType, ID, Strategy string
	// CreateLines and ConvertLines are the checks of the literals of an enum.
	CreateLines, ConvertLines []string
}

// factoryTestFile is the test of the factory implementation of a package:
// create by class id, each creator and creator from a container, the
// invalid data types and the conversions of the serializable data types.
func (g *gen) factoryTestFile(p ecore.EPackage) engine.File {
	s := g.newScope()
	g.scanTestImports(s, false)
	factory := g.interfaceIdentifier(p, g.factoryImplementationName(p))
	if g.isSharedPackage(p) {
		factory = g.implementationIdentifier(p, g.factoryImplementationName(p))
	}
	base := factoryTestView{names: names{s}, Factory: s.qualify(factory)}
	view := testsView{Describe: g.factoryImplementationName(p)}
	add := func(template string, data factoryTestView, tight bool) {
		view.Tests = append(view.Tests, member{Template: template, Data: data, Tight: tight})
	}
	var classes []ecore.EClass
	for _, c := range sortedByName(g.m.Package(p).AllClasses) {
		if !c.IsInterface() && !c.IsAbstract() {
			classes = append(classes, c)
		}
	}
	create := base
	for _, c := range classes {
		create.Classes = append(create.Classes, s.classifierID(c))
	}
	add("tests/factory.create", create, false)
	// v1 writes no blank line after the test of a creator with a class id
	afterClassID := false
	tight := func() bool {
		t := afterClassID
		afterClassID = false
		return t
	}
	for _, c := range classes {
		creator := base
		creator.Class = c.GetName()
		add("tests/factory.creator", creator, tight())
		for _, container := range sortedByName(g.m.Class(c).AllContainers) {
			fromContainer := creator
			if t, ok := container.GetEType().(ecore.EClass); ok {
				fromContainer.Container = s.qualify(g.classInterfaceIdentifier(t))
			}
			if opposite := container.GetEOpposite(); opposite != nil && opposite.GetUpperBound() != 1 {
				fromContainer.ContainerList = s.qualify(g.typeOf(opposite.GetEType()))
				fromContainer.ContainerGet = g.getCall(opposite)
			}
			fromContainer.Test = "create" + c.GetName() + "FromContainer"
			add("tests/factory.fromcontainer", fromContainer, tight())
			if g.m.ElementIDAccessorName(c) != "" {
				fromContainer.WithClassID = true
				fromContainer.Test = "create" + c.GetName() + "FromContainerAndClassID"
				add("tests/factory.fromcontainer", fromContainer, false)
				afterClassID = true
			}
		}
	}
	add("tests/factory.invalidcreate", base, tight())
	add("tests/factory.invalidconvert", base, true)
	for _, dt := range sortedByName(g.allDataTypes(p)) {
		if !dt.IsSerializable() {
			continue
		}
		test := base
		test.DataType, test.ID = dt.GetName(), s.classifierID(dt)
		if e, ok := dt.(ecore.EEnum); ok {
			test.Strategy = "enum"
			if e.GetELiterals().Empty() {
				test.Strategy = "none"
			}
			for _, l := range sortedByName(enumLiterals(e)) {
				constant := s.qualify(g.enumIdentifier(e) + "." + literalConstantName(l))
				test.CreateLines = append(test.CreateLines, "expect(factory.createFromString(eDataType, "+quote(l.GetName())+")).toStrictEqual("+constant+")")
				test.ConvertLines = append(test.ConvertLines, "expect(factory.convertToString(eDataType, "+constant+")).toStrictEqual("+quote(l.GetName())+")")
			}
			add("tests/factory.createfrom", test, true)
			// v1 leaves a blank line after the creation test of an enum
			add("tests/factory.convertto", test, false)
			continue
		}
		switch t := g.instanceTypeName(dt); t {
		case "string", "number", "byte", "Uint8Array", "boolean", "Date":
			test.Strategy = t
		default:
			test.Strategy = "unknown"
		}
		add("tests/factory.createfrom", test, true)
		add("tests/factory.convertto", test, true)
	}
	return g.file(g.factoryTestFileName(p), s, "tests.ts", view)
}

// enumLiterals returns the literals of an enum.
func enumLiterals(e ecore.EEnum) []ecore.EEnumLiteral {
	var literals []ecore.EEnumLiteral
	for it := e.GetELiterals().Iterator(); it.HasNext(); {
		literals = append(literals, it.Next().(ecore.EEnumLiteral))
	}
	return literals
}
