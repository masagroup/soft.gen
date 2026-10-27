package golang

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/model"
)

// packageImplView is the view model of the package implementation file
// (template package_impl.go, v1 generatePackageImplementation).
type packageImplView struct {
	// Instance is the GetPackage singleton, nil when the interface file holds it.
	Instance *packageInstanceView
	// Name is the implementation struct.
	Name string
	// InterfaceName is the package interface, named in the comment.
	InterfaceName string
	// Constructor is new<Name>.
	Constructor string
	// EPackageExt is the embedded ecore.EPackageExt.
	EPackageExt string
	// MemberGroups lists the meta object members: the classes, then the data
	// types (enums included), each sorted by name; a blank line separates the
	// groups.
	MemberGroups [][]packageMemberView
	// PackageFactory is the factory of the package given to Initialize:
	// GetFactory(), or the constructor of the factory of a shared package.
	PackageFactory string
	// EcoreFactory is the ecore.GetFactory() call.
	EcoreFactory string
	// PackageFactoryType is the type of the packageFactory parameter: the
	// factory interface, ecore.EFactory for a shared package.
	PackageFactoryType string
	// EcoreFactoryType is ecore.EcoreFactory.
	EcoreFactoryType string
	// EcoreFactoryParam is the ecoreFactory parameter of
	// initializePackageContents: "_" in the ecore package, which creates no annotation.
	EcoreFactoryParam string
	// NameConstant, NsPrefixConstant and NsURIConstant are the NAME, NS_PREFIX and NS_URI constants.
	NameConstant, NsPrefixConstant, NsURIConstant string
	// Getters lists the getters of the meta objects.
	Getters []packageGetterView
	// Creations lists the creation of the classes and their contents, sorted by name.
	Creations []classCreationView
	// DataTypeCreations lists the creation of the data types, sorted by name.
	DataTypeCreations []creationView
	// SuperTypes lists the super type registrations.
	SuperTypes []superTypeView
	// Initializations lists the initialization of the classes, sorted by name.
	Initializations []classInitView
	// DataTypeInitializations lists the initialization of the data types, sorted by name.
	DataTypeInitializations []dataTypeInitView
	// AnnotationSources lists the annotation initializers, one per source.
	AnnotationSources []annotationSourceView
}

// packageMemberView is one meta object member of the package implementation.
type packageMemberView struct {
	// Name is the member (v1 getElementInstance).
	Name string
	// Type is the Ecore interface of the meta object.
	Type string
}

// packageGetterView is one getter of a meta object (template package/getter.<kind>).
type packageGetterView struct {
	// Name is the getter.
	Name string
	// Type is the Ecore interface returned.
	Type string
	// Kind selects the body: "member" returns the member, "feature" and
	// "operation" index the features or operations of the class member.
	Kind string
	// Member is the member of the meta object, or of its class.
	Member string
	// Index is the index of the feature or the operation in its class.
	Index int
}

// classCreationView is the creation of a class and of its features and operations.
type classCreationView struct {
	creationView
	// Contents lists the creation of the features and operations, in model order.
	Contents []creationView
}

// creationView is the creation of one meta object with the ecore factory.
type creationView struct {
	// Member is the member receiving a classifier, or the class member a
	// feature or an operation is added to.
	Member string
	// Create is the creation method of the ecore factory: Create<Kind>FromContainerAndClassID.
	Create string
	// ID is the id constant of the meta object.
	ID string
}

// superTypeView is the registration of a super type of a class.
type superTypeView struct {
	// Member is the class member.
	Member string
	// Super is the getter of the super type.
	Super string
}

// classInitView is the initialization of a class and of its features and operations.
type classInitView struct {
	// Member, Name, InstanceTypeName, Abstract and Interface are the InitEClass arguments.
	Member, Name, InstanceTypeName, Abstract, Interface string
	// Features lists the initialization of the attributes, then of the references.
	Features []featureInitView
	// Operations lists the initialization of the operations.
	Operations []operationInitView
}

// featureInitView is the initialization of a feature (template package/init.<kind>).
type featureInitView struct {
	// Kind is "attribute" or "reference".
	Kind string
	// Getter is the getter of the feature.
	Getter string
	// Type is the getter of the type of the feature.
	Type string
	// Opposite is the getter of the opposite of a reference, nil when none.
	Opposite string
	// Name and DefaultValueLiteral are the name and the default value literal, escaped.
	Name, DefaultValueLiteral string
	// Lower and Upper are the bounds.
	Lower, Upper string
	// Flags lists the boolean arguments: transient, volatile, changeable,
	// unsettable, unique, derived, ordered, iD for an attribute; transient,
	// volatile, changeable, containment, resolveProxies, unsettable, unique,
	// derived, ordered for a reference.
	Flags []string
}

