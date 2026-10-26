package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	genGoSource = "http://net.masagroup/soft/2019/GenGo"
	genTSSource = "http://net.masagroup/soft/2020/GenTS"
)

// loadModel loads the fixture model testdata/models/<name>.ecore.
func loadModel(t *testing.T, name string) *Model {
	t.Helper()
	m, err := Load("../../testdata/models/"+name+".ecore", genGoSource)
	require.NoError(t, err)
	require.NotNil(t, m)
	return m
}

// ecoreMetamodel returns the model of the embedded ecore metamodel (Model.Ecore()):
// the real Ecore classes with the generator annotations of ecore.ecore.
func ecoreMetamodel(t *testing.T) *Model {
	t.Helper()
	p := FromPackage(nil, genGoSource).Ecore()
	require.NotNil(t, p)
	m := FromPackage(p, genGoSource)
	m.ecore = p
	return m
}

// loadDocument writes the .ecore document to a temporary file and loads it:
// for the loading facts no fixture model has (an empty package, a reference
// to an ecore class).
func loadDocument(t *testing.T, document string) *Model {
	t.Helper()
	path := filepath.Join(t.TempDir(), "document.ecore")
	require.NoError(t, os.WriteFile(path, []byte(document), 0o600))
	m, err := Load(path, genGoSource)
	require.NoError(t, err)
	return m
}

func classNamed(t *testing.T, m *Model, name string) ecore.EClass {
	t.Helper()
	c, _ := m.Root().GetEClassifier(name).(ecore.EClass)
	require.NotNil(t, c, "class %s", name)
	return c
}

func featureNamed(t *testing.T, m *Model, className, featureName string) ecore.EStructuralFeature {
	t.Helper()
	f := classNamed(t, m, className).GetEStructuralFeatureFromName(featureName)
	require.NotNil(t, f, "feature %s.%s", className, featureName)
	return f
}

// TestLoadCrossReferences checks that the references of a loaded model to the
// ecore package resolve to the embedded ecore.ecore (its Java instance type
// names, its generator annotations), not to the runtime package of soft.go.
func TestLoadCrossReferences(t *testing.T) {
	m := loadModel(t, "attributes")
	assert.Equal(t, "attributes", m.Root().GetName())
	assert.Equal(t, genGoSource, m.AnnotationSource())
	assert.False(t, m.IsEcore(m.Root()))

	ecorePackage := m.Ecore()
	require.NotNil(t, ecorePackage)
	assert.True(t, m.IsEcore(ecorePackage))
	assert.Equal(t, EcoreNsURI, ecorePackage.GetNsURI())
	eString := ecorePackage.GetEClassifier("EString")
	require.NotNil(t, eString)
	assert.Equal(t, "java.lang.String", eString.GetInstanceTypeName())
	assert.NotSame(t, ecore.GetPackage().GetEString(), eString)

	label := featureNamed(t, m, "Holder", "label")
	assert.Same(t, eString, label.GetEType())
	assert.True(t, m.IsEcoreClassifier(label.GetEType()))
	assert.False(t, m.IsEcoreClassifier(classNamed(t, m, "Holder")))
	// a data type of the model is not an ecore classifier
	assert.False(t, m.IsEcoreClassifier(featureNamed(t, m, "Holder", "temperature").GetEType()))
	assert.True(t, m.Annotation(ecorePackage).HasKey("lazy"))
	assert.Equal(t, "list:function", m.Annotation(ecorePackage).Value("lazy"))
}

