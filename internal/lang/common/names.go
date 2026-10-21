package common

import "github.com/masagroup/soft.go/ecore"

// ReflectiveGetterName is the getter of a model element in the package
// interface (v1 getReflectiveGetterName): get<Accessor>, suffixed by the kind
// of the element when EPackage already has an operation of that name
// (getEClassifier -> getEClassifierClass). Go upper-cases its first letter.
func (g Gen) ReflectiveGetterName(el ecore.ENamedElement) string {
	getter := "get" + g.Model.ElementAccessorName(el)
	if g.ecorePackageHasOperation(getter) {
		getter += elementTypeName(el)
	}
	return getter
}

// ecorePackageHasOperation reports whether the EPackage class of the Ecore
// metamodel has an operation of that name.
func (g Gen) ecorePackageHasOperation(name string) bool {
	class, _ := g.Model.Ecore().GetEClassifier("EPackage").(ecore.EClass)
	if class == nil {
		return false
	}
	for it := class.GetEAllOperations().Iterator(); it.HasNext(); {
		if it.Next().(ecore.EOperation).GetName() == name {
			return true
		}
	}
	return false
}

// elementTypeName is the suffix of a reflective getter by kind of element (v1
// getElementTypeName).
func elementTypeName(el ecore.ENamedElement) string {
	switch el.(type) {
	case ecore.EAttribute:
		return "Attribute"
	case ecore.EReference:
		return "Reference"
	case ecore.EOperation:
		return "Operation"
	case ecore.EClass:
		return "Class"
	}
	return ""
}
