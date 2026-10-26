package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageClassifiers checks the split of the classifiers of a package into
// classes, enums and data types, in model order, and that the All* lists of a
// package without sub package are the same.
func TestPackageClassifiers(t *testing.T) {
	tests := []struct {
		model          string
		classes, enums []string
		dataTypes      []string
	}{
		{
			model:   "shop",
			classes: []string{"Named", "Tagged", "Shop", "Product", "Customer", "Order"},
			enums:   []string{"Category"},
		},
		{
			model:   "enums",
			classes: []string{"Lamp"},
			enums:   []string{"Color", "Light", "Empty", "Single", "Status", "Level"},
		},
		{
			model:   "attributes",
			classes: []string{"Holder"},
			dataTypes: []string{
				"StringType", "BoolType", "IntType", "LongType", "Int32Type", "Int8Type", "ShortType", "ByteType",
				"BytesType", "DoubleType", "FloatType", "ClassType", "DateType", "ObjectType", "UnknownType",
				"Celsius", "Point", "EChar", "Octet", "Opaque",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.model, func(t *testing.T) {
			m := loadModel(t, test.model)
			info := m.Package(m.Root())
			assert.Same(t, m.Root(), info.Package)
			assert.Equal(t, orEmpty(test.classes), names(info.Classes))
			assert.Equal(t, orEmpty(test.enums), names(info.Enums))
			assert.Equal(t, orEmpty(test.dataTypes), names(info.DataTypes))
			assert.Empty(t, info.SubPackages)
			assert.Equal(t, names(info.Classes), names(info.AllClasses))
			assert.Equal(t, names(info.Enums), names(info.AllEnums))
			assert.Equal(t, names(info.DataTypes), names(info.AllDataTypes))
		})
	}

	// the real ecore package
	me := ecoreMetamodel(t)
	infoEcore := me.Package(me.Root())
	assert.Len(t, infoEcore.Classes, 20)
	assert.Len(t, infoEcore.Enums, 0)
	assert.Len(t, infoEcore.DataTypes, 33)
	assert.Equal(t, 53, len(infoEcore.Classes)+len(infoEcore.DataTypes))
	assert.Equal(t, "EBigDecimal", infoEcore.DataTypes[0].GetName())
	assert.Equal(t, names(infoEcore.DataTypes), names(infoEcore.AllDataTypes))
}

// TestPackageSubPackages checks the sub packages and the All* lists over them.
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

// TestOrderedClasses checks that the first super type chain of a class
// precedes it, the classes keeping the model order otherwise, while a mixin
// may follow the class (as in v1).
func TestOrderedClasses(t *testing.T) {
	// operations.ecore with Circle (Shape, Measurable) moved first: Shape is
	// placed before it, Measurable, a mixin, stays after it
	m := loadModel(t, "operations")
	m.Root().GetEClassifiers().Move(5, 0)
	require.Equal(t, "Circle", m.Root().GetEClassifiers().Get(0).(ecore.EClass).GetName())
	info := m.Package(m.Root())
	ordered := info.OrderedClasses
	assert.Equal(t, []string{"Shape", "Circle", "Calculator", "Greeter", "Polite", "Measurable"}, names(ordered))
	assert.ElementsMatch(t, names(info.Classes), names(ordered))
	// stable
	assert.Equal(t, names(ordered), names(orderedClasses(m.Root())))

	// the classes already ordered keep the model order
	for _, name := range []string{"inheritance", "shop", "metamodel"} {
		mf := loadModel(t, name)
		assert.Equal(t, names(mf.Package(mf.Root()).Classes), names(mf.Package(mf.Root()).OrderedClasses), name)
	}

	// the real ecore package: EAttribute is declared first
	me := ecoreMetamodel(t)
	orderedEcore := me.Package(me.Root()).OrderedClasses
	assert.Equal(t, []string{"EModelElement", "ENamedElement", "ETypedElement", "EStructuralFeature", "EAttribute", "EAnnotation", "EClassifier", "EClass"}, names(orderedEcore[:8]))
	assert.Len(t, orderedEcore, 20)
	position := map[ecore.EClass]int{}
	for i, c := range orderedEcore {
		position[c] = i
	}
	for _, c := range orderedEcore {
		if base := firstSuperType(c); base != nil {
			assert.Less(t, position[base], position[c], "%s before %s", base.GetName(), c.GetName())
		}
	}
}

// TestOrderedClassesForeignSuperType checks that a super type of another package is skipped.
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

// TestNestedAnnotations checks the annotations nested in an annotation, in
// EMF order: its eAnnotations, then the annotations of its contents, each with
// its own; the contents that are not annotations are left out.
func TestNestedAnnotations(t *testing.T) {
	m := loadModel(t, "annotations")
	doc := classNamed(t, m, "Document").GetEAnnotation("http://soft.gen/test/doc")
	require.NotNil(t, doc)
	require.Equal(t, 3, doc.GetContents().Size()+doc.GetEAnnotations().Size())

	nested := NestedAnnotations(doc)
	require.Len(t, nested, 2)
	assert.Equal(t, "http://soft.gen/test/nested", nested[0].Annotation.GetSource())
	assert.False(t, nested[0].Content)
	assert.Empty(t, nested[0].Nested)
	assert.Equal(t, "http://soft.gen/test/content", nested[1].Annotation.GetSource())
	assert.True(t, nested[1].Content)
	require.Len(t, nested[1].Nested, 1)
	assert.Equal(t, "http://soft.gen/test/deep", nested[1].Nested[0].Annotation.GetSource())
	assert.True(t, nested[1].Nested[0].Content)

	for _, a := range m.Package(m.Root()).Annotations {
		assert.NotNil(t, a.GetEModelElement(), "the nested annotation %s is listed", a.GetSource())
	}
}

// TestAnnotationSources checks the annotation sources of a package: the
// distinct sources of the annotations of its contents, sorted, minus those the
// supressAnnotations key of the generator annotation of the model source
// lists. As in EMF, the sources of the annotations nested in an annotation
// (nested, content, deep) are not listed: they are created with their parent.
func TestAnnotationSources(t *testing.T) {
	m := loadModel(t, "annotations")
	// GenGo suppresses http://soft.gen/test/hidden
	assert.Equal(t, []AnnotationSource{
		{URI: genGoSource, Identifier: "GenGo"},
		{URI: "http://soft.gen/test/doc", Identifier: "doc"},
	}, m.Package(m.Root()).AnnotationSources)

	// the suppression is read from the annotation of the model source: the
	// GenGo key does not apply to a GenTS model
	ts := FromPackage(m.Root(), genTSSource)
	assert.Equal(t, []AnnotationSource{
		{URI: genGoSource, Identifier: "GenGo"},
		{URI: "http://soft.gen/test/doc", Identifier: "doc"},
		{URI: "http://soft.gen/test/hidden", Identifier: "hidden"},
	}, ts.Package(m.Root()).AnnotationSources)

	// the real ecore package: sorted by uri, http:///org... < http://net... < http://www...
	me := ecoreMetamodel(t)
	var identifiers []string
	for _, s := range me.Package(me.Root()).AnnotationSources {
		identifiers = append(identifiers, s.Identifier)
	}
	assert.Equal(t, []string{"ExtendedMetaData", "GenCpp", "GenGo", "GenTS", "Ecore"}, identifiers)
}

// TestAnnotationSourceIdentifier checks the identifier of a source: its last path segment.
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
