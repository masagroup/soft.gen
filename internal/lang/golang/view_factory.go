package golang

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// factoryInterfaceView is the view model of the factory interface file
// (template factory.go, v1 generateFactoryInterface).
type factoryInterfaceView struct {
	// Name is the factory interface.
	Name string
	// EFactory is the embedded ecore.EFactory interface.
	EFactory string
	// Creates lists the creation methods of the concrete classes.
	Creates []factoryCreateView
	// Instance is the GetFactory singleton, nil when the implementation is
	// in another package (it then holds the singleton).
	Instance *factoryInstanceView
}

// factoryInstanceView is the GetFactory singleton (template factory/instance,
// v1 generateFactoryInstance).
type factoryInstanceView struct {
	// Name is the package name of the comment.
	Name string
	// Type is the factory interface.
	Type string
	// Once is sync.Once.
	Once string
	// Constructor is the constructor of the factory implementation (or of
	// its hand-written extension, annotation factory=extension).
	Constructor string
	// SharedPackage is the GetPackage() call of the package shared by the
	// implementations, given to the factory; "" for an unshared package.
	SharedPackage string
}

// factoryCreateView is one creation method of a factory: Create<X>,
// Create<X>FromContainer or Create<X>FromContainerAndClassID.
type factoryCreateView struct {
	// Name is the method.
	Name string
	// Params is the parameter list: "", "eContainer T" or "eContainer T, classID int".
	Params string
	// Type is the interface returned.
	Type string
	// Kind selects the body of the implementation: "new" returns the
	// constructor, "container" creates the element and adds it to the container.
	Kind string
	// Receiver is the receiver of the implementation method ("" in the interface).
	Receiver string
	// Constructor is the constructor of the implementation (kind new).
	Constructor string
	// Create is the plain creation method the container form calls (kind container).
	Create string
	// SetID is the id setter called on the element (Set<Kind>ID) in the
	// FromContainerAndClassID form; "" otherwise.
	SetID string
	// Opposite is the accessor of the opposite of the container reference
	// the element is added to (its getter) or set as (its setter); "" when
	// the reference has no opposite.
	Opposite string
	// OppositeMany is whether the opposite is many-valued (Add instead of Set).
	OppositeMany bool
}

// factoryCaseView is one case of the Create switch of the factory implementation.
type factoryCaseView struct {
	// ID is the classifier id constant of the class.
	ID string
	// Create is the creation method of the class.
	Create string
}

// factoryImplView is the view model of a factory implementation file
// (template factory_impl.go, v1 generateFactory): the factory of a package,
// or the factory of the data types of a shared package (DataTypesOnly).
type factoryImplView struct {
	// Instance is the GetFactory singleton, nil when the interface file holds it.
	Instance *factoryInstanceView
	// Internal is the internal interface of the conversions of the data
	// types; nil when the conversions are delegated to the shared package.
	Internal *factoryInternalView
	// Name is the implementation struct.
	Name string
	// Constructor is new<Name>.
	Constructor string
	// Receiver is the receiver of the methods: the struct name, lower-cased first.
	Receiver string
	// EFactoryExt is the embedded ecore.EFactoryExt.
	EFactoryExt string
	// Install is the Install function of the implementation of a shared
	// package; nil otherwise.
	Install *factoryInstallView
	// CreateKind selects the creation section: "classes" (AsEFactory, the
	// Create switch and the creation methods) or "panic" (the data type
	// factory creates no object).
	CreateKind string
	// Interface is the factory interface (kind classes).
	Interface string
	// Cases lists the cases of the Create switch (kind classes).
	Cases []factoryCaseView
	// Creates lists the creation methods (kind classes).
	Creates []factoryCreateView
	// NsURI is the NS_URI constant of the package (kind panic).
	NsURI string
	// ConversionKind selects the conversion section: "switch" (the
	// CreateFromString and ConvertToString switches and the conversion of
	// every data type) or "delegated" (to the factory of the shared package).
	ConversionKind string
	// PackageInstance is the GetPackage() call of the shared package (kind delegated).
	PackageInstance string
	// Conversions lists the conversions of the serializable data types,
	// sorted by name (kind switch).
	Conversions []conversionView
	// EClass, EObject and EDataType are the Ecore interfaces.
	EClass, EObject, EDataType string
	// Any is the empty interface type.
	Any string
	// Itoa is strconv.Itoa.
	Itoa string
}

