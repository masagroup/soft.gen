package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFeatureBidirectionalReference checks a single resolving bidirectional
// non-containment reference (references Husband.wife <-> Wife.husband): a
// field and a proxy with every single valued accessor but unset.
func TestFeatureBidirectionalReference(t *testing.T) {
	m := loadModel(t, "references")
	f := m.Feature(featureNamed(t, m, "Husband", "wife"))
	assert.Same(t, classNamed(t, m, "Husband"), f.Class)
	require.NotNil(t, f.Reference)
	assert.Nil(t, f.Attribute)
	assert.Same(t, featureNamed(t, m, "Wife", "husband"), f.Opposite)
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
	assert.Equal(t, LazyInfo{Mode: "no", InitMethodName: "initWife", Deferred: true}, f.Lazy)
	assert.False(t, f.Lazy.IsLazy())

	// the many side of a bidirectional reference is a list without single accessor
	friends := m.Feature(featureNamed(t, m, "Husband", "friends"))
	assert.True(t, friends.IsBidirectional)
	assert.True(t, friends.IsList)
	assert.True(t, friends.IsProxy)
	assert.False(t, friends.Set)
	assert.False(t, friends.BasicSet)
	assert.False(t, friends.BasicGet)
	assert.Equal(t, "reference list", friends.Kind)
}

// TestFeatureContainmentList checks the containment references: a list has
// no single accessor and no notification, a single one has set and basic set
// but no basic get; a containment is never a proxy.
func TestFeatureContainmentList(t *testing.T) {
	m := loadModel(t, "references")
	// a containment list without opposite (resolveProxies defaults to true)
	f := m.Feature(featureNamed(t, m, "Root", "items"))
	assert.True(t, f.IsContainment)
	assert.False(t, f.IsContainer)
	assert.False(t, f.IsBidirectional)
	assert.Nil(t, f.Opposite)
	assert.True(t, f.IsList)
	assert.False(t, f.IsMap)
	assert.False(t, f.IsProxy) // a containment
	assert.True(t, f.IsField)
	assert.False(t, f.IsBackReference)
	assert.Equal(t, "containment reference list", f.Kind)
	assert.False(t, f.Set)
	assert.False(t, f.BasicSet)
	assert.False(t, f.BasicGet)
	assert.False(t, f.BasicUnset)
	assert.False(t, f.RequiresNotification)
	assert.True(t, f.Lazy.Deferred)

	// a bidirectional containment list
	children := m.Feature(featureNamed(t, m, "Root", "children"))
	assert.True(t, children.IsContainment)
	assert.True(t, children.IsBidirectional)
	assert.True(t, children.IsList)
	assert.False(t, children.BasicSet)
	assert.Equal(t, "containment reference list", children.Kind)

	// a single containment
	optional := m.Feature(featureNamed(t, m, "Root", "optional"))
	assert.True(t, optional.IsContainment)
	assert.False(t, optional.IsList)
	assert.False(t, optional.IsProxy)
	assert.True(t, optional.Set)
	assert.True(t, optional.BasicSet)
	assert.False(t, optional.BasicGet)
	assert.Equal(t, "containment reference", optional.Kind)
}

// TestFeatureContainer checks a container reference (the opposite of a
// containment): a back reference, neither a field nor a proxy, with a basic
// set only when changeable.
func TestFeatureContainer(t *testing.T) {
	m := loadModel(t, "references")
	f := m.Feature(featureNamed(t, m, "Child", "parent"))
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

	// a non changeable container has neither set nor basic set
	owner := m.Feature(featureNamed(t, m, "Fixed", "owner"))
	assert.True(t, owner.IsContainer)
	assert.False(t, owner.IsChangeable)
	assert.False(t, owner.Set)
	assert.False(t, owner.BasicSet)
	assert.False(t, owner.RequiresNotification)
}