// TestLoadEcoreClassReference checks that a class reference to the ecore
// package resolves to a real EClass of the embedded package, with the
// instance type name of ecore.ecore (no fixture model references an ecore
// class).
func TestLoadEcoreClassReference(t *testing.T) {
	m := loadDocument(t, `<?xml version="1.0" encoding="UTF-8"?>
<ecore:EPackage xmi:version="2.0" xmlns:xmi="http://www.omg.org/XMI" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:ecore="http://www.eclipse.org/emf/2002/Ecore" name="refs" nsURI="http://soft.gen/test/refs" nsPrefix="refs">
  <eClassifiers xsi:type="ecore:EClass" name="Root">
    <eStructuralFeatures xsi:type="ecore:EReference" name="entries" upperBound="-1"
        eType="ecore:EClass http://www.eclipse.org/emf/2002/Ecore#//EStringToStringMapEntry" containment="true"/>
  </eClassifiers>
</ecore:EPackage>
`)
	entries := featureNamed(t, m, "Root", "entries")
	mapEntry, isClass := entries.GetEType().(ecore.EClass)
	require.True(t, isClass)
	assert.Same(t, m.Ecore().GetEClassifier("EStringToStringMapEntry"), mapEntry)
	assert.Equal(t, "java.util.Map$Entry", mapEntry.GetInstanceTypeName())
	assert.True(t, m.Class(mapEntry).IsMapEntry)
	assert.True(t, m.Feature(entries).IsMap)
	assert.Equal(t, "map", m.Feature(entries).Kind)
}

// TestLoadEcoreNamedPackage checks that a loaded package named "ecore" is an
// ecore package distinct from the embedded one: its local references stay
// local, its references to the real ecore resolve to the embedded one.
func TestLoadEcoreNamedPackage(t *testing.T) {
	m := loadModel(t, "metamodel")
	root := m.Root()
	assert.Equal(t, "ecore", root.GetName())
	assert.True(t, m.IsEcore(root))
	assert.NotSame(t, m.Ecore(), root)
	assert.Equal(t, 3, root.GetEClassifiers().Size())
	// Tree.nodes types to the local Node
	assert.Same(t, classNamed(t, m, "Node"), featureNamed(t, m, "Tree", "nodes").GetEType())
	// Node.label types to the embedded EString
	assert.Same(t, m.Ecore().GetEClassifier("EString"), featureNamed(t, m, "Node", "label").GetEType())
	// the classes of a package named ecore are ecore classifiers
	assert.True(t, m.IsEcoreClassifier(classNamed(t, m, "Node")))
}

// TestLoadEmpty checks that a package without classifier has no class, no
// ordered class and no annotation source.
func TestLoadEmpty(t *testing.T) {
	m := loadDocument(t, `<?xml version="1.0" encoding="UTF-8"?>
<ecore:EPackage xmi:version="2.0" xmlns:xmi="http://www.omg.org/XMI" xmlns:ecore="http://www.eclipse.org/emf/2002/Ecore"
    name="empty" nsURI="http://soft.gen/test/empty" nsPrefix="empty"/>
`)
	assert.Equal(t, "empty", m.Root().GetName())
	info := m.Package(m.Root())
	assert.Empty(t, info.Classes)
	assert.Empty(t, info.OrderedClasses)
	assert.Empty(t, info.AnnotationSources)
}

// TestLoadSchemaLocation checks that the package of the xsi:schemaLocation of
// a model loads, its generics stripped, when the contents of an annotation
// need it.
func TestLoadSchemaLocation(t *testing.T) {
	m := loadModel(t, "annotations")
	doc := classNamed(t, m, "Document").GetEAnnotation("http://soft.gen/test/doc")
	require.NotNil(t, doc)
	note := doc.GetContents().Get(1).(ecore.EObject)
	assert.Equal(t, "Note", note.EClass().GetName())
	assert.Equal(t, "http://soft.gen/test/notes", note.EClass().GetEPackage().GetNsURI())
	// the eType of text comes from its stripped eGenericType
	text := note.EClass().GetEStructuralFeatureFromName("text")
	require.NotNil(t, text)
	assert.Equal(t, "EString", text.GetEType().GetName())
	assert.Equal(t, "dropped by the generators, as in EMF", note.EGet(text))
}