// operationInitView is the initialization of an operation (template package/init.<kind>).
type operationInitView struct {
	// Kind is "operation" for a single InitEOperation call,
	// "operation.parameters" for the block adding the parameters.
	Kind string
	// Getter is the getter of the operation.
	Getter string
	// Type is the getter of the return type, nil for a void operation.
	Type string
	// Name is the name of the operation.
	Name string
	// Lower, Upper, Unique and Ordered are the bounds and the flags.
	Lower, Upper, Unique, Ordered string
	// Parameters lists the parameters; an operation with parameters is
	// initialized in a block.
	Parameters []parameterInitView
}

// parameterInitView is the registration of a parameter of an operation.
type parameterInitView struct {
	// Type is the getter of the type of the parameter: nil for a void
	// operation, as v1 did.
	Type string
	// Name is the name of the parameter.
	Name string
	// Unique and Ordered are the flags.
	Unique, Ordered string
}

// dataTypeInitView is the initialization of a data type (template package/init.<kind>).
type dataTypeInitView struct {
	// Kind is "enum" or "datatype".
	Kind string
	// Getter is the getter of the data type.
	Getter string
	// Name is the name of the data type.
	Name string
	// InstanceTypeName is the Go type name of the data type.
	InstanceTypeName string
	// Default is the default value of a data type, unquoted ("" for nil).
	Default string
	// Serializable is whether the data type is serializable.
	Serializable string
	// Literals lists the literals of an enum, in model order.
	Literals []enumLiteralInitView
	// BlankAfter leaves a blank line after the initialization: v1 ends the
	// block of an enum with one, not the line of a data type.
	BlankAfter bool
}

// enumLiteralInitView is the registration of a literal of an enum.
type enumLiteralInitView struct {
	// Name, Literal and Value are the name, the literal and the value of the literal.
	Name, Literal, Value string
	// Constant is the constant of the literal.
	Constant string
}

// annotationSourceView is the initializer of the annotations of one source.
type annotationSourceView struct {
	// Identifier is the identifier of the source: initialize<Identifier>Annotations.
	Identifier string
	// URI is the source.
	URI string
	// Annotations lists the annotations of the source, in model order.
	Annotations []annotationView
}

// annotationView is the creation of one annotation.
type annotationView struct {
	// Container is the getter of the annotated element.
	Container string
	// Details lists the details, in model order.
	Details []detailView
	// Nested lists the annotations nested in the annotation and, with the
	// property annotationContents, the objects of its contents.
	Nested []nestedView
}

// nestedView is the creation of an element held by an annotation: a nested
// annotation (template package/annotation) or an object of its contents
// (template package/object).
type nestedView interface {
	// Template is the template of the creation.
	Template() string
}

// nestedAnnotationView is the creation of an annotation nested in another one,
// in its eAnnotations or its contents (EMF addAnnotation with a depth or a
// path).
type nestedAnnotationView struct {
	// Var is the variable of the annotation, eAnnotation<depth>; Parent the
	// one of its parent.
	Var, Parent string
	// List is the getter of the list of its parent it is added to:
	// GetEAnnotations or GetContents.
	List string
	// Source is the source, escaped.
	Source string
	// Details lists the details, in model order.
	Details []detailView
	// Nested lists the annotations nested in the annotation and the objects
	// of its contents.
	Nested []nestedView
}

// Template is package/annotation.
func (nestedAnnotationView) Template() string { return "package/annotation" }

// contentObjectView is the creation of an object of the contents of an
// annotation, or of an object it contains, by the factory of its package.
type contentObjectView struct {
	// Var is the variable of the object, eObject<depth>.
	Var string
	// Create is the creation of the object: <factory>.Create<Class>().
	Create string
	// Values lists the statements setting its attributes, in feature order.
	Values []string
	// Children lists the creation of the objects it contains, in feature order.
	Children []contentObjectView
	// Add is the statement adding it to its parent.
	Add string
}

// Template is package/object.
func (contentObjectView) Template() string { return "package/object" }

