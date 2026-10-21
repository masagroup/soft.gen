package model

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStripGenerics checks that the generic elements give their eType and are
// dropped, and that a document without generics is returned as is.
func TestStripGenerics(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<ecore:EPackage xmlns:ecore="http://www.eclipse.org/emf/2002/Ecore" name="p">
  <eClassifiers xsi:type="ecore:EClass" name="A">
    <eOperations name="typed" eType="#//EInt">
      <eGenericType eClassifier="#//EInt"/>
    </eOperations>
    <eOperations name="generic">
      <eAnnotations source="s"><details key="k" value="v"/></eAnnotations>
      <eGenericType eClassifier="#//EEList">
        <eTypeArguments eClassifier="#//EObject"/>
      </eGenericType>
      <eParameters name="arguments">
        <eGenericType eClassifier="#//EEList">
          <eTypeArguments/>
        </eGenericType>
      </eParameters>
    </eOperations>
    <eStructuralFeatures xsi:type="ecore:EAttribute" name="instanceClass">
      <eGenericType eClassifier="#//EJavaClass">
        <eTypeArguments/>
      </eGenericType>
    </eStructuralFeatures>
    <eOperations name="untyped">
      <eGenericType eTypeParameter="#//EEList/E"/>
    </eOperations>
  </eClassifiers>
  <eClassifiers xsi:type="ecore:EDataType" name="EEList" instanceClassName="org.eclipse.emf.common.util.EList">
    <eTypeParameters name="E"/>
  </eClassifiers>
</ecore:EPackage>
`
	output, err := stripGenerics([]byte(input))
	require.NoError(t, err)
	result := string(output)
	assert.NotContains(t, result, "eGenericType")
	assert.NotContains(t, result, "eTypeArguments")
	assert.NotContains(t, result, "eTypeParameters")
	assert.Contains(t, result, `<eOperations name="typed" eType="#//EInt">`)
	assert.Contains(t, result, `<eOperations name="generic" eType="#//EEList">`)
	assert.Contains(t, result, `<eParameters name="arguments" eType="#//EEList">`)
	assert.Contains(t, result, `<eStructuralFeatures xsi:type="ecore:EAttribute" name="instanceClass" eType="#//EJavaClass">`)
	assert.Contains(t, result, `<eOperations name="untyped">`)
	assert.Contains(t, result, `<eAnnotations source="s"><details key="k" value="v"/></eAnnotations>`)
	assert.Contains(t, result, `<eClassifiers xsi:type="ecore:EDataType" name="EEList" instanceClassName="org.eclipse.emf.common.util.EList">`)
	assert.Equal(t, 1, strings.Count(result, `eType="#//EInt"`))

	// a document without generics is returned as is
	plain, err := os.ReadFile("../testdata/models/references.ecore")
	require.NoError(t, err)
	same, err := stripGenerics(plain)
	require.NoError(t, err)
	assert.Equal(t, plain, same)

	_, err = stripGenerics([]byte("<a><b></a>"))
	assert.Error(t, err)
}
