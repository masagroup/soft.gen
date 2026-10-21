package model

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/masagroup/soft.go/codec/xmi"
	"github.com/masagroup/soft.go/codec/xml"
	"github.com/masagroup/soft.go/ecore"
)

// ecoreModel is the ecore metamodel the generators see (v1 getEcorePackage,
// EcorePackage.eINSTANCE of EMF): the Java instance type names and the
// generator annotations of ecore.ecore, not the runtime package of soft.go.
//
//go:embed ecore.ecore
var ecoreModel []byte

// EcoreNsURI is the namespace URI of the ecore package.
const EcoreNsURI = "http://www.eclipse.org/emf/2002/Ecore"

// ecoreModelURI is the uri of the resource of the embedded ecore package: its
// scheme never clashes with a loaded file, its extension selects the xmi codec.
var ecoreModelURI = ecore.NewURI("embedded:/soft.gen/model/ecore.ecore")

// Model is a loaded Ecore model with the generator facts of its elements. The
// facts are memoized per element: Package, Class, Feature and Operation return
// the same *Info for the same element.
type Model struct {
	root             ecore.EPackage
	annotationSource string
	ecore            ecore.EPackage
	packages         map[ecore.EPackage]*PackageInfo
	classes          map[ecore.EClass]*ClassInfo
	features         map[ecore.EStructuralFeature]*FeatureInfo
	operations       map[ecore.EOperation]*OperationInfo
}

// Load loads the .ecore file at path and returns its root package as a Model
// whose generator annotations have the given source. The cross references to
// the ecore package (http://www.eclipse.org/emf/2002/Ecore#//EString...)
// resolve to the embedded ecore.ecore, as in v1 where they resolved to
// EcorePackage.eINSTANCE. The generic elements of the file, which soft.go
// does not model, give the elements their eType and are dropped (see
// stripGenerics). It fails with the diagnostics of the resource when the file
// can not be loaded or has no EPackage root.
func Load(path, annotationSource string) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("model: loading %s: %w", path, err)
	}
	resourceSet := ecore.NewEResourceSetImpl()
	processor := xmi.NewProcessor(xml.ProcessorResourceSet(resourceSet))
	ecorePackage, err := loadEcoreModel(resourceSet, processor)
	if err != nil {
		return nil, err
	}
	// the resource set registry answers before the global one: the proxies to
	// the ecore package resolve in the embedded package, while the objects of
	// the ecore namespace are still created by the ecore factory of soft.go
	registry := resourceSet.GetPackageRegistry()
	registry.PutPackage(EcoreNsURI, ecorePackage)
	registry.PutFactory(EcoreNsURI, ecore.GetFactory())

	root, err := loadPackage(resourceSet, processor, ecore.CreateFileURI(path), data)
	if err != nil {
		return nil, fmt.Errorf("model: loading %s: %w", path, err)
	}
	m := FromPackage(root, annotationSource)
	m.ecore = ecorePackage
	return m, nil
}

// FromPackage wraps an already loaded package (tests, embedding).
func FromPackage(root ecore.EPackage, annotationSource string) *Model {
	return &Model{
		root:             root,
		annotationSource: annotationSource,
		packages:         map[ecore.EPackage]*PackageInfo{},
		classes:          map[ecore.EClass]*ClassInfo{},
		features:         map[ecore.EStructuralFeature]*FeatureInfo{},
		operations:       map[ecore.EOperation]*OperationInfo{},
	}
}

// loadEcoreModel loads the embedded ecore.ecore in resourceSet and returns its
// package.
func loadEcoreModel(resourceSet ecore.EResourceSet, processor *xmi.Processor) (ecore.EPackage, error) {
	root, err := loadPackage(resourceSet, processor, ecoreModelURI, ecoreModel)
	if err != nil {
		return nil, fmt.Errorf("model: loading the embedded ecore.ecore: %w", err)
	}
	return root, nil
}

// loadPackage decodes the .ecore document data as the resource uri of
// resourceSet, its generics stripped, and returns its root package.
func loadPackage(resourceSet ecore.EResourceSet, processor *xmi.Processor, uri *ecore.URI, data []byte) (ecore.EPackage, error) {
	data, err := stripGenerics(data)
	if err != nil {
		return nil, err
	}
	resource := resourceSet.CreateResource(uri)
	options := map[string]any{xml.OPTION_EXTENDED_META_DATA: processor.GetExtendedMetaData()}
	resource.LoadWithReader(bytes.NewReader(data), options)
	return rootPackage(resource)
}