// TestLoadSchemaLocationMissing checks that a schema location naming a
// missing file fails with the namespace and the location.
func TestLoadSchemaLocationMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.ecore")
	require.NoError(t, os.WriteFile(path, []byte(`<?xml version="1.0" encoding="UTF-8"?>
<ecore:EPackage xmi:version="2.0" xmlns:xmi="http://www.omg.org/XMI" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:ecore="http://www.eclipse.org/emf/2002/Ecore" xmlns:gone="http://soft.gen/test/gone"
    xsi:schemaLocation="http://soft.gen/test/gone gone.ecore"
    name="document" nsURI="http://soft.gen/test/document" nsPrefix="document">
  <eAnnotations source="meta"><contents xsi:type="gone:Note"/></eAnnotations>
</ecore:EPackage>
`), 0o600))
	_, err := Load(path, genGoSource)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http://soft.gen/test/gone")
	assert.Contains(t, err.Error(), "gone.ecore")
}

// TestLoadErrors checks that a missing file fails with its path.
func TestLoadErrors(t *testing.T) {
	_, err := Load("../../testdata/models/missing.ecore", genGoSource)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing.ecore")

	_, err = Load("../../testdata/props/does-not-exist.ecore", genGoSource)
	require.Error(t, err)
}

// TestFromPackage checks that a wrapped package loads the embedded ecore
// package on demand, once.
func TestFromPackage(t *testing.T) {
	m := FromPackage(ecore.GetPackage(), genGoSource)
	assert.Same(t, ecore.GetPackage(), m.Root())
	assert.True(t, m.IsEcore(m.Root()))
	// the embedded ecore package is loaded on demand
	assert.NotSame(t, ecore.GetPackage(), m.Ecore())
	assert.Equal(t, "java.lang.String", m.Ecore().GetEClassifier("EString").GetInstanceTypeName())
	assert.Same(t, m.Ecore(), m.Ecore())
}

// TestIsAny checks the classifiers any type can stand for, on the ecore metamodel.
func TestIsAny(t *testing.T) {
	m := ecoreMetamodel(t)
	assert.True(t, m.IsAny(nil))
	for _, name := range []string{"EEnumerator", "EDiagnosticChain", "EJavaObject"} {
		assert.True(t, m.IsAny(m.Root().GetEClassifier(name)), name)
	}
	assert.False(t, m.IsAny(m.Root().GetEClassifier("EString")))
	assert.False(t, m.IsAny(m.Root().GetEClassifier("EClass")))
}

// TestEntryNames checks the map entry and feature map entry instance type names.
func TestEntryNames(t *testing.T) {
	m := FromPackage(ecore.GetPackage(), genGoSource)
	assert.True(t, m.IsMapEntryName("java.util.Map$Entry"))
	assert.True(t, m.IsMapEntryName("java.util.Map.Entry"))
	assert.False(t, m.IsMapEntryName("java.lang.String"))
	assert.True(t, m.IsFeatureMapEntryName("org.eclipse.emf.ecore.util.FeatureMap$Entry"))
	assert.True(t, m.IsFeatureMapEntryName("org.eclipse.emf.ecore.util.FeatureMap.Entry"))
	assert.False(t, m.IsFeatureMapEntryName("java.util.Map$Entry"))
}

// TestMemoization checks that the facts of an element are computed once.
func TestMemoization(t *testing.T) {
	m := loadModel(t, "operations")
	circle := classNamed(t, m, "Circle")
	assert.Same(t, m.Class(circle), m.Class(circle))
	assert.Same(t, m.Package(m.Root()), m.Package(m.Root()))
	radius := featureNamed(t, m, "Circle", "radius")
	assert.Same(t, m.Feature(radius), m.Feature(radius))
	compute := operationNamed(t, m, "Calculator", "compute", 0)
	assert.Same(t, m.Operation(compute), m.Operation(compute))
}