// factoryInternalView is the internal interface of the conversions (template factory/internal).
type factoryInternalView struct {
	// Name is the interface.
	Name string
	// DataTypes lists the names of the serializable data types, sorted by name.
	DataTypes []string
	// EDataType is the Ecore interface.
	EDataType string
	// Any is the empty interface type.
	Any string
}

// factoryInstallView is the Install function registering the implementation
// of a shared package.
type factoryInstallView struct {
	// Registry is ecore.EPackageRegistry.
	Registry string
	// Package is the GetPackage() call of the shared package.
	Package string
}

// conversionView is the conversion of one data type to and from a string
// (templates factory/createfromstring.<kind> and factory/converttostring.<kind>).
type conversionView struct {
	// Receiver and Impl are the receiver and the struct of the methods.
	Receiver, Impl string
	// Name is the data type name: create<Name>FromString, convert<Name>ToString.
	Name string
	// ID is the classifier id constant of the data type.
	ID string
	// Kind is the conversion kind, by Go type: enum, string, bool, int,
	// int64, int32, int16, int8, byte, bytes, float64, float32, reflect, time
	// or unsupported.
	Kind string
	// EDataType is the Ecore interface.
	EDataType string
	// Any is the empty interface type.
	Any string
	// Type is the Go type of the value: the enum type, or *time.Time.
	Type string
	// Var is the variable holding an enum value: its name, lower-cased first.
	Var string
	// Parse is the standard function the creation uses (strconv.ParseBool,
	// time.Parse...) or the value it references (reflect.Ptr).
	Parse string
	// Format is the standard function the conversion uses (strconv.FormatBool...).
	Format string
}

// factoryInterfaceFile is the factory interface: <interface dir>/<package>factory.go,
// with the GetFactory singleton when the implementation is in the same
// package (v1 generateFactoryInterface).
func (g *gen) factoryInterfaceFile(p ecore.EPackage) []engine.File {
	packagePath := g.interfacePackagePath(p)
	scope := newFileScope(packagePath)
	view := factoryInterfaceView{
		Name:     g.factoryName(p),
		EFactory: scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), "EFactory")),
		Creates:  g.factoryCreates(p, scope, false),
	}
	if g.implementationPackagePath(p) == packagePath {
		view.Instance = g.factoryInstanceView(p, scope)
	}
	return []engine.File{{
		Path:   filePath(g.interfaceOutputDirectory(p), g.interfaceFileName(p, g.factoryName(p))),
		Body:   goSource(scope, "factory.go", view),
		Header: true,
	}}
}

// factoryImplementationFile is the factory implementation:
// <implementation dir>/<package>factory_impl.go (v1 generateFactoryImplementation).
func (g *gen) factoryImplementationFile(p ecore.EPackage) []engine.File {
	return []engine.File{g.factoryFile(p, false)}
}

// dataTypeFactoryFile is the factory of the data types of a shared package,
// generated with the interfaces: it converts the values of the data types,
// the same for all implementations, and creates no object (v1
// generateDataTypeFactory).
func (g *gen) dataTypeFactoryFile(p ecore.EPackage) []engine.File {
	return []engine.File{g.factoryFile(p, true)}
}

