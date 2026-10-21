package typescript

import (
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/model"
)

// dataTypeRow maps the instance type names of a data type to its TypeScript
// type, default value and test value (v1 getAllDataTypes). As in v1, an
// instance type name matches a row when it is any of its values, the
// TypeScript type, default and test values included.
type dataTypeRow struct {
	names        []string
	tsType       string
	defaultValue string
	testValue    string
}

// matches reports whether the row has the instance type name among its values.
func (r dataTypeRow) matches(name string) bool {
	if name == r.tsType || name == r.defaultValue || name == r.testValue {
		return true
	}
	for _, n := range r.names {
		if n == name {
			return true
		}
	}
	return false
}

// dataTypes returns the data type table; the ecore types are qualified
// identifiers of the ecore module.
func (g *gen) dataTypes() []dataTypeRow {
	ecoreType := g.ecoreIdentifier
	return []dataTypeRow{
		{[]string{"java.lang.Boolean", "boolean"}, "boolean", "false", "true"},
		{[]string{"java.lang.Integer", "int"}, "number", "0", "45"},
		{[]string{"java.lang.Byte", "java.lang.Character", "char"}, "number", "''", "'a'"},
		{[]string{"byte[]"}, "Uint8Array", "null", "null"},
		{[]string{"java.util.Date"}, "Date", "null", "new Date()"},
		{[]string{"java.lang.Double", "java.math.BigDecimal", "double"}, "number", "0.0", "3.5"},
		{[]string{"java.lang.Float", "float"}, "number", "0.0", "1.2"},
		{[]string{"java.lang.Class"}, "any", "null", ""},
		{[]string{"java.lang.Object"}, "any", "null", "null"},
		{[]string{"java.lang.Long", "java.math.BigInteger", "long"}, "number", "0", "45"},
		{[]string{"java.util.Map"}, ecoreType("EMap"), "null", "null"},
		{[]string{"java.util.Map$Entry"}, ecoreType("EMapEntry"), "null", "null"},
		{[]string{"java.lang.Short", "short"}, "number", "0", "2"},
		{[]string{"java.lang.String"}, "string", `""`, `"Test String"`},
		{[]string{"org.eclipse.emf.common.util.DiagnosticChain"}, ecoreType("EDiagnosticChain"), "null", "null"},
		{[]string{"org.eclipse.emf.common.util.EList"}, ecoreType("EList"), "null", "null"},
		{[]string{"org.eclipse.emf.common.util.Enumerator"}, "any", "null", "null"},
		{[]string{"org.eclipse.emf.ecore.util.FeatureMap"}, "any", "null", "null"},
		{[]string{"org.eclipse.emf.ecore.util.FeatureMap$Entry"}, "any", "null", "null"},
		{[]string{"org.eclipse.emf.ecore.resource.Resource"}, ecoreType("EResource"), "null", "null"},
		{[]string{"org.eclipse.emf.ecore.resource.ResourceSet"}, ecoreType("EResourceSet"), "null", "null"},
		{[]string{"org.eclipse.emf.common.util.TreeIterator"}, "IterableIterator", "null", "null"},
	}
}

// findDataType returns the first row matching an instance type name.
func (g *gen) findDataType(name string) (dataTypeRow, bool) {
	for _, row := range g.dataTypes() {
		if row.matches(name) {
			return row, true
		}
	}
	return dataTypeRow{}, false
}

// instanceTypeName is the instance type of a classifier (v1
// getInstanceTypeName): its key instanceTypeName, else its instance type
// name, else the interface of a class; mapped by the data type table.
func (g *gen) instanceTypeName(c ecore.EClassifier) string {
	name, ok := g.KeyValue(c, "instanceTypeName")
	if !ok {
		name = c.GetInstanceTypeName()
		if name == "" {
			if class, isClass := c.(ecore.EClass); isClass {
				name = g.classInterfaceIdentifier(class)
			}
		}
	}
	if row, found := g.findDataType(name); found {
		return row.tsType
	}
	return name
}

