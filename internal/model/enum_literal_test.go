package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnumLiteral checks that a loaded explicit literal="" stays "" while an
// absent literal gives the name, as in EMF: soft.go gave the name for both
// until EEnumLiteral tracked whether its literal is set.
func TestEnumLiteral(t *testing.T) {
	m := loadDocument(t, `<?xml version="1.0" encoding="UTF-8"?>
<ecore:EPackage xmi:version="2.0" xmlns:xmi="http://www.omg.org/XMI" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:ecore="http://www.eclipse.org/emf/2002/Ecore" name="p" nsURI="http://p" nsPrefix="p">
  <eClassifiers xsi:type="ecore:EEnum" name="Types">
    <eLiterals name="_" literal=""/>
    <eLiterals name="Absent" value="1"/>
    <eLiterals name="Named" value="2" literal="named"/>
  </eClassifiers>
</ecore:EPackage>
`)
	e, _ := m.Root().GetEClassifier("Types").(ecore.EEnum)
	require.NotNil(t, e)
	literal := func(name string) string {
		l := e.GetEEnumLiteralByName(name)
		require.NotNil(t, l, "literal %s", name)
		return l.GetLiteral()
	}
	assert.Equal(t, "", literal("_"))
	assert.Equal(t, "Absent", literal("Absent"))
	assert.Equal(t, "named", literal("Named"))
}