// detailView is one detail of an annotation.
type detailView struct {
	// Key is the key, verbatim; Value is the value, escaped.
	Key, Value string
}

// packageImplementationFile is the package implementation:
// <package instance dir>/<package>package_impl.go, with the GetPackage
// singleton when the interface file does not hold it (v1 generatePackageImplementation).
func (g *gen) packageImplementationFile(p ecore.EPackage) ([]engine.File, error) {
	packagePath := g.packageInstancePath(p)
	scope := newFileScope(packagePath)
	e := g.m.Ecore()
	name := g.packageImplementationName(p)
	shared := g.isSharedPackage(p)
	constant := func(c string) string { return scope.qualify(g.qualifiedInterfaceIdentifier(p, c)) }
	view := packageImplView{
		Name:               name,
		InterfaceName:      g.packageInterfaceName(p),
		Constructor:        constructorName(name),
		EPackageExt:        scope.qualify(g.qualifiedImplementationIdentifier(e, "EPackageExt")),
		MemberGroups:       g.packageMemberGroups(p, scope),
		PackageFactory:     "GetFactory()",
		EcoreFactory:       scope.qualify(g.qualifiedImplementationIdentifier(e, "GetFactory")) + "()",
		PackageFactoryType: scope.qualify(g.qualifiedInterfaceIdentifier(p, g.factoryName(p))),
		EcoreFactoryType:   scope.qualify(g.qualifiedInterfaceIdentifier(e, "EcoreFactory")),
		EcoreFactoryParam:  "ecoreFactory",
		NameConstant:       constant("NAME"),
		NsPrefixConstant:   constant("NS_PREFIX"),
		NsURIConstant:      constant("NS_URI"),
		Getters:            g.packageGetters(p, scope),
	}
	if g.interfacePackagePath(p) != packagePath || shared {
		view.Instance = g.packageInstanceView(p, scope)
	}
	if shared {
		// the factory of a shared package is its data type factory (or its
		// hand-written extension), an ecore.EFactory
		factory := g.factoryImplementationName(p)
		if g.KeyIs(p, "factory", "extension") {
			factory = g.factoryExtensionName(p)
		}
		view.PackageFactory = constructorName(factory) + "()"
		view.PackageFactoryType = scope.qualify(g.qualifiedInterfaceIdentifier(e, "EFactory"))
	}
	if g.m.IsEcore(p) {
		view.EcoreFactoryParam = "_"
	}
	classes := sortedByName(g.m.Package(p).AllClasses)
	for _, c := range classes {
		view.Creations = append(view.Creations, g.classCreation(c, scope))
		view.SuperTypes = append(view.SuperTypes, g.superTypes(c, scope)...)
		view.Initializations = append(view.Initializations, g.classInitialization(c, scope))
	}
	for _, dt := range g.allDataTypesByName(p) {
		view.DataTypeCreations = append(view.DataTypeCreations, g.creation(dt, scope))
		view.DataTypeInitializations = append(view.DataTypeInitializations, g.dataTypeInitialization(dt, scope))
	}
	if !g.m.IsEcore(p) {
		sources, err := g.annotationInitializers(p, scope)
		if err != nil {
			return nil, err
		}
		view.AnnotationSources = sources
	}
	return []engine.File{{
		Path:   filePath(g.packageInstanceOutputDirectory(p), g.implementationFileName(p, g.packageName(p))),
		Body:   goSource(scope, "package_impl.go", view),
		Header: true,
	}}, nil
}

// packageMemberGroups lists the members of the package implementation: the
// classes sorted by name, then the data types (enums included) sorted by name
// (v1 generateDataMember); an empty group is left out.
func (g *gen) packageMemberGroups(p ecore.EPackage, scope *fileScope) [][]packageMemberView {
	var groups [][]packageMemberView
	group := func(elements []ecore.ENamedElement) {
		var members []packageMemberView
		for _, el := range elements {
			members = append(members, packageMemberView{Name: g.m.MemberName(el), Type: scope.qualify(g.metaType(el))})
		}
		if len(members) > 0 {
			groups = append(groups, members)
		}
	}
	group(asNamedElements(sortedByName(g.m.Package(p).AllClasses)))
	group(asNamedElements(g.allDataTypesByName(p)))
	return groups
}

