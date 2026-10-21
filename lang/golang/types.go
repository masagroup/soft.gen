package golang

import (
	"strconv"
	"strings"

	"github.com/masagroup/soft.go/ecore"
)

// dataTypeRow maps the instance type names of a data type to its Go type,
// default value, test value and comparison function (v1 getAllDataTypes).
// The Go type is itself a name of the row: a model loaded with Go type names
// maps the same way.
type dataTypeRow struct {
	names        []string
	goType       string
	defaultValue string
	testValue    string
	comparison   string
}

// dataTypes returns the data type table, with the ecore package path and the
// any type of the run.
func (g *gen) dataTypes() []dataTypeRow {
	e := g.m.Ecore()
	anyType := g.anyType()
	ecoreType := func(name string) string { return g.qualifiedInterfaceIdentifier(e, name) }
	return []dataTypeRow{
		{[]string{"java.lang.Boolean", "boolean"}, "bool", "false", "true", ""},
		{[]string{"java.lang.Integer"}, "int", "0", "45", ""},
		{[]string{"java.lang.Byte", "java.lang.Character", "char"}, "byte", "''", "'a'", ""},
		{[]string{"byte[]"}, "[]byte", "nil", "nil", ""},
		{[]string{"java.util.Date"}, "*time/time.Time", "nil", "nil", ""},
		{[]string{"java.lang.Double", "java.math.BigDecimal", "double"}, "float64", "0.0", "3", ""},
		{[]string{"java.lang.Float", "float"}, "float32", "0.0", "1", ""},
		{[]string{"java.lang.Class"}, "reflect/reflect.Type", "nil", `reflect/reflect.TypeOf("")`, ""},
		{[]string{"java.lang.Object"}, anyType, "nil", "nil", ""},
		{[]string{"java.lang.Long", "java.math.BigInteger", "long"}, "int64", "0", "45", ""},
		{[]string{"java.util.Map"}, ecoreType("EMap"), "nil", "nil", ""},
		{[]string{"java.util.Map$Entry"}, ecoreType("EMapEntry"), "nil", "nil", ""},
		{[]string{"java.lang.Short", "short"}, "int16", "0", "2", ""},
		// Go types without a Java name, reached through instanceTypeName (v1
		// had no row: its test values T() did not compile)
		{nil, "int32", "0", "32", ""},
		{nil, "int8", "0", "8", ""},
		{[]string{"java.lang.String"}, "string", `""`, `"Test String"`, ""},
		{[]string{"java.lang.Error"}, "error", "nil", `errors/errors.New("Error")`, ""},
		{[]string{"com.google.common.primitives.UnsignedLong"}, "uint64", "0", "1", ""},
		{[]string{"org.eclipse.emf.common.util.DiagnosticChain"}, ecoreType("EDiagnosticChain"), "nil", "nil", ""},
		{[]string{"org.eclipse.emf.common.util.EList"}, ecoreType("EList"), "nil", ecoreType("NewEmptyBasicEList()"), ""},
		{[]string{"org.eclipse.emf.common.util.Enumerator"}, anyType, "nil", "nil", ""},
		{[]string{"org.eclipse.emf.ecore.util.FeatureMap"}, anyType, "nil", "nil", ""},
		{[]string{"org.eclipse.emf.ecore.util.FeatureMap$Entry"}, anyType, "nil", "nil", ""},
		{[]string{"org.eclipse.emf.ecore.resource.Resource"}, ecoreType("EResource"), "nil", "nil", ""},
		{[]string{"org.eclipse.emf.ecore.resource.ResourceSet"}, ecoreType("EResourceSet"), "nil", "nil", ""},
		{[]string{"org.eclipse.emf.common.util.TreeIterator"}, ecoreType("EIterator"), "nil", "nil", ""},
	}
}

// findDataType returns the row naming an instance type (the Go type counts as a name).
func (g *gen) findDataType(instanceTypeName string) (dataTypeRow, bool) {
	if instanceTypeName == "" {
		return dataTypeRow{}, false
	}
	for _, row := range g.dataTypes() {
		if row.goType == instanceTypeName {
			return row, true
		}
		for _, name := range row.names {
			if name == instanceTypeName {
				return row, true
			}
		}
	}
	return dataTypeRow{}, false
}

// goVersion is the goVersion property, default 1.19.
func (g *gen) goVersion() string {
	if v, ok := g.PackageProperty("goVersion"); ok {
		return v
	}
	return "1.19"
}

