package typescript

import (
	"sort"
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/naming"
)

// sectionsView is the view model of a class or interface made of sections of
// lines, separated by a blank line.
type sectionsView struct {
	Declaration string
	Sections    [][]string
	// Function is the function following the declaration, if any.
	Function []string
}

// constantsFile is the class of the constants of a package (v1
// generatePackageConstants): its name, nsURI and prefix, then for every class
// in the order of their first super types its id, its feature ids, its feature
// count, its operation ids and its operation count, then the ids of the data
// types sorted by name.
func (g *gen) constantsFile(p ecore.EPackage) engine.File {
	constant := func(comment, name, typ, value string) []string {
		lines := []string{}
		if comment != "" {
			lines = append(lines, "// "+comment)
		}
		return append(lines, "public static readonly "+name+": "+typ+" = "+value)
	}
	view := sectionsView{Declaration: "export class " + g.constantsName(p)}
	view.Sections = append(view.Sections,
		constant("", "eNAME", "string", quote(p.GetName())),
		constant("", "eNS_URI", "string", quote(p.GetNsURI())),
		constant("", "eNS_PREFIX", "string", quote(p.GetNsPrefix())))
	for _, c := range g.m.Package(p).OrderedClasses {
		ids := g.m.Class(c).IDs
		view.Sections = append(view.Sections, constant(
			ids.ClassifierIDName+" is the meta object id for the class "+c.GetName()+".",
			ids.ClassifierIDName, "number", strconv.Itoa(ids.ClassifierIDValue)))
		for _, id := range ids.FeatureIDs {
			f := id.Element.(ecore.EStructuralFeature)
			view.Sections = append(view.Sections, constant(
				id.Name+" is the feature id for the "+c.GetName()+" '"+naming.UpperFirst(propertyName(f))+"' "+g.m.FeatureKind(f)+".",
				id.Name, "number", strconv.Itoa(id.Value.Int())))
		}
		view.Sections = append(view.Sections, constant(
			ids.FeatureCountIDName+" is the number of structural features of the class "+c.GetName()+".",
			ids.FeatureCountIDName, "number", strconv.Itoa(ids.FeatureCountIDValue.Int())))
		// the operation ids are one section
		var operations []string
		for _, id := range ids.OperationIDs {
			op := id.Element.(ecore.EOperation)
			operations = append(operations, constant(
				id.Name+" is the operation id for the '"+op.GetName()+"' operation.",
				id.Name, "number", strconv.Itoa(id.Value.Int()))...)
		}
		if len(operations) > 0 {
			view.Sections = append(view.Sections, operations)
		}
		view.Sections = append(view.Sections, constant(
			ids.OperationCountName+" is the number of "+ids.OperationCountName+".",
			ids.OperationCountName, "number", strconv.Itoa(ids.OperationCountValue.Int())))
	}
	// the data type ids are one section
	var dataTypes []string
	for _, dt := range sortedByName(g.allDataTypes(p)) {
		kind := "data type"
		if _, isEnum := dt.(ecore.EEnum); isEnum {
			kind = "enum"
		}
		name := g.m.ClassifierIDName(dt)
		dataTypes = append(dataTypes, constant(
			name+" The meta object id for the "+kind+" "+dt.GetName()+".",
			name, "number", strconv.Itoa(g.m.ClassifierIDValue(dt)))...)
	}
	if len(dataTypes) > 0 {
		view.Sections = append(view.Sections, dataTypes)
	}
	return g.file(g.constantsFileName(p), g.newScope(), "sections.ts", view)
}

// packageInterfaceFile is the package interface (v1 generatePackageInterface):
// for every class sorted by name the getters of its meta object, of its
// attributes, references and operations sorted by name, then of the data
// types; and the function returning the package instance.
func (g *gen) packageInterfaceFile(p ecore.EPackage) engine.File {
	s := g.newScope()
	name := g.packageInterfaceName(p)
	view := sectionsView{Declaration: "export interface " + name + " extends " + s.qualify(g.ecoreIdentifier("EPackage"))}
	getter := func(el ecore.ENamedElement) []string {
		return []string{
			"// Returns the meta object for the " + naming.UpperFirst(el.GetName()),
			g.ReflectiveGetterName(el) + "(): " + s.qualify(g.ecoreIdentifier(el.EClass().GetName())),
		}
	}
	getters := func(elements []ecore.ENamedElement) []string {
		var lines []string
		for _, el := range sortedByName(elements) {
			lines = append(lines, getter(el)...)
		}
		return lines
	}
	addSection := func(lines []string) {
		if len(lines) > 0 {
			view.Sections = append(view.Sections, lines)
		}
	}
	for _, c := range sortedByName(g.m.Package(p).AllClasses) {
		addSection(getter(c))
		addSection(getters(namedElements(c.GetEAttributes())))
		addSection(getters(namedElements(c.GetEReferences())))
		addSection(getters(namedElements(c.GetEOperations())))
	}
	var dataTypes []ecore.ENamedElement
	for _, dt := range g.allDataTypes(p) {
		dataTypes = append(dataTypes, dt)
	}
	addSection(getters(dataTypes))
	implementation := g.implementationIdentifier(p, g.packageImplementationName(p))
	if g.isSharedPackage(p) {
		implementation = g.interfaceIdentifier(p, g.packageImplementationName(p))
	}
	view.Function = []string{
		"export function get" + name + "(): " + name + " {",
		indentUnit + "return " + s.qualify(implementation) + ".getInstance()",
		"}",
	}
	return g.file(g.interfaceFileName(p, g.packageName(p)), s, "sections.ts", view)
}

// namedElements returns the elements of a list.
func namedElements(list ecore.EList) []ecore.ENamedElement {
	var elements []ecore.ENamedElement
	for it := list.Iterator(); it.HasNext(); {
		elements = append(elements, it.Next().(ecore.ENamedElement))
	}
	return elements
}

// sortedByName returns the elements sorted by name, keeping the order of
// equal names (OCL sortedBy).
func sortedByName[T ecore.ENamedElement](elements []T) []T {
	sorted := append([]T(nil), elements...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].GetName() < sorted[j].GetName() })
	return sorted
}

// allDataTypes are the data types of a package, enums included (v1
// eAllContents(EDataType)).
func (g *gen) allDataTypes(p ecore.EPackage) []ecore.EDataType {
	info := g.m.Package(p)
	dataTypes := make([]ecore.EDataType, 0, len(info.AllEnums)+len(info.AllDataTypes))
	for _, e := range info.AllEnums {
		dataTypes = append(dataTypes, e)
	}
	return append(dataTypes, info.AllDataTypes...)
}