// packageGetters lists the getters of the meta objects: for each class
// sorted by name, the class, then its attributes, references and operations,
// each sorted by name; then the data types sorted by name (v1 generateGetter).
// A feature or an operation is found by its index in the features or the
// operations of its class.
func (g *gen) packageGetters(p ecore.EPackage, scope *fileScope) []packageGetterView {
	var getters []packageGetterView
	member := func(el ecore.ENamedElement) {
		getters = append(getters, packageGetterView{
			Name: g.reflectiveGetterName(el), Type: scope.qualify(g.metaType(el)), Kind: "member", Member: g.m.MemberName(el),
		})
	}
	for _, c := range sortedByName(g.m.Package(p).AllClasses) {
		member(c)
		for _, f := range sortedByName(namedElements(c.GetEAttributes())) {
			getters = append(getters, packageGetterView{
				Name: g.reflectiveGetterName(f), Type: scope.qualify(g.metaType(f)), Kind: "feature",
				Member: g.m.MemberName(c), Index: c.GetEStructuralFeatures().IndexOf(f),
			})
		}
		for _, f := range sortedByName(namedElements(c.GetEReferences())) {
			getters = append(getters, packageGetterView{
				Name: g.reflectiveGetterName(f), Type: scope.qualify(g.metaType(f)), Kind: "feature",
				Member: g.m.MemberName(c), Index: c.GetEStructuralFeatures().IndexOf(f),
			})
		}
		for _, op := range sortedByName(namedElements(c.GetEOperations())) {
			getters = append(getters, packageGetterView{
				Name: g.reflectiveGetterName(op), Type: scope.qualify(g.metaType(op)), Kind: "operation",
				Member: g.m.MemberName(c), Index: c.GetEOperations().IndexOf(op),
			})
		}
	}
	for _, dt := range g.allDataTypesByName(p) {
		member(dt)
	}
	return getters
}

// creation is the creation of a classifier member with the ecore factory (v1 generateCreation).
func (g *gen) creation(c ecore.EClassifier, scope *fileScope) creationView {
	return creationView{
		Member: g.m.MemberName(c),
		Create: "Create" + c.EClass().GetName() + "FromContainerAndClassID",
		ID:     scope.qualify(g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.m.ClassifierIDName(c))),
	}
}

// classCreation is the creation of a class, then of its features and
// operations in model order (v1 generateCreation).
func (g *gen) classCreation(c ecore.EClass, scope *fileScope) classCreationView {
	creation := classCreationView{creationView: g.creation(c, scope)}
	idOf := func(name string) string { return scope.qualify(g.qualifiedInterfaceIdentifier(c.GetEPackage(), name)) }
	for _, f := range namedElements(c.GetEStructuralFeatures()) {
		creation.Contents = append(creation.Contents, creationView{
			Member: creation.Member,
			Create: "Create" + f.EClass().GetName() + "FromContainerAndClassID",
			ID:     idOf(g.m.FeatureIDName(c, f.(ecore.EStructuralFeature))),
		})
	}
	for _, op := range namedElements(c.GetEOperations()) {
		creation.Contents = append(creation.Contents, creationView{
			Member: creation.Member,
			Create: "CreateEOperationFromContainerAndClassID",
			ID:     idOf(g.m.OperationIDName(c, op.(ecore.EOperation))),
		})
	}
	return creation
}

// superTypes lists the super type registrations of a class, in model order (v1 initSuperTypes).
func (g *gen) superTypes(c ecore.EClass, scope *fileScope) []superTypeView {
	var supers []superTypeView
	for _, super := range namedElements(c.GetESuperTypes()) {
		supers = append(supers, superTypeView{Member: g.m.MemberName(c), Super: g.classifierGetter(super.(ecore.EClassifier), c, scope)})
	}
	return supers
}

// elementGetter is the getter of a meta object of the package: p.Get<X>() (v1 getPackageGetter).
func (g *gen) elementGetter(el ecore.ENamedElement) string {
	return "p." + g.reflectiveGetterName(el) + "()"
}

// classifierGetter is the getter of a classifier seen from the package of a
// classifier: p.Get<X>() when both packages have the same name (v1 compares
// the names, kept), else <package>.GetPackage().Get<X>(); nil without a
// classifier (v1 getPackageGetter(EClassifier, EClassifier)).
func (g *gen) classifierGetter(c ecore.EClassifier, from ecore.EClassifier, scope *fileScope) string {
	if c == nil {
		return "nil"
	}
	instance := "p"
	if c.GetEPackage().GetName() != from.GetEPackage().GetName() {
		instance = scope.qualify(g.packageInstance(c.GetEPackage()))
	}
	return instance + "." + g.reflectiveGetterName(c) + "()"
}

