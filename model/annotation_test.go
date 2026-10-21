package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotation(t *testing.T) {
	m := loadModel(t, "ecore")
	root := m.Annotation(m.Root())
	assert.True(t, root.Exists())
	require.NotNil(t, root.EAnnotation())
	assert.Equal(t, genGoSource, root.EAnnotation().GetSource())
	assert.True(t, root.HasKey("lazy"))
	assert.True(t, root.HasKey("exported"))
	assert.False(t, root.HasKey("factory")) // a GenCpp key
	assert.Equal(t, "list:function", root.Value("lazy"))
	assert.Equal(t, "", root.Value("factory"))

	eClass := m.Annotation(classNamed(t, m, "EClass"))
	assert.True(t, eClass.Exists())
	assert.Equal(t, "true", eClass.Value("extension"))

	// an element without a generator annotation
	eString := m.Annotation(m.Root().GetEClassifier("EString"))
	assert.False(t, eString.Exists())
	assert.Nil(t, eString.EAnnotation())
	assert.False(t, eString.HasKey("exported"))
	assert.Equal(t, "", eString.Value("exported"))

	none := m.Annotation(nil)
	assert.False(t, none.Exists())
	_, found := none.Lookup("exported")
	assert.False(t, found)

	// the source selects the annotation
	ts := FromPackage(m.Root(), genTSSource)
	assert.Equal(t, "getEStructuralFeatureFromName", ts.Annotation(operationNamed(t, ts, "EClass", "getEStructuralFeature", 1)).Value("operationName"))
	assert.Equal(t, "GetEStructuralFeatureFromName", m.Annotation(operationNamed(t, m, "EClass", "getEStructuralFeature", 1)).Value("operationName"))
}

func TestAnnotationLookup(t *testing.T) {
	m := loadModel(t, "ecore")
	eSuperTypes := featureNamed(t, m, "EClass", "eSuperTypes")
	eAttributes := featureNamed(t, m, "EClass", "eAttributes")

	// the key of the feature
	value, found := m.Annotation(eAttributes).Lookup("lazy")
	assert.True(t, found)
	assert.Equal(t, "action", value)
	// the key of the package, through the class
	value, found = m.Annotation(eSuperTypes).Lookup("lazy")
	assert.True(t, found)
	assert.Equal(t, "list:function", value)
	value, found = m.Annotation(eSuperTypes).Lookup("exported")
	assert.True(t, found)
	assert.Equal(t, "true", value)
	// the key of the class
	value, found = m.Annotation(eSuperTypes).Lookup("extension")
	assert.True(t, found)
	assert.Equal(t, "true", value)
	// an operation goes through its class
	value, found = m.Annotation(operationNamed(t, m, "EClass", "getFeatureCount", 0)).Lookup("extension")
	assert.True(t, found)
	assert.Equal(t, "true", value)
	// an absent key
	_, found = m.Annotation(eSuperTypes).Lookup("unknown")
	assert.False(t, found)
	// a data type has no ancestor: the package key is not found
	_, found = m.Annotation(m.Root().GetEClassifier("EString")).Lookup("exported")
	assert.False(t, found)
	// nor has a package
	_, found = m.Annotation(m.Root()).Lookup("extension")
	assert.False(t, found)
	value, found = m.Annotation(m.Root()).Lookup("exported")
	assert.True(t, found)
	assert.Equal(t, "true", value)
}

// operationNamed returns the index-th operation of the class named name.
func operationNamed(t *testing.T, m *Model, className, name string, index int) ecore.EOperation {
	t.Helper()
	for _, op := range elements[ecore.EOperation](classNamed(t, m, className).GetEOperations()) {
		if op.GetName() == name {
			if index == 0 {
				return op
			}
			index--
		}
	}
	require.Fail(t, "operation not found", "%s.%s", className, name)
	return nil
}
