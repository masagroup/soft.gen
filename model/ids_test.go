package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// packageConstants returns the package constants of the model, by name: the
// classifier ids as literals, the feature, operation and count ids of the
// classes. It fails on a duplicate name.
func packageConstants(t *testing.T, m *Model) map[string]Expr {
	t.Helper()
	info := m.Package(m.Root())
	constants := map[string]Expr{}
	put := func(name string, value Expr) {
		_, duplicate := constants[name]
		assert.False(t, duplicate, "duplicate constant %s", name)
		constants[name] = value
	}
	for _, c := range info.OrderedClasses {
		put(m.ClassifierIDName(c), literalExpr(m.ClassifierIDValue(c)))
		for _, f := range elements[ecore.EStructuralFeature](c.GetEAllStructuralFeatures()) {
			put(m.FeatureIDName(c, f), m.FeatureIDValue(c, f))
		}
		put(m.FeatureCountIDName(c), m.FeatureCountIDValue(c))
		for _, op := range elements[ecore.EOperation](c.GetEAllOperations()) {
			put(m.OperationIDName(c, op), m.OperationIDValue(c, op))
		}
		put(m.OperationCountIDName(c), m.OperationCountIDValue(c))
	}
	for _, d := range info.DataTypes {
		put(m.ClassifierIDName(d), literalExpr(m.ClassifierIDValue(d)))
	}
	for _, e := range info.Enums {
		put(m.ClassifierIDName(e), literalExpr(m.ClassifierIDValue(e)))
	}
	return constants
}

// TestIDsMatchEcore checks, for every class of the ecore metamodel and of
// fixtures, that the values of the feature, operation and count constants
// are the ids ecore computes (GetFeatureID, GetFeatureCount, GetOperationID,
// GetOperationCount), that the constant names are unique in the package, that
// a value expressed from a base constant names a constant of the package whose
// value plus the offset is the value, and that the memoized ids of a class
// are the computed ones.
func TestIDsMatchEcore(t *testing.T) {
	models := map[string]*Model{"ecore metamodel": ecoreMetamodel(t)}
	for _, name := range []string{"inheritance", "operations", "metamodel", "shop", "references", "maps", "enums", "attributes"} {
		models[name] = loadModel(t, name)
	}
	for name, m := range models {
		t.Run(name, func(t *testing.T) {
			info := m.Package(m.Root())
			require.NotEmpty(t, info.OrderedClasses)
			for _, c := range info.OrderedClasses {
				for _, f := range elements[ecore.EStructuralFeature](c.GetEAllStructuralFeatures()) {
					assert.Equal(t, c.GetFeatureID(f), m.FeatureIDValue(c, f).Int(), "%s.%s", c.GetName(), f.GetName())
				}
				assert.Equal(t, c.GetFeatureCount(), m.FeatureCountIDValue(c).Int(), c.GetName())
				for _, op := range elements[ecore.EOperation](c.GetEAllOperations()) {
					assert.Equal(t, c.GetOperationID(op), m.OperationIDValue(c, op).Int(), "%s.%s", c.GetName(), op.GetName())
				}
				assert.Equal(t, c.GetOperationCount(), m.OperationCountIDValue(c).Int(), c.GetName())
			}
			for _, classifier := range elements[ecore.EClassifier](m.Root().GetEClassifiers()) {
				if id := classifier.GetClassifierID(); id != 0 {
					assert.Equal(t, id, m.ClassifierIDValue(classifier), classifier.GetName())
				}
			}

			constants := packageConstants(t, m)
			for constantName, value := range constants {
				if !value.HasBase {
					continue
				}
				base, found := constants[value.Base]
				if assert.True(t, found, "%s = %s: unknown base", constantName, value) {
					assert.Equal(t, value.Int(), base.Int()+value.Offset, "%s = %s", constantName, value)
				}
			}

			// the memoized ids of a class are the computed ones
			for _, c := range info.OrderedClasses {
				ids := m.Class(c).IDs
				assert.Equal(t, m.ClassifierIDName(c), ids.ClassifierIDName)
				assert.Equal(t, m.ClassifierIDValue(c), ids.ClassifierIDValue)
				require.Equal(t, c.GetEAllStructuralFeatures().Size(), len(ids.FeatureIDs))
				require.Equal(t, c.GetEAllOperations().Size(), len(ids.OperationIDs))
				for _, id := range ids.FeatureIDs {
					assert.Equal(t, constants[id.Name], id.Value, id.Name)
					assert.Same(t, id.Element, c.GetEAllStructuralFeatures().Get(id.Value.Int()))
				}
				for _, id := range ids.OperationIDs {
					assert.Equal(t, constants[id.Name], id.Value, id.Name)
					assert.Same(t, id.Element, c.GetEAllOperations().Get(id.Value.Int()))
				}
				assert.Equal(t, constants[ids.FeatureCountIDName], ids.FeatureCountIDValue)
				assert.Equal(t, constants[ids.OperationCountName], ids.OperationCountValue)
			}
		})
	}
}