// typeOf is the qualified type of a classifier (v1 getType): any for none,
// the interface of a class, the enum, the instance type of a data type.
func (g *gen) typeOf(c ecore.EClassifier) string {
	switch c := c.(type) {
	case nil:
		return "any"
	case ecore.EClass:
		return g.classInterfaceIdentifier(c)
	case ecore.EEnum:
		return g.enumIdentifier(c)
	}
	return g.instanceTypeName(c)
}

// typedElementType is the type of a feature, a parameter or an operation as
// the scope writes it (v1 generateType): void without type, EMap<K, V> for a
// map, EList<T> for a list, else the type of its classifier.
func (s *scope) typedElementType(el ecore.ETypedElement) string {
	return s.qualify(s.g.typedElementQID(el))
}

// typedInfo tells whether a typed element is a list or a map.
type typedInfo struct{ isList, isMap bool }

// typedInfo classifies a typed element as the model does for features and
// operations (v1 isListType, isMapType).
func (g *gen) typedInfo(el ecore.ETypedElement) typedInfo {
	isMap := g.m.IsMapType(el)
	return typedInfo{isMap: isMap, isList: !isMap && model.IsListType(el)}
}

// mapTypeParameters are the key and value types of a map-typed element (v1
// generateMapTypeParameter): its two type arguments, else the types of the
// key and value features of its map entry class, else any.
func (g *gen) mapTypeParameters(el ecore.ETypedElement) (key, value string) {
	if class, ok := el.GetEType().(ecore.EClass); ok && g.m.Class(class).IsMapEntry {
		return g.mapEntryFeatureType(class, "key"), g.mapEntryFeatureType(class, "value")
	}
	return "any", "any"
}

// mapEntryFeatureType is the type of the key or value feature of a map entry class.
func (g *gen) mapEntryFeatureType(class ecore.EClass, name string) string {
	f := class.GetEStructuralFeatureFromName(name)
	if f == nil || f.GetEType() == nil {
		return "any"
	}
	return g.typeOf(f.GetEType())
}

// defaultValue is the initial value of a feature (v1 getDefaultValue): null
// for a reference, nothing for a many attribute, the default literal of an
// attribute, else the default value of its type.
func (s *scope) defaultValue(f ecore.EStructuralFeature) string {
	attribute, ok := f.(ecore.EAttribute)
	if !ok {
		return "null"
	}
	if attribute.IsMany() {
		return ""
	}
	if literal := attribute.GetDefaultValueLiteral(); literal != "" {
		return s.defaultValueFromLiteral(attribute.GetEAttributeType(), literal)
	}
	return s.classifierDefaultValue(attribute.GetEAttributeType())
}

// defaultValueFromLiteral is the value of a default literal: the constant of
// the enum literal of that name, else the literal as is.
func (s *scope) defaultValueFromLiteral(t ecore.EClassifier, literal string) string {
	if enum, ok := t.(ecore.EEnum); ok {
		for it := enum.GetELiterals().Iterator(); it.HasNext(); {
			if l := it.Next().(ecore.EEnumLiteral); l.GetName() == literal {
				return s.qualify(s.g.enumIdentifier(enum) + "." + literalConstantName(l))
			}
		}
	}
	return literal
}

// classifierDefaultValue is the default value of a type: the first literal of
// an enum (0 without literal), the key defaultValue or the default value of a
// data type (but EChar), the default of its row, null for a class.
func (s *scope) classifierDefaultValue(t ecore.EClassifier) string {
	g := s.g
	switch t := t.(type) {
	case ecore.EEnum:
		if t.GetELiterals().Empty() {
			return "0"
		}
		first := t.GetELiterals().Get(0).(ecore.EEnumLiteral)
		return s.qualify(g.enumIdentifier(t) + "." + literalConstantName(first))
	case ecore.EDataType:
		if value, ok := g.KeyValue(t, "defaultValue"); ok {
			return value
		}
		if t.GetName() != "EChar" {
			if value, ok := javaPrimitiveDefaults[t.GetInstanceTypeName()]; ok {
				return value
			}
		}
		// as v1, the row is found by the TypeScript type: number finds the
		// row of java.lang.Integer
		if value := g.dataTypeValue(g.instanceTypeName(t), func(r dataTypeRow) string { return r.defaultValue }); value != "" {
			return s.qualify(value)
		}
		return ""
	}
	return "null"
}

