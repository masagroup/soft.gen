package typescript

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// factoryInterfaceFile is the factory interface of a package (v1
// generateFactoryInterface): per concrete class sorted by name its creator
// and its creators from a container, then the function returning the factory
// instance, but for a shared package.
func (g *gen) factoryInterfaceFile(p ecore.EPackage) engine.File {
	s := g.newScope().declare(g.interfaceIdentifier(p, g.factoryName(p)))
	name := g.factoryName(p)
	view := sectionsView{Declaration: "export interface " + name + " extends " + s.qualify(g.ecoreIdentifier("EFactory"))}
	for _, c := range sortedByName(g.m.Package(p).AllClasses) {
		if c.IsAbstract() {
			continue
		}
		iface := s.qualify(g.classInterfaceIdentifier(c))
		section := []string{"create" + c.GetName() + "(): " + iface}
		for _, container := range sortedByName(g.m.Class(c).AllContainers) {
			containerType := s.typedElementType(container)
			section = append(section, "create"+c.GetName()+"FromContainer(eContainer: "+containerType+"): "+iface)
			if g.m.ElementIDAccessorName(c) != "" {
				section = append(section, "create"+c.GetName()+"FromContainerAndClassID(eContainer: "+containerType+", classID: number): "+iface)
			}
		}
		view.Sections = append(view.Sections, section)
	}
	if !g.isSharedPackage(p) {
		factory := g.factoryImplementationName(p)
		if g.KeyIs(p, "factory", "extension") {
			factory = g.factoryExtensionName(p)
		}
		view.Function = []string{
			"export function get" + name + "(): " + name + " {",
			indentUnit + "return " + s.qualify(g.interfaceIdentifier(p, factory)) + ".getInstance()",
			"}",
		}
	}
	return g.file(g.interfaceFileName(p, name), s, "sections.ts", view)
}

// factoryView is the data of the members of a factory implementation.
type factoryView struct {
	names
	Name, Factory, Constants string
	// Shared is whether the factory is the one of an implementation of a
	// shared package.
	Shared bool
	pkg    ecore.EPackage
	// Classes are the concrete classes and their creators.
	Classes []creatorView
	// DataTypes are the serializable data types and their conversions.
	DataTypes []dataTypeView
	// Creator is the creator of a member: a class and a container.
	Creator creatorView
	// DataType is a data type of a member.
	DataType dataTypeView
}

// creatorView is a creator of a factory.
type creatorView struct {
	// Class is the name of the class; Interface its interface; Implementation
	// the class built (its extension when it has one); ID its classifier id.
	Class, Interface, Implementation, ID string
	// Signature is the signature of a creator from a container.
	Signature string
	// Container is the type of the container; WithClassID whether the creator
	// sets the class id; IDSetter the statement setting it; AddToContainer
	// the statement adding the element to its container ("" without opposite).
	Container, IDSetter, AddToContainer string
	WithClassID                         bool
}

// dataTypeView is a data type with its conversions.
type dataTypeView struct {
	Name, ID string
	// Create and Convert are the variants of the conversions: enum, string,
	// boolean, number, byte, Uint8Array, Date, unknown.
	Create, Convert string
	// Enum is the enum name and Var its first-lower name, for an enum.
	Enum, Var string
}

// factoryImplementationFile is the implementation of the factory of a
// package (v1 generateFactoryImplementation): its instance (but with
// factory=extension), the creation of the objects by class id and per class,
// the creators from a container, the conversions of the data types; for an
// implementation of a shared package, the data types are converted by the
// factory of the shared package and the file ends with the getFactory and
// install functions.
func (g *gen) factoryImplementationFile(p ecore.EPackage) engine.File {
	s := g.newScope().declare(g.implementationIdentifier(p, g.factoryImplementationName(p)))
	n := names{s}
	shared := g.isSharedPackage(p)
	extension := !shared && g.KeyIs(p, "factory", "extension")
	name := g.factoryImplementationName(p)
	view := classView{Declaration: "export class " + name + " extends " + s.qualify(g.implementationIdentifier(g.m.Ecore(), "EFactoryExt")) +
		" implements " + s.qualify(g.interfaceIdentifier(p, g.factoryName(p)))}
	data := factoryView{names: n, Name: name, Factory: g.factoryName(p), Shared: shared, pkg: p}
	if !extension {
		view.add("factory/impl.instance", data)
	}
	var classes []creatorView
	for _, c := range g.m.Package(p).AllClasses {
		if c.IsInterface() || c.IsAbstract() {
			continue
		}
		implementation := g.classImplementationIdentifier(c)
		if g.HasExtension(c) {
			implementation = g.classExtensionIdentifier(c)
		}
		creator := creatorView{Class: c.GetName(), Interface: s.qualify(g.classInterfaceIdentifier(c)),
			Implementation: s.qualify(implementation), ID: s.classifierID(c)}
		classes = append(classes, creator)
	}
	data.Classes = classes
	view.add("factory/impl.create", data)
	i := 0
	for _, c := range g.m.Package(p).AllClasses {
		if c.IsInterface() || c.IsAbstract() {
			continue
		}
		creator := classes[i]
		i++
		member := data
		member.Creator = creator
		view.add("factory/impl.creator", member)
		for _, container := range sortedByName(g.m.Class(c).AllContainers) {
			fromContainer := creator
			fromContainer.Container = s.qualify(g.typeOf(container.GetEType()))
			fromContainer.AddToContainer = g.addToContainer(container)
			fromContainer.Signature = "create" + c.GetName() + "FromContainer(eContainer: " + fromContainer.Container + "): " + creator.Interface
			member.Creator = fromContainer
			view.add("factory/impl.fromcontainer", member)
			if accessor := g.m.ElementIDAccessorName(c); accessor != "" {
				withID := fromContainer
				withID.WithClassID = true
				withID.Signature = "create" + c.GetName() + "FromContainerAndClassID(eContainer: " + fromContainer.Container + ", classID: number): " + creator.Interface
				withID.IDSetter = "element.set" + accessor + "ID(classID)"
				if !g.isMethodAccessors(container) && g.isPropertyAccessors(container) {
					withID.IDSetter = "element." + naming.LowerFirst(accessor) + "ID = classID"
				}
				member.Creator = withID
				view.add("factory/impl.fromcontainer", member)
			}
		}
	}
	if shared {
		view.add("factory/impl.shareddatatypes", data)
	} else {
		g.addDataTypeConversions(&view, data, p, s)
	}
	file := g.file(g.implementationFileName(p, g.factoryName(p)), s, "class.ts", view)
	if shared {
		file.Body = tsSource(s, "factory/impl.shared", struct {
			Class    classView
			Data     factoryView
			Registry string
		}{view, data, s.qualify(g.ecoreIdentifier("EPackageRegistry"))})
	}
	return file
}

