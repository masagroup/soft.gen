package golang

import (
	"sort"
	"strconv"
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/model"
	"github.com/masagroup/soft.gen/v2/internal/naming"
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
	// Hybrid is the typed fields of a hybrid class, nil without any.
	Hybrid *hybridView
}

// hybridView is the typed fields of a hybrid class (property storage=hybrid): the
// block holding them, a cache of the values of the store read by the getters,
// declared to the hybrid root (its package: NewLevel, Value, List, NewLayout),
// which implements the reflective methods with them.
type hybridView struct {
	// Type is the struct of the block; Header the header it embeds;
	// FieldsType the Fields of the root holding it; Any the any type.
	Type, Header, FieldsType, Any string
	// Offsetof qualifies unsafe.Offsetof: the descriptors find the fields
	// and the blocks by offset.
	Offsetof string
	// Level and Layout are the variables of the fields of the class and of
	// its layout; ParentLayout the layout of the closest class it extends
	// that has fields, nil for none.
	Level, Layout, ParentLayout string
	// NewLevel, NewLayout and LayoutType qualify the identifiers of the root
	// package; Root is the struct of the hybrid root, embedded by the class.
	NewLevel, NewLayout, LayoutType, Root string
	// Fields lists the fields of the block, one per hybrid feature.
	Fields []fieldView
	// Values lists the descriptors of the fields, in bit order.
	Values []hybridFieldView
}

// hybridFieldView is the descriptor of a typed field.
type hybridFieldView struct {
	// Name is the variable of the descriptor; Ctor the constructor, with
	// its type argument (Value[T] or List); ID the feature id; Index the bit;
	// Var and Type the field.
	Name, Ctor, ID, Index, Var, Type string
	// Flags is the FieldFlags expression of a Value, "" for a List.
	Flags string
	// Getter is the typed getter of the feature; Reference and Nillable
	// tell how EGetFromID returns it: a reference only when resolving, a
	// nil pointer as an untyped nil.
	Getter              string
	Reference, Nillable bool
}

// goComparableTypes are the Go types of the attributes held in typed fields:
// comparable with ==.
var goComparableTypes = map[string]bool{
	"bool": true, "string": true, "byte": true, "rune": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true,
}

// isHybridFeature reports whether a feature of a reflective class is held
// by a typed field: stored (neither a container, volatile nor transient), not
// a map, of a type other than the any type, and a list, a reference, an
// enum, a basic type or a pointer, whose values compare with ==.
func (g *gen) isHybridFeature(f ecore.EStructuralFeature) bool {
	info := g.m.Feature(f)
	if !info.IsField || f.IsTransient() || info.IsMap || info.IsFeatureMap || g.isAnyType(g.typeOf(f)) {
		return false
	}
	if info.IsList || info.Reference != nil {
		return true
	}
	if _, isEnum := f.GetEType().(ecore.EEnum); isEnum {
		return true
	}
	t := g.typeOf(f)
	return goComparableTypes[t] || strings.HasPrefix(t, "*")
}

// hasHybridFields reports whether the implementation of c has typed fields.
func (g *gen) hasHybridFields(c ecore.EClass) bool {
	for _, f := range g.m.Class(c).ImplementedFeatures {
		if g.isHybridFeature(f) {
			return true
		}
	}
	return false
}

// hybridRoot is the struct the root hybrid classes embed: the rootExtendsClass property, else
// ecore.HybridEObjectImpl.
func (g *gen) hybridRoot() string {
	if root, ok := g.ExplicitPackageProperty("rootExtendsClass"); ok {
		return root
	}
	return g.qualifiedImplementationIdentifier(g.m.Ecore(), "HybridEObjectImpl")
}

// hybridPrefix is the prefix of the identifiers of the typed fields of c.
func (g *gen) hybridPrefix(c ecore.EClass) string {
	return naming.LowerFirst(g.interfaceName(c))
}