// TestIDsInheritance pins the constant rules on inheritance.ecore: a class
// without base has literal values; an inherited feature is the constant of
// the base class; an added feature, and the count, is the count constant of
// the base class plus an offset (printed even when 0); in a diamond the
// features of the second branch are added features; the names are the upper
// snake case of the class and feature names.
func TestIDsInheritance(t *testing.T) {
	constants := packageConstants(t, loadModel(t, "inheritance"))
	tests := []struct {
		name, value string
		integer     int
	}{
		{"IDENTIFIED", "15", 15}, // id 0: the classifier count
		{"IDENTIFIED__ID", "0", 0},
		{"IDENTIFIED_FEATURE_COUNT", "1", 1},
		{"IDENTIFIED_OPERATION_COUNT", "0", 0},
		{"BASE", "1", 1},
		{"BASE__ID", "IDENTIFIED__ID", 0},
		{"BASE__BASE", "IDENTIFIED_FEATURE_COUNT + 0", 1},
		{"BASE_FEATURE_COUNT", "IDENTIFIED_FEATURE_COUNT + 1", 2},
		{"BASE_OPERATION_COUNT", "IDENTIFIED_OPERATION_COUNT + 0", 0},
		{"LEFT__ID", "BASE__ID", 0},
		{"LEFT__LEFT", "BASE_FEATURE_COUNT + 0", 2},
		{"LEFT_FEATURE_COUNT", "BASE_FEATURE_COUNT + 1", 3},
		{"RIGHT__RIGHT", "BASE_FEATURE_COUNT + 0", 2},
		{"RIGHT__PARTNER", "BASE_FEATURE_COUNT + 1", 3},
		{"BOTTOM__LEFT", "LEFT__LEFT", 2},
		{"BOTTOM__RIGHT", "LEFT_FEATURE_COUNT + 0", 3},
		{"BOTTOM__PARTNER", "LEFT_FEATURE_COUNT + 1", 4},
		{"BOTTOM__BOTTOM", "LEFT_FEATURE_COUNT + 2", 5},
		{"BOTTOM_FEATURE_COUNT", "LEFT_FEATURE_COUNT + 3", 6},
		{"ITEM__TAGS", "IDENTIFIED_FEATURE_COUNT + 0", 1},
		{"ITEM__WEIGHT", "IDENTIFIED_FEATURE_COUNT + 1", 2},
		{"SUB_EXT", "8", 8},
		{"SUB_EXT__VALUE", "EXT__VALUE", 0},
		{"SUB_EXT__EXTRA", "EXT_FEATURE_COUNT + 0", 1},
		{"ECLASSIFIER__CLASSIFIER_ID", "0", 0},
		{"ECLASS__CLASSIFIER_ID", "ECLASSIFIER__CLASSIFIER_ID", 0},
		{"ECLASS__HOLDER", "ECLASSIFIER_FEATURE_COUNT + 0", 1},
		{"EATTRIBUTE__FEATURE_ID", "0", 0},
		{"EOPERATION__OPERATION_ID", "0", 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := constants[test.name]
			require.True(t, found)
			assert.Equal(t, test.value, value.String())
			assert.Equal(t, test.integer, value.Int())
		})
	}
}