// factoryFile builds a factory implementation file (v1 generateFactory): in
// the interface package for the data type factory of a shared package, in the
// implementation package otherwise.
func (g *gen) factoryFile(p ecore.EPackage, dataTypesOnly bool) engine.File {
	packagePath, outputDirectory := g.implementationPackagePath(p), g.implementationOutputDirectory(p)
	if dataTypesOnly {
		packagePath, outputDirectory = g.interfacePackagePath(p), g.interfaceOutputDirectory(p)
	}
	scope := newFileScope(packagePath)
	e := g.m.Ecore()
	shared := g.isSharedPackage(p)
	delegatesDataTypes := shared && !dataTypesOnly
	name := g.factoryImplementationName(p)
	view := factoryImplView{
		Name:        name,
		Constructor: constructorName(name),
		Receiver:    g.factoryReceiverName(p),
		EFactoryExt: scope.qualify(g.qualifiedInterfaceIdentifier(e, "EFactoryExt")),
		EClass:      scope.qualify(g.qualifiedInterfaceIdentifier(e, "EClass")),
		EObject:     scope.qualify(g.qualifiedInterfaceIdentifier(e, "EObject")),
		EDataType:   scope.qualify(g.qualifiedInterfaceIdentifier(e, "EDataType")),
		Any:         g.anyType(),
		Itoa:        scope.qualify("strconv/strconv.Itoa"),
	}
	if g.interfacePackagePath(p) != packagePath {
		view.Instance = g.factoryInstanceView(p, scope)
	}
	if delegatesDataTypes {
		view.ConversionKind = "delegated"
		view.PackageInstance = scope.qualify(g.packageInstance(p))
	} else {
		view.ConversionKind = "switch"
		view.Conversions = g.conversions(p, scope, view)
		view.Internal = &factoryInternalView{Name: g.factoryInternalName(p), EDataType: view.EDataType, Any: view.Any}
		for _, c := range view.Conversions {
			view.Internal.DataTypes = append(view.Internal.DataTypes, c.Name)
		}
	}
	if shared && !dataTypesOnly {
		view.Install = &factoryInstallView{
			Registry: scope.qualify(g.qualifiedInterfaceIdentifier(e, "EPackageRegistry")),
			Package:  scope.qualify(g.packageInstance(p)),
		}
	}
	if dataTypesOnly {
		view.CreateKind = "panic"
		view.NsURI = scope.qualify(g.qualifiedInterfaceIdentifier(p, "NS_URI"))
	} else {
		view.CreateKind = "classes"
		view.Interface = scope.qualify(g.qualifiedInterfaceIdentifier(p, g.factoryName(p)))
		view.Creates = g.factoryCreates(p, scope, true)
		for i := range view.Creates {
			view.Creates[i].Receiver = view.Receiver
		}
		for _, c := range g.concreteClassesByName(p) {
			view.Cases = append(view.Cases, factoryCaseView{
				ID:     scope.qualify(g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.m.ClassifierIDName(c))),
				Create: "Create" + c.GetName(),
			})
		}
	}
	return engine.File{
		Path:   filePath(outputDirectory, g.implementationFileName(p, g.factoryName(p))),
		Body:   goSource(scope, "factory_impl.go", view),
		Header: true,
	}
}

// factoryReceiverName is the receiver of the methods of the factory
// implementation: its name, lower-cased first (v1 getFactoryReceiverName).
func (g *gen) factoryReceiverName(p ecore.EPackage) string {
	return naming.LowerFirst(g.factoryImplementationName(p))
}

// factoryInstanceView builds the GetFactory singleton of a package: the
// constructor of the factory extension when the annotation factory=extension
// is set on an unshared package, else the factory implementation; a shared
// package is given to the factory (v1 generateFactoryInstance).
func (g *gen) factoryInstanceView(p ecore.EPackage, scope *fileScope) *factoryInstanceView {
	shared := g.isSharedPackage(p)
	view := &factoryInstanceView{
		Name:        p.GetName(),
		Type:        scope.qualify(g.qualifiedInterfaceIdentifier(p, g.factoryName(p))),
		Once:        scope.qualify("sync/sync.Once"),
		Constructor: constructorName(g.factoryImplementationName(p)),
	}
	if g.KeyIs(p, "factory", "extension") && !shared {
		view.Constructor = constructorName(g.factoryExtensionName(p))
	}
	if shared {
		view.SharedPackage = scope.qualify(g.packageInstance(p))
	}
	return view
}

// concreteClassesByName lists the non-abstract classes of a package (sub
// packages included), sorted by name.
func (g *gen) concreteClassesByName(p ecore.EPackage) []ecore.EClass {
	var concrete []ecore.EClass
	for _, c := range g.m.Package(p).AllClasses {
		if !c.IsAbstract() {
			concrete = append(concrete, c)
		}
	}
	return sortedByName(concrete)
}

