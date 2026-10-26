package typescript

import (
	"sort"
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// The tests of the implementation of a class (v1
// tests/generateImplementationTests.mtl): per feature the tests of its
// accessors, per operation the test of its default implementation, then the
// tests of the reflective methods. They run with vitest (or jest) and
// ts-mockito.

// testsView is the view model of a test file: the mock interfaces it
// declares and its tests, separated by a blank line.
type testsView struct {
	Interfaces []string
	Describe   string
	Tests      []member
}

// testView is the data of a test: the class, its feature or operation.
type testView struct {
	*featureView
	// Class is the implementation tested; Test the name of the test.
	Class, Test string
	// GetCall reads the feature (getX() or the property).
	GetCall string
	// Strategy is the variant of the test.
	Strategy string
	// ReverseClass is the getter of the class of the opposite and ReverseID the
	// id of the opposite in it.
	ReverseClass, ReverseID string
	// Expected is a value the test expects; Lines lines built in Go.
	Expected string
	Lines    []string
	// BlankAfter is whether a blank line follows the case (v1).
	BlankAfter bool
}

// SetCall writes the feature (setX(value) or the property = value).
func (t testView) SetCall(value string) string { return t.s.g.setCall(t.f, value) }

// TestValue declares a test value of the type of the feature named after
// varname (v1 generateTestValue).
func (t testView) TestValue(varname string) string {
	return t.s.testValueDeclaration(t.f.GetEType(), varname)
}

// testValueDeclaration declares a test value of a type: a mock of the
// internal interface of a class, the test value of a data type or an enum,
// null without type.
func (s *scope) testValueDeclaration(t ecore.EClassifier, varname string) string {
	switch t := t.(type) {
	case ecore.EClass:
		mockName := "mock" + naming.UpperFirst(varname)
		return "const " + mockName + " = " + s.mockingName("mock") + "<" + s.g.interfaceName(t) + "Internal>()\n" +
			"const " + varname + " = " + s.mockingName("instance") + "(" + mockName + ")"
	case nil:
		return "const " + varname + " = null"
	}
	return "const " + varname + " = " + s.testValue(t)
}

// mockingName writes a function of ts-mockito: the renderer imports the ones
// the file uses.
func (s *scope) mockingName(name string) string { return name }

// implementationTestFile is the test file of the implementation of a class.
func (g *gen) implementationTestFile(c ecore.EClass) engine.File {
	s := g.newScope()
	g.scanTestImports(s, true)
	class := g.implementationName(c)
	view := testsView{Describe: class}
	features := g.implementedFeatures(c)
	// the mock interfaces of the types of the references
	var mockTypes []ecore.EClass
	seen := map[ecore.EClass]bool{}
	for _, f := range features {
		if r, ok := f.(ecore.EReference); ok {
			if t, ok := r.GetEType().(ecore.EClass); ok && t.GetName() != "EObject" && !seen[t] {
				seen[t] = true
				mockTypes = append(mockTypes, t)
			}
		}
	}
	sort.SliceStable(mockTypes, func(i, j int) bool { return mockTypes[i].GetName() < mockTypes[j].GetName() })
	for _, t := range mockTypes {
		s.declare(g.interfaceIdentifier(t.GetEPackage(), g.interfaceName(t)+"Internal"))
		view.Interfaces = append(view.Interfaces, "interface "+g.interfaceName(t)+"Internal extends "+s.qualify(g.classInterfaceIdentifier(t))+
			", "+s.qualify(g.ecoreIdentifier("EObjectInternal"))+" {}")
	}
	implementation := g.interfaceIdentifier(c.GetEPackage(), class)
	if g.isSharedPackage(c.GetEPackage()) {
		implementation = g.classImplementationIdentifier(c)
	}
	newTest := func(f ecore.EStructuralFeature, test, strategy string) testView {
		t := testView{featureView: &featureView{names: names{s}}, Class: s.qualify(implementation), Test: test, Strategy: strategy}
		if f != nil {
			t.featureView = g.featureView(c, f, s)
			t.GetCall = g.getCall(f)
			if opposite := g.m.Feature(f).Opposite; opposite != nil {
				reverse := opposite.GetEContainingClass()
				t.ReverseClass = s.qualify(g.packageGetterIdentifier(reverse.GetEPackage())) + "()." + g.ReflectiveGetterName(reverse) + "()"
				t.ReverseID = s.featureID(reverse, opposite)
			}
		}
		return t
	}
	add := func(template string, t testView) {
		view.Tests = append(view.Tests, member{Template: template, Data: t})
	}

	static := newTest(nil, "eStaticClass", "")
	static.Expected = s.qualify(g.packageGetterIdentifier(c.GetEPackage())) + "()." + g.ReflectiveGetterName(c) + "()"
	add("tests/feature.static", static)
	for _, f := range features {
		info := g.m.Feature(f)
		name := naming.UpperFirst(f.GetName())
		if info.Get {
			if strategy := g.getTestStrategy(f); strategy != "" {
				t := newTest(f, "get"+name, strategy)
				if strategy == "attribute" {
					t.Expected = s.defaultValue(f)
				}
				add("tests/feature.get", t)
			}
		}
		if info.Set {
			t := newTest(f, "set"+name, g.setTestStrategy(f))
			t.Expected = s.defaultValue(f)
			add("tests/feature.set", t)
		}
		if info.BasicSet {
			t := newTest(f, "basicSet"+name, "basicset")
			t.Expected = s.defaultValue(f)
			add("tests/feature.basicset", t)
		}
		if info.Unset && !g.isPropertyAccessors(f) {
			strategy := "field"
			switch {
			case info.IsVolatile:
				strategy = "volatile"
			case info.IsList:
				strategy = "list"
			}
			t := newTest(f, "unset"+name, strategy)
			t.Expected = s.defaultValue(f)
			if strategy == "field" {
				t.Expected = t.UnsetValue()
			}
			add("tests/feature.unset", t)
		}
		if info.IsSet {
			strategy := "falsy"
			switch {
			case info.IsVolatile:
				strategy = "volatile"
			case g.hasESet(f) && info.Unset:
				strategy = "flag"
			}
			add("tests/feature.isset", newTest(f, "isSet"+name, strategy))
		}
	}
	operations := g.implementedOperations(c)
	firstOperation := true
	for _, op := range operations {
		if !g.isDefaultOperation(op) {
			continue
		}
		t := newTest(nil, g.operationName(op), "operation")
		var args []string
		for it := op.GetEParameters().Iterator(); it.HasNext(); {
			args = append(args, s.classifierDefaultValue(it.Next().(ecore.EParameter).GetEType()))
		}
		t.Expected = g.operationName(op) + "(" + strings.Join(args, ", ") + ")"
		// v1 writes the tests of the operations without blank lines between them
		view.Tests = append(view.Tests, member{Template: "tests/feature.operation", Data: t, Tight: !firstOperation})
		firstOperation = false
	}
	g.addReflectiveTests(&view, c, s, newTest)
	return g.file(g.implementationTestFileName(c, c.GetName()), s, "tests.ts", view)
}

// getTestStrategy is the variant of the test of a getter (the first v1 guard
// that holds): volatile, container, map, list, lazy (a lazy reference),
// proxy, reference, attribute.
func (g *gen) getTestStrategy(f ecore.EStructuralFeature) string {
	info := g.m.Feature(f)
	switch {
	case info.IsVolatile:
		return "volatile"
	case info.IsContainer:
		return "container"
	case info.IsMap:
		return "map"
	case info.IsList:
		if f.GetEType() != nil && f.GetEType().GetName() == "EFeatureMapEntry" {
			return "featuremap"
		}
		return "list"
	case info.Reference != nil && info.Lazy.IsLazy():
		return "lazy"
	case info.Reference != nil && info.IsProxy:
		return "proxy"
	case info.Reference != nil:
		return "reference"
	case info.Attribute != nil:
		return "attribute"
	}
	return ""
}

// setTestStrategy is the variant of the test of a setter: volatile,
// container, bidirectional, field.
func (g *gen) setTestStrategy(f ecore.EStructuralFeature) string {
	info := g.m.Feature(f)
	switch {
	case info.IsVolatile:
		return "volatile"
	case info.IsContainer:
		return "container"
	case info.IsBidirectional:
		return "bidirectional"
	}
	return "field"
}

// addReflectiveTests adds the tests of the mapping of the mixin ids and of the
// reflective methods.
func (g *gen) addReflectiveTests(view *testsView, c ecore.EClass, s *scope, newTest func(ecore.EStructuralFeature, string, string) testView) {
	info := g.m.Class(c)
	add := func(template string, t testView) {
		view.Tests = append(view.Tests, member{Template: template, Data: t})
	}
	p := c.GetEPackage()
	packageGetter := s.qualify(g.packageGetterIdentifier(p)) + "()"
	derived := func(name string, ids func(mixin ecore.EClass) [][2]string) {
		t := newTest(nil, name, "")
		for _, mixin := range info.Mixins {
			t.Lines = append(t.Lines, "{", indentUnit+"const cls = "+packageGetter+".get"+g.m.ElementAccessorName(mixin)+"()",
				indentUnit+"expect(o."+name+"(cls, -1)).toBe(-1)")
			for _, id := range ids(mixin) {
				t.Lines = append(t.Lines, indentUnit+"expect(o."+name+"(cls, "+id[0]+")).toBe("+id[1]+")")
			}
			t.Lines = append(t.Lines, "}")
		}
		add("tests/reflectives.derived", t)
	}
	hasFeatures, hasOperations := false, false
	for _, mixin := range info.Mixins {
		hasFeatures = hasFeatures || len(g.m.Class(mixin).ImplementedFeatures) > 0
		hasOperations = hasOperations || len(g.m.Class(mixin).ImplementedOperations) > 0
	}
	if hasFeatures {
		derived("eDerivedFeatureID", func(mixin ecore.EClass) [][2]string {
			var ids [][2]string
			for _, f := range g.m.Class(mixin).ImplementedFeatures {
				ids = append(ids, [2]string{s.featureID(mixin, f), s.featureID(c, f)})
			}
			return ids
		})
	}
	if hasOperations {
		derived("eDerivedOperationID", func(mixin ecore.EClass) [][2]string {
			var ids [][2]string
			for _, op := range g.m.Class(mixin).ImplementedOperations {
				ids = append(ids, [2]string{s.operationID(mixin, op), s.operationID(c, op)})
			}
			return ids
		})
	}
	cases := func(features []ecore.EStructuralFeature, exclude func(ecore.EStructuralFeature) bool, strategy func(ecore.EStructuralFeature) string) []testView {
		var tests []testView
		for _, f := range sortedByName(features) {
			if exclude != nil && exclude(f) {
				continue
			}
			tests = append(tests, newTest(f, "", strategy(f)))
		}
		return tests
	}
	notFeatureMap := func(f ecore.EStructuralFeature) bool { return g.m.Feature(f).IsFeatureMap }
	if egets := cases(info.EGetFeatures, nil, func(f ecore.EStructuralFeature) string {
		fi := g.m.Feature(f)
		switch {
		case fi.IsVolatile || (fi.Lazy.HasInitMethod && fi.Reference != nil && !fi.IsList && !fi.IsProxy):
			return "throw"
		case !fi.IsMap && fi.IsList && fi.Reference != nil:
			return "unresolved"
		}
		return "get"
	}); len(egets) > 0 {
		view.Tests = append(view.Tests, member{Template: "tests/reflectives.eget", Data: reflectiveTests{testView: newTest(nil, "eGetFromID", ""), Cases: egets}})
	}
	if esets := cases(info.ESetFeatures, notFeatureMap, func(f ecore.EStructuralFeature) string {
		fi := g.m.Feature(f)
		switch {
		case fi.IsVolatile:
			return "volatile"
		case fi.IsContainer:
			return "container"
		case fi.IsMap:
			if class, ok := f.GetEType().(ecore.EClass); ok && g.m.Class(class).IsMapEntry {
				return "map"
			}
			return "none"
		case fi.IsList && fi.Reference != nil:
			return "referencelist"
		case fi.IsList:
			return "list"
		case fi.IsBidirectional:
			return "bidirectional"
		case fi.IsContainment:
			return "containment"
		}
		return "field"
	}); len(esets) > 0 {
		// v1 leaves a blank line after a list, but the last case
		for i := range esets[:len(esets)-1] {
			esets[i].BlankAfter = esets[i].Strategy == "referencelist" || esets[i].Strategy == "list"
		}
		view.Tests = append(view.Tests, member{Template: "tests/reflectives.eset", Data: reflectiveTests{testView: newTest(nil, "eSetFromID", ""), Cases: esets}})
	}
	if issets := cases(info.EIsSetFeatures, notFeatureMap, func(f ecore.EStructuralFeature) string {
		if g.m.Feature(f).IsVolatile {
			return "throw"
		}
		return "falsy"
	}); len(issets) > 0 {
		view.Tests = append(view.Tests, member{Template: "tests/reflectives.eisset", Data: reflectiveTests{testView: newTest(nil, "eIsSetFromID", ""), Cases: issets}})
	}
	if unsets := cases(info.EUnsetFeatures, notFeatureMap, func(f ecore.EStructuralFeature) string {
		fi := g.m.Feature(f)
		switch {
		case fi.IsVolatile:
			return "throw"
		case fi.IsList:
			return "list"
		case fi.Reference != nil:
			return "reference"
		}
		switch s.defaultValue(f) {
		case "":
			return "unset"
		case "null":
			return "null"
		}
		return "default"
	}); len(unsets) > 0 {
		for i := range unsets {
			unsets[i].Expected = s.defaultValue(unsets[i].f)
		}
		view.Tests = append(view.Tests, member{Template: "tests/reflectives.eunset", Data: reflectiveTests{testView: newTest(nil, "eUnsetFromID", ""), Cases: unsets}})
	}
	if c.GetName() != "EObject" && len(info.EInvokeOperations) > 0 {
		operations := append([]ecore.EOperation(nil), info.EInvokeOperations...)
		sort.SliceStable(operations, func(i, j int) bool {
			return g.m.OperationIDName(operations[i].GetEContainingClass(), operations[i]) < g.m.OperationIDName(operations[j].GetEContainingClass(), operations[j])
		})
		t := newTest(nil, "eInvokeFromID", "")
		for _, op := range operations {
			t.Lines = append(t.Lines, "expect(() => o.eInvokeFromID("+s.operationID(op.GetEContainingClass(), op)+", null)).toThrow(Error)")
		}
		add("tests/reflectives.einvoke", t)
	}
	if adds := cases(info.EInverseAddFeatures, nil, func(f ecore.EStructuralFeature) string {
		fi := g.m.Feature(f)
		switch {
		case fi.IsMap:
			return "none"
		case fi.IsList:
			return "list"
		case fi.IsContainer:
			return "container"
		}
		return "reference"
	}); len(adds) > 0 {
		view.Tests = append(view.Tests, member{Template: "tests/reflectives.einverseadd", Data: reflectiveTests{testView: newTest(nil, "eBasicInverseAdd", ""), Cases: adds}})
	}
	if removes := cases(info.EInverseRemoveFeatures, nil, func(f ecore.EStructuralFeature) string {
		fi := g.m.Feature(f)
		switch {
		case fi.IsMap:
			return "map"
		case fi.IsList:
			return "list"
		}
		return "reference"
	}); len(removes) > 0 {
		view.Tests = append(view.Tests, member{Template: "tests/reflectives.einverseremove", Data: reflectiveTests{testView: newTest(nil, "eBasicInverseRemove", ""), Cases: removes}})
	}
}

// reflectiveTests is a test of a reflective method: one case per feature.
type reflectiveTests struct {
	testView
	Cases []testView
}

// IsContainment reports whether the feature is a containment.
func (t testView) IsContainment() bool { return t.f != nil && t.s.g.m.Feature(t.f).IsContainment }

// IsProxy reports whether the feature is a proxy.
func (t testView) IsProxy() bool { return t.f != nil && t.s.g.m.Feature(t.f).IsProxy }

// HasOpposite reports whether the feature has an opposite.
func (t testView) HasOpposite() bool { return t.f != nil && t.s.g.m.Feature(t.f).Opposite != nil }

// ElementType is the type of the elements of a list.
func (t testView) ElementType() string { return t.s.qualify(t.s.g.typeOf(t.f.GetEType())) }

// MapTypes is the key and value types of a map, "K, V".
func (t testView) MapTypes() string {
	key, value := t.s.g.mapTypeParameters(t.f)
	return t.s.qualify(key) + ", " + t.s.qualify(value)
}

// MapEntryTestValue declares a test value of the key or value feature of the
// map entry class of the feature.
func (t testView) MapEntryTestValue(name string) string {
	class := t.f.GetEType().(ecore.EClass)
	var classifier ecore.EClassifier
	if f := class.GetEStructuralFeatureFromName(name); f != nil {
		classifier = f.GetEType()
	}
	return t.s.testValueDeclaration(classifier, name)
}

// Preamble is the mock interfaces the file declares, followed by a blank
// line; "" without interfaces.
func (v testsView) Preamble() string {
	if len(v.Interfaces) == 0 {
		return ""
	}
	return strings.Join(v.Interfaces, "\n") + "\n\n"
}

// scanTestImports adds the modules of the test functions: vitest (with the
// vitest test framework), deep-equal (in the tests of an implementation) and
// ts-mockito, in the v1 order.
func (g *gen) scanTestImports(s *scope, deepEqual bool) {
	framework, _ := g.PackageProperty("testFramework")
	if strings.ToLower(framework) == "vitest" {
		s.scan("vitest", "", "describe", "expect", "test")
	}
	if deepEqual {
		s.scan("deep-equal", "deepEqual")
	}
	s.scan("ts-mockito", "", "anything", "capture", "instance", "mock", "reset", "verify", "when")
}

// MockType is the mock interface of the type of a reference: <Type>Internal,
// declared by the test file.
func (t testView) MockType() string {
	if class, ok := t.f.GetEType().(ecore.EClass); ok {
		return t.s.g.interfaceName(class) + "Internal"
	}
	return "any"
}

// IsContainer reports whether the feature is a container.
func (t testView) IsContainer() bool { return t.f != nil && t.s.g.m.Feature(t.f).IsContainer }

// inverseAddID is the id the mock of an added object receives in eInverseAdd:
// EOPPOSITE_FEATURE_BASE - ID for a containment without opposite, the id of
// the opposite otherwise; "" for a reference without opposite.
func (t testView) inverseAddID() string {
	info := t.s.g.m.Feature(t.f)
	switch {
	case info.IsContainment && info.Opposite == nil:
		return t.Ecore("EOPPOSITE_FEATURE_BASE") + " - " + t.ID()
	case info.Reference != nil && info.Opposite != nil:
		return t.ReverseID
	}
	return ""
}

// AddedExpectations are the expectations on the mock of an object added to a
// list (v1 generateFeatureMockAddedExpectations).
func (t testView) AddedExpectations() string {
	lines := []string{"when(mockValue.eClass()).thenReturn(null)", "when(mockValue.eStaticClass()).thenReturn(null)"}
	if id := t.inverseAddID(); id != "" {
		lines = append(lines, "when(mockValue.eInverseAdd(o, "+id+", anything())).thenReturn(null)")
	}
	// v1 leaves a blank line after them
	return strings.Join(lines, "\n") + "\n"
}

// AddedVerifications are the verifications on the mock of an object added to
// a list (v1 generateFeatureMockAddedVerifications, which starts with the
// expectations).
func (t testView) AddedVerifications() string {
	lines := []string{"when(mockValue.eClass()).thenReturn(null)", "when(mockValue.eStaticClass()).thenReturn(null)"}
	if id := t.inverseAddID(); id != "" {
		lines = append(lines, "verify(mockValue.eInverseAdd(o, "+id+", anything())).once()")
	}
	return strings.Join(lines, "\n")
}

// SetExpression writes the feature of o in an expression: o.setX(value), or
// (o.x = value) for a property, parenthesized as prettier does in an arrow
// function.
func (t testView) SetExpression(value string) string {
	if t.s.g.isMethodAccessors(t.f) {
		return "o." + t.s.g.setterName(t.f) + "(" + value + ")"
	}
	return "(o." + propertyName(t.f) + " = " + value + ")"
}