// TestIDsOperations pins the operation constant rules on operations.ecore: the
// name is suffixed by the type of the single parameter, or by the types of
// the first and the last parameters, in upper case; an inherited operation is
// the constant of the base class, an added one the operation count constant
// of the base class plus an offset, the operations of a mixin being added.
func TestIDsOperations(t *testing.T) {
	constants := packageConstants(t, loadModel(t, "operations"))
	tests := []struct {
		name, value string
		integer     int
	}{
		{"CALCULATOR", "6", 6}, // id 0: the classifier count
		{"CALCULATOR__CLEAR_EBOOLEAN", "0", 0},
		{"CALCULATOR__COMPUTE_EINT_EDOUBLE", "1", 1},
		{"CALCULATOR__NAMES", "4", 4},
		{"CALCULATOR__PROCESS_EJAVAOBJECT_EJAVAOBJECT", "5", 5},
		{"CALCULATOR__RESET", "6", 6},
		{"CALCULATOR__PEER_GREETER", "9", 9},
		{"CALCULATOR_OPERATION_COUNT", "10", 10},
		{"CALCULATOR_FEATURE_COUNT", "0", 0},
		{"GREETER", "1", 1},
		{"POLITE__GREET_ESTRING", "GREETER__GREET_ESTRING", 0},
		{"POLITE__WAVE", "GREETER__WAVE", 1},
		{"POLITE_OPERATION_COUNT", "GREETER_OPERATION_COUNT + 0", 2},
		{"CIRCLE__LABEL", "SHAPE__LABEL", 0},
		{"CIRCLE__RADIUS", "SHAPE_FEATURE_COUNT + 0", 1},
		{"CIRCLE__DRAW", "SHAPE__DRAW", 0},
		{"CIRCLE__SCALE_EDOUBLE", "SHAPE_OPERATION_COUNT + 0", 1},
		{"CIRCLE__AREA", "SHAPE_OPERATION_COUNT + 1", 2},
		{"CIRCLE_OPERATION_COUNT", "SHAPE_OPERATION_COUNT + 2", 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := constants[test.name]
			require.True(t, found)
			assert.Equal(t, test.value, value.String())
			assert.Equal(t, test.integer, value.Int())
		})
	}
}

// TestExpr checks the literal, base and offset forms of a constant value.
func TestExpr(t *testing.T) {
	assert.Equal(t, "3", literalExpr(3).String())
	assert.Equal(t, 3, literalExpr(3).Int())
	assert.Equal(t, "A__B", baseExpr("A__B", 2).String())
	assert.Equal(t, 2, baseExpr("A__B", 2).Int())
	assert.Equal(t, "A_FEATURE_COUNT + 0", offsetExpr("A_FEATURE_COUNT", 4, 0).String())
	assert.Equal(t, 4, offsetExpr("A_FEATURE_COUNT", 4, 0).Int())
	assert.Equal(t, "A_FEATURE_COUNT + 2", offsetExpr("A_FEATURE_COUNT", 4, 2).String())
	assert.Equal(t, 6, offsetExpr("A_FEATURE_COUNT", 4, 2).Int())
}

// TestClassifierIDValueQuirk checks that the first classifier, of id 0, has
// the number of classifiers of its package as value (v1).
func TestClassifierIDValueQuirk(t *testing.T) {
	m := loadModel(t, "operations")
	calculator := classNamed(t, m, "Calculator")
	assert.Equal(t, 0, calculator.GetClassifierID())
	assert.Equal(t, 6, m.Root().GetEClassifiers().Size())
	assert.Equal(t, 6, m.ClassifierIDValue(calculator))
	assert.Equal(t, 1, m.ClassifierIDValue(classNamed(t, m, "Greeter")))
	assert.Equal(t, 5, m.ClassifierIDValue(classNamed(t, m, "Circle")))
	// the first classifier of enums.ecore, a class followed by enums
	me := loadModel(t, "enums")
	assert.Equal(t, 7, me.ClassifierIDValue(classNamed(t, me, "Lamp")))
	assert.Equal(t, 1, me.ClassifierIDValue(me.Root().GetEClassifier("Color")))
}

