package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	genGoSource = "http://net.masagroup/soft/2019/GenGo"
	genTSSource = "http://net.masagroup/soft/2020/GenTS"
)

func loadModel(t *testing.T, name string) *Model {
	t.Helper()
	m, err := Load("../testdata/models/"+name+".ecore", genGoSource)
	require.NoError(t, err)
	require.NotNil(t, m)
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

func TestLoadLibrary(t *testing.T) {
	m := loadModel(t, "library")
	assert.Equal(t, "library", m.Root().GetName())
	assert.Equal(t, genGoSource, m.AnnotationSource())
	assert.False(t, m.IsEcore(m.Root()))

	// the cross references resolve to the embedded ecore package, with the
	// Java instance type names of ecore.ecore
	ecorePackage := m.Ecore()
	require.NotNil(t, ecorePackage)
	assert.True(t, m.IsEcore(ecorePackage))
	assert.Equal(t, EcoreNsURI, ecorePackage.GetNsURI())
	eString := ecorePackage.GetEClassifier("EString")
	require.NotNil(t, eString)
	assert.Equal(t, "java.lang.String", eString.GetInstanceTypeName())
	assert.NotSame(t, ecore.GetPackage().GetEString(), eString)

	title := featureNamed(t, m, "Book", "title")
	assert.Same(t, eString, title.GetEType())
	assert.Equal(t, "java.lang.String", title.GetEType().GetInstanceTypeName())
	assert.True(t, m.IsEcoreClassifier(title.GetEType()))
	assert.False(t, m.IsEcoreClassifier(classNamed(t, m, "Book")))

	// a class reference to the ecore package resolves to a real EClass of the
	// embedded package, with the generator annotations of ecore.ecore
	mapEntry := featureNamed(t, m, "DocumentRoot", "xMLNSPrefixMap").GetEType()
	require.NotNil(t, mapEntry)
	assert.Same(t, ecorePackage.GetEClassifier("EStringToStringMapEntry"), mapEntry)
	assert.Equal(t, "java.util.Map$Entry", mapEntry.GetInstanceTypeName())
	assert.True(t, m.Annotation(ecorePackage).HasKey("lazy"))
	assert.Equal(t, "list:function", m.Annotation(ecorePackage).Value("lazy"))
}

func TestLoadEcore(t *testing.T) {
	m := loadModel(t, "ecore")
	root := m.Root()
	assert.Equal(t, "ecore", root.GetName())
	assert.True(t, m.IsEcore(root))
	assert.NotSame(t, m.Ecore(), root)
	// the references of ecore.ecore are local: EClass.eSuperTypes types to its own EClass
	eClass := classNamed(t, m, "EClass")
	assert.Same(t, eClass, eClass.GetEStructuralFeatureFromName("eSuperTypes").GetEType())
	assert.Equal(t, 53, root.GetEClassifiers().Size())
}

func TestLoadEmpty(t *testing.T) {
	m := loadModel(t, "empty")
	info := m.Package(m.Root())
	assert.Empty(t, info.Classes)
	assert.Empty(t, info.OrderedClasses)
	assert.Empty(t, info.AnnotationSources)
}

func TestLoadErrors(t *testing.T) {
	_, err := Load("../testdata/models/missing.ecore", genGoSource)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing.ecore")

	_, err = Load("../testdata/props/does-not-exist.ecore", genGoSource)
	require.Error(t, err)
}

func TestFromPackage(t *testing.T) {
	m := FromPackage(ecore.GetPackage(), genGoSource)
	assert.Same(t, ecore.GetPackage(), m.Root())
	assert.True(t, m.IsEcore(m.Root()))
	// the embedded ecore package is loaded on demand
	assert.NotSame(t, ecore.GetPackage(), m.Ecore())
	assert.Equal(t, "java.lang.String", m.Ecore().GetEClassifier("EString").GetInstanceTypeName())
	assert.Same(t, m.Ecore(), m.Ecore())
}

func TestIsAny(t *testing.T) {
	m := loadModel(t, "ecore")
	assert.True(t, m.IsAny(nil))
	for _, name := range []string{"EEnumerator", "EDiagnosticChain", "EJavaObject"} {
		assert.True(t, m.IsAny(m.Root().GetEClassifier(name)), name)
	}
	assert.False(t, m.IsAny(m.Root().GetEClassifier("EString")))
	assert.False(t, m.IsAny(m.Root().GetEClassifier("EClass")))
}

func TestEntryNames(t *testing.T) {
	m := FromPackage(ecore.GetPackage(), genGoSource)
	assert.True(t, m.IsMapEntryName("java.util.Map$Entry"))
	assert.True(t, m.IsMapEntryName("java.util.Map.Entry"))
	assert.False(t, m.IsMapEntryName("java.lang.String"))
	assert.True(t, m.IsFeatureMapEntryName("org.eclipse.emf.ecore.util.FeatureMap$Entry"))
	assert.True(t, m.IsFeatureMapEntryName("org.eclipse.emf.ecore.util.FeatureMap.Entry"))
	assert.False(t, m.IsFeatureMapEntryName("java.util.Map$Entry"))
}

func TestMemoization(t *testing.T) {
	m := loadModel(t, "library")
	book := classNamed(t, m, "Book")
	assert.Same(t, m.Class(book), m.Class(book))
	assert.Same(t, m.Package(m.Root()), m.Package(m.Root()))
	title := featureNamed(t, m, "Book", "title")
	assert.Same(t, m.Feature(title), m.Feature(title))
	op := classNamed(t, loadModel(t, "ecore"), "EClass").GetEOperations().Get(0).(ecore.EOperation)
	m2 := FromPackage(op.GetEContainingClass().GetEPackage(), genGoSource)
	assert.Same(t, m2.Operation(op), m2.Operation(op))
}
