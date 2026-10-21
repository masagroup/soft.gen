package model

import "github.com/masagroup/soft.go/ecore"

// Annotation is the generator annotation of an element: the EAnnotation whose
// source is the annotation source of the model (v1 getGenAnnotation).
type Annotation struct {
	model      *Model
	element    ecore.EModelElement
	annotation ecore.EAnnotation // nil when the element has no generator annotation
}

// Annotation returns the generator annotation of el.
func (m *Model) Annotation(el ecore.EModelElement) Annotation {
	var annotation ecore.EAnnotation
	if el != nil {
		annotation = el.GetEAnnotation(m.annotationSource)
	}
	return Annotation{model: m, element: el, annotation: annotation}
}

// Exists returns whether the element has a generator annotation (v1 hasGenAnnotation).
func (a Annotation) Exists() bool {
	return a.annotation != nil
}

// EAnnotation returns the annotation object, nil when the element has none.
func (a Annotation) EAnnotation() ecore.EAnnotation {
	return a.annotation
}

// HasKey returns whether the annotation has the detail key (v1 hasKey).
func (a Annotation) HasKey(key string) bool {
	return a.Exists() && a.annotation.GetDetails().ContainsKey(key)
}

// Value returns the value of the detail key, "" when absent (v1 valueKey).
func (a Annotation) Value(key string) string {
	if !a.Exists() {
		return ""
	}
	value, _ := a.annotation.GetDetails().GetValue(key).(string)
	return value
}

// Lookup returns the value of the detail key on the element, else on its
// ancestors (v1 getAnnotationValue of generateCommon.mtl):
// a feature or an operation is looked up, then its containing class, then the
// package of the class. The lookup stops at the package: a parent package, a
// data type or an enum has no ancestor (v1 getAncestor).
func (a Annotation) Lookup(key string) (string, bool) {
	for el := a.element; el != nil; el = annotationAncestor(el) {
		if annotation := a.model.Annotation(el); annotation.HasKey(key) {
			return annotation.Value(key), true
		}
	}
	return "", false
}

// annotationAncestor returns the element a key is looked up on when el does
// not have it (v1 getAncestor): the containing class of an operation or a
// feature, the package of a class, nil otherwise.
func annotationAncestor(el ecore.EModelElement) ecore.EModelElement {
	switch el := el.(type) {
	case ecore.EOperation:
		if c := el.GetEContainingClass(); c != nil {
			return c
		}
	case ecore.EStructuralFeature:
		if c := el.GetEContainingClass(); c != nil {
			return c
		}
	case ecore.EClass:
		if p := el.GetEPackage(); p != nil {
			return p
		}
	}
	return nil
}