// TestNames checks the constant, member, accessor and identifier names of the
// real Ecore elements, overloaded operations included.
func TestNames(t *testing.T) {
	m := ecoreMetamodel(t)
	eClass := classNamed(t, m, "EClass")
	eModelElement := classNamed(t, m, "EModelElement")
	assert.Equal(t, "ECLASS", m.ClassifierIDName(eClass))
	assert.Equal(t, "EMODEL_ELEMENT", m.ClassifierIDName(eModelElement))
	assert.Equal(t, "ESTRING_TO_STRING_MAP_ENTRY", m.ClassifierIDName(m.Root().GetEClassifier("EStringToStringMapEntry")))

	eSuperTypes := featureNamed(t, m, "EClass", "eSuperTypes")
	assert.Equal(t, "eClass", m.MemberName(eClass))
	assert.Equal(t, "eClass_ESuperTypes", m.MemberName(eSuperTypes))
	assert.Equal(t, "EClass", m.ElementAccessorName(eClass))
	assert.Equal(t, "EClass_ESuperTypes", m.ElementAccessorName(eSuperTypes))
	assert.Equal(t, "eSuperTypes", m.ElementIdentifier(eSuperTypes))
	assert.Equal(t, "ECLASS__ESUPER_TYPES", m.FeatureIDName(eClass, eSuperTypes))

	var getFeatureCount, getFeatureFromID, getFeatureFromName, isSuperTypeOf ecore.EOperation
	for _, op := range elements[ecore.EOperation](eClass.GetEOperations()) {
		switch {
		case op.GetName() == "getFeatureCount":
			getFeatureCount = op
		case op.GetName() == "isSuperTypeOf":
			isSuperTypeOf = op
		case op.GetName() == "getEStructuralFeature" && getFeatureFromID == nil:
			getFeatureFromID = op
		case op.GetName() == "getEStructuralFeature":
			getFeatureFromName = op
		}
	}
	require.NotNil(t, getFeatureCount)
	require.NotNil(t, getFeatureFromID)
	require.NotNil(t, getFeatureFromName)
	require.NotNil(t, isSuperTypeOf)
	assert.Equal(t, "eClass_GetFeatureCount", m.MemberName(getFeatureCount))
	assert.Equal(t, "getFeatureCount", m.ElementIdentifier(getFeatureCount))
	assert.Equal(t, "eClass_GetEStructuralFeature_EInt", m.MemberName(getFeatureFromID))
	assert.Equal(t, "getEStructuralFeature_EInt", m.ElementIdentifier(getFeatureFromID))
	assert.Equal(t, "EClass_GetEStructuralFeature_EString", m.ElementAccessorName(getFeatureFromName))
	assert.Equal(t, "ECLASS__IS_SUPER_TYPE_OF_ECLASS", m.OperationIDName(eClass, isSuperTypeOf))
	assert.Equal(t, "isSuperTypeOf_EClass", m.Operation(isSuperTypeOf).Identifier)

	// an operation with several parameters is suffixed by the first and the last types
	eOperation := classNamed(t, m, "EOperation")
	eObject := classNamed(t, m, "EObject")
	var eInvoke ecore.EOperation
	for _, op := range elements[ecore.EOperation](eObject.GetEOperations()) {
		if op.GetName() == "eInvoke" {
			eInvoke = op
		}
	}
	require.NotNil(t, eInvoke)
	assert.Equal(t, "EOBJECT__EINVOKE_EOPERATION_EELIST", m.OperationIDName(eObject, eInvoke))
	assert.Equal(t, "eInvoke_EOperation_EEList", m.ElementIdentifier(eInvoke))

	assert.Equal(t, "Classifier", m.ElementIDAccessorName(eClass))
	assert.Equal(t, "Classifier", m.ElementIDAccessorName(classNamed(t, m, "EEnum")))
	assert.Equal(t, "Classifier", m.ElementIDAccessorName(classNamed(t, m, "EDataType")))
	assert.Equal(t, "Feature", m.ElementIDAccessorName(classNamed(t, m, "EAttribute")))
	assert.Equal(t, "Feature", m.ElementIDAccessorName(classNamed(t, m, "EReference")))
	assert.Equal(t, "Operation", m.ElementIDAccessorName(eOperation))
	assert.Equal(t, "", m.ElementIDAccessorName(eModelElement))

	// a non classifier/feature/operation named element
	literal := m.Root().GetEClassifier("EString")
	assert.Equal(t, "eString", m.MemberName(literal))
	param := isSuperTypeOf.GetEParameters().Get(0).(ecore.ENamedElement)
	assert.Equal(t, "someClass", m.MemberName(param))
	assert.Equal(t, "SomeClass", m.ElementAccessorName(param))
}

