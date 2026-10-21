package golang

import (
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// goKeywords are the reserved words a feature name must not be used as a variable (v1 getReservedKeywords).
var goKeywords = map[string]bool{
	"break": true, "default": true, "func": true, "interface": true, "select": true,
	"case": true, "defer": true, "go": true, "map": true, "struct": true,
	"chan": true, "else": true, "goto": true, "package": true, "switch": true,
	"const": true, "fallthrough": true, "if": true, "range": true, "type": true,
	"continue": true, "for": true, "import": true, "return": true, "var": true,
}

// factoryName is the name of the factory interface of a package: rule factoryName, default <Name>Factory.
func (g *gen) factoryName(p ecore.EPackage) string {
	return naming.UpperFirst(g.ElementName(p, p.GetName(), "factoryName", "$1Factory"))
}

// factoryImplementationName is the name of the factory implementation of a package.
func (g *gen) factoryImplementationName(p ecore.EPackage) string {
	return g.implementationNameOf(p, g.factoryName(p))
}

// factoryInternalName is the name of the internal factory interface of a package.
func (g *gen) factoryInternalName(p ecore.EPackage) string {
	return p.GetName() + "FactoryInternal"
}

// factoryExtensionName is the name of the hand-written factory extension of a package.
func (g *gen) factoryExtensionName(p ecore.EPackage) string {
	return g.extensionNameOf(p, g.factoryName(p))
}

// packageName is the name of the package interface: rule packageName, default <Name>Package.
func (g *gen) packageName(p ecore.EPackage) string {
	return naming.UpperFirst(g.ElementName(p, p.GetName(), "packageName", "$1Package"))
}

// interfaceName is the name of the interface of a class: rule interfaceName, default the class name.
func (g *gen) interfaceName(c ecore.EClass) string {
	return naming.UpperFirst(g.ElementName(c, c.GetName(), "interfaceName", ""))
}

// packageInterfaceName is the name of the package interface after the interfaceName rule.
func (g *gen) packageInterfaceName(p ecore.EPackage) string {
	return naming.UpperFirst(g.ElementName(p, g.packageName(p), "interfaceName", ""))
}

// implementationName is the name of the implementation of a class: rule
// implementationName, default <Name>Impl, lower-cased first unless exported.
func (g *gen) implementationName(c ecore.EClass) string {
	return g.implementationNameOf(c, c.GetName())
}

// packageImplementationName is the name of the package implementation.
func (g *gen) packageImplementationName(p ecore.EPackage) string {
	return g.implementationNameOf(p, g.packageName(p))
}

// implementationNameOf applies the implementationName rule and the export case to a name.
func (g *gen) implementationNameOf(el ecore.ENamedElement, name string) string {
	return g.exportCase(el, g.ElementName(el, name, "implementationName", "$1Impl"))
}

// extensionName is the name of the hand-written extension of a class: rule extensionName, default <Name>Ext.
func (g *gen) extensionName(c ecore.EClass) string {
	return g.extensionNameOf(c, c.GetName())
}

// extensionNameOf applies the extensionName rule and the export case to a name.
func (g *gen) extensionNameOf(el ecore.ENamedElement, name string) string {
	return g.exportCase(el, g.ElementName(el, name, "extensionName", "$1Ext"))
}

// exportCase upper-cases the first letter of an implementation name when the
// element is exported, lower-cases it otherwise.
func (g *gen) exportCase(el ecore.ENamedElement, name string) string {
	if g.isExported(el) {
		return naming.UpperFirst(name)
	}
	return naming.LowerFirst(name)
}

// mockName is the name of the mock of a class: rule mockName, default Mock<Name>.
func (g *gen) mockName(c ecore.EClass) string {
	return g.ElementName(c, c.GetName(), "mockName", "Mock$1")
}

// interfaceFileName is the file of an interface: rule interfaceFileName, default <name>.go.
func (g *gen) interfaceFileName(el ecore.ENamedElement, name string) string {
	return strings.ToLower(g.ElementName(el, name, "interfaceFileName", "$1.go"))
}

// implementationFileName is the file of an implementation: rule implementationFileName, default <name>_impl.go.
func (g *gen) implementationFileName(el ecore.ENamedElement, name string) string {
	return strings.ToLower(g.ElementName(el, name, "implementationFileName", "$1_impl.go"))
}

// mockFileName is the file of a mock: rule mockFileName, default <name>_mock.go.
func (g *gen) mockFileName(el ecore.ENamedElement, name string) string {
	return strings.ToLower(g.ElementName(el, name, "mockFileName", "$1_mock.go"))
}

// mockTestFileName is the file of the tests of a mock: rule mockTestFileName, default <name>_mock_test.go.
func (g *gen) mockTestFileName(el ecore.ENamedElement, name string) string {
	return strings.ToLower(g.ElementName(el, name, "mockTestFileName", "$1_mock_test.go"))
}

// implementationTestFileName is the file of the tests of an implementation: rule implementationTestFileName, default <name>_impl_test.go.
func (g *gen) implementationTestFileName(el ecore.ENamedElement, name string) string {
	return strings.ToLower(g.ElementName(el, name, "implementationTestFileName", "$1_impl_test.go"))
}

// factoryTestFileName is the file of the tests of the factory: rule factoryTestFileName on "<package>factory".
func (g *gen) factoryTestFileName(p ecore.EPackage) string {
	return strings.ToLower(g.ElementName(p, p.GetName()+"factory", "factoryTestFileName", "$1_impl_test.go"))
}

// packageTestFileName is the file of the tests of the package: rule packageTestFileName on "<package>package".
func (g *gen) packageTestFileName(p ecore.EPackage) string {
	return strings.ToLower(g.ElementName(p, p.GetName()+"package", "packageTestFileName", "$1_impl_test.go"))
}

// reflectiveGetterName is the getter of a model element in the package
// interface: the shared one, upper-cased (GetEClassifierClass).
func (g *gen) reflectiveGetterName(el ecore.ENamedElement) string {
	return naming.UpperFirst(g.ReflectiveGetterName(el))
}

// accessorName is the base of the accessor names of a feature: its name,
// upper-cased first, prefixed by Typed in a map entry class (v1 getAccessorName).
func (g *gen) accessorName(f ecore.EStructuralFeature) string {
	name := naming.UpperFirst(f.GetName())
	if g.m.Class(f.GetEContainingClass()).IsMapEntry {
		return "Typed" + name
	}
	return name
}

// getterName is the getter of a feature: rule getterName, default Get<Name>, or Is<Name> for an EBoolean.
func (g *gen) getterName(f ecore.EStructuralFeature) string {
	prefix := "Get"
	if t := f.GetEType(); t != nil && t.GetName() == "EBoolean" {
		prefix = "Is"
	}
	return g.ElementName(f, g.accessorName(f), "getterName", prefix+"$1")
}

// setterName is the setter of a feature: rule setterName, default Set<Name>.
func (g *gen) setterName(f ecore.EStructuralFeature) string {
	return g.ElementName(f, g.accessorName(f), "setterName", "Set$1")
}

// unsetterName is the unsetter of a feature: rule unSetterName, default Unset<Name>.
func (g *gen) unsetterName(f ecore.EStructuralFeature) string {
	return g.ElementName(f, g.accessorName(f), "unSetterName", "Unset$1")
}

// operationName is the method of an operation: rule operationName, default the name, upper-cased first.
func (g *gen) operationName(op ecore.EOperation) string {
	return naming.UpperFirst(g.ElementName(op, op.GetName(), "operationName", ""))
}

// isConstantsQualified reports whether the constants of an enum are prefixed
// by the enum name (v1 isConstantsQualified: another enum of the package
// shares literal names; v1 compares the enum with itself, so any other enum
// qualifies them, kept for fidelity).
func (g *gen) isConstantsQualified(e ecore.EEnum) bool {
	if e.GetELiterals().Empty() {
		return false
	}
	for _, other := range g.m.Package(e.GetEPackage()).AllEnums {
		if other != e {
			return true
		}
	}
	return false
}

// constantsName is the constant of an enum literal: its name in UPPER_SNAKE,
// prefixed by ENUM__ when the constants are qualified (v1 getConstantsName).
func (g *gen) constantsName(literal ecore.EEnumLiteral) string {
	name := naming.UpperSnake(literal.GetName())
	if g.isConstantsQualified(literal.GetEEnum()) {
		return naming.UpperSnake(literal.GetEEnum().GetName()) + "__" + name
	}
	return name
}

// receiverName is the receiver of the methods of a class: its first letter, lower-cased.
func receiverName(c ecore.EClass) string {
	return strings.ToLower(c.GetName()[:1])
}

// variableName is the field of a feature in its implementation: is<Name> for
// a bool, <name>_ for a keyword, else the name (v1 getVariableName).
func (g *gen) variableName(f ecore.EStructuralFeature) string {
	if g.typeOfClassifier(f.GetEType()) == "bool" {
		return "is" + naming.UpperFirst(f.GetName())
	}
	return naming.AvoidKeywords(f.GetName(), goKeywords, "_")
}

// constructorName is the constructor of a qualified implementation type:
// "<path>.new<Name>" (v1 getConstructorName).
func constructorName(qualified string) string {
	path := identifierPackagePath(qualified)
	name := "new" + naming.UpperFirst(simpleIdentifier(qualified))
	if path == "" {
		return name
	}
	return qualifiedIdentifier(path, name)
}

// interfaceExtends lists the qualified interfaces a class interface embeds (v1
// getInterfaceExtends): EObject and EMapEntry for a map entry, ENotifier for
// EObject itself, the rootExtendsInterface property or ecore.EObject for a
// root class, else the interfaces of its super types.
func (g *gen) interfaceExtends(c ecore.EClass) []string {
	e := g.m.Ecore()
	switch {
	case g.m.Class(c).IsMapEntry:
		return []string{g.qualifiedInterfaceIdentifier(e, "EObject"), g.qualifiedInterfaceIdentifier(e, "EMapEntry")}
	case c.GetESuperTypes().Empty():
		if c.GetName() == "EObject" {
			return []string{g.qualifiedInterfaceIdentifier(e, "ENotifier")}
		}
		if root, ok := g.ExplicitPackageProperty("rootExtendsInterface"); ok {
			return []string{root}
		}
		return []string{g.qualifiedInterfaceIdentifier(e, "EObject")}
	}
	var extends []string
	for it := c.GetESuperTypes().Iterator(); it.HasNext(); {
		extends = append(extends, g.classInterface(it.Next().(ecore.EClass)))
	}
	return extends
}

// classExtends is the qualified struct a class implementation embeds (v1
// getClassExtends): the extension or implementation of the extended class,
// else BasicEObjectImpl for EObject itself, the rootExtendsClass property,
// ReflectiveEObjectImpl in reflective delegation or EObjectImpl.
func (g *gen) classExtends(c ecore.EClass) string {
	e := g.m.Ecore()
	extends := g.m.Class(c).ExtendsClass
	if extends == nil {
		switch {
		case c.GetName() == "EObject":
			return g.qualifiedImplementationIdentifier(e, "BasicEObjectImpl")
		}
		if root, ok := g.ExplicitPackageProperty("rootExtendsClass"); ok {
			return root
		}
		if g.IsReflectiveDelegation() {
			return g.qualifiedImplementationIdentifier(e, "ReflectiveEObjectImpl")
		}
		return g.qualifiedImplementationIdentifier(e, "EObjectImpl")
	}
	if g.HasExtension(extends) {
		return g.classExtension(extends)
	}
	return g.classImplementation(extends)
}

// qualifiedInterfaceIdentifier qualifies an identifier of the interface package of a package.
func (g *gen) qualifiedInterfaceIdentifier(p ecore.EPackage, identifier string) string {
	return qualifiedIdentifier(g.interfacePackagePath(p), identifier)
}

// qualifiedImplementationIdentifier qualifies an identifier of the implementation package of a package.
func (g *gen) qualifiedImplementationIdentifier(p ecore.EPackage, identifier string) string {
	return qualifiedIdentifier(g.implementationPackagePath(p), identifier)
}

// qualifiedMockIdentifier qualifies an identifier of the mock package of a package.
func (g *gen) qualifiedMockIdentifier(p ecore.EPackage, identifier string) string {
	return qualifiedIdentifier(g.mockPackagePath(p), identifier)
}

// classInterface is the qualified interface of a class.
func (g *gen) classInterface(c ecore.EClass) string {
	return g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.interfaceName(c))
}

// classImplementation is the qualified implementation of a class.
func (g *gen) classImplementation(c ecore.EClass) string {
	return g.qualifiedImplementationIdentifier(c.GetEPackage(), g.implementationName(c))
}

// classExtension is the qualified hand-written extension of a class.
func (g *gen) classExtension(c ecore.EClass) string {
	return g.qualifiedImplementationIdentifier(c.GetEPackage(), g.extensionName(c))
}

// classMock is the qualified mock of a class.
func (g *gen) classMock(c ecore.EClass) string {
	return g.qualifiedMockIdentifier(c.GetEPackage(), g.mockName(c))
}

// packageInstance is the qualified GetPackage() call of a package (v1 getPackageInstanceIdentifier).
func (g *gen) packageInstance(p ecore.EPackage) string {
	return qualifiedIdentifier(g.packageInstancePath(p), "GetPackage()")
}

// metaType is the qualified Ecore interface of a model element (EClass,
// EAttribute, EReference, EOperation, EDataType, EEnum) (v1 getMetaType).
func (g *gen) metaType(el ecore.ENamedElement) string {
	return g.qualifiedInterfaceIdentifier(g.m.Ecore(), el.EClass().GetName())
}