// addHybridFeature makes the getter of a feature read a typed field of the
// block of the class (isHybridFeature), among the first 64.
func (g *gen) addHybridFeature(view *implementationView, c ecore.EClass, fv *featureView, f ecore.EStructuralFeature, info *model.FeatureInfo, scope *fileScope) {
	if fv.GetterStrategy != "reflective" || !g.isHybridFeature(f) {
		return
	}
	ecoreID := func(name string) string { return scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), name)) }
	if view.Hybrid == nil {
		prefix := g.hybridPrefix(c)
		name := prefix + "Fields"
		parent := "nil"
		for e := g.m.Class(c).ExtendsClass; e != nil; e = g.m.Class(e).ExtendsClass {
			if g.hasHybridFields(e) {
				parent = g.hybridPrefix(e) + "Layout"
				break
			}
		}
		view.Hybrid = &hybridView{
			Type:         name,
			Header:       ecoreID("FieldsHeader"),
			FieldsType:   ecoreID("Fields") + "[" + name + "]",
			Any:          g.anyType(),
			Offsetof:     scope.qualify("unsafe/unsafe.Offsetof"),
			Level:        prefix + "Level",
			Layout:       prefix + "Layout",
			ParentLayout: parent,
			NewLevel:     ecoreID("NewLevel"),
			NewLayout:    ecoreID("NewLayout"),
			LayoutType:   ecoreID("Layout"),
			Root:         simpleIdentifier(g.hybridRoot()),
		}
	}
	h := view.Hybrid
	if len(h.Fields) >= 64 {
		return
	}
	index := strconv.Itoa(len(h.Fields))
	fv.HybridBit = "1 << " + index
	fv.HybridField = g.hybridPrefix(c) + naming.UpperFirst(fv.Var) + "Field"
	fv.HybridRoot = h.Root
	fv.GetterStrategy = "hybrid"
	d := hybridFieldView{Name: fv.HybridField, ID: fv.ID, Index: index, Var: fv.Var, Type: fv.Type,
		Getter: fv.Getter, Reference: info.Reference != nil, Nillable: strings.HasPrefix(fv.Type, "*")}
	if info.IsList {
		d.Ctor = ecoreID("List")
	} else {
		d.Ctor = ecoreID("Value") + "[" + fv.Type + "]"
		var flags []string
		if info.Reference != nil {
			flags = append(flags, ecoreID("Reference"))
			fv.HybridReference = true
		}
		if info.Set && !info.IsContainment && !info.IsBidirectional && !info.IsContainer {
			flags = append(flags, ecoreID("Plain"))
			fv.SetterStrategy = "hybrid"
		}
		if strings.HasPrefix(fv.Type, "*") {
			flags = append(flags, ecoreID("Nillable"))
		}
		d.Flags = "0"
		if len(flags) > 0 {
			d.Flags = strings.Join(flags, "|")
		}
	}
	h.Values = append(h.Values, d)
	h.Fields = append(h.Fields, fieldView{Name: fv.Var, Type: fv.Type})
	if len(h.Fields) > compactHybridFields {
		// more fields than an ecore.FieldsHeader holds: the 16-byte header of up to 64
		h.Header = ecoreID("WideFieldsHeader")
	}
}

// compactHybridFields is the number of typed fields of a block with an ecore.FieldsHeader, of 8
// bytes; a block of more, up to 64, has an ecore.WideFieldsHeader, of 16.
const compactHybridFields = 32

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
		if g.IsHybridStorage() {
			g.addHybridFeature(view, c, fv, f, g.m.Feature(f), scope)
		}
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
// field, its ESet flag, and initialization (v1 isField, generateVariableDeclaration,
// generateVariableInitialization), its lazy initialization method (v1
// hasInitializationMethod) and its basic accessors (v1 basicFeatures).
func (g *gen) addFeature(view *implementationView, fv *featureView, info *model.FeatureInfo, reflective bool) {
	if info.IsField && !reflective {
		view.HasInitialize = true
		view.Fields = append(view.Fields, fieldView{Name: fv.Var, Type: fv.Type})
		if fv.ESetVar != "" {
			view.Fields = append(view.Fields, fieldView{Name: fv.ESetVar, Type: "bool"})
		}
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
