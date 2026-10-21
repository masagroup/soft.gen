package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
)

func TestOperation(t *testing.T) {
	m := loadModel(t, "ecore")
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
