package golang

import (
	"sort"
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/model"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// packageInterfaceView is the view model of the package interface file (template package.go).
type packageInterfaceView struct {
	// Name, NsURI and NsPrefix are the NAME, NS_URI and NS_PREFIX constants.
	Name, NsURI, NsPrefix string
	// InterfaceName is the package interface.
	InterfaceName string
	// EPackage is the embedded ecore.EPackage interface.
	EPackage string
	// ConstantGroups lists the classifier, feature and operation ids and
	// counts, by group: a blank line separates the groups.
	ConstantGroups [][]constantView
	// GetterGroups lists the getters of the meta objects, by group: a blank
	// line separates the groups.
	GetterGroups [][]getterView
	// Instance is the GetPackage singleton, nil when the package instance is
	// generated with the implementation.
	Instance *packageInstanceView
}

// constantView is one constant of the package.
type constantView struct {
	Comment, Name, Value string
}

// getterView is one getter of a meta object.
type getterView struct {
	// Comment is the element name after "Returns the meta object for the".
	Comment string
	// Name is the getter.
	Name string
	// Type is the Ecore interface returned.
	Type string
}

// packageInstanceView is the GetPackage singleton (template package/instance).
type packageInstanceView struct {
	// Name is the package name of the comment.
	Name string
	// Type is the package interface.
	Type string
	// Once is sync.Once.
	Once string
	// Constructor is the constructor of the package implementation.
	Constructor string
}

// packageInterfaceFile is the package interface: <interface dir>/<package>.go,
// with the GetPackage singleton when the implementation is in the same package.
func (g *gen) packageInterfaceFile(p ecore.EPackage) engine.File {
	packagePath := g.interfacePackagePath(p)
	scope := newFileScope(packagePath)
	view := packageInterfaceView{
		Name:           p.GetName(),
		NsURI:          p.GetNsURI(),
		NsPrefix:       p.GetNsPrefix(),
		InterfaceName:  g.packageInterfaceName(p),
		EPackage:       scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), "EPackage")),
		ConstantGroups: g.packageConstantGroups(p, scope),
		GetterGroups:   g.packageGetterGroups(p, scope),
	}
	if g.implementationPackagePath(p) == packagePath {
		view.Instance = g.packageInstanceView(p, scope)
	}
	return engine.File{
		Path:   filePath(g.interfaceOutputDirectory(p), g.interfaceFileName(p, g.packageName(p))),
		Body:   goSource(scope, "package.go", view),
		Header: true,
	}
}

// packageInstanceView builds the GetPackage singleton of a package.
func (g *gen) packageInstanceView(p ecore.EPackage, scope *fileScope) *packageInstanceView {
	return &packageInstanceView{
		Name:        p.GetName(),
		Type:        scope.qualify(g.qualifiedInterfaceIdentifier(p, g.packageInterfaceName(p))),
		Once:        scope.qualify("sync/sync.Once"),
		Constructor: constructorName(naming.UpperFirst(g.packageImplementationName(p))),
	}
}

