package golang

import (
	"sort"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/model"
	"github.com/masagroup/soft.gen/v2/naming"
)

// interfaceView is the view model of the interface file of a class (template
// interface.go, v1 generateInterface).
type interfaceView struct {
	// Name is the interface.
	Name string
	// ClassName is the model class named by the comment.
	ClassName string
	// Extends lists the embedded interfaces.
	Extends []string
	// Operations lists the operation signatures, in model order.
	Operations []string
	// Features lists the accessor declarations of the attributes then of the
	// references, one group per feature; a blank line separates the groups.
	Features [][]string
}

// interfaceFile is the interface of a class: <interface dir>/<class>.go (v1
// generateInterface). A diamond class (more than one super type) embeds only
// EObject and redeclares every inherited operation, attribute and reference,
// as v1 did. The file ends with the user code region of the interface.
func (g *gen) interfaceFile(c ecore.EClass) []engine.File {
	pkg := c.GetEPackage()
	scope := newFileScope(g.interfacePackagePath(pkg))
	view := interfaceView{Name: g.interfaceName(c), ClassName: c.GetName()}
	diamond := c.GetESuperTypes().Size() > 1
	operations, attributes, references := c.GetEOperations(), c.GetEAttributes(), c.GetEReferences()
	if diamond {
		view.Extends = []string{scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), "EObject"))}
		operations, attributes, references = c.GetEAllOperations(), c.GetEAllAttributes(), c.GetEAllReferences()
	} else {
		for _, extends := range g.interfaceExtends(c) {
			view.Extends = append(view.Extends, scope.qualify(extends))
		}
	}
	for it := operations.Iterator(); it.HasNext(); {
		view.Operations = append(view.Operations, g.operationSignature(it.Next().(ecore.EOperation), scope))
	}
	for _, list := range []ecore.EList{attributes, references} {
		for it := list.Iterator(); it.HasNext(); {
			view.Features = append(view.Features, g.featureDeclarations(it.Next().(ecore.EStructuralFeature), scope))
		}
	}
	return []engine.File{{
		Path:      filePath(g.interfaceOutputDirectory(pkg), g.interfaceFileName(c, c.GetName())),
		Body:      goSource(scope, "interface.go", view),
		Header:    true,
		Protected: true,
	}}
}

// ecoreNames is the identifiers of the ecore package an implementation uses,
// qualified for the file (bare inside the ecore package: v1 ecorePrefix).
type ecoreNames struct {
	EObject, EClass, EList, EFactory, ENotificationChain, EObjectInternal, EObjectList, ENotifyingList string
	NewNotificationByFeatureID, SET, UNSET, RESOLVE, NO_INDEX, EOPPOSITE_FEATURE_BASE                  string
}

// ecoreNames qualifies the ecore identifiers for the scope.
func (g *gen) ecoreNames(scope *fileScope) ecoreNames {
	q := func(name string) string { return scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), name)) }
	return ecoreNames{
		EObject: q("EObject"), EClass: q("EClass"), EList: q("EList"), EFactory: q("EFactory"),
		ENotificationChain: q("ENotificationChain"), EObjectInternal: q("EObjectInternal"),
		EObjectList: q("EObjectList"), ENotifyingList: q("ENotifyingList"),
		NewNotificationByFeatureID: q("NewNotificationByFeatureID"),
		SET:                        q("SET"), UNSET: q("UNSET"), RESOLVE: q("RESOLVE"), NO_INDEX: q("NO_INDEX"),
		EOPPOSITE_FEATURE_BASE: q("EOPPOSITE_FEATURE_BASE"),
	}
}