// anyType is the empty interface type of the run: any from Go 1.18, else interface{}.
func (g *gen) anyType() string {
	if compareVersions(g.goVersion(), "1.18") >= 0 {
		return "any"
	}
	return "interface{}"
}

// isAnyType reports whether a type is the empty interface type.
func (g *gen) isAnyType(t string) bool {
	return t == g.anyType()
}

// compareVersions compares dotted versions part by part, numerically when
// both parts are numbers.
func compareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var pa, pb string
		if i < len(as) {
			pa = as[i]
		}
		if i < len(bs) {
			pb = bs[i]
		}
		na, errA := strconv.Atoi(pa)
		nb, errB := strconv.Atoi(pb)
		switch {
		case errA == nil && errB == nil && na != nb:
			if na < nb {
				return -1
			}
			return 1
		case (errA != nil || errB != nil) && pa != pb:
			return strings.Compare(pa, pb)
		}
	}
	return 0
}

// instanceTypeName is the Go type name of a classifier (v1 getInstanceTypeName):
// the instanceTypeName annotation, else the instance type name of the model,
// else the interface of a class; mapped through the data type table.
func (g *gen) instanceTypeName(c ecore.EClassifier) string {
	name := ""
	if v, ok := g.KeyValue(c, "instanceTypeName"); ok {
		name = v
	} else if n := c.GetInstanceTypeName(); n != "" {
		name = n
	} else if class, ok := c.(ecore.EClass); ok {
		name = g.classInterface(class)
	}
	if row, ok := g.findDataType(name); ok {
		return row.goType
	}
	return name
}

// typeOfClassifier is the qualified Go type of a classifier (v1 getType): the
// any type when nil, the interface of a class, the enum type, or the instance
// type name of a data type.
func (g *gen) typeOfClassifier(c ecore.EClassifier) string {
	switch c := c.(type) {
	case nil:
		return g.anyType()
	case ecore.EClass:
		return g.classInterface(c)
	case ecore.EEnum:
		return g.qualifiedInterfaceIdentifier(c.GetEPackage(), c.GetName())
	case ecore.EDataType:
		return g.instanceTypeName(c)
	}
	return ""
}

// typeOf is the qualified Go type of a typed element (v1 generateType): ""
// without a type, EMap for a map, EList for a list, else the type of its classifier.
func (g *gen) typeOf(el ecore.ETypedElement) string {
	if el == nil || el.GetEType() == nil {
		return ""
	}
	switch {
	case g.isMapType(el):
		return g.qualifiedInterfaceIdentifier(g.m.Ecore(), "EMap")
	case isListType(el):
		return g.qualifiedInterfaceIdentifier(g.m.Ecore(), "EList")
	}
	return g.typeOfClassifier(el.GetEType())
}

// isListType reports whether a typed element is many-valued (v1 isListType).
func isListType(el ecore.ETypedElement) bool {
	return el.IsMany() || el.GetUpperBound() != 1
}

// isMapType reports whether a typed element is a list of map entries (v1 isMapType).
func (g *gen) isMapType(el ecore.ETypedElement) bool {
	if !isListType(el) || el.GetEType() == nil {
		return false
	}
	return g.m.IsMapEntryName(el.GetEType().GetInstanceTypeName())
}

// defaultValueOfAttribute is the default value of an attribute (v1
// getDefaultValue(EAttribute)): "" when many, the default value literal
// converted to the type, else the default value of the type.
func (g *gen) defaultValueOfAttribute(a ecore.EAttribute) string {
	if a.IsMany() {
		return ""
	}
	if literal := a.GetDefaultValueLiteral(); literal != "" {
		return g.defaultValueFromLiteral(a.GetEType(), literal)
	}
	return g.defaultValueOfClassifier(a.GetEType())
}