// javaPrimitiveDefaults are the default values EMF gives the data types of a
// Java primitive instance class, as Java writes them (v1 ran on EMF:
// defaultValue.toString()).
var javaPrimitiveDefaults = map[string]string{
	"boolean": "false", "byte": "0", "char": "\x00", "double": "0.0", "float": "0.0", "int": "0", "long": "0", "short": "0",
}

// testValue is the value the tests set (v1 getTestValue): a comment for a
// class, the second literal of an enum (the first one when alone, 0 without),
// the key testValue or the test value of the row of a data type.
func (s *scope) testValue(t ecore.EClassifier) string {
	g := s.g
	switch t := t.(type) {
	case nil:
		return "null"
	case ecore.EClass:
		return "// must be defined"
	case ecore.EEnum:
		literals := t.GetELiterals()
		if literals.Size() == 0 {
			return "0"
		}
		index := 0
		if literals.Size() > 1 {
			index = 1
		}
		return s.qualify(g.enumIdentifier(t) + "." + literalConstantName(literals.Get(index).(ecore.EEnumLiteral)))
	case ecore.EDataType:
		if value, ok := g.KeyValue(t, "testValue"); ok {
			return value
		}
		if value := g.dataTypeValue(g.instanceTypeName(t), func(r dataTypeRow) string { return r.testValue }); value != "" {
			return s.qualify(value)
		}
		return ""
	}
	return ""
}

// typedElementQID is the qualified type of a typed element, as
// typedElementType writes it through a scope.
func (g *gen) typedElementQID(el ecore.ETypedElement) string {
	t := el.GetEType()
	if t == nil {
		return "void"
	}
	info := g.typedInfo(el)
	switch {
	case info.isMap:
		key, value := g.mapTypeParameters(el)
		return g.ecoreIdentifier("EMap") + "<" + key + "," + value + ">"
	case info.isList:
		return g.ecoreIdentifier("EList") + "<" + g.typeOf(t) + ">"
	}
	return g.typeOf(t) + g.typeArguments(el)
}

// typeArguments are the type arguments of the generic type of a single typed
// element, "<A,B>", but for EJavaClass (v1 generateType): any for a wildcard.
func (g *gen) typeArguments(el ecore.ETypedElement) string {
	args := g.m.TypeArguments(el)
	if len(args) == 0 || el.GetEType().GetName() == "EJavaClass" {
		return ""
	}
	types := make([]string, len(args))
	for i, arg := range args {
		types[i] = g.typeOf(arg)
	}
	return "<" + strings.Join(types, ",") + ">"
}

// mapEntryFeatureQID is the qualified type of the key or value feature of a
// map entry class, lists included (v1 getMapTypeParameter).
func (g *gen) mapEntryFeatureQID(class ecore.EClass, name string) string {
	f := class.GetEStructuralFeatureFromName(name)
	if f == nil {
		return "any"
	}
	return g.typedElementQID(f)
}

// dataTypeValue is a value of the first row matching an instance type name
// whose value is not empty: v1 iterated the rows with an empty result as
// "not found" (the test value of any is null, from java.lang.Object, the row
// of java.lang.Class having none).
func (g *gen) dataTypeValue(name string, value func(dataTypeRow) string) string {
	for _, row := range g.dataTypes() {
		if row.matches(name) {
			if v := value(row); v != "" {
				return v
			}
		}
	}
	return ""
}