// implementationView is the view model of the implementation file of a class
// (template impl.go, v1 generateImplementation).
type implementationView struct {
	ecoreNames
	// Name is the implementation struct; ClassName the model class of the comment.
	Name, ClassName string
	// Receiver is the receiver of the methods.
	Receiver string
	// Extends is the embedded struct; ExtendsSimple its bare name, used to
	// call the embedded methods.
	Extends, ExtendsSimple string
	// Fields lists the fields, sorted by name; none in reflective delegation.
	Fields []fieldView
	// Initializers is the interface of the lazy initialization methods, nil without any.
	Initializers *methodSetView
	// Basics is the interface of the basic accessors, nil without any.
	Basics *methodSetView
	// Constructor is the constructor function.
	Constructor string
	// HasInitialize is whether the struct has fields to initialize (not in
	// reflective delegation); Initializations lists the attribute assignments.
	HasInitialize   bool
	Initializations []string
	// Interface is the interface of the class; AsInterface the accessor
	// casting the object to it.
	Interface, AsInterface string
	// StaticClass is the EClass expression of EStaticClass.
	StaticClass string
	// FeatureCount is the value of EStaticFeatureCount: the count constant, 0
	// in reflective delegation.
	FeatureCount string
	// IsShared is whether the package instance is shared: EFactory is generated.
	IsShared bool
	// MapEntry is the EMapEntry accessors of a map entry class, nil otherwise.
	MapEntry *mapEntryView
	// Operations lists the default implementations, sorted by name.
	Operations []*operationView
	// Features lists the implemented features, sorted by variable name.
	Features []*featureView
	// InitMethods lists the lazy initialization methods, sorted by variable name.
	InitMethods []initMethodView
	// Reflectives lists the reflective methods, in v1 order.
	Reflectives []*reflectiveMethodView
}

// fieldView is one field of the implementation struct.
type fieldView struct {
	Name, Type string
}

// methodSetView is one of the private interfaces of the implementation
// (initializers, basics): its name and its method declarations.
type methodSetView struct {
	Name    string
	Methods []string
}

// mapEntryView is the EMapEntry accessors of a map entry class (template
// impl/mapentry), delegating to the typed accessors of its key and value.
type mapEntryView struct {
	Receiver, Impl, Any, EList string
	// KeyGetter, KeySetter, ValueGetter and ValueSetter are the typed accessors.
	KeyGetter, KeySetter, ValueGetter, ValueSetter string
	// KeyList and ValueList are whether the feature is many: the list is filled.
	KeyList, ValueList bool
	// KeyCast and ValueCast are the type assertions of the values (".(T)", or
	// "" for the any type).
	KeyCast, ValueCast string
}

// implementationFile is the implementation of a class: <implementation
// dir>/<class>_impl.go (v1 generateImplementation).
func (g *gen) implementationFile(c ecore.EClass) []engine.File {
	pkg := c.GetEPackage()
	scope := newFileScope(g.implementationPackagePath(pkg))
	info := g.m.Class(c)
	reflective := g.IsReflectiveDelegation()
	extends := g.classExtends(c)
	view := &implementationView{
		ecoreNames:    g.ecoreNames(scope),
		Name:          g.implementationName(c),
		ClassName:     c.GetName(),
		Receiver:      receiverName(c),
		Extends:       scope.qualify(extends),
		ExtendsSimple: simpleIdentifier(extends),
		Constructor:   constructorName(g.implementationName(c)),
		Interface:     scope.qualify(g.classInterface(c)),
		AsInterface:   asInterfaceName(g.interfaceName(c)),
		StaticClass:   scope.qualify(g.packageInstance(pkg)) + "." + g.reflectiveGetterName(c) + "()",
		FeatureCount:  "0",
		IsShared:      g.isSharedPackage(pkg),
	}
	if !reflective {
		view.FeatureCount = scope.qualify(g.qualifiedInterfaceIdentifier(pkg, g.m.FeatureCountIDName(c)))
	}
	views := map[ecore.EStructuralFeature]*featureView{}
	for _, f := range g.sortedByVariable(info.ImplementedFeatures) {
		fv := g.featureView(c, f, scope)
		views[f] = fv
		view.Features = append(view.Features, fv)
		g.addFeature(view, fv, g.m.Feature(f), reflective)
	}
	for _, op := range g.sortedByOperationName(info.ImplementedOperations) {
		if g.m.Operation(op).IsDefault {
			view.Operations = append(view.Operations, g.operationView(c, op, scope))
		}
	}
	if info.IsMapEntry {
		view.MapEntry = g.mapEntryView(c, scope)
	}
	view.Reflectives = g.reflectiveMethods(c, views, scope)
	return []engine.File{{
		Path:   filePath(g.implementationOutputDirectory(pkg), g.implementationFileName(c, c.GetName())),
		Body:   goSource(scope, "impl.go", view),
		Header: true,
	}}
}

