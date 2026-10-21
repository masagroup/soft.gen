package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageClassifiers(t *testing.T) {
	m := loadModel(t, "library")
	info := m.Package(m.Root())
	assert.Same(t, m.Root(), info.Package)
	assert.Len(t, info.Classes, 16)
	assert.Equal(t, []string{"BookCategory"}, names(info.Enums))
	assert.Empty(t, info.DataTypes)
	assert.Empty(t, info.SubPackages)
	assert.Equal(t, names(info.Classes), names(info.AllClasses))
	assert.Equal(t, names(info.Enums), names(info.AllEnums))
	assert.Empty(t, info.AllDataTypes)

	me := loadModel(t, "ecore")
	infoEcore := me.Package(me.Root())
	assert.Len(t, infoEcore.Classes, 20)
	assert.Len(t, infoEcore.Enums, 0)
	assert.Len(t, infoEcore.DataTypes, 33)
	assert.Equal(t, 53, len(infoEcore.Classes)+len(infoEcore.DataTypes))
	assert.Equal(t, "EBigDecimal", infoEcore.DataTypes[0].GetName())
	assert.Equal(t, names(infoEcore.DataTypes), names(infoEcore.AllDataTypes))
}

func TestPackageSubPackages(t *testing.T) {
	p := ecore.GetFactory().CreateEPackage()
	p.SetName("p")
	sub := ecore.GetFactory().CreateEPackage()
	sub.SetName("sub")
	p.GetESubPackages().Add(sub)
	a := ecore.GetFactory().CreateEClassFromContainer(p)
	a.SetName("A")
	b := ecore.GetFactory().CreateEClassFromContainer(sub)
	b.SetName("B")
	e := ecore.GetFactory().CreateEEnumFromContainer(sub)
	e.SetName("E")
	annotation := ecore.GetFactory().CreateEAnnotationFromContainer(b)
	annotation.SetSource("http://example.com/sub/")
	m := FromPackage(p, genGoSource)
	info := m.Package(p)
	assert.Equal(t, []ecore.EPackage{sub}, info.SubPackages)
	assert.Equal(t, []ecore.EClass{a}, info.Classes)
	assert.Equal(t, []ecore.EClass{a, b}, info.AllClasses)
	assert.Empty(t, info.Enums)
	assert.Equal(t, []ecore.EEnum{e}, info.AllEnums)
	assert.Equal(t, []ecore.EClass{a}, info.OrderedClasses)
	assert.Equal(t, []AnnotationSource{{URI: "http://example.com/sub/", Identifier: "sub"}}, info.AnnotationSources)
	assert.Equal(t, []ecore.EClass{b}, m.Package(sub).Classes)
}

func TestOrderedClasses(t *testing.T) {
	m := loadModel(t, "library")
	info := m.Package(m.Root())
	ordered := info.OrderedClasses
	assert.Equal(t, []string{
		"Item", "CirculatingItem", "Book", "BookIndex", "Addressable", "Library", "Person", "Writer",
		"Lendable", "Periodical", "AudioVisualItem", "BookOnTape", "VideoCassette", "Borrower", "Employee", "DocumentRoot",
	}, names(ordered))
	assert.ElementsMatch(t, names(info.Classes), names(ordered))
	// the first super type chain precedes a class (a mixin may not: Lendable
	// comes after CirculatingItem, as in v1)
	position := map[ecore.EClass]int{}
	for i, c := range ordered {
		position[c] = i
	}
	for _, c := range ordered {
		if base := firstSuperType(c); base != nil {
			assert.Less(t, position[base], position[c], "%s before %s", base.GetName(), c.GetName())
		}
	}
	// stable
	assert.Equal(t, names(ordered), names(orderedClasses(m.Root())))

	me := loadModel(t, "ecore")
	orderedEcore := me.Package(me.Root()).OrderedClasses
	assert.Equal(t, []string{"EModelElement", "ENamedElement", "ETypedElement", "EStructuralFeature", "EAttribute", "EAnnotation", "EClassifier", "EClass"}, names(orderedEcore[:8]))
	assert.Len(t, orderedEcore, 20)
}

func TestOrderedClassesForeignSuperType(t *testing.T) {
	// a super type of another package is skipped
	p := ecore.GetFactory().CreateEPackage()
	p.SetName("p")
	other := ecore.GetFactory().CreateEPackage()
	other.SetName("other")
	base := ecore.GetFactory().CreateEClassFromContainer(other)
	base.SetName("Base")
	mid := ecore.GetFactory().CreateEClassFromContainer(p)
	mid.SetName("Mid")
	mid.GetESuperTypes().Add(base)
	leaf := ecore.GetFactory().CreateEClassFromContainer(p)
	leaf.SetName("Leaf")
	leaf.GetESuperTypes().Add(mid)
	// Leaf is declared first: Mid is placed before it anyway
	p.GetEClassifiers().Move(1, 0)
	require.Equal(t, "Leaf", p.GetEClassifiers().Get(0).(ecore.EClass).GetName())
	assert.Equal(t, []ecore.EClass{mid, leaf}, orderedClasses(p))
}

func TestAnnotationSources(t *testing.T) {
	m := loadModel(t, "library")
	assert.Equal(t, []AnnotationSource{
		{URI: "http:///org/eclipse/emf/ecore/util/ExtendedMetaData", Identifier: "ExtendedMetaData"},
		{URI: genGoSource, Identifier: "GenGo"},
	}, m.Package(m.Root()).AnnotationSources)

	ts := FromPackage(m.Root(), genTSSource)
	assert.Equal(t, []AnnotationSource{
		{URI: "http:///org/eclipse/emf/ecore/util/ExtendedMetaData", Identifier: "ExtendedMetaData"},
		{URI: genTSSource, Identifier: "GenTS"},
	}, ts.Package(m.Root()).AnnotationSources)

	// without a generator annotation nothing is suppressed
	all := FromPackage(m.Root(), "http://example.com/none")
	assert.Equal(t, []string{
		"http:///org/eclipse/emf/ecore/util/ExtendedMetaData",
		"http://net.masagroup/soft/2018/GenCpp",
		genGoSource,
		genTSSource,
	}, func() []string {
		var uris []string
		for _, s := range all.Package(m.Root()).AnnotationSources {
			uris = append(uris, s.URI)
		}
		return uris
	}())

	me := loadModel(t, "ecore")
	var identifiers []string
	for _, s := range me.Package(me.Root()).AnnotationSources {
		identifiers = append(identifiers, s.Identifier)
	}
	// sorted by uri: http:///org... < http://net... < http://www...
	assert.Equal(t, []string{"ExtendedMetaData", "GenCpp", "GenGo", "GenTS", "Ecore"}, identifiers)
}

func TestAnnotationSourceIdentifier(t *testing.T) {
	assert.Equal(t, "GenGo", annotationSourceIdentifier("http://net.masagroup/soft/2019/GenGo"))
	assert.Equal(t, "GenGo", annotationSourceIdentifier("http://net.masagroup/soft/2019/GenGo/"))
	assert.Equal(t, "Ecore", annotationSourceIdentifier("http://www.eclipse.org/emf/2002/Ecore"))
	assert.Equal(t, "plain", annotationSourceIdentifier("plain"))
	assert.Equal(t, "plain", annotationSourceIdentifier("plain/"))
	assert.Equal(t, "x", annotationSourceIdentifier("http://x/"))
	assert.Equal(t, "", annotationSourceIdentifier(""))
	assert.Equal(t, "", annotationSourceIdentifier("/"))
}