// classInitialization is the initialization of a class, of its attributes
// then its references in model order, then of its operations (v1 generateInitialization).
func (g *gen) classInitialization(c ecore.EClass, scope *fileScope) classInitView {
	init := classInitView{
		Member:           g.m.MemberName(c),
		Name:             c.GetName(),
		InstanceTypeName: g.instanceTypeName(c),
		Abstract:         strconv.FormatBool(c.IsAbstract()),
		Interface:        strconv.FormatBool(c.IsInterface()),
	}
	for _, a := range namedElements(c.GetEAttributes()) {
		init.Features = append(init.Features, g.attributeInitialization(a.(ecore.EAttribute), scope))
	}
	for _, r := range namedElements(c.GetEReferences()) {
		init.Features = append(init.Features, g.referenceInitialization(r.(ecore.EReference), scope))
	}
	for _, op := range namedElements(c.GetEOperations()) {
		init.Operations = append(init.Operations, g.operationInitialization(op.(ecore.EOperation), scope))
	}
	return init
}

// attributeInitialization is the InitEAttribute call of an attribute (v1 generateInitialization(EAttribute)).
func (g *gen) attributeInitialization(a ecore.EAttribute, scope *fileScope) featureInitView {
	return featureInitView{
		Kind:                "attribute",
		Getter:              g.elementGetter(a),
		Type:                g.classifierGetter(a.GetEType(), a.GetEContainingClass(), scope),
		Name:                a.GetName(),
		DefaultValueLiteral: escapeLiteral(a.GetDefaultValueLiteral()),
		Lower:               strconv.Itoa(a.GetLowerBound()),
		Upper:               strconv.Itoa(a.GetUpperBound()),
		Flags: bools(a.IsTransient(), a.IsVolatile(), a.IsChangeable(), a.IsUnsettable(),
			a.IsUnique(), a.IsDerived(), a.IsOrdered(), a.IsID()),
	}
}

// referenceInitialization is the InitEReference call of a reference (v1 generateInitialization(EReference)).
func (g *gen) referenceInitialization(r ecore.EReference, scope *fileScope) featureInitView {
	opposite := "nil"
	if r.GetEOpposite() != nil {
		opposite = g.elementGetter(r.GetEOpposite())
	}
	return featureInitView{
		Kind:                "reference",
		Getter:              g.elementGetter(r),
		Type:                g.classifierGetter(r.GetEType(), r.GetEContainingClass(), scope),
		Opposite:            opposite,
		Name:                r.GetName(),
		DefaultValueLiteral: escapeLiteral(r.GetDefaultValueLiteral()),
		Lower:               strconv.Itoa(r.GetLowerBound()),
		Upper:               strconv.Itoa(r.GetUpperBound()),
		Flags: bools(r.IsTransient(), r.IsVolatile(), r.IsChangeable(), r.IsContainment(), r.IsResolveProxies(),
			r.IsUnsettable(), r.IsUnique(), r.IsDerived(), r.IsOrdered()),
	}
}

// operationInitialization is the InitEOperation call of an operation and the
// AddEParameter calls of its parameters (v1 generateInitialization(EOperation)).
// The type of a void operation is nil, and so is the type of its parameters,
// as v1 did (its getter of a parameter type checked the type of the operation).
func (g *gen) operationInitialization(op ecore.EOperation, scope *fileScope) operationInitView {
	c := op.GetEContainingClass()
	void := op.GetEType() == nil
	typeGetter := func(t ecore.EClassifier) string {
		if void {
			return "nil"
		}
		return g.classifierGetter(t, c, scope)
	}
	init := operationInitView{
		Kind:    "operation",
		Getter:  g.elementGetter(op),
		Type:    typeGetter(op.GetEType()),
		Name:    op.GetName(),
		Lower:   strconv.Itoa(op.GetLowerBound()),
		Upper:   strconv.Itoa(op.GetUpperBound()),
		Unique:  strconv.FormatBool(op.IsUnique()),
		Ordered: strconv.FormatBool(op.IsOrdered()),
	}
	for _, el := range namedElements(op.GetEParameters()) {
		parameter := el.(ecore.EParameter)
		init.Parameters = append(init.Parameters, parameterInitView{
			Type:    typeGetter(parameter.GetEType()),
			Name:    parameter.GetName(),
			Unique:  strconv.FormatBool(parameter.IsUnique()),
			Ordered: strconv.FormatBool(parameter.IsOrdered()),
		})
	}
	if len(init.Parameters) > 0 {
		init.Kind = "operation.parameters"
	}
	return init
}

