package model

import (
	"sort"
	"strings"

	"github.com/masagroup/soft.go/ecore"
)

// PackageInfo is the generator facts of a package.
type PackageInfo struct {
	Package     ecore.EPackage
	Classes     []ecore.EClass    // eClassifiers that are EClass, model order
	Enums       []ecore.EEnum     // eClassifiers that are EEnum, model order
	DataTypes   []ecore.EDataType // eClassifiers that are non-enum data types, model order
	SubPackages []ecore.EPackage  // eSubpackages, model order
	// AllClasses, AllEnums and AllDataTypes are the eAllContents of the package
	// (sub packages included), model order.
	AllClasses   []ecore.EClass
	AllEnums     []ecore.EEnum
	AllDataTypes []ecore.EDataType
	// OrderedClasses is the classes of the package ordered so that the first
	// super type chain of a class precedes it (v1 getOrderedClasses).
	OrderedClasses []ecore.EClass
	// Annotations is every annotation of the package and of its contents, in
	// the containment order of EMF (see emfAnnotations); the annotations
	// nested in an annotation are not listed (see NestedAnnotations).
	Annotations []ecore.EAnnotation
	// AnnotationSources is the distinct sources of Annotations minus the
	// suppressed ones, sorted (v1 and EMF getAnnotationSources): a source of
	// nested annotations only has no initializer of its own.
	AnnotationSources []AnnotationSource
}

// AnnotationSource is the source of the annotations of a package.
type AnnotationSource struct {
	URI        string // the source
	Identifier string // its last path segment, a trailing '/' ignored (v1 getAnnotationSourceIdentifier)
}

// Package returns the facts of p, computed once.
func (m *Model) Package(p ecore.EPackage) *PackageInfo {
	if info := m.packages[p]; info != nil {
		return info
	}
	info := &PackageInfo{Package: p}
	m.packages[p] = info
	for _, classifier := range p.GetEClassifiers().ToArray() {
		switch classifier := classifier.(type) {
		case ecore.EClass:
			info.Classes = append(info.Classes, classifier)
		case ecore.EEnum:
			info.Enums = append(info.Enums, classifier)
		case ecore.EDataType:
			info.DataTypes = append(info.DataTypes, classifier)
		}
	}
	info.SubPackages = elements[ecore.EPackage](p.GetESubPackages())
	for it := p.EAllContents(); it.HasNext(); {
		switch content := it.Next().(type) {
		case ecore.EClass:
			info.AllClasses = append(info.AllClasses, content)
		case ecore.EEnum:
			info.AllEnums = append(info.AllEnums, content)
		case ecore.EDataType:
			info.AllDataTypes = append(info.AllDataTypes, content)
		}
	}
	info.OrderedClasses = orderedClasses(p)
	info.Annotations = emfAnnotations(p, nil)
	sources := map[string]struct{}{}
	for _, a := range info.Annotations {
		sources[a.GetSource()] = struct{}{}
	}
	info.AnnotationSources = m.annotationSources(p, sources)
	return info
}

// emfAnnotations appends the annotations of el and of its contents to list, in
// the order of EMF's eAllContents, which v1 walked: an element's annotations,
// then its classifiers and sub-packages, its operations before its structural
// features, the parameters of an operation, the literals of an enum. soft.go's
// eAllContents orders the features of a class before its operations. The walk
// stops at an annotation, as EMF getAllAnnotations does: the annotations nested
// in it are created with it (see NestedAnnotations).
func emfAnnotations(el ecore.EModelElement, list []ecore.EAnnotation) []ecore.EAnnotation {
	list = append(list, elements[ecore.EAnnotation](el.GetEAnnotations())...)
	var children []ecore.EModelElement
	switch el := el.(type) {
	case ecore.EPackage:
		children = append(elements[ecore.EModelElement](el.GetEClassifiers()), elements[ecore.EModelElement](el.GetESubPackages())...)
	case ecore.EClass:
		children = append(elements[ecore.EModelElement](el.GetEOperations()), elements[ecore.EModelElement](el.GetEStructuralFeatures())...)
	case ecore.EOperation:
		children = elements[ecore.EModelElement](el.GetEParameters())
	case ecore.EEnum:
		children = elements[ecore.EModelElement](el.GetELiterals())
	}
	for _, child := range children {
		list = emfAnnotations(child, list)
	}
	return list
}

// NestedAnnotation is an annotation nested in another one, in its
// eAnnotations or in its contents, with its own nested annotations.
type NestedAnnotation struct {
	Annotation ecore.EAnnotation
	// Content is true for an annotation of the contents of its parent, false
	// for one of its eAnnotations.
	Content bool
	Nested  []NestedAnnotation
}

// NestedAnnotations returns the annotations nested in a, in the order of EMF's
// eAllContents: its eAnnotations, then the annotations of its contents. As in
// EMF GenPackage.getAllNestedAnnotations, the contents that are not annotations
// are left out, with what they hold: the generated packages do not have them.
func NestedAnnotations(a ecore.EAnnotation) []NestedAnnotation {
	var nested []NestedAnnotation
	for _, child := range elements[ecore.EAnnotation](a.GetEAnnotations()) {
		nested = append(nested, NestedAnnotation{Annotation: child, Nested: NestedAnnotations(child)})
	}
	for _, content := range a.GetContents().ToArray() {
		if child, ok := content.(ecore.EAnnotation); ok {
			nested = append(nested, NestedAnnotation{Annotation: child, Content: true, Nested: NestedAnnotations(child)})
		}
	}
	return nested
}

// annotationSources returns the sorted annotation sources of p minus the
// suppressed ones (v1 getAnnotationSources).
func (m *Model) annotationSources(p ecore.EPackage, sources map[string]struct{}) []AnnotationSource {
	for _, suppressed := range m.suppressedAnnotations(p) {
		delete(sources, suppressed)
	}
	uris := make([]string, 0, len(sources))
	for source := range sources {
		uris = append(uris, source)
	}
	sort.Strings(uris)
	result := make([]AnnotationSource, len(uris))
	for i, uri := range uris {
		result[i] = AnnotationSource{URI: uri, Identifier: annotationSourceIdentifier(uri)}
	}
	return result
}

// suppressedAnnotations returns the annotation sources the generator
// annotation of p suppresses: the space separated tokens of its
// supressAnnotations (sic) key (v1 getSuppressedAnnotations).
func (m *Model) suppressedAnnotations(p ecore.EPackage) []string {
	annotation := m.Annotation(p)
	if !annotation.HasKey("supressAnnotations") {
		return nil
	}
	return strings.Fields(annotation.Value("supressAnnotations"))
}

// annotationSourceIdentifier returns the identifier of an annotation source:
// its last path segment, a trailing '/' ignored (v1
// getAnnotationSourceIdentifier). v1 looked for the segment before a trailing
// '/' but kept the '/' in the result ("GenGo/" for ".../GenGo/", an unusable
// identifier): the intent is ported, not that bug.
func annotationSourceIdentifier(source string) string {
	source = strings.TrimSuffix(source, "/")
	if index := strings.LastIndex(source, "/"); index != -1 {
		return source[index+1:]
	}
	return source
}
