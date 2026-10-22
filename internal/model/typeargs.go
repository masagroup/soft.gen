package model

import (
	"strconv"
	"strings"

	"github.com/masagroup/soft.go/ecore"
)

// TypeArguments returns the classifiers of the type arguments of the generic
// type of a typed element (EEList<EObject> gives [EObject]), the generics the
// loader removed from the model: nil for an element without them, a nil
// entry for a wildcard or a type parameter.
func (m *Model) TypeArguments(el ecore.ETypedElement) []ecore.EClassifier {
	return m.typeArguments[el]
}

// resolveTypeArguments finds the elements and the classifiers of the type
// arguments recorded while loading.
func (m *Model) resolveTypeArguments(refs typeArgumentRefs) map[ecore.ETypedElement][]ecore.EClassifier {
	resolved := map[ecore.ETypedElement][]ecore.EClassifier{}
	if len(refs) == 0 {
		return resolved
	}
	m.visitPositions(func(position string, el ecore.EObject) {
		if typed, ok := el.(ecore.ETypedElement); ok {
			if args, found := refs[position]; found {
				classifiers := make([]ecore.EClassifier, len(args))
				for i, ref := range args {
					classifiers[i] = m.classifierOfRef(ref)
				}
				resolved[typed] = classifiers
			}
		}
	})
	return resolved
}

// visitPositions calls visit with the elements of the model that can be
// recorded while loading and their position in the document (see
// elementPosition).
func (m *Model) visitPositions(visit func(position string, el ecore.EObject)) {
	var walk func(position string, el ecore.EObject)
	children := func(position, local string, list ecore.EList) {
		for i := 0; i < list.Size(); i++ {
			walk(elementPosition(position, local, i), list.Get(i).(ecore.EObject))
		}
	}
	walk = func(position string, el ecore.EObject) {
		visit(position, el)
		switch el := el.(type) {
		case ecore.EPackage:
			children(position, "eClassifiers", el.GetEClassifiers())
			children(position, "eSubpackages", el.GetESubPackages())
		case ecore.EClass:
			children(position, "eStructuralFeatures", el.GetEStructuralFeatures())
			children(position, "eOperations", el.GetEOperations())
		case ecore.EEnum:
			children(position, "eLiterals", el.GetELiterals())
		case ecore.EOperation:
			children(position, "eParameters", el.GetEParameters())
		}
	}
	walk("", m.root)
}

// classifierOfRef returns the classifier of a reference of an .ecore document
// ("#//EObject", "ecore:EDataType http://www.eclipse.org/emf/2002/Ecore#//EString"):
// in the model or in the ecore package; nil when not found.
func (m *Model) classifierOfRef(ref string) ecore.EClassifier {
	if ref == "" {
		return nil
	}
	if i := strings.LastIndexByte(ref, ' '); i >= 0 {
		ref = ref[i+1:]
	}
	uri, fragment, ok := strings.Cut(ref, "#")
	if !ok {
		return nil
	}
	var p ecore.EPackage
	switch {
	case uri == "" || uri == m.root.GetNsURI():
		p = m.root
	case uri == EcoreNsURI && m.ecore != nil:
		p = m.ecore
	default:
		return nil
	}
	segments := strings.Split(strings.TrimPrefix(fragment, "//"), "/")
	for _, sub := range segments[:len(segments)-1] {
		p = subPackageNamed(p, sub)
		if p == nil {
			return nil
		}
	}
	return p.GetEClassifier(segments[len(segments)-1])
}

// subPackageNamed returns the sub package of that name (or index "@n").
func subPackageNamed(p ecore.EPackage, name string) ecore.EPackage {
	if index, err := strconv.Atoi(strings.TrimPrefix(name, "@eSubpackages.")); err == nil && index < p.GetESubPackages().Size() {
		return p.GetESubPackages().Get(index).(ecore.EPackage)
	}
	for it := p.GetESubPackages().Iterator(); it.HasNext(); {
		if sub := it.Next().(ecore.EPackage); sub.GetName() == name {
			return sub
		}
	}
	return nil
}