// dataTypeInitialization is the InitEEnum call of an enum and the
// AddEEnumLiteral calls of its literals, or the InitEDataType call of a data
// type, whose default value is its default value with the quotes removed, ""
// for nil (v1 generateInitialization(EEnum), generateInitialization(EDataType)).
func (g *gen) dataTypeInitialization(dt ecore.EDataType, scope *fileScope) dataTypeInitView {
	init := dataTypeInitView{Getter: g.elementGetter(dt), Name: dt.GetName()}
	if enum, ok := dt.(ecore.EEnum); ok {
		init.Kind = "enum"
		init.BlankAfter = true
		init.InstanceTypeName = "int32"
		if enum.GetInstanceTypeName() != "" {
			init.InstanceTypeName = g.instanceTypeName(enum)
		}
		for _, el := range namedElements(enum.GetELiterals()) {
			literal := el.(ecore.EEnumLiteral)
			init.Literals = append(init.Literals, enumLiteralInitView{
				Name:     literal.GetName(),
				Literal:  literal.GetLiteral(),
				Value:    strconv.Itoa(literal.GetValue()),
				Constant: scope.qualify(g.qualifiedInterfaceIdentifier(enum.GetEPackage(), g.constantsName(literal))),
			})
		}
		return init
	}
	init.Kind = "datatype"
	init.InstanceTypeName = g.instanceTypeName(dt)
	init.Default = strings.NewReplacer("'", "", `"`, "").Replace(g.defaultValueOfDataType(dt))
	if init.Default == "nil" {
		init.Default = ""
	}
	init.Serializable = strconv.FormatBool(dt.IsSerializable())
	return init
}

// annotationInitializers lists, per annotation source of the package, the
// annotations of the source in model order (v1 initialize<Source>Annotations).
func (g *gen) annotationInitializers(p ecore.EPackage, scope *fileScope) ([]annotationSourceView, error) {
	info := g.m.Package(p)
	var sources []annotationSourceView
	for _, source := range info.AnnotationSources {
		view := annotationSourceView{Identifier: source.Identifier, URI: source.URI}
		for _, a := range info.Annotations {
			if a.GetSource() != source.URI {
				continue
			}
			nested, err := g.nestedViews(a, model.NestedAnnotations(a), "eAnnotation", 1, p, scope)
			if err != nil {
				return nil, err
			}
			view.Annotations = append(view.Annotations, annotationView{
				Container: g.annotatedElementGetter(a.GetEModelElement(), scope),
				Details:   annotationDetails(a),
				Nested:    nested,
			})
		}
		sources = append(sources, view)
	}
	return sources, nil
}

// annotationDetails lists the details of an annotation, in model order.
func annotationDetails(a ecore.EAnnotation) []detailView {
	var details []detailView
	for _, entry := range a.GetDetails().ToArray() {
		entry := entry.(ecore.EMapEntry)
		key, _ := entry.GetKey().(string)
		value, _ := entry.GetValue().(string)
		details = append(details, detailView{Key: key, Value: escapeLiteral(value)})
	}
	return details
}

// nestedViews lists the creations of the elements held by the annotation a of
// the variable parent, at depth: the annotations of its eAnnotations, then its
// contents in model order, the annotations and, with the property
// annotationContents, the other objects (v1 dropped them: it created an
// annotation nested in eAnnotations in nil, and none of the contents). nested
// is model.NestedAnnotations(a).
func (g *gen) nestedViews(a ecore.EAnnotation, nested []model.NestedAnnotation, parent string, depth int, p ecore.EPackage, scope *fileScope) ([]nestedView, error) {
	var views []nestedView
	annotation := func(n model.NestedAnnotation) error {
		v := nestedAnnotationView{
			Var:     "eAnnotation" + strconv.Itoa(depth),
			Parent:  parent,
			List:    "GetEAnnotations",
			Source:  escapeLiteral(n.Annotation.GetSource()),
			Details: annotationDetails(n.Annotation),
		}
		if n.Content {
			v.List = "GetContents"
		}
		children, err := g.nestedViews(n.Annotation, n.Nested, v.Var, depth+1, p, scope)
		if err != nil {
			return err
		}
		v.Nested = children
		views = append(views, v)
		return nil
	}
	var contents []model.NestedAnnotation
	for _, n := range nested {
		if n.Content {
			contents = append(contents, n)
		} else if err := annotation(n); err != nil {
			return nil, err
		}
	}
	// nested lists the annotations of the contents in the order of the contents
	objects := g.PackagePropertyIs("annotationContents", "true")
	for _, content := range a.GetContents().ToArray() {
		if _, ok := content.(ecore.EAnnotation); ok {
			if err := annotation(contents[0]); err != nil {
				return nil, err
			}
			contents = contents[1:]
			continue
		}
		if !objects {
			continue
		}
		v, err := g.contentObject(content.(ecore.EObject), parent+".GetContents().Add(%s)", depth, p, scope)
		if err != nil {
			return nil, fmt.Errorf("golang: the contents of the annotation %s of %s: %w", a.GetSource(), elementLabel(a.GetEModelElement()), err)
		}
		views = append(views, v)
	}
	return views, nil
}

