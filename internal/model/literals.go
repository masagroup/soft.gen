package model

import "github.com/masagroup/soft.go/ecore"

// EnumLiteral returns the literal of an enum literal as EMF gives it: "" for
// an explicit literal="" of the document, where the GetLiteral of soft.go
// gives the name, as for an unset literal.
func (m *Model) EnumLiteral(l ecore.EEnumLiteral) string {
	if m.emptyLiterals[l] {
		return ""
	}
	return l.GetLiteral()
}

// resolveEmptyLiterals finds the enum literals with an explicit empty literal
// recorded while loading.
func (m *Model) resolveEmptyLiterals(refs map[string]bool) map[ecore.EEnumLiteral]bool {
	resolved := map[ecore.EEnumLiteral]bool{}
	if len(refs) == 0 {
		return resolved
	}
	m.visitPositions(func(position string, el ecore.EObject) {
		if literal, ok := el.(ecore.EEnumLiteral); ok && refs[position] {
			resolved[literal] = true
		}
	})
	return resolved
}