// TestNamesFixture checks the member, accessor and identifier names of the
// elements of operations.ecore: a feature is class_Feature, an operation
// class_Operation suffixed by the type of its single parameter or of its first
// and last parameters, unsuffixed without parameter.
func TestNamesFixture(t *testing.T) {
	m := loadModel(t, "operations")
	tests := []struct {
		element                    ecore.ENamedElement
		member, accessor, identity string
	}{
		{classNamed(t, m, "Calculator"), "calculator", "Calculator", "Calculator"},
		{featureNamed(t, m, "Circle", "radius"), "circle_Radius", "Circle_Radius", "radius"},
		{operationNamed(t, m, "Calculator", "reset", 0), "calculator_Reset", "Calculator_Reset", "reset"},
		{operationNamed(t, m, "Calculator", "clear", 0), "calculator_Clear_EBoolean", "Calculator_Clear_EBoolean", "clear_EBoolean"},
		{operationNamed(t, m, "Calculator", "compute", 0), "calculator_Compute_EInt_EDouble", "Calculator_Compute_EInt_EDouble", "compute_EInt_EDouble"},
		{operationNamed(t, m, "Calculator", "peer", 0), "calculator_Peer_Greeter", "Calculator_Peer_Greeter", "peer_Greeter"},
		{operationNamed(t, m, "Measurable", "scale", 0), "measurable_Scale_EDouble", "Measurable_Scale_EDouble", "scale_EDouble"},
		// a parameter is named after its name
		{operationNamed(t, m, "Calculator", "compute", 0).GetEParameters().Get(1).(ecore.ENamedElement), "y", "Y", "y"},
	}
	for _, test := range tests {
		t.Run(test.member, func(t *testing.T) {
			assert.Equal(t, test.member, m.MemberName(test.element))
			assert.Equal(t, test.accessor, m.ElementAccessorName(test.element))
			assert.Equal(t, test.identity, m.ElementIdentifier(test.element))
		})
	}
	// the operation of a mixin is named in the class it is inherited by
	circle := classNamed(t, m, "Circle")
	assert.Equal(t, "CIRCLE__SCALE_EDOUBLE", m.OperationIDName(circle, operationNamed(t, m, "Measurable", "scale", 0)))
	assert.Equal(t, "CIRCLE__RADIUS", m.FeatureIDName(circle, featureNamed(t, m, "Circle", "radius")))
	assert.Equal(t, "CIRCLE_FEATURE_COUNT", m.FeatureCountIDName(circle))
	assert.Equal(t, "CIRCLE_OPERATION_COUNT", m.OperationCountIDName(circle))
}

// TestFeatureKind checks the kind of a feature: attribute, reference,
// containment reference, container reference or map, " list" appended to a
// many non-map feature.
func TestFeatureKind(t *testing.T) {
	tests := []struct {
		model, class, feature string
		kind                  string
	}{
		{"attributes", "Holder", "label", "attribute"},
		{"attributes", "Holder", "values", "attribute list"},
		{"references", "Root", "favorite", "reference"},
		{"references", "Root", "links", "reference list"},
		{"references", "Root", "items", "containment reference list"},
		{"references", "Root", "optional", "containment reference"},
		{"references", "Child", "parent", "container reference"},
		{"maps", "Holder", "values", "map"},
		{"maps", "OwnedEntry", "holder", "container reference"},
	}
	for _, test := range tests {
		t.Run(test.model+"/"+test.class+"."+test.feature, func(t *testing.T) {
			m := loadModel(t, test.model)
			assert.Equal(t, test.kind, m.FeatureKind(featureNamed(t, m, test.class, test.feature)))
		})
	}

	// a non containment list of map entries is a reference list (no fixture
	// has one: added to the loaded model)
	m := loadModel(t, "maps")
	indexes := ecore.GetFactory().CreateEReferenceFromContainer(classNamed(t, m, "Holder"))
	indexes.SetName("indexes")
	indexes.SetUpperBound(-1)
	indexes.SetEType(classNamed(t, m, "NamesToAnyEntry"))
	assert.Equal(t, "reference list", m.FeatureKind(indexes))
}