// contentObject is the creation of the object o of the contents of an
// annotation of the package p, or of an object it contains, at depth: created
// by the factory of its package, its attributes set and its contained objects
// created in feature order, then added to its parent by add, a format of its
// variable. The references other than containments are not written: the
// generation fails on one that is set.
func (g *gen) contentObject(o ecore.EObject, add string, depth int, p ecore.EPackage, scope *fileScope) (contentObjectView, error) {
	c := o.EClass()
	v := contentObjectView{Var: "eObject" + strconv.Itoa(depth)}
	v.Add = fmt.Sprintf(add, v.Var)
	factory, err := g.contentFactory(c, p, scope)
	if err != nil {
		return v, err
	}
	v.Create = factory + ".Create" + c.GetName() + "()"
	for _, el := range c.GetEAllStructuralFeatures().ToArray() {
		f := el.(ecore.EStructuralFeature)
		if !o.EIsSet(f) {
			continue
		}
		if r, ok := f.(ecore.EReference); ok && r.IsContainer() {
			// set by the containment of its opposite
			continue
		}
		if !f.IsChangeable() || g.isMapType(f) || g.m.IsFeatureMapEntryName(f.GetEType().GetInstanceTypeName()) {
			return v, fmt.Errorf("the feature %s of %s can not be written", f.GetName(), c.GetName())
		}
		var values []any
		if isListType(f) {
			values = o.EGet(f).(ecore.EList).ToArray()
		} else {
			values = []any{o.EGet(f)}
		}
		for _, value := range values {
			if r, ok := f.(ecore.EReference); ok {
				if !r.IsContainment() {
					return v, fmt.Errorf("the reference %s of %s is not a containment", f.GetName(), c.GetName())
				}
				child, err := g.contentObject(value.(ecore.EObject), g.contentAccessor(v.Var, f), depth+1, p, scope)
				if err != nil {
					return v, err
				}
				v.Children = append(v.Children, child)
				continue
			}
			literal, err := g.contentValue(f.GetEType().(ecore.EDataType), value, isListType(f), p, scope)
			if err != nil {
				return v, fmt.Errorf("the attribute %s of %s: %w", f.GetName(), c.GetName(), err)
			}
			v.Values = append(v.Values, fmt.Sprintf(g.contentAccessor(v.Var, f), literal))
		}
	}
	return v, nil
}

// contentAccessor is the format of the statement giving a value to the
// feature f of the object of the variable owner: <owner>.Get<F>().Add(%s) for
// a many-valued feature, <owner>.Set<F>(%s) otherwise.
func (g *gen) contentAccessor(owner string, f ecore.EStructuralFeature) string {
	if isListType(f) {
		return owner + "." + g.getterName(f) + "().Add(%s)"
	}
	return owner + "." + g.setterName(f) + "(%s)"
}

// contentFactory is the factory creating the objects of the class c in the
// annotations of the package p: GetFactory() of the package of c. A class of
// p itself is refused: p is not ready while it initializes.
func (g *gen) contentFactory(c ecore.EClass, p ecore.EPackage, scope *fileScope) (string, error) {
	if c.GetEPackage().GetNsURI() == p.GetNsURI() {
		return "", fmt.Errorf("the class %s is of the package %s itself", c.GetName(), p.GetName())
	}
	return scope.qualify(g.qualifiedImplementationIdentifier(c.GetEPackage(), "GetFactory()")), nil
}