// rootPackage returns the root package of a loaded resource, or the error
// made of the diagnostics of the resource.
func rootPackage(resource ecore.EResource) (ecore.EPackage, error) {
	if errs := resource.GetErrors(); !errs.Empty() {
		return nil, diagnosticsError(errs)
	}
	if !resource.IsLoaded() {
		return nil, errors.New("the resource is not loaded")
	}
	if resource.GetContents().Empty() {
		return nil, errors.New("the resource is empty")
	}
	root, _ := resource.GetContents().Get(0).(ecore.EPackage)
	if root == nil {
		return nil, errors.New("the root of the resource is not an EPackage")
	}
	return root, nil
}

// diagnosticsError returns the error made of the diagnostics of a resource.
func diagnosticsError(diagnostics ecore.EList) error {
	var messages []string
	for _, d := range diagnostics.ToArray() {
		switch d := d.(type) {
		case ecore.EDiagnostic:
			message := d.GetMessage()
			if location := d.GetLocation(); location != "" {
				message = fmt.Sprintf("%s (%s:%d:%d)", message, location, d.GetLine(), d.GetColumn())
			}
			messages = append(messages, message)
		case error:
			messages = append(messages, d.Error())
		default:
			messages = append(messages, fmt.Sprint(d))
		}
	}
	return errors.New(strings.Join(messages, "; "))
}

// Root returns the root package of the model.
func (m *Model) Root() ecore.EPackage {
	return m.root
}

// AnnotationSource returns the source of the generator annotations
// (http://net.masagroup/soft/2019/GenGo...).
func (m *Model) AnnotationSource() string {
	return m.annotationSource
}

// Ecore returns the embedded ecore package (v1 getEcorePackage): the package
// the cross references of a loaded model resolve to.
func (m *Model) Ecore() ecore.EPackage {
	if m.ecore == nil {
		resourceSet := ecore.NewEResourceSetImpl()
		processor := xmi.NewProcessor(xml.ProcessorResourceSet(resourceSet))
		ecorePackage, err := loadEcoreModel(resourceSet, processor)
		if err != nil {
			panic(err)
		}
		m.ecore = ecorePackage
	}
	return m.ecore
}

// IsEcore returns whether p is the ecore package (v1 isEcorePackage): its name
// is "ecore".
func (m *Model) IsEcore(p ecore.EPackage) bool {
	return p != nil && p.GetName() == "ecore"
}

// IsEcoreClassifier returns whether c belongs to the ecore package.
func (m *Model) IsEcoreClassifier(c ecore.EClassifier) bool {
	return c != nil && m.IsEcore(c.GetEPackage())
}

// IsAny returns whether c is a classifier any type can stand for (v1 isAny):
// undefined, or named EEnumerator, EDiagnosticChain or EJavaObject.
func (m *Model) IsAny(c ecore.EClassifier) bool {
	if c == nil {
		return true
	}
	switch c.GetName() {
	case "EEnumerator", "EDiagnosticChain", "EJavaObject":
		return true
	}
	return false
}

// IsMapEntryName returns whether instanceTypeName is the Java map entry type
// (v1 isMap).
func (m *Model) IsMapEntryName(instanceTypeName string) bool {
	return instanceTypeName == "java.util.Map.Entry" || instanceTypeName == "java.util.Map$Entry"
}

// IsFeatureMapEntryName returns whether instanceTypeName is the EMF feature
// map entry type (v1 isFeatureMap).
func (m *Model) IsFeatureMapEntryName(instanceTypeName string) bool {
	return instanceTypeName == "org.eclipse.emf.ecore.util.FeatureMap.Entry" || instanceTypeName == "org.eclipse.emf.ecore.util.FeatureMap$Entry"
}

// elements returns the elements of list that are T, in list order.
func elements[T any](list ecore.EList) []T {
	result := make([]T, 0, list.Size())
	for _, e := range list.ToArray() {
		if t, ok := e.(T); ok {
			result = append(result, t)
		}
	}
	return result
}

// appendUnique appends to dst the elements of src not in dst yet.
func appendUnique[T comparable](dst []T, src ...T) []T {
	for _, e := range src {
		found := false
		for _, d := range dst {
			if d == e {
				found = true
				break
			}
		}
		if !found {
			dst = append(dst, e)
		}
	}
	return dst
}