// defaultValueFromLiteral converts a default value literal: the constant of an
// enum literal, a quoted string for a string type, else the literal itself.
func (g *gen) defaultValueFromLiteral(c ecore.EClassifier, literal string) string {
	switch c := c.(type) {
	case ecore.EEnum:
		for it := c.GetELiterals().Iterator(); it.HasNext(); {
			if l := it.Next().(ecore.EEnumLiteral); l.GetName() == literal {
				return g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.constantsName(l))
			}
		}
		return literal
	case ecore.EDataType:
		if g.instanceTypeName(c) == "string" {
			return `"` + strings.ReplaceAll(strings.ReplaceAll(literal, `\`, `\\`), `"`, `\"`) + `"`
		}
	}
	return literal
}

// defaultValueOfClassifier is the default value of a classifier (v1
// getDefaultValue(EClassifier)): the first literal of an enum, the default of
// a data type, nil for a class.
func (g *gen) defaultValueOfClassifier(c ecore.EClassifier) string {
	switch c := c.(type) {
	case ecore.EEnum:
		if c.GetELiterals().Empty() {
			return "0"
		}
		return g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.constantsName(c.GetELiterals().Get(0).(ecore.EEnumLiteral)))
	case ecore.EDataType:
		return g.defaultValueOfDataType(c)
	}
	return "nil"
}

// defaultValueOfDataType is the default value of a data type (v1
// getDefaultValue(EDataType)): the defaultValue annotation, the Java primitive
// default of its instance class (except EChar), nil for a pointer type, else
// the default of the data type table.
func (g *gen) defaultValueOfDataType(dt ecore.EDataType) string {
	if v, ok := g.KeyValue(dt, "defaultValue"); ok {
		return v
	}
	if dt.GetName() != "EChar" {
		if v, ok := javaPrimitiveDefault(dt.GetInstanceClassName()); ok {
			return v
		}
	}
	typeName := g.instanceTypeName(dt)
	if strings.HasPrefix(typeName, "*") {
		return "nil"
	}
	if row, ok := g.findDataType(typeName); ok {
		return row.defaultValue
	}
	return ""
}

// javaPrimitiveDefault is the default value EMF gives to a data type whose
// instance class is a Java primitive.
func javaPrimitiveDefault(instanceClassName string) (string, bool) {
	switch instanceClassName {
	case "boolean":
		return "false", true
	case "byte", "short", "int", "long":
		return "0", true
	case "float", "double":
		return "0.0", true
	}
	return "", false
}

// defaultValueWithType is the default value of a feature converted to its type
// (v1 getDefaultValueWithType): T(v), (T)(v) for a pointer type, nil for a
// reference, "" when the attribute has no default.
func (g *gen) defaultValueWithType(f ecore.EStructuralFeature) string {
	a, ok := f.(ecore.EAttribute)
	if !ok {
		return "nil"
	}
	value := g.defaultValueOfAttribute(a)
	if value == "" || value == "nil" {
		return value
	}
	return convertedValue(g.typeOfClassifier(f.GetEType()), value)
}

// convertedValue is "T(v)", with T between parentheses when it is a pointer type.
func convertedValue(t, v string) string {
	if strings.HasPrefix(t, "*") {
		return "(" + t + ")(" + v + ")"
	}
	return t + "(" + v + ")"
}

// testValue is the value the tests give to a classifier (v1 getTestValue): a
// mock for a class, the second literal of an enum, the testValue annotation
// or the table value converted to the type for a data type.
func (g *gen) testValue(c ecore.EClassifier) string {
	switch c := c.(type) {
	case nil:
		return "nil"
	case ecore.EClass:
		if c.GetName() == "EObject" {
			return g.qualifiedMockIdentifier(g.m.Ecore(), "NewMockEObjectInternal") + "(t)"
		}
		return g.qualifiedMockIdentifier(c.GetEPackage(), "New"+g.mockName(c)) + "(t)"
	case ecore.EEnum:
		literals := c.GetELiterals()
		if literals.Empty() {
			return "0"
		}
		index := 0
		if literals.Size() > 1 {
			index = 1
		}
		return g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.constantsName(literals.Get(index).(ecore.EEnumLiteral)))
	case ecore.EDataType:
		if raw := g.testValueOfDataType(c); raw != "" {
			return convertedValue(g.typeOfClassifier(c), raw)
		}
		// a type the type table does not know, without testValue: its zero value
		return zeroValueOf(g.typeOfClassifier(c))
	}
	return ""
}

// testValueOfDataType is the raw test value of a data type: the testValue
// annotation, nil for a pointer type, else the table value.
func (g *gen) testValueOfDataType(dt ecore.EDataType) string {
	if v, ok := g.KeyValue(dt, "testValue"); ok {
		return v
	}
	typeName := g.instanceTypeName(dt)
	if strings.HasPrefix(typeName, "*") {
		return "nil"
	}
	if row, ok := g.findDataType(typeName); ok {
		return row.testValue
	}
	return ""
}

// comparisonFunction is the function comparing two values of a data type ("" for !=).
func (g *gen) comparisonFunction(dt ecore.EDataType) string {
	if row, ok := g.findDataType(g.instanceTypeName(dt)); ok {
		return row.comparison
	}
	return ""
}

// zeroValueOf is the zero value of any type t, *new(t): the value of a data
// type the type table does not know (T{} would only suit a struct type).
func zeroValueOf(t string) string {
	return "*new(" + t + ")"
}
