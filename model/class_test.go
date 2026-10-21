package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func names[T ecore.ENamedElement](elements []T) []string {
	result := make([]string, len(elements))
	for i, e := range elements {
		result[i] = e.GetName()
	}
	return result
}

// TestClassHierarchyEcore checks the base, extended and mixin classes, the
// implemented features and operations, the containers and the list/map
// requirements of the real Ecore classes.
func TestClassHierarchyEcore(t *testing.T) {
	m := ecoreMetamodel(t)
	eClass := m.Class(classNamed(t, m, "EClass"))
	assert.Same(t, classNamed(t, m, "EClassifier"), eClass.BaseClass)
	assert.Same(t, classNamed(t, m, "EClassifier"), eClass.ExtendsClass)
	assert.Empty(t, eClass.Mixins)
	assert.False(t, eClass.HasElementID)
	// the implemented features and operations are the own ones
	assert.Equal(t, names(elements[ecore.EStructuralFeature](eClass.Class.GetEStructuralFeatures())), names(eClass.ImplementedFeatures))
	assert.Equal(t, names(elements[ecore.EOperation](eClass.Class.GetEOperations())), names(eClass.ImplementedOperations))
	assert.True(t, eClass.RequiresList)
	assert.False(t, eClass.RequiresMap)

	eOperation := m.Class(classNamed(t, m, "EOperation"))
	assert.Same(t, classNamed(t, m, "ETypedElement"), eOperation.BaseClass)
	assert.Same(t, classNamed(t, m, "ETypedElement"), eOperation.ExtendsClass)
	assert.Empty(t, eOperation.Mixins)
	assert.Equal(t, []string{"eContainingClass", "eParameters", "eExceptions", "operationID"}, names(eOperation.ImplementedFeatures))
	assert.Equal(t, []string{"isOverrideOf"}, names(eOperation.ImplementedOperations))
	assert.Equal(t, []string{"eContainingClass"}, names(eOperation.AllContainers))

	eAttribute := m.Class(classNamed(t, m, "EAttribute"))
	assert.Same(t, classNamed(t, m, "EStructuralFeature"), eAttribute.BaseClass)
	assert.Same(t, classNamed(t, m, "EStructuralFeature"), eAttribute.ExtendsClass)
	assert.Empty(t, eAttribute.Mixins)
	assert.Equal(t, []string{"iD", "eAttributeType"}, names(eAttribute.ImplementedFeatures))
	assert.Empty(t, eAttribute.ImplementedOperations)
	assert.Equal(t, []string{"eContainingClass"}, names(eAttribute.AllContainers))
	assert.False(t, eAttribute.RequiresList)

	eModelElement := m.Class(classNamed(t, m, "EModelElement"))
	assert.Nil(t, eModelElement.BaseClass)
	assert.Nil(t, eModelElement.ExtendsClass)
	assert.Equal(t, []string{"eAnnotations"}, names(eModelElement.ImplementedFeatures))
	assert.Equal(t, []string{"getEAnnotation"}, names(eModelElement.ImplementedOperations))
	assert.Empty(t, eModelElement.AllContainers)

	eObject := m.Class(classNamed(t, m, "EObject"))
	assert.True(t, eObject.RequiresList) // eContents() EEList
	assert.False(t, eObject.RequiresMap)
	assert.False(t, m.Class(classNamed(t, m, "ENamedElement")).RequiresList)
	// EStringToStringMapEntry is a map entry
	assert.True(t, m.Class(classNamed(t, m, "EStringToStringMapEntry")).IsMapEntry)
}

