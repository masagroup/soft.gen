package typescript

import (
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// The names of the generated types, ports of the v1 naming queries
// (generateCommon.mtl of soft.generator.ts): each one applies the naming
// rule of its key, found on the element or its ancestors, else the default
// rule.

// factoryName is the factory of a package: rule factoryName, default
// $1Factory, upper-cased first.
func (g *gen) factoryName(p ecore.EPackage) string {
	return naming.UpperFirst(g.ElementName(p, p.GetName(), "factoryName", "$1Factory"))
}

// factoryImplementationName is the implementation of the factory.
func (g *gen) factoryImplementationName(p ecore.EPackage) string {
	return g.ElementName(p, g.factoryName(p), "implementationName", "$1Impl")
}

// factoryExtensionName is the hand-written extension of the factory
// (factory=extension).
func (g *gen) factoryExtensionName(p ecore.EPackage) string {
	return g.ElementName(p, g.factoryName(p), "extensionName", "$1Ext")
}

// packageName is the package interface of a package: rule packageName,
// default $1Package, upper-cased first.
func (g *gen) packageName(p ecore.EPackage) string {
	return naming.UpperFirst(g.ElementName(p, p.GetName(), "packageName", "$1Package"))
}

// packageInterfaceName is the package interface, through the interfaceName rule.
func (g *gen) packageInterfaceName(p ecore.EPackage) string {
	return naming.UpperFirst(g.ElementName(p, g.packageName(p), "interfaceName", ""))
}

// packageImplementationName is the implementation of the package.
func (g *gen) packageImplementationName(p ecore.EPackage) string {
	return g.ElementName(p, g.packageName(p), "implementationName", "$1Impl")
}

// constantsName is the class of the constants of a package: rule
// constantsName on the upper-cased name, default $1Constants.
func (g *gen) constantsName(p ecore.EPackage) string {
	return g.ElementName(p, naming.UpperFirst(p.GetName()), "constantsName", "$1Constants")
}

// interfaceName is the interface of a class: rule interfaceName, upper-cased first.
func (g *gen) interfaceName(c ecore.EClass) string {
	return naming.UpperFirst(g.ElementName(c, c.GetName(), "interfaceName", ""))
}

// implementationName is the implementation of a class: rule
// implementationName, default $1Impl.
func (g *gen) implementationName(c ecore.EClass) string {
	return g.ElementName(c, c.GetName(), "implementationName", "$1Impl")
}

// extensionName is the hand-written extension of a class (extension=true):
// rule extensionName, default $1Ext.
func (g *gen) extensionName(c ecore.EClass) string {
	return g.ElementName(c, c.GetName(), "extensionName", "$1Ext")
}

// The file names, each with its rule and default.

func (g *gen) interfaceFileName(el ecore.ENamedElement, name string) string {
	return g.ElementName(el, name, "interfaceFileName", "$1.ts")
}

func (g *gen) implementationFileName(el ecore.ENamedElement, name string) string {
	return g.ElementName(el, name, "implementationFileName", "$1Impl.ts")
}

func (g *gen) implementationTestFileName(el ecore.ENamedElement, name string) string {
	return g.ElementName(el, name, "implementationTestsFileName", "$1Impl.test.ts")
}

func (g *gen) factoryTestFileName(p ecore.EPackage) string {
	return g.ElementName(p, naming.UpperFirst(p.GetName())+"Factory", "factoryTestFileName", "$1Impl.test.ts")
}

func (g *gen) constantsFileName(p ecore.EPackage) string {
	return g.ElementName(p, g.constantsName(p), "constantsFileName", "$1.ts")
}

func (g *gen) internalFileName(p ecore.EPackage) string {
	return g.ElementName(p, g.packageName(p), "internalFileName", "internal.ts")
}

func (g *gen) indexFileName(p ecore.EPackage) string {
	return g.ElementName(p, g.packageName(p), "indexFileName", "index.ts")
}

// moduleName is a file name without its .ts extension: the name exported by
// the internal module ("./Book.js").
func (g *gen) moduleName(fileName string) string {
	name := strings.TrimSuffix(fileName, ".ts")
	if g.isExplicitFileExtension() {
		return name + ".js"
	}
	return name
}

// The names of the members of the features and operations.

// isBoolean reports whether a feature is typed by the EBoolean data type.
func isBoolean(f ecore.EStructuralFeature) bool {
	return f.GetEType() != nil && f.GetEType().GetName() == "EBoolean"
}

// getterPrefix is "is" for a boolean feature with method accessors, else "get".
func (g *gen) getterPrefix(f ecore.EStructuralFeature) string {
	if isBoolean(f) && !g.isPropertyAccessors(f) {
		return "is"
	}
	return "get"
}

func (g *gen) getterName(f ecore.EStructuralFeature) string {
	return g.ElementName(f, naming.UpperFirst(f.GetName()), "getterName", g.getterPrefix(f)+"$1")
}

func (g *gen) setterName(f ecore.EStructuralFeature) string {
	return g.ElementName(f, naming.UpperFirst(f.GetName()), "setterName", "set$1")
}

func (g *gen) unsetterName(f ecore.EStructuralFeature) string {
	return g.ElementName(f, naming.UpperFirst(f.GetName()), "unSetterName", "unSet$1")
}

func (g *gen) isSetterName(f ecore.EStructuralFeature) string {
	return g.ElementName(f, naming.UpperFirst(f.GetName()), "isSetName", "isSet$1")
}

func (g *gen) operationName(op ecore.EOperation) string {
	return g.ElementName(op, op.GetName(), "operationName", "")
}

// parameterName is the name of a parameter: "arguments" is reserved.
func parameterName(p ecore.EParameter) string {
	if p.GetName() == "arguments" {
		return "args"
	}
	return p.GetName()
}

// reservedKeywords are the words a property may not be named after (v1
// getReservedKeywords). v1 lists "null," with its comma: null is not one of
// them.
var reservedKeywords = map[string]bool{}

func init() {
	for _, keyword := range strings.Fields("break case catch class const continue debugger default delete do else enum export extends " +
		"false finally for function if import in instanceof new null, return super switch this " +
		"throw true try typeof var void while with as implements interface let package private " +
		"protected public static yield any boolean constructor declare get module require number " +
		"set string symbol type from of") {
		reservedKeywords[keyword] = true
	}
}

// propertyName is the name of the property of a feature: is<Name> for a
// boolean, <name>_ for a reserved word, else its name.
func propertyName(f ecore.EStructuralFeature) string {
	switch {
	case isBoolean(f):
		return "is" + naming.UpperFirst(f.GetName())
	case reservedKeywords[f.GetName()]:
		return f.GetName() + "_"
	}
	return f.GetName()
}

// variableName is the field of a feature: _<name>, _is<Name> for a boolean.
func variableName(f ecore.EStructuralFeature) string {
	if isBoolean(f) {
		return "_is" + naming.UpperFirst(f.GetName())
	}
	return "_" + f.GetName()
}

// literalConstantName is the constant of an enum literal: its words in upper
// case joined by "_" (an "_" literal stays "_").
func literalConstantName(l ecore.EEnumLiteral) string {
	if l.GetName() == "_" {
		return "_"
	}
	return strings.ToUpper(naming.SplitAndCombine(l.GetName(), "_"))
}

// The accessors of the features.

// accessorsTypes are the kinds of accessors of a feature: its key accessors,
// else the package property, comma separated (method, property).
func (g *gen) accessorsTypes(f ecore.EStructuralFeature) []string {
	value, ok := g.KeyValue(f, "accessors")
	if !ok {
		value, _ = g.PackageProperty("accessors")
	}
	if value == "" {
		value = "method"
	}
	return strings.FieldsFunc(value, func(r rune) bool { return r == ',' })
}

func (g *gen) hasAccessors(f ecore.EStructuralFeature, kind string) bool {
	for _, k := range g.accessorsTypes(f) {
		if k == kind {
			return true
		}
	}
	return false
}

// isPropertyAccessors reports whether a feature has get/set properties.
func (g *gen) isPropertyAccessors(f ecore.EStructuralFeature) bool {
	return g.hasAccessors(f, "property")
}

// isMethodAccessors reports whether a feature has get/set methods.
func (g *gen) isMethodAccessors(f ecore.EStructuralFeature) bool { return g.hasAccessors(f, "method") }

// isAsync reports whether a typed element has an Async variant: its key
// async, else the package property.
func (g *gen) isAsync(el ecore.ETypedElement) bool {
	if value, ok := g.KeyValue(el, "async"); ok {
		return value == "true"
	}
	return g.propertyIsTrue("async")
}