// addFeature records what the implementation declares for a feature: its
// field and initialization (v1 isField, generateVariableDeclaration,
// generateVariableInitialization), its lazy initialization method (v1
// hasInitializationMethod) and its basic accessors (v1 basicFeatures).
func (g *gen) addFeature(view *implementationView, fv *featureView, info *model.FeatureInfo, reflective bool) {
	if info.IsField && !reflective {
		view.HasInitialize = true
		view.Fields = append(view.Fields, fieldView{Name: fv.Var, Type: fv.Type})
		if info.Attribute != nil && fv.Default != "" {
			view.Initializations = append(view.Initializations, fv.Receiver+"."+fv.Var+" = "+fv.Default)
		}
	}
	if info.Lazy.HasInitMethod {
		if view.Initializers == nil {
			view.Initializers = &methodSetView{Name: naming.LowerFirst(simpleIdentifier(view.Interface)) + "Initializers"}
		}
		view.Initializers.Methods = append(view.Initializers.Methods, fv.initMethodDeclaration(info))
		view.InitMethods = append(view.InitMethods, g.initMethodView(fv, info))
	}
	if !reflective && (info.BasicGet || info.BasicSet || info.BasicUnset) {
		if view.Basics == nil {
			view.Basics = &methodSetView{Name: naming.LowerFirst(simpleIdentifier(view.Interface)) + "Basics"}
		}
		view.Basics.Methods = append(view.Basics.Methods, fv.basicsDeclarations(info)...)
	}
}

// asInterfaceName is the accessor casting an implementation to its interface: as<Interface>.
func asInterfaceName(interfaceName string) string {
	return "as" + interfaceName
}

// mapEntryView builds the EMapEntry accessors of a map entry class (v1
// generateImplementation, isMapEntry): GetKey/SetKey and GetValue/SetValue
// delegate to the typed accessors of the key and value features.
func (g *gen) mapEntryView(c ecore.EClass, scope *fileScope) *mapEntryView {
	key := c.GetEStructuralFeatureFromName("key")
	value := c.GetEStructuralFeatureFromName("value")
	view := &mapEntryView{
		Receiver: receiverName(c), Impl: g.implementationName(c), Any: g.anyType(),
		EList:     scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), "EList")),
		KeyGetter: g.getterName(key), KeySetter: g.setterName(key),
		ValueGetter: g.getterName(value), ValueSetter: g.setterName(value),
		KeyList: isListType(key), ValueList: isListType(value),
		KeyCast: g.castTo(key, scope), ValueCast: g.castTo(value, scope),
	}
	return view
}

// castTo is the type assertion converting the any type to the type of a
// feature: ".(T)", or "" when the type of the classifier is the any type (v1
// isAnyType(eType.getType())).
func (g *gen) castTo(f ecore.EStructuralFeature, scope *fileScope) string {
	if g.isAnyType(g.typeOfClassifier(f.GetEType())) {
		return ""
	}
	return ".(" + scope.qualify(g.typeOf(f)) + ")"
}

// sortedByVariable sorts features by variable name (v1 sortedBy(getVariableName())).
func (g *gen) sortedByVariable(features []ecore.EStructuralFeature) []ecore.EStructuralFeature {
	sorted := append([]ecore.EStructuralFeature(nil), features...)
	sort.SliceStable(sorted, func(i, j int) bool { return g.variableName(sorted[i]) < g.variableName(sorted[j]) })
	return sorted
}

// sortedByFeatureName sorts features by model name (v1 sortedBy(name)).
func sortedByFeatureName(features []ecore.EStructuralFeature) []ecore.EStructuralFeature {
	return sortedByName(features)
}

// sortedByOperationName sorts operations by method name (v1 sortedBy(getOperationName())).
func (g *gen) sortedByOperationName(operations []ecore.EOperation) []ecore.EOperation {
	sorted := append([]ecore.EOperation(nil), operations...)
	sort.SliceStable(sorted, func(i, j int) bool { return g.operationName(sorted[i]) < g.operationName(sorted[j]) })
	return sorted
}
