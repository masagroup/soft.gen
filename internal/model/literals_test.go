package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnumLiteral checks that an explicit literal="" stays "" while an absent
// literal gives the name, as in EMF.
func TestEnumLiteral(t *testing.T) {
	m := loadDocument(t, `<?xml version="1.0" encoding="UTF-8"?>
<ecore:EPackage xmi:version="2.0" xmlns:xmi="http://www.omg.org/XMI" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:ecore="http://www.eclipse.org/emf/2002/Ecore" name="p" nsURI="http://p" nsPrefix="p">
  <eClassifiers xsi:type="ecore:EEnum" name="Types">
    <eLiterals name="_" literal=""/>
    <eLiterals name="Absent" value="1"/>
    <eLiterals name="Named" value="2" literal="named"/>
  </eClassifiers>
  <eClassifiers xsi:type="ecore:EEnum" name="Other">
    <eLiterals name="First"/>
    <eLiterals name="Empty" value="1" literal=""/>
  </eClassifiers>
</ecore:EPackage>
`)
	literal := func(p ecore.EPackage, enum, name string) ecore.EEnumLiteral {
		e, _ := p.GetEClassifier(enum).(ecore.EEnum)
		require.NotNil(t, e, "enum %s", enum)
		l := e.GetEEnumLiteralByName(name)
		require.NotNil(t, l, "literal %s.%s", enum, name)
		return l
	}
	root := m.Root()
	assert.Equal(t, "", m.EnumLiteral(literal(root, "Types", "_")))
	assert.Equal(t, "Absent", m.EnumLiteral(literal(root, "Types", "Absent")))
	assert.Equal(t, "named", m.EnumLiteral(literal(root, "Types", "Named")))
	assert.Equal(t, "First", m.EnumLiteral(literal(root, "Other", "First")))
	assert.Equal(t, "", m.EnumLiteral(literal(root, "Other", "Empty")))
}
