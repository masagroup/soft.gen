package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureBidirectionalReference(t *testing.T) {
	m := loadModel(t, "library")
	f := m.Feature(featureNamed(t, m, "Book", "author"))
	assert.Same(t, classNamed(t, m, "Book"), f.Class)
	require.NotNil(t, f.Reference)
	assert.Nil(t, f.Attribute)
	assert.Same(t, featureNamed(t, m, "Writer", "books"), f.Opposite)
	assert.True(t, f.IsBidirectional)
	assert.False(t, f.IsContainer)
	assert.False(t, f.IsContainment)
	assert.True(t, f.IsChangeable)
	assert.False(t, f.IsUnsettable)
	assert.False(t, f.IsVolatile)
	assert.False(t, f.IsList)
	assert.False(t, f.IsMap)
	assert.False(t, f.IsFeatureMap)
	assert.True(t, f.IsProxy)
	assert.True(t, f.IsField)
	assert.False(t, f.IsBackReference)
	assert.Equal(t, "reference", f.Kind)
	assert.True(t, f.Get)
	assert.True(t, f.Set)
	assert.False(t, f.Unset)
	assert.False(t, f.IsSet)
	assert.True(t, f.BasicGet)
	assert.True(t, f.BasicSet)
	assert.False(t, f.BasicUnset)
	assert.True(t, f.RequiresNotification)
	assert.Equal(t, LazyInfo{Mode: "no", InitMethodName: "initAuthor", Deferred: true}, f.Lazy)
	assert.False(t, f.Lazy.IsLazy())
}

func TestFeatureContainmentList(t *testing.T) {
	m := loadModel(t, "library")
	f := m.Feature(featureNamed(t, m, "Library", "writers"))
	assert.True(t, f.IsContainment)
	assert.False(t, f.IsContainer)
	assert.False(t, f.IsBidirectional)
	assert.Nil(t, f.Opposite)
	assert.True(t, f.IsList)
	assert.False(t, f.IsMap)
	assert.False(t, f.IsProxy) // resolveProxies="false", and a containment anyway
	assert.True(t, f.IsField)
	assert.False(t, f.IsBackReference)
	assert.Equal(t, "containment reference list", f.Kind)
	assert.False(t, f.Set)
	assert.False(t, f.BasicSet)
	assert.False(t, f.BasicGet)
	assert.False(t, f.BasicUnset)
	assert.False(t, f.RequiresNotification)
	assert.True(t, f.Lazy.Deferred)

	// a bidirectional containment
	branches := m.Feature(featureNamed(t, m, "Library", "branches"))
	assert.True(t, branches.IsContainment)
	assert.True(t, branches.IsBidirectional)
	assert.True(t, branches.IsList)
	assert.False(t, branches.BasicSet)
	assert.Equal(t, "containment reference list", branches.Kind)

	// a single containment
	proprietary := m.Feature(featureNamed(t, m, "Library", "proprietary"))
	assert.True(t, proprietary.IsContainment)
	assert.False(t, proprietary.IsList)
	assert.False(t, proprietary.IsProxy)
	assert.True(t, proprietary.Set)
	assert.True(t, proprietary.BasicSet)
	assert.False(t, proprietary.BasicGet)
	assert.Equal(t, "containment reference", proprietary.Kind)
}

func TestFeatureContainer(t *testing.T) {
	m := loadModel(t, "library")
	f := m.Feature(featureNamed(t, m, "Library", "parentBranch"))
	assert.True(t, f.IsContainer)
	assert.True(t, f.IsBidirectional)
	assert.False(t, f.IsContainment)
	assert.False(t, f.IsProxy)
	assert.False(t, f.IsField)
	assert.True(t, f.IsBackReference)
	assert.False(t, f.IsList)
	assert.Equal(t, "container reference", f.Kind)
	assert.True(t, f.Set)
	assert.True(t, f.BasicSet)
	assert.False(t, f.BasicGet)
	assert.False(t, f.Unset)
	assert.False(t, f.IsSet)
	assert.False(t, f.Lazy.Deferred)
}

func TestFeatureMap(t *testing.T) {
	m := loadModel(t, "library")
	indexes := m.Feature(featureNamed(t, m, "Book", "indexes"))
	assert.True(t, indexes.IsList)
	assert.True(t, indexes.IsMap)
	assert.False(t, indexes.IsFeatureMap)
	assert.False(t, indexes.IsContainment)
	assert.Equal(t, "reference list", indexes.Kind)

	xmlns := m.Feature(featureNamed(t, m, "DocumentRoot", "xMLNSPrefixMap"))
	assert.True(t, xmlns.IsMap)
	assert.True(t, xmlns.IsContainment)
	assert.Equal(t, "map", xmlns.Kind)

	people := m.Feature(featureNamed(t, m, "Library", "people"))
	assert.True(t, people.IsFeatureMap)
	assert.False(t, people.IsMap)
	assert.True(t, people.IsList)
	assert.NotNil(t, people.Attribute)
	assert.Equal(t, "attribute list", people.Kind)

	bookIndex := m.Class(classNamed(t, m, "BookIndex"))
	assert.True(t, bookIndex.IsMapEntry)
	assert.False(t, m.Class(classNamed(t, m, "Book")).IsMapEntry)
	assert.True(t, m.Feature(featureNamed(t, m, "BookIndex", "value")).IsList)
}