// packageConstantGroups lists the constants of a package: for each class in
// the order of the class hierarchy, its id, the ids of all its features, its
// feature count, the ids of all its operations, its operation count; then the
// ids of the data types (enums included) sorted by name. Each id of a class or
// a feature, each count, the operation ids together and the data type ids
// together form a group. A constant expressed from the constant of a base class
// of another package qualifies it.
func (g *gen) packageConstantGroups(p ecore.EPackage, scope *fileScope) [][]constantView {
	var groups [][]constantView
	one := func(c constantView) { groups = append(groups, []constantView{c}) }
	for _, c := range g.m.Package(p).OrderedClasses {
		one(constantView{
			Comment: g.m.ClassifierIDName(c) + " is the meta object id for the class " + c.GetName() + ".",
			Name:    g.m.ClassifierIDName(c),
			Value:   strconv.Itoa(g.m.ClassifierIDValue(c)),
		})
		for it := c.GetEAllStructuralFeatures().Iterator(); it.HasNext(); {
			f := it.Next().(ecore.EStructuralFeature)
			one(constantView{
				Comment: g.m.FeatureIDName(c, f) + " is the feature id for the " + c.GetName() + " '" + naming.UpperFirst(f.GetName()) + "' " + g.m.FeatureKind(f) + ".",
				Name:    g.m.FeatureIDName(c, f),
				Value:   g.constantValue(p, scope, g.m.FeatureIDValue(c, f)),
			})
		}
		one(constantView{
			Comment: g.m.FeatureCountIDName(c) + " is the number of structural features of the class " + c.GetName() + ".",
			Name:    g.m.FeatureCountIDName(c),
			Value:   g.constantValue(p, scope, g.m.FeatureCountIDValue(c)),
		})
		var operations []constantView
		for it := c.GetEAllOperations().Iterator(); it.HasNext(); {
			op := it.Next().(ecore.EOperation)
			operations = append(operations, constantView{
				Comment: g.m.OperationIDName(c, op) + " is the operation id for the '" + op.GetName() + "' operation.",
				Name:    g.m.OperationIDName(c, op),
				Value:   g.constantValue(p, scope, g.m.OperationIDValue(c, op)),
			})
		}
		if len(operations) > 0 {
			groups = append(groups, operations)
		}
		one(constantView{
			Comment: g.m.OperationCountIDName(c) + " is the number of " + g.m.OperationCountIDName(c),
			Name:    g.m.OperationCountIDName(c),
			Value:   g.constantValue(p, scope, g.m.OperationCountIDValue(c)),
		})
	}
	var dataTypes []constantView
	for _, dt := range g.allDataTypesByName(p) {
		kind := "data type"
		if _, ok := dt.(ecore.EEnum); ok {
			kind = "enum"
		}
		dataTypes = append(dataTypes, constantView{
			Comment: g.m.ClassifierIDName(dt) + " The meta object id for the " + kind + " " + dt.GetName() + ".",
			Name:    g.m.ClassifierIDName(dt),
			Value:   strconv.Itoa(g.m.ClassifierIDValue(dt)),
		})
	}
	if len(dataTypes) > 0 {
		groups = append(groups, dataTypes)
	}
	return groups
}

// constantValue is the value of a constant of the package p: its base
// constant, declared by the package interface of the base class, is
// qualified when that is another package (a class extending a class of
// another model).
func (g *gen) constantValue(p ecore.EPackage, scope *fileScope, e model.Expr) string {
	if !e.HasBase || e.Package == nil || e.Package == p {
		return e.String()
	}
	base := model.Expr{Value: e.Value, Base: scope.qualify(g.qualifiedInterfaceIdentifier(e.Package, e.Base)), Offset: e.Offset, HasBase: true, HasOffset: e.HasOffset}
	return base.String()
}

// allDataTypesByName lists the data types of a package, enums included, sorted by name.
func (g *gen) allDataTypesByName(p ecore.EPackage) []ecore.EDataType {
	info := g.m.Package(p)
	all := append([]ecore.EDataType(nil), info.AllDataTypes...)
	for _, e := range info.AllEnums {
		all = append(all, e)
	}
	return sortedByName(all)
}

// packageGetterGroups lists the getters of the meta objects of a package: for
// each class sorted by name, the class, then its attributes, references and
// operations, each sorted by name; then the data types sorted by name.
func (g *gen) packageGetterGroups(p ecore.EPackage, scope *fileScope) [][]getterView {
	var groups [][]getterView
	group := func(elements []ecore.ENamedElement) {
		if len(elements) == 0 {
			return
		}
		var getters []getterView
		for _, el := range elements {
			getters = append(getters, getterView{
				Comment: naming.UpperFirst(el.GetName()),
				Name:    g.reflectiveGetterName(el),
				Type:    scope.qualify(g.metaType(el)),
			})
		}
		groups = append(groups, getters)
	}
	for _, c := range sortedByName(g.m.Package(p).AllClasses) {
		group([]ecore.ENamedElement{c})
		group(sortedByName(namedElements(c.GetEAttributes())))
		group(sortedByName(namedElements(c.GetEReferences())))
		group(sortedByName(namedElements(c.GetEOperations())))
	}
	group(asNamedElements(g.allDataTypesByName(p)))
	return groups
}

// namedElements converts an EList of named elements to a slice.
func namedElements(list ecore.EList) []ecore.ENamedElement {
	var elements []ecore.ENamedElement
	for it := list.Iterator(); it.HasNext(); {
		elements = append(elements, it.Next().(ecore.ENamedElement))
	}
	return elements
}

// sortedByName returns a copy of the elements sorted by name (stable).
func sortedByName[T ecore.ENamedElement](elements []T) []T {
	sorted := append([]T(nil), elements...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].GetName() < sorted[j].GetName() })
	return sorted
}

// asNamedElements converts a slice of named elements to []ecore.ENamedElement.
func asNamedElements[T ecore.ENamedElement](elements []T) []ecore.ENamedElement {
	named := make([]ecore.ENamedElement, len(elements))
	for i, el := range elements {
		named[i] = el
	}
	return named
}
