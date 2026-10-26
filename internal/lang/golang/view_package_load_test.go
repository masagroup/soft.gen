package golang

import (
	"bytes"
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/masagroup/soft.gen/v2/internal/model"
)

// The serialized package of a package with loadInitialization loads back with
// all the contents of its annotations, the Note of the notes package included,
// without the annotations of its suppressed source, its references to the
// ecore package written with its namespace URI.
func TestSerializedPackageLoads(t *testing.T) {
	m, err := model.Load("../../../testdata/models/annotations.ecore", AnnotationSource)
	require.NoError(t, err)
	data, err := (&gen{m: m}).serialize(m.Root())
	require.NoError(t, err)
	assert.Contains(t, string(data), `eType="ecore:EDataType http://www.eclipse.org/emf/2002/Ecore#//EString"`)

	notes := m.Root().EResource().GetResourceSet().GetPackageRegistry().GetPackage("http://soft.gen/test/notes")
	require.NotNil(t, notes)
	resourceSet := ecore.NewEResourceSetImpl()
	resourceSet.GetPackageRegistry().RegisterPackage(notes)
	resource := resourceSet.CreateResource(ecore.NewURI("annotations.ecore"))
	resource.LoadWithReader(bytes.NewReader(data), nil)
	require.True(t, resource.GetErrors().Empty())
	ecore.ResolveAllInResource(resource)

	document := resource.GetContents().Get(0).(ecore.EPackage).GetEClassifier("Document").(ecore.EClass)
	annotation := document.GetEAnnotation("http://soft.gen/test/doc")
	require.NotNil(t, annotation)
	require.Equal(t, 2, annotation.GetContents().Size())
	nested, _ := annotation.GetContents().Get(0).(ecore.EAnnotation)
	require.NotNil(t, nested)
	assert.Equal(t, "http://soft.gen/test/content", nested.GetSource())
	note := annotation.GetContents().Get(1).(ecore.EObject)
	assert.Equal(t, notes, note.EClass().GetEPackage())
	assert.Equal(t, "dropped by the generators, as in EMF", note.EGet(note.EClass().GetEStructuralFeatureFromName("text")))
	assert.Nil(t, document.GetEAnnotation("http://soft.gen/test/hidden"))
	title := document.GetEStructuralFeatureFromName("title")
	require.NotNil(t, title)
	assert.Equal(t, ecore.GetPackage().GetEString(), title.GetEType())
}