// TestClassHierarchyFixtures checks, on the fixture classes, the rules of the
// class facts: BaseClass is the first super type; ExtendsClass the first
// non-interface class along the first super type chain (nil for an interface
// base without base); Mixins the mixins of an interface base then the all
// super types after the base; HasElementID whether a mixin implements a
// feature or an operation; the implemented features and operations those not
// inherited from ExtendsClass; AllContainers the own then inherited container
// references; RequiresList/RequiresMap whether an implemented feature or
// operation is a list/map or returns an EEList/EMap.
func TestClassHierarchyFixtures(t *testing.T) {
	tests := []struct {
		name                      string
		model, class              string
		base, extends             string // "" for nil
		mixins                    []string
		hasElementID              bool
		features, operations      []string
		containers                []string
		requiresList, requiresMap bool
	}{
		{
			name: "concrete base extended, interface mixin", model: "operations", class: "Circle",
			base: "Shape", extends: "Shape", mixins: []string{"Measurable"}, hasElementID: true,
			features: []string{"radius"}, operations: []string{"scale", "area"},
		},
		{
			name: "diamond: the second branch is a mixin", model: "inheritance", class: "Bottom",
			base: "Left", extends: "Left", mixins: []string{"Right"}, hasElementID: true,
			features: []string{"right", "partner", "bottom"},
		},
		{
			name: "single concrete base: no mixin", model: "inheritance", class: "Left",
			base: "Base", extends: "Base", features: []string{"left"},
		},
		{
			name: "interface base without base is not extended", model: "inheritance", class: "Base",
			base: "Identified", features: []string{"id", "base"},
		},
		{
			name: "interface bases: the second one is a mixin", model: "inheritance", class: "Item",
			base: "Identified", mixins: []string{"Tagged"}, hasElementID: true,
			features: []string{"id", "tags", "weight"}, requiresList: true,
		},
		{
			name: "interface without super type", model: "inheritance", class: "Identified",
			features: []string{"id"},
		},
		{
			name: "interface base and abstract class mixin, own container", model: "shop", class: "Product",
			base: "Named", mixins: []string{"Tagged"}, hasElementID: true,
			features: []string{"name", "label", "price", "category", "shop"}, operations: []string{"describe", "clearLabel"},
			containers: []string{"shop"},
		},
		{
			name: "inherited container", model: "metamodel", class: "Leaf",
			base: "Node", extends: "Node", features: []string{"weight", "next"}, operations: []string{"touch"},
			containers: []string{"parent"},
		},
		{
			name: "own container", model: "references", class: "Child",
			features: []string{"parent"}, containers: []string{"parent"},
		},
		{
			name: "map features", model: "maps", class: "Holder",
			features: []string{"values", "owned", "pseudos"}, requiresList: true, requiresMap: true,
		},
		{
			name: "map entry with a many key", model: "maps", class: "NamesToAnyEntry",
			features: []string{"key", "value"}, requiresList: true,
		},
		{
			name: "list and map required by operations", model: "operations", class: "Calculator",
			operations:   []string{"clear", "compute", "custom", "filter", "names", "process", "reset", "sum", "table", "peer"},
			requiresList: true, requiresMap: true,
		},
		{
			name: "operations without list", model: "operations", class: "Greeter",
			operations: []string{"greet", "wave"},
		},
		{
			name: "everything inherited from the extended class", model: "operations", class: "Polite",
			base: "Greeter", extends: "Greeter",
		},
		{
			name: "interface operations", model: "operations", class: "Measurable",
			operations: []string{"scale", "area"},
		},
	}
	nameOf := func(c ecore.EClass) string {
		if c == nil {
			return ""
		}
		return c.GetName()
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := loadModel(t, test.model)
			info := m.Class(classNamed(t, m, test.class))
			assert.Equal(t, test.base, nameOf(info.BaseClass), "BaseClass")
			assert.Equal(t, test.extends, nameOf(info.ExtendsClass), "ExtendsClass")
			assert.Equal(t, orEmpty(test.mixins), names(info.Mixins), "Mixins")
			assert.Equal(t, test.hasElementID, info.HasElementID, "HasElementID")
			assert.Equal(t, orEmpty(test.features), names(info.ImplementedFeatures), "ImplementedFeatures")
			assert.Equal(t, orEmpty(test.operations), names(info.ImplementedOperations), "ImplementedOperations")
			assert.Equal(t, orEmpty(test.containers), names(info.AllContainers), "AllContainers")
			assert.Equal(t, test.requiresList, info.RequiresList, "RequiresList")
			assert.Equal(t, test.requiresMap, info.RequiresMap, "RequiresMap")
		})
	}
}

// orEmpty returns names, an empty slice for nil (names returns non-nil slices).
func orEmpty(names []string) []string {
	if names == nil {
		return []string{}
	}
	return names
}

