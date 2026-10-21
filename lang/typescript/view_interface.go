package typescript

import (
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/naming"
)

// interfaceFile is the file of the interface of a class (v1
// generateInterface): its extends clause and its sections, the property
// declarations, one section per feature with method accessors, the
// operations.
func (g *gen) interfaceFile(c ecore.EClass) engine.File {
	s := g.newScope().declare(g.classInterfaceIdentifier(c))
	view := sectionsView{Declaration: "export interface " + g.interfaceName(c) + s.extendsClause(g.interfaceExtends(c))}
	var properties []string
	var methods [][]string
	for it := c.GetEStructuralFeatures().Iterator(); it.HasNext(); {
		f := it.Next().(ecore.EStructuralFeature)
		info := g.m.Feature(f)
		if !info.Get && !info.Set {
			continue
		}
		if g.isPropertyAccessors(f) {
			properties = append(properties, s.propertyDeclaration(f)...)
		}
		if g.isMethodAccessors(f) {
			methods = append(methods, s.methodsDeclaration(f))
		}
	}
	if len(properties) > 0 {
		view.Sections = append(view.Sections, properties)
	}
	view.Sections = append(view.Sections, methods...)
	if c.GetEOperations().Size() > 0 {
		operations := []string{"// operations"}
		for it := c.GetEOperations().Iterator(); it.HasNext(); {
			operations = append(operations, s.operationDeclarations(it.Next().(ecore.EOperation))...)
		}
		view.Sections = append(view.Sections, operations)
	}
	return g.file(g.interfaceFileName(c, c.GetName()), s, "sections.ts", view)
}

// interfaceExtends are the interfaces the interface of a class extends (v1
// getInterfaceExtends): EObject and EMapEntry<K, V> for a map entry; for a
// root class ENotifier (the class EObject), the rootExtendsInterface property
// or EObject; else its super types.
func (g *gen) interfaceExtends(c ecore.EClass) []string {
	switch {
	case g.m.Class(c).IsMapEntry:
		key, value := g.mapEntryFeatureQID(c, "key"), g.mapEntryFeatureQID(c, "value")
		return []string{g.ecoreIdentifier("EObject"), g.ecoreIdentifier("EMapEntry<" + key + "," + value + ">")}
	case c.GetESuperTypes().Empty():
		if c.GetName() == "EObject" {
			return []string{g.ecoreIdentifier("ENotifier")}
		}
		if root, ok := g.ExplicitPackageProperty("rootExtendsInterface"); ok {
			return []string{root}
		}
		return []string{g.ecoreIdentifier("EObject")}
	}
	var extends []string
	for it := c.GetESuperTypes().Iterator(); it.HasNext(); {
		extends = append(extends, g.classInterfaceIdentifier(it.Next().(ecore.EClass)))
	}
	return extends
}

// extendsClause is " extends A, B", or nothing.
func (s *scope) extendsClause(identifiers []string) string {
	if len(identifiers) == 0 {
		return ""
	}
	written := make([]string, len(identifiers))
	for i, id := range identifiers {
		written[i] = s.qualify(id)
	}
	return " extends " + strings.Join(written, ", ")
}

// propertyDeclaration declares the property of a feature with property
// accessors: readonly without setter.
func (s *scope) propertyDeclaration(f ecore.EStructuralFeature) []string {
	info := s.g.m.Feature(f)
	readonly := ""
	if info.Get && !info.Set {
		readonly = "readonly "
	}
	return []string{"// " + f.GetName() + " property", readonly + propertyName(f) + ": " + s.typedElementType(f)}
}

// methodsDeclaration declares the accessor methods of a feature: its getter,
// the async getter of a single proxy, its setter and its unsetter.
func (s *scope) methodsDeclaration(f ecore.EStructuralFeature) []string {
	g := s.g
	info := g.m.Feature(f)
	t := s.typedElementType(f)
	lines := []string{"// " + f.GetName()}
	if info.Get {
		lines = append(lines, g.getterName(f)+"(): "+t)
	}
	if g.isAsyncGet(f) {
		lines = append(lines, g.getterName(f)+"Async(): Promise<"+t+">")
	}
	if info.Set {
		lines = append(lines, g.setterName(f)+"(new"+naming.UpperFirst(propertyName(f))+": "+t+"): void")
	}
	if info.Unset {
		lines = append(lines, g.unsetterName(f)+"(): void")
	}
	return lines
}

// isAsyncGet reports whether a feature has an async getter: a single proxy,
// not a container (v1 isAsyncGet).
func (g *gen) isAsyncGet(f ecore.EStructuralFeature) bool {
	info := g.m.Feature(f)
	return info.Get && info.IsProxy && !info.IsList && !info.IsContainer
}

// operationDeclarations declares an operation, and its async variant.
func (s *scope) operationDeclarations(op ecore.EOperation) []string {
	g := s.g
	parameters, result := s.parameterList(op), s.typedElementType(op)
	lines := []string{g.operationName(op) + "(" + parameters + "): " + result}
	if g.isAsync(op) {
		lines = append(lines, g.operationName(op)+"Async("+parameters+"): Promise<"+result+">")
	}
	return lines
}

// parameterList is the parameters of an operation, "name: type, ...".
func (s *scope) parameterList(op ecore.EOperation) string {
	var parameters []string
	for it := op.GetEParameters().Iterator(); it.HasNext(); {
		p := it.Next().(ecore.EParameter)
		parameters = append(parameters, parameterName(p)+": "+s.typedElementType(p))
	}
	return strings.Join(parameters, ", ")
}
