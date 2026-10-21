package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
)

// TestOperation checks the operation facts on the real Ecore operations, and the
// default key.
func TestOperation(t *testing.T) {
	m := ecoreMetamodel(t)
	eClass := classNamed(t, m, "EClass")

	isSuperTypeOf := m.Operation(operationNamed(t, m, "EClass", "isSuperTypeOf", 0))
	assert.Same(t, eClass, isSuperTypeOf.Class)
	assert.False(t, isSuperTypeOf.IsVoid)
	assert.True(t, isSuperTypeOf.IsDefault)
	assert.Equal(t, "isSuperTypeOf_EClass", isSuperTypeOf.Identifier)
	assert.Equal(t, []string{"EBoolean", "EClass"}, names(isSuperTypeOf.ReferencedTypes))
	assert.False(t, isSuperTypeOf.IsList)
	assert.False(t, isSuperTypeOf.IsMap)

	getFeatureCount := m.Operation(operationNamed(t, m, "EClass", "getFeatureCount", 0))
	assert.Equal(t, "getFeatureCount", getFeatureCount.Identifier)
	assert.Equal(t, []string{"EInt"}, names(getFeatureCount.ReferencedTypes))

	// the generic return type is the type (v1 ran on EMF)
	eContents := m.Operation(operationNamed(t, m, "EObject", "eContents", 0))
	assert.False(t, eContents.IsVoid)
	assert.False(t, eContents.IsDefault) // default=false
	assert.Equal(t, []string{"EEList"}, names(eContents.ReferencedTypes))
	eInvoke := m.Operation(operationNamed(t, m, "EObject", "eInvoke", 0))
	assert.Equal(t, "eInvoke_EOperation_EEList", eInvoke.Identifier)
	assert.Equal(t, []string{"EJavaObject", "EOperation", "EEList"}, names(eInvoke.ReferencedTypes))

	// a void operation; a duplicate parameter type is listed once
	eUnset := m.Operation(operationNamed(t, m, "EObject", "eUnset", 0))
	assert.True(t, eUnset.IsVoid)
	assert.Equal(t, []string{"EStructuralFeature"}, names(eUnset.ReferencedTypes))
	eSet := m.Operation(operationNamed(t, m, "EObject", "eSet", 0))
	assert.True(t, eSet.IsVoid)
	assert.Equal(t, "eSet_EStructuralFeature_EJavaObject", eSet.Identifier)
	assert.Equal(t, []string{"EStructuralFeature", "EJavaObject"}, names(eSet.ReferencedTypes))

	// default = true is explicit default
	p := ecore.GetFactory().CreateEPackage()
	c := ecore.GetFactory().CreateEClassFromContainer(p)
	op := ecore.GetFactory().CreateEOperationFromContainer(c)
	op.SetName("op")
	op.SetUpperBound(-1)
	annotation := ecore.GetFactory().CreateEAnnotationFromContainer(op)
	annotation.SetSource(genGoSource)
	annotation.GetDetails().Put("default", "true")
	mp := FromPackage(p, genGoSource)
	assert.True(t, mp.Operation(op).IsDefault)
	assert.True(t, mp.Operation(op).IsList)
	assert.True(t, mp.Operation(op).IsVoid)
	assert.Empty(t, mp.Operation(op).ReferencedTypes)
	annotation.GetDetails().Put("default", "no")
	assert.False(t, FromPackage(p, genGoSource).Operation(op).IsDefault)
}

// TestOperationFixtures checks the operation facts on operations.ecore: void
// without type, default unless default=false, the identifier suffixed by the
// type of the single parameter or of the first and last parameters, the
// referenced types (return then parameters, without duplicate), list/map.
func TestOperationFixtures(t *testing.T) {
	m := loadModel(t, "operations")
	tests := []struct {
		class, operation string
		identifier       string
		isVoid           bool
		isDefault        bool
		referencedTypes  []string
		isList           bool
	}{
		{"Calculator", "clear", "clear_EBoolean", true, true, []string{"EBoolean"}, false},
		{"Calculator", "compute", "compute_EInt_EDouble", false, true, []string{"EDouble", "EInt"}, false},
		{"Calculator", "custom", "custom_EString", false, false, []string{"EString"}, false},
		{"Calculator", "filter", "filter_EEList", false, true, []string{"EEList"}, false},
		{"Calculator", "names", "names", false, true, []string{"EString"}, true},
		{"Calculator", "process", "process_EJavaObject_EJavaObject", false, true, []string{"EJavaObject"}, false},
		{"Calculator", "reset", "reset", true, true, nil, false},
		{"Calculator", "sum", "sum_EInt", false, true, []string{"EInt"}, false},
		{"Calculator", "table", "table", false, true, []string{"EMap"}, false},
		{"Calculator", "peer", "peer_Greeter", false, true, []string{"Greeter"}, false},
		{"Greeter", "greet", "greet_EString", false, true, []string{"EString"}, false},
	}
	for _, test := range tests {
		t.Run(test.class+"."+test.operation, func(t *testing.T) {
			info := m.Operation(operationNamed(t, m, test.class, test.operation, 0))
			assert.Same(t, classNamed(t, m, test.class), info.Class)
			assert.Equal(t, test.identifier, info.Identifier)
			assert.Equal(t, test.isVoid, info.IsVoid, "IsVoid")
			assert.Equal(t, test.isDefault, info.IsDefault, "IsDefault")
			assert.Equal(t, orEmpty(test.referencedTypes), names(info.ReferencedTypes), "ReferencedTypes")
			assert.Equal(t, test.isList, info.IsList, "IsList")
			assert.False(t, info.IsMap) // no list of map entries
		})
	}
}