// TestClassReflectiveSets checks the reflective feature sets: eGet/eIsSet are
// the implemented features, eSet/eUnset the changeable ones, eInverseAdd the
// bidirectional non-volatile ones, eInverseRemove the containments and the
// bidirectional non-volatile ones, eInvoke the implemented operations.
func TestClassReflectiveSets(t *testing.T) {
	tests := []struct {
		name                        string
		model, class                string
		eGet, eSet                  []string
		eInverseAdd, eInverseRemove []string
		eInvoke                     []string
	}{
		{
			name: "containments and references", model: "references", class: "Root",
			eGet:           []string{"children", "items", "fixed", "optional", "favorite", "direct", "shared", "links", "locals", "marks"},
			eSet:           []string{"children", "items", "fixed", "optional", "favorite", "direct", "shared", "links", "locals", "marks"},
			eInverseAdd:    []string{"children", "fixed"},
			eInverseRemove: []string{"children", "items", "fixed", "optional"},
		},
		{
			name: "non changeable container", model: "references", class: "Fixed",
			eGet: []string{"owner"}, eInverseAdd: []string{"owner"}, eInverseRemove: []string{"owner"},
		},
		{
			name: "bidirectional references", model: "references", class: "Husband",
			eGet: []string{"wife", "friends"}, eSet: []string{"wife", "friends"},
			eInverseAdd: []string{"wife", "friends"}, eInverseRemove: []string{"wife", "friends"},
		},
		{
			name: "volatile containments are inverse removed only", model: "volatile", class: "Holder",
			eGet:           []string{"label", "code", "linked", "target", "owned", "kept"},
			eSet:           []string{"label", "code", "linked", "target", "owned", "kept"},
			eInverseRemove: []string{"owned", "kept"},
		},
		{
			name: "features and operations of the mixins", model: "shop", class: "Product",
			eGet:        []string{"name", "label", "price", "category", "shop"},
			eSet:        []string{"name", "label", "price", "category", "shop"},
			eInverseAdd: []string{"shop"}, eInverseRemove: []string{"shop"},
			eInvoke: []string{"describe", "clearLabel"},
		},
		{
			name: "members of the extended class excluded", model: "operations", class: "Circle",
			eGet: []string{"radius"}, eSet: []string{"radius"}, eInvoke: []string{"scale", "area"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := loadModel(t, test.model)
			info := m.Class(classNamed(t, m, test.class))
			assert.Equal(t, orEmpty(test.eGet), names(info.EGetFeatures), "EGetFeatures")
			assert.Equal(t, names(info.ImplementedFeatures), names(info.EGetFeatures))
			assert.Equal(t, names(info.EGetFeatures), names(info.EIsSetFeatures))
			assert.Equal(t, orEmpty(test.eSet), names(info.ESetFeatures), "ESetFeatures")
			assert.Equal(t, names(info.ESetFeatures), names(info.EUnsetFeatures))
			assert.Equal(t, orEmpty(test.eInverseAdd), names(info.EInverseAddFeatures), "EInverseAddFeatures")
			assert.Equal(t, orEmpty(test.eInverseRemove), names(info.EInverseRemoveFeatures), "EInverseRemoveFeatures")
			assert.Equal(t, orEmpty(test.eInvoke), names(info.EInvokeOperations), "EInvokeOperations")
		})
	}

	// the real Ecore classes
	me := ecoreMetamodel(t)
	eClass := me.Class(classNamed(t, me, "EClass"))
	// eAttributes, eReferences and the eAll* features are not changeable
	assert.Equal(t, []string{"abstract", "interface", "eStructuralFeatures", "eSuperTypes", "eOperations"}, names(eClass.ESetFeatures))
	assert.Equal(t, []string{"eStructuralFeatures", "eOperations"}, names(eClass.EInverseAddFeatures))
	assert.Equal(t, []string{"eStructuralFeatures", "eOperations"}, names(eClass.EInverseRemoveFeatures))
	assert.Equal(t, names(eClass.ImplementedOperations), names(eClass.EInvokeOperations))
	eAttribute := me.Class(classNamed(t, me, "EAttribute"))
	require.Len(t, eAttribute.ESetFeatures, 1)
	assert.Equal(t, "iD", eAttribute.ESetFeatures[0].GetName())
}

func TestClassMixinsOfInterfaceBase(t *testing.T) {
	// C extends I (interface) extends J (interface), and mixes K: the mixins
	// of C are those of I then the all super types after I. As in v1, I has
	// no mixin (J is its base class, and nothing follows J in its all super
	// types) and J, which precedes I in the all super types of C, is not a
	// mixin of C either: the mixins of C are [K].
	p := ecore.GetFactory().CreateEPackage()
	p.SetName("p")
	newClass := func(name string, isInterface bool, superTypes ...ecore.EClass) ecore.EClass {
		c := ecore.GetFactory().CreateEClassFromContainer(p)
		c.SetName(name)
		c.SetInterface(isInterface)
		c.SetAbstract(isInterface)
		for _, s := range superTypes {
			c.GetESuperTypes().Add(s)
		}
		return c
	}
	j := newClass("J", true)
	i := newClass("I", true, j)
	k := newClass("K", true)
	c := newClass("C", false, i, k)
	f := ecore.GetFactory().CreateEAttributeFromContainer(k)
	f.SetName("f")
	m := FromPackage(p, genGoSource)
	info := m.Class(c)
	assert.Same(t, i, info.BaseClass)
	assert.Nil(t, info.ExtendsClass)
	assert.Equal(t, []ecore.EClass{k}, info.Mixins)
	assert.True(t, info.HasElementID)
	assert.Empty(t, m.Class(i).Mixins)
	assert.Empty(t, m.Class(j).Mixins)
	assert.Equal(t, []ecore.EClass{j, i, k}, elements[ecore.EClass](c.GetEAllSuperTypes()))

	// a concrete class along the chain is the extended one
	b := newClass("B", false)
	ib := newClass("IB", true, b)
	d := newClass("D", false, ib)
	assert.Same(t, ib, m.Class(d).BaseClass)
	assert.Same(t, b, m.Class(d).ExtendsClass)
	assert.Empty(t, m.Class(d).Mixins) // nothing follows IB in [B, IB]

	// a second interface after a concrete base class is a mixin, with its own mixins first
	l := newClass("L", true, k)
	e := newClass("E", false, b, l)
	assert.Same(t, b, m.Class(e).ExtendsClass)
	assert.Equal(t, []ecore.EClass{k, l}, m.Class(e).Mixins)
	assert.True(t, m.Class(e).HasElementID) // K has a feature
}