func TestFeatureUnsettable(t *testing.T) {
	m := loadModel(t, "library")
	f := m.Feature(featureNamed(t, m, "Book", "category"))
	require.NotNil(t, f.Attribute)
	assert.Nil(t, f.Reference)
	assert.True(t, f.IsUnsettable)
	assert.True(t, f.IsSet)
	assert.True(t, f.Unset)
	assert.True(t, f.Set)
	assert.False(t, f.BasicSet)
	assert.False(t, f.BasicUnset) // not a reference
	assert.True(t, f.RequiresNotification)
	assert.Equal(t, "attribute", f.Kind)
	assert.Equal(t, "BookCategory", f.Feature.GetEType().GetName())
}

func TestFeatureVolatile(t *testing.T) {
	m := loadModel(t, "library")
	name := m.Feature(featureNamed(t, m, "Writer", "name"))
	assert.True(t, name.IsVolatile)
	assert.False(t, name.IsField)
	assert.True(t, name.Set)
	assert.False(t, name.Lazy.HasInitMethod)

	// a feature is volatile when its opposite is
	me := loadModel(t, "ecore")
	eAttributeType := me.Feature(featureNamed(t, me, "EAttribute", "eAttributeType"))
	assert.True(t, eAttributeType.IsVolatile)
	assert.False(t, eAttributeType.IsChangeable)
	assert.False(t, eAttributeType.Set)
	assert.False(t, eAttributeType.BasicSet)
	assert.True(t, eAttributeType.IsProxy)
	assert.True(t, eAttributeType.BasicGet)
}

func TestFeatureBasicUnset(t *testing.T) {
	// a changeable unsettable single bidirectional reference has a basic unset
	p := ecore.GetFactory().CreateEPackage()
	p.SetName("p")
	a := ecore.GetFactory().CreateEClassFromContainer(p)
	a.SetName("A")
	b := ecore.GetFactory().CreateEClassFromContainer(p)
	b.SetName("B")
	ab := ecore.GetFactory().CreateEReferenceFromContainer(a)
	ab.SetName("b")
	ab.SetEType(b)
	ab.SetUnsettable(true)
	ba := ecore.GetFactory().CreateEReferenceFromContainer(b)
	ba.SetName("a")
	ba.SetEType(a)
	ab.SetEOpposite(ba)
	ba.SetEOpposite(ab)
	m := FromPackage(p, genGoSource)
	f := m.Feature(ab)
	assert.True(t, f.BasicUnset)
	assert.True(t, f.Unset)
	assert.True(t, f.BasicSet)
	assert.False(t, m.Feature(ba).BasicUnset)

	// a shared aggregation is a back reference
	annotation := ecore.GetFactory().CreateEAnnotationFromContainer(ba)
	annotation.SetSource(genGoSource)
	annotation.GetDetails().Put("aggregation", "shared")
	m = FromPackage(p, genGoSource)
	assert.True(t, m.Feature(ba).IsBackReference)
	assert.False(t, m.Feature(ba).IsContainer)
	assert.False(t, m.Feature(ab).IsBackReference)
}

func TestFeatureEcoreAccessors(t *testing.T) {
	m := loadModel(t, "ecore")
	// EStructuralFeature.eContainingClass is a container
	eContainingClass := m.Feature(featureNamed(t, m, "EStructuralFeature", "eContainingClass"))
	assert.True(t, eContainingClass.IsContainer)
	assert.True(t, eContainingClass.IsBackReference)
	assert.False(t, eContainingClass.IsUnsettable) // unsettable needs not container
	assert.False(t, eContainingClass.IsChangeable)
	assert.False(t, eContainingClass.BasicSet) // a non changeable container
	assert.False(t, eContainingClass.Set)
	assert.False(t, eContainingClass.IsField)
	// EModelElement.eAnnotations is a bidirectional containment list
	eAnnotations := m.Feature(featureNamed(t, m, "EModelElement", "eAnnotations"))
	assert.True(t, eAnnotations.IsContainment)
	assert.True(t, eAnnotations.IsBidirectional)
	assert.Equal(t, "containment reference list", eAnnotations.Kind)
	// EReference.eOpposite is a proxy
	eOpposite := m.Feature(featureNamed(t, m, "EReference", "eOpposite"))
	assert.True(t, eOpposite.IsProxy)
	assert.True(t, eOpposite.BasicGet)
	assert.False(t, eOpposite.BasicSet)
	assert.True(t, eOpposite.Lazy.Deferred)
}
