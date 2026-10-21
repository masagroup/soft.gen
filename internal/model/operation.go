package model

import "github.com/masagroup/soft.go/ecore"

// OperationInfo is the generator facts of an operation.
type OperationInfo struct {
	Operation ecore.EOperation
	Class     ecore.EClass // eContainingClass
	IsVoid    bool         // no type (v1 isVoid)
	IsDefault bool         // no `default` key, or "true" (v1 isDefault)
	// Identifier is the name of the operation, suffixed by the parameter type
	// names when it has parameters (v1 getElementIdentifier).
	Identifier string
	// ReferencedTypes is the return type and the parameter types, nil dropped,
	// in order, without duplicate (v1 getReferencedTypes).
	ReferencedTypes []ecore.EClassifier
	IsList          bool // many or upperBound != 1 (v1 isListType)
	IsMap           bool // a list whose type is a Map$Entry (v1 isMapType)
}

// Operation returns the facts of op, computed once.
func (m *Model) Operation(op ecore.EOperation) *OperationInfo {
	if info := m.operations[op]; info != nil {
		return info
	}
	info := &OperationInfo{Operation: op, Class: op.GetEContainingClass()}
	m.operations[op] = info
	info.IsVoid = op.GetEType() == nil
	annotation := m.Annotation(op)
	info.IsDefault = !annotation.HasKey("default") || annotation.Value("default") == "true"
	info.Identifier = m.ElementIdentifier(op)
	info.ReferencedTypes = referencedTypes(op)
	info.IsList = isListType(op)
	info.IsMap = m.isMapType(op)
	return info
}

// referencedTypes returns the types op references (v1 getReferencedTypes).
func referencedTypes(op ecore.EOperation) []ecore.EClassifier {
	var result []ecore.EClassifier
	add := func(c ecore.EClassifier) {
		if c != nil {
			result = appendUnique(result, c)
		}
	}
	// v1 also adds eGenericType.eClassifier, which soft.go does not model: it
	// equals eType for the non-generic operations soft.go loads
	add(op.GetEType())
	for _, parameter := range elements[ecore.EParameter](op.GetEParameters()) {
		add(parameter.GetEType())
	}
	return result
}