// contentValue is the Go expression of the value of an attribute of type dt
// of an object of the contents of an annotation of the package p: the
// constant of an enum literal, a literal of a string, a boolean or a number
// (converted to its Go type when typed, for a list), else the value created
// from its string by the factory of dt.
func (g *gen) contentValue(dt ecore.EDataType, value any, typed bool, p ecore.EPackage, scope *fileScope) (string, error) {
	if enum, ok := dt.(ecore.EEnum); ok {
		var literal ecore.EEnumLiteral
		switch value := value.(type) {
		case ecore.EEnumLiteral:
			literal = value
		case int:
			literal = enum.GetEEnumLiteralByValue(value)
		}
		if literal == nil {
			return "", fmt.Errorf("the value %v is no literal of %s", value, enum.GetName())
		}
		return scope.qualify(g.qualifiedInterfaceIdentifier(enum.GetEPackage(), g.constantsName(literal))), nil
	}
	goType := scope.qualify(g.typeOfClassifier(dt))
	var literal string
	switch value := value.(type) {
	case string:
		if goType == "string" {
			return strconv.Quote(value), nil
		}
		return g.contentValueFromString(dt, value, goType, p, scope)
	case bool:
		literal = strconv.FormatBool(value)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		literal = fmt.Sprint(value)
	case float32, float64:
		f := reflect.ValueOf(value).Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return "", fmt.Errorf("the value %v has no Go literal", value)
		}
		literal = strconv.FormatFloat(f, 'g', -1, 64)
	default:
		return "", fmt.Errorf("the value %v of type %T has no Go literal", value, value)
	}
	switch goType {
	case "bool", "int", "float64":
		return literal, nil
	}
	if typed {
		return goType + "(" + literal + ")", nil
	}
	return literal, nil
}

// contentValueFromString is the value of the data type dt created from its
// string s by the factory of its package, asserted to its Go type.
func (g *gen) contentValueFromString(dt ecore.EDataType, s, goType string, p ecore.EPackage, scope *fileScope) (string, error) {
	if g.m.IsEcoreClassifier(dt) {
		return "", fmt.Errorf("the value %q of %s has no Go literal", s, dt.GetName())
	}
	instance := "p"
	if dt.GetEPackage().GetNsURI() != p.GetNsURI() {
		instance = scope.qualify(g.packageInstance(dt.GetEPackage()))
	}
	value := instance + ".GetEFactoryInstance().CreateFromString(" + instance + "." + g.reflectiveGetterName(dt) + "(), " + strconv.Quote(s) + ")"
	if g.isAnyType(goType) {
		return value, nil
	}
	return value + ".(" + goType + ")", nil
}

// elementLabel names a model element in a message: its name, or the name of
// its type.
func elementLabel(el ecore.EModelElement) string {
	if named, ok := el.(ecore.ENamedElement); ok {
		return named.GetName()
	}
	return el.EClass().GetName()
}

// annotatedElementGetter is the expression of the element an annotation is
// created in (v1 getPackageGetter(EModelElement)): p for a package, the
// literal found by name in its enum, the getter of a named element, nil otherwise.
// A parameter is reached through its operation, where v1 called a getter named
// after the parameter, which does not exist or is another element's. A nested
// annotation is created in its parent (see nestedAnnotations).
func (g *gen) annotatedElementGetter(el ecore.EModelElement, scope *fileScope) string {
	switch el := el.(type) {
	case ecore.EPackage:
		return "p"
	case ecore.EEnumLiteral:
		return g.elementGetter(el.GetEEnum()) + `.GetEEnumLiteralByName("` + el.GetName() + `")`
	case ecore.EParameter:
		op := el.GetEOperation()
		index := op.GetEParameters().IndexOf(el)
		parameter := scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), "EParameter"))
		return g.elementGetter(op) + ".GetEParameters().Get(" + strconv.Itoa(index) + ").(" + parameter + ")"
	case ecore.ENamedElement:
		return g.elementGetter(el)
	}
	return "nil"
}

// bools formats booleans as Go literals.
func bools(values ...bool) []string {
	flags := make([]string, len(values))
	for i, v := range values {
		flags[i] = strconv.FormatBool(v)
	}
	return flags
}

// escapeLiteral escapes a string for a double-quoted Go literal as v1 did:
// backslashes, double quotes, newlines and carriage returns.
func escapeLiteral(s string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`).Replace(s)
}
