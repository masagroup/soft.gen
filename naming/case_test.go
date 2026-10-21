package naming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpperSnake(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", ""},
		{"a", "A"},
		{"ab", "AB"},
		{"aB", "AB"},
		{"EClass", "ECLASS"},
		{"EInt", "EINT"},
		{"URI2", "URI2"},
		{"XMLSchema", "XML_SCHEMA"},
		{"name_with_underscore", "NAME_WITH_UNDERSCORE"},
		{"EStringToStringMapEntry", "ESTRING_TO_STRING_MAP_ENTRY"},
		{"EFeatureMapEntry", "EFEATURE_MAP_ENTRY"},
		{"eAllStructuralFeatures", "EALL_STRUCTURAL_FEATURES"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, UpperSnake(test.name))
		})
	}
}

// TestUpperSnakeEcoreOracle checks UpperSnake against the feature and operation
// constants that the v1 generator produced in soft.go/ecore/ecorepackage.go:
// the suffix after "__" of ECLASS__... and EOPERATION__... is UpperSnake of the
// feature (or operation) name of the ecore model.
func TestUpperSnakeEcoreOracle(t *testing.T) {
	tests := []struct {
		constant string
		name     string
	}{
		{"ECLASS__EALL_STRUCTURAL_FEATURES", "eAllStructuralFeatures"},
		{"ECLASS__EID_ATTRIBUTE", "eIDAttribute"},
		{"ECLASS__INSTANCE_CLASS_NAME", "instanceClassName"},
		{"ECLASS__ESUPER_TYPES", "eSuperTypes"},
		{"ECLASS__ECROSS_REFERENCE_FEATURES", "eCrossReferenceFeatures"},
		{"ECLASS__CLASSIFIER_ID", "classifierID"},
		{"ECLASS__ABSTRACT", "abstract"},
		{"ECLASS__IS_SUPER_TYPE_OF", "isSuperTypeOf"},
		{"ECLASS__GET_ESTRUCTURAL_FEATURE", "getEStructuralFeature"},
		{"ECLASS__GET_FEATURE_ID", "getFeatureID"},
		{"EOPERATION__ECONTAINING_CLASS", "eContainingClass"},
		{"EOPERATION__EPARAMETERS", "eParameters"},
		{"EOPERATION__LOWER_BOUND", "lowerBound"},
		{"EOPERATION__ETYPE", "eType"},
		{"EOPERATION__OPERATION_ID", "operationID"},
		{"EOPERATION__IS_OVERRIDE_OF", "isOverrideOf"},
	}
	for _, test := range tests {
		t.Run(test.constant, func(t *testing.T) {
			expected := test.constant[len("ECLASS__"):]
			if len(test.constant) > len("EOPERATION__") && test.constant[:len("EOPERATION__")] == "EOPERATION__" {
				expected = test.constant[len("EOPERATION__"):]
			}
			assert.Equal(t, expected, UpperSnake(test.name))
		})
	}
}

func TestUpperFirst(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", ""},
		{"b", "B"},
		{"B", "B"},
		{"book", "Book"},
		{"Book", "Book"},
		{"bookShelf", "BookShelf"},
		{"BOOK", "BOOK"},
		{"_book", "_book"},
		{"1book", "1book"},
		{"élève", "Élève"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, UpperFirst(test.name))
		})
	}
}

func TestLowerFirst(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", ""},
		{"B", "b"},
		{"b", "b"},
		{"Book", "book"},
		{"book", "book"},
		{"EClass", "eClass"},
		{"BOOK", "bOOK"},
		{"_Book", "_Book"},
		{"Élève", "élève"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, LowerFirst(test.name))
		})
	}
}