// factoryCreates lists the creation methods of the concrete classes sorted
// by name: Create<X>, then for each container reference sorted by name
// Create<X>FromContainer and, for the ecore classes whose instances have an
// id, Create<X>FromContainerAndClassID (v1 generateFactoryInterface,
// generateFactoryImplementation, generateCreateFunctionContent). The
// constructors, used by the implementation only, are qualified (and their
// packages imported) when implementation is set.
func (g *gen) factoryCreates(p ecore.EPackage, scope *fileScope, implementation bool) []factoryCreateView {
	var creates []factoryCreateView
	for _, c := range g.concreteClassesByName(p) {
		typ := scope.qualify(g.classInterface(c))
		create := "Create" + c.GetName()
		constructor := constructorName(g.classImplementation(c))
		if g.HasExtension(c) {
			constructor = constructorName(g.classExtension(c))
		}
		if implementation {
			constructor = scope.qualify(constructor)
		}
		creates = append(creates, factoryCreateView{Name: create, Type: typ, Kind: "new", Constructor: constructor})
		for _, ref := range sortedByName(g.m.Class(c).AllContainers) {
			// v1 typed eContainer with the bare name of the container class in
			// the implementation; its qualified interface is identical when the
			// container is in the same package and compiles otherwise.
			container := scope.qualify(g.typeOf(ref))
			fromContainer := factoryCreateView{
				Name:   create + "FromContainer",
				Params: "eContainer " + container,
				Type:   typ,
				Kind:   "container",
				Create: create,
			}
			if opposite := ref.GetEOpposite(); opposite != nil {
				// v1 named the accessor Get/Set<Name> from the raw name, without
				// the getter or setter naming rules.
				fromContainer.OppositeMany = opposite.GetUpperBound() != 1
				fromContainer.Opposite = g.oppositeAccessor(opposite)
			}
			creates = append(creates, fromContainer)
			if accessor := g.m.ElementIDAccessorName(c); accessor != "" {
				withID := fromContainer
				withID.Name = create + "FromContainerAndClassID"
				withID.Params = "eContainer " + container + ", classID int"
				withID.SetID = "Set" + accessor + "ID"
				creates = append(creates, withID)
			}
		}
	}
	return creates
}

// oppositeAccessor is the accessor of the opposite of a container reference
// a created element goes in: the getter of a many-valued opposite (the element
// is added to its list), else the setter.
func (g *gen) oppositeAccessor(opposite ecore.EReference) string {
	if opposite.GetUpperBound() != 1 {
		return g.getterName(opposite)
	}
	return g.setterName(opposite)
}

// conversions lists the conversions of the serializable data types of a
// package (sub packages included), sorted by name (v1 generateFactory dataTypes).
func (g *gen) conversions(p ecore.EPackage, scope *fileScope, view factoryImplView) []conversionView {
	var serializable []ecore.EDataType
	for _, dt := range g.allDataTypesByName(p) {
		if dt.IsSerializable() {
			serializable = append(serializable, dt)
		}
	}
	var conversions []conversionView
	for _, dt := range serializable {
		conversions = append(conversions, g.conversionView(dt, scope, view))
	}
	return conversions
}

// conversionView builds the conversion of a data type: its kind and the
// identifiers its templates use (v1 generateCreateFromString, generateConvertToString).
func (g *gen) conversionView(dt ecore.EDataType, scope *fileScope, view factoryImplView) conversionView {
	c := conversionView{
		Receiver:  view.Receiver,
		Impl:      view.Name,
		Name:      dt.GetName(),
		ID:        scope.qualify(g.qualifiedInterfaceIdentifier(dt.GetEPackage(), g.m.ClassifierIDName(dt))),
		EDataType: view.EDataType,
		Any:       view.Any,
	}
	strconvFunc := func(name string) string { return scope.qualify("strconv/strconv." + name) }
	if enum, ok := dt.(ecore.EEnum); ok {
		c.Kind = "enum"
		c.Type = scope.qualify(g.typeOfClassifier(enum))
		c.Var = naming.LowerFirst(enum.GetName())
		return c
	}
	c.Kind = conversionKind(g.instanceTypeName(dt))
	switch c.Kind {
	case "bool":
		c.Parse, c.Format = strconvFunc("ParseBool"), strconvFunc("FormatBool")
	case "int":
		c.Parse, c.Format = strconvFunc("Atoi"), strconvFunc("Itoa")
	case "int64", "int32", "int16", "int8":
		c.Parse, c.Format = strconvFunc("ParseInt"), strconvFunc("FormatInt")
	case "float64", "float32":
		c.Parse, c.Format = strconvFunc("ParseFloat"), strconvFunc("FormatFloat")
	case "reflect":
		c.Parse = scope.qualify("reflect/reflect.Ptr")
	case "time":
		c.Parse = scope.qualify("time/time.Parse")
		c.Type = scope.qualify("*time/time.Time")
	}
	return c
}

// conversionKind is the conversion kind of a Go type (v1 isFactoryImplemented
// and the guards of generateCreateFromString): the type itself for the
// primitive types, bytes for []byte, reflect for reflect.Type, time for
// *time.Time, unsupported otherwise (the conversion panics).
func conversionKind(goType string) string {
	switch goType {
	case "string", "bool", "int", "int64", "int32", "int16", "int8", "byte", "float64", "float32":
		return goType
	case "[]byte":
		return "bytes"
	case "reflect/reflect.Type":
		return "reflect"
	case "*time/time.Time":
		return "time"
	}
	return "unsupported"
}