// addToContainer is the statement adding the created element to its
// container, through the opposite of the container reference: add to a list,
// else set (v1 generateCreateFunctionContent); "" without opposite.
func (g *gen) addToContainer(container ecore.EReference) string {
	opposite := container.GetEOpposite()
	if opposite == nil {
		return ""
	}
	if opposite.GetUpperBound() != 1 {
		return "eContainer." + g.getCall(opposite) + ".add(element)"
	}
	return "eContainer." + g.setCall(opposite, "element")
}

// getCall reads a feature: its getter, else its property (v1 generateGetCall).
func (g *gen) getCall(f ecore.EStructuralFeature) string {
	if g.isMethodAccessors(f) {
		return g.getterName(f) + "()"
	}
	return propertyName(f)
}

// setCall writes a feature: its setter, else its property (v1 generateSetCall).
func (g *gen) setCall(f ecore.EStructuralFeature, value string) string {
	if g.isMethodAccessors(f) {
		return g.setterName(f) + "(" + value + ")"
	}
	return propertyName(f) + " = " + value
}

// dataTypeFactoryFile is the factory of the data types of a shared package
// (v1 generateDataTypeFactory): it creates no object, each implementation has
// a factory of its own.
func (g *gen) dataTypeFactoryFile(p ecore.EPackage) engine.File {
	s := g.newScope().declare(g.interfaceIdentifier(p, g.factoryImplementationName(p)))
	n := names{s}
	name := g.factoryImplementationName(p)
	view := classView{Declaration: "export class " + name + " extends " + s.qualify(g.implementationIdentifier(g.m.Ecore(), "EFactoryExt"))}
	data := factoryView{names: n, Name: name, Constants: s.qualify(g.interfaceIdentifier(p, g.constantsName(p)))}
	view.add("factory/impl.nocreate", data)
	g.addDataTypeConversions(&view, data, p, s)
	return g.file(g.implementationFileName(p, g.factoryName(p)), s, "class.ts", view)
}

// addDataTypeConversions adds createFromString and convertToString and the
// conversions of the serializable data types sorted by name (v1
// generateDataTypeMethods).
func (g *gen) addDataTypeConversions(view *classView, data factoryView, p ecore.EPackage, s *scope) {
	for _, dt := range sortedByName(g.allDataTypes(p)) {
		if !dt.IsSerializable() {
			continue
		}
		v := dataTypeView{Name: dt.GetName(), ID: s.classifierID(dt)}
		if e, ok := dt.(ecore.EEnum); ok {
			v.Create, v.Convert = "enum", "enum"
			v.Enum = s.qualify(g.enumIdentifier(e))
			v.Var = naming.LowerFirst(e.GetName())
			s.qualify(g.interfaceIdentifier(p, v.Var+"FromString"))
			s.qualify(g.interfaceIdentifier(p, v.Var+"ToString"))
		} else {
			switch t := g.instanceTypeName(dt); t {
			case "string", "boolean", "number", "byte", "Uint8Array", "Date":
				v.Create, v.Convert = t, t
			default:
				v.Create, v.Convert = "unknown", "unknown"
			}
			if v.Convert == "boolean" {
				v.Convert = "number"
			}
		}
		data.DataTypes = append(data.DataTypes, v)
	}
	view.add("factory/impl.fromstring", data)
	view.add("factory/impl.tostring", data)
	for _, dt := range data.DataTypes {
		member := data
		member.DataType = dt
		view.add("factory/impl.createfromstring", member)
		view.add("factory/impl.converttostring", member)
	}
}

// PackageGetter is the call of the getter of the package (getLibraryPackage()).
func (v factoryView) PackageGetter() string {
	return v.s.qualify(v.s.g.packageGetterIdentifier(v.pkg)) + "()"
}
