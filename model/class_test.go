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

func TestClassHierarchyEcore(t *testing.T) {
	m := loadModel(t, "ecore")
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

func TestClassHierarchyLibrary(t *testing.T) {
	m := loadModel(t, "library")
	item := classNamed(t, m, "Item")
	lendable := classNamed(t, m, "Lendable")
	addressable := classNamed(t, m, "Addressable")

	circulating := m.Class(classNamed(t, m, "CirculatingItem"))
	assert.Same(t, item, circulating.BaseClass)
	assert.Same(t, item, circulating.ExtendsClass)
	assert.Equal(t, []ecore.EClass{lendable}, circulating.Mixins)
	assert.True(t, circulating.HasElementID)
	assert.Equal(t, []string{"copies", "borrowers"}, names(circulating.ImplementedFeatures))

	book := m.Class(classNamed(t, m, "Book"))
	assert.Same(t, circulating.Class, book.BaseClass)
	assert.Same(t, circulating.Class, book.ExtendsClass)
	assert.Empty(t, book.Mixins)
	assert.False(t, book.HasElementID)
	assert.Equal(t, []string{"title", "pages", "category", "author", "tableOfContents", "indexes"}, names(book.ImplementedFeatures))
	assert.Empty(t, book.ImplementedOperations)
	assert.True(t, book.RequiresList)
	assert.True(t, book.RequiresMap)
	assert.Empty(t, book.AllContainers)

	// an interface base class without base is not extended
	library := m.Class(classNamed(t, m, "Library"))
	assert.Same(t, addressable, library.BaseClass)
	assert.Nil(t, library.ExtendsClass)
	assert.Empty(t, library.Mixins)
	assert.False(t, library.HasElementID)
	assert.Equal(t, []string{"address", "name", "writers", "employees", "borrowers", "stock", "books", "branches", "parentBranch", "people", "proprietary"}, names(library.ImplementedFeatures))
	assert.Equal(t, []string{"parentBranch"}, names(library.AllContainers))
	assert.True(t, library.RequiresList)
	assert.False(t, library.RequiresMap)

	person := m.Class(classNamed(t, m, "Person"))
	assert.Same(t, addressable, person.BaseClass)
	assert.Nil(t, person.ExtendsClass)
	assert.False(t, person.RequiresList)
	writer := m.Class(classNamed(t, m, "Writer"))
	assert.Same(t, person.Class, writer.BaseClass)
	assert.Same(t, person.Class, writer.ExtendsClass)
	assert.Equal(t, []string{"name", "books"}, names(writer.ImplementedFeatures))

	lendableInfo := m.Class(lendable)
	assert.Nil(t, lendableInfo.BaseClass)
	assert.Nil(t, lendableInfo.ExtendsClass)
	assert.Empty(t, lendableInfo.Mixins)
	assert.Equal(t, []string{"copies", "borrowers"}, names(lendableInfo.ImplementedFeatures))

	assert.False(t, m.Class(classNamed(t, m, "Employee")).RequiresList)
	assert.True(t, m.Class(classNamed(t, m, "DocumentRoot")).RequiresMap)
	assert.True(t, m.Class(classNamed(t, m, "BookIndex")).RequiresList)
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

func TestClassReflectiveSets(t *testing.T) {
	m := loadModel(t, "library")
	book := m.Class(classNamed(t, m, "Book"))
	assert.Equal(t, names(book.ImplementedFeatures), names(book.EGetFeatures))
	assert.Equal(t, names(book.ImplementedFeatures), names(book.ESetFeatures))
	assert.Equal(t, names(book.EGetFeatures), names(book.EIsSetFeatures))
	assert.Equal(t, names(book.ESetFeatures), names(book.EUnsetFeatures))
	assert.Equal(t, []string{"author"}, names(book.EInverseAddFeatures))
	assert.Equal(t, []string{"author"}, names(book.EInverseRemoveFeatures))
	assert.Empty(t, book.EInvokeOperations)

	library := m.Class(classNamed(t, m, "Library"))
	assert.Equal(t, []string{"branches", "parentBranch"}, names(library.EInverseAddFeatures))
	assert.Equal(t, []string{"writers", "employees", "borrowers", "stock", "books", "branches", "parentBranch", "proprietary"}, names(library.EInverseRemoveFeatures))

	me := loadModel(t, "ecore")
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