// TestFeatureMap checks the maps (lists of Map$Entry typed elements), the
// feature maps (FeatureMap$Entry typed attributes) and the map entry classes
// (a Map$Entry instance type with a key and a value feature).
func TestFeatureMap(t *testing.T) {
	m := loadModel(t, "maps")
	values := m.Feature(featureNamed(t, m, "Holder", "values"))
	assert.True(t, values.IsMap)
	assert.True(t, values.IsList)
	assert.True(t, values.IsContainment)
	assert.False(t, values.IsFeatureMap)
	assert.Equal(t, "map", values.Kind)
	// the type of a map is a Map$Entry instance type, whatever its features
	pseudos := m.Feature(featureNamed(t, m, "Holder", "pseudos"))
	assert.True(t, pseudos.IsMap)
	assert.Equal(t, "map", pseudos.Kind)

	assert.True(t, m.Class(classNamed(t, m, "NamesToAnyEntry")).IsMapEntry)
	assert.True(t, m.Class(classNamed(t, m, "OwnedEntry")).IsMapEntry)
	assert.False(t, m.Class(classNamed(t, m, "PseudoEntry")).IsMapEntry) // no key nor value
	assert.False(t, m.Class(classNamed(t, m, "Holder")).IsMapEntry)
	assert.True(t, m.Feature(featureNamed(t, m, "NamesToAnyEntry", "key")).IsList)

	// a non containment list of map entries is a map but a reference list
	// (no fixture has one: added to the loaded model)
	indexes := ecore.GetFactory().CreateEReferenceFromContainer(classNamed(t, m, "Holder"))
	indexes.SetName("indexes")
	indexes.SetUpperBound(-1)
	indexes.SetEType(classNamed(t, m, "NamesToAnyEntry"))
	indexesInfo := m.Feature(indexes)
	assert.True(t, indexesInfo.IsList)
	assert.True(t, indexesInfo.IsMap)
	assert.False(t, indexesInfo.IsContainment)
	assert.Equal(t, "reference list", indexesInfo.Kind)

	// a feature map attribute
	ma := loadModel(t, "attributes")
	mixed := ma.Feature(featureNamed(t, ma, "Holder", "mixed"))
	assert.True(t, mixed.IsFeatureMap)
	assert.False(t, mixed.IsMap)
	assert.True(t, mixed.IsList)
	assert.NotNil(t, mixed.Attribute)
	assert.Equal(t, "attribute list", mixed.Kind)
	assert.False(t, ma.Feature(featureNamed(t, ma, "Holder", "tags")).IsFeatureMap)
}

// TestFeatureUnsettable checks an unsettable single attribute (shop
// Product.category): set, unset and is set accessors, no basic accessor.
func TestFeatureUnsettable(t *testing.T) {
	m := loadModel(t, "shop")
	f := m.Feature(featureNamed(t, m, "Product", "category"))
	require.NotNil(t, f.Attribute)
	assert.Nil(t, f.Reference)
	assert.True(t, f.IsUnsettable)
	assert.True(t, f.IsSet)
	assert.True(t, f.HasESet)
	assert.True(t, f.Unset)
	assert.True(t, f.Set)
	assert.False(t, f.BasicSet)
	assert.False(t, f.BasicUnset) // not a reference
	assert.True(t, f.RequiresNotification)
	assert.Equal(t, "attribute", f.Kind)
	assert.Equal(t, "Category", f.Feature.GetEType().GetName())
}

// TestFeatureVolatile checks the volatile features: not a field, no
// initialization method; a feature is volatile when its opposite is.
func TestFeatureVolatile(t *testing.T) {
	m := loadModel(t, "volatile")
	label := m.Feature(featureNamed(t, m, "Holder", "label"))
	assert.True(t, label.IsVolatile)
	assert.False(t, label.IsField)
	assert.True(t, label.Set)
	assert.False(t, label.Lazy.HasInitMethod)
	// a volatile containment still has a basic set (its body is user code)
	owned := m.Feature(featureNamed(t, m, "Holder", "owned"))
	assert.True(t, owned.IsVolatile)
	assert.True(t, owned.IsContainment)
	assert.False(t, owned.IsField)
	assert.True(t, owned.BasicSet)

	// a feature is volatile when its opposite is
	me := ecoreMetamodel(t)
	eAttributeType := me.Feature(featureNamed(t, me, "EAttribute", "eAttributeType"))
	assert.True(t, eAttributeType.IsVolatile)
	assert.False(t, eAttributeType.IsChangeable)
	assert.False(t, eAttributeType.Set)
	assert.False(t, eAttributeType.BasicSet)
	assert.True(t, eAttributeType.IsProxy)
	assert.True(t, eAttributeType.BasicGet)
}

// TestFeatureBasicUnset checks that a changeable unsettable single bidirectional
// reference has a basic unset, and that a shared aggregation is a back reference.
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

// TestFeatureEcoreAccessors checks the accessors of real Ecore features.
func TestFeatureEcoreAccessors(t *testing.T) {
	m := ecoreMetamodel(t)
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
