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
	// the containment order of EMF (see emfAnnotations).
	Annotations []ecore.EAnnotation
	// AnnotationSources is the distinct sources of the annotations of the
	// eAllContents of the package minus the suppressed ones, sorted (v1
	// getAnnotationSources).
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
	sources := map[string]struct{}{}
	for it := p.EAllContents(); it.HasNext(); {
		switch content := it.Next().(type) {
		case ecore.EClass:
			info.AllClasses = append(info.AllClasses, content)
		case ecore.EEnum:
			info.AllEnums = append(info.AllEnums, content)
		case ecore.EDataType:
			info.AllDataTypes = append(info.AllDataTypes, content)
		case ecore.EAnnotation:
			sources[content.GetSource()] = struct{}{}
		}
	}
	info.OrderedClasses = orderedClasses(p)
	info.Annotations = emfAnnotations(p, nil)
	info.AnnotationSources = m.annotationSources(p, sources)
	return info
}

// emfAnnotations appends the annotations of el and of its contents to list, in
// the order of EMF's eAllContents, which v1 walked: an element's annotations
// (each followed by its own nested annotations and contents), then its
// classifiers and sub-packages, its operations before its structural features,
// the parameters of an operation, the literals of an enum. soft.go's
// eAllContents orders the features of a class before its operations.
func emfAnnotations(el ecore.EModelElement, list []ecore.EAnnotation) []ecore.EAnnotation {
	for _, a := range elements[ecore.EAnnotation](el.GetEAnnotations()) {
		list = append(list, a)
		list = emfAnnotations(a, list)
		for _, content := range a.GetContents().ToArray() {
			if content, ok := content.(ecore.EModelElement); ok {
				list = emfAnnotations(content, list)
			}
		}
	}
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
