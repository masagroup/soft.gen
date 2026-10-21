package golang

import (
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/model"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// featureView is the view model of one structural feature of an
// implementation (templates feature/*, v1 lib/generateStructuralFeature): the
// names and identifiers every accessor uses, and the strategy of each accessor.
type featureView struct {
	ecoreNames
	// Receiver and Impl are the receiver and the struct of the methods;
	// AsInterface the accessor casting the object to its interface.
	Receiver, Impl, AsInterface string
	// Name is the model name of the feature; Var its field (v1 getVariableName).
	Name, Var string
	// Type is the Go type of the feature; Cast the type assertion from the any
	// type (".(Type)", "" when Type is the any type).
	Type, Cast string
	// The accessor names.
	Getter, Setter, Unsetter, BasicGetter, BasicSetter, BasicUnsetter, InitMethod string
	// NewVar and OldVar are the local variables of the setters.
	NewVar, OldVar string
	// ID is the feature id constant in the class; ReverseID the id of the
	// opposite in its class ("" without opposite); InverseID the id given to
	// EInverseAdd/EInverseRemove: ReverseID, else EOPPOSITE_FEATURE_BASE - ID.
	ID, ReverseID, InverseID string
	// Default is the default value of an attribute ("" when many), nil for a
	// reference (v1 getDefaultValue).
	Default string
	// Guard is the value a lazily initialized field is compared with before
	// calling the initializer: the attribute default, else nil.
	Guard string
	// Constructor is the list or map constructor expression of a many feature.
	Constructor string
	// CheckImplementation is the ecore.CheckSameImplementation statement of the
	// setter of a shared package, "" when none.
	CheckImplementation string
	// GetterBody, SetterBody and UnsetterBody are the hand-written bodies of
	// the annotation keys getterBody, setterBody and unsetterBody.
	GetterBody, SetterBody, UnsetterBody string
	// Accessors lists the accessors to generate, in v1 order: getter,
	// basicget, setter, basicset, unsetter, basicunset.
	Accessors []string
	// The strategy of each accessor: the variant of its template.
	GetterStrategy, BasicGetterStrategy, SetterStrategy, BasicSetterStrategy, UnsetterStrategy, BasicUnsetterStrategy string
}

// initMethodView is the lazy initialization method of a feature (template
// feature/init, v1 generateVariableInitializationMethodDefinition).
type initMethodView struct {
	Receiver, Impl string
	// Name is the method; Var the field; Type the type returned by a function.
	Name, Var, Type string
	// Mode is the lazy mode: function (returns the value) or action (sets the field).
	Mode string
	// Strategy is value (Value is returned or assigned), none (an action
	// initializing a proxy does nothing) or panic (not implemented).
	Strategy string
	// Value is the constructor or default value.
	Value string
}

// featureView builds the view of a feature implemented by the class.
func (g *gen) featureView(c ecore.EClass, f ecore.EStructuralFeature, scope *fileScope) *featureView {
	info := g.m.Feature(f)
	e := g.ecoreNames(scope)
	v := &featureView{
		ecoreNames:   e,
		Receiver:     receiverName(c),
		Impl:         g.implementationName(c),
		AsInterface:  asInterfaceName(g.interfaceName(c)),
		Name:         f.GetName(),
		Var:          g.variableName(f),
		Type:         scope.qualify(g.typeOf(f)),
		Getter:       g.getterName(f),
		Setter:       g.setterName(f),
		Unsetter:     g.unsetterName(f),
		InitMethod:   info.Lazy.InitMethodName,
		ID:           g.featureID(c, f, scope),
		Default:      g.defaultValue(f, scope),
		GetterBody:   g.m.Annotation(f).Value("getterBody"),
		SetterBody:   g.m.Annotation(f).Value("setterBody"),
		UnsetterBody: g.m.Annotation(f).Value("unsetterBody"),
	}
	v.BasicGetter = "basic" + naming.UpperFirst(v.Getter)
	v.BasicSetter = "basic" + naming.UpperFirst(v.Setter)
	v.BasicUnsetter = "basic" + naming.UpperFirst(v.Unsetter)
	v.NewVar = "new" + naming.UpperFirst(v.Var)
	v.OldVar = "old" + naming.UpperFirst(v.Var)
	// v1 tests isAnyType(eType.getType()): the type of the classifier, even
	// for a list; the generated type is tested instead, so that a list of the
	// any type is still asserted to EList (v1 did not compile there).
	if !g.isAnyType(v.Type) {
		v.Cast = ".(" + v.Type + ")"
	}
	if info.Opposite != nil {
		v.ReverseID = g.featureID(info.Opposite.GetEContainingClass(), info.Opposite, scope)
		v.InverseID = v.ReverseID
	} else {
		v.InverseID = e.EOPPOSITE_FEATURE_BASE + " - " + v.ID
	}
	v.Guard = "nil"
	if info.Attribute != nil && v.Default != "" {
		v.Guard = v.Default
	}
	if info.IsList {
		v.Constructor = g.listConstructor(c, f, scope)
	}
	g.featureStrategies(v, c, info)
	return v
}

// featureID is the constant of a feature in a class, qualified and resolved
// for the file (v1 getImportedIDName).
func (g *gen) featureID(c ecore.EClass, f ecore.EStructuralFeature, scope *fileScope) string {
	return scope.qualify(g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.m.FeatureIDName(c, f)))
}

// defaultValue is the default value of a feature for the file: the default of
// an attribute, nil for a reference (v1 getDefaultValue(EAttribute), and nil
// where v1 assigns a reference).
func (g *gen) defaultValue(f ecore.EStructuralFeature, scope *fileScope) string {
	if a, ok := f.(ecore.EAttribute); ok {
		return scope.qualify(g.defaultValueOfAttribute(a))
	}
	return "nil"
}

// featureStrategies decides the accessors of the feature and the variant of
// each (v1 generateImplementation(EStructuralFeature) and the
// generate*Implementation templates).
func (g *gen) featureStrategies(v *featureView, c ecore.EClass, info *model.FeatureInfo) {
	reflective := g.IsReflectiveDelegation()
	v.GetterStrategy = g.getterStrategy(v, info, reflective)
	v.Accessors = append(v.Accessors, "getter")
	if info.BasicGet && !reflective {
		v.BasicGetterStrategy = g.basicGetterStrategy(v, info)
		v.Accessors = append(v.Accessors, "basicget")
	}
	if info.Set {
		v.SetterStrategy = g.setterStrategy(v, c, info, reflective)
		v.Accessors = append(v.Accessors, "setter")
	}
	if info.BasicSet && !reflective {
		v.BasicSetterStrategy = g.basicSetterStrategy(v, info)
		v.Accessors = append(v.Accessors, "basicset")
	}
	if info.Unset {
		v.UnsetterStrategy = g.unsetterStrategy(v, info, reflective)
		v.Accessors = append(v.Accessors, "unsetter")
	}
	if info.BasicUnset && !reflective {
		v.BasicUnsetterStrategy = g.basicUnsetterStrategy(info)
		v.Accessors = append(v.Accessors, "basicunset")
	}
}

// getterStrategy is the variant of the getter (v1 generateGetImplementation):
// reflective (EGetFromID), container (the container of the object), proxy
// (resolve a single proxy; v1 also had an inverse add/remove block for a
// containment, dead code since a proxy is never a containment), lazyfunction
// and lazyaction (call the initializer; the lambda mode is lazy without
// initializer and reads the field), list (build the list or map on first
// read), field, body (the getterBody key) or panic.
func (g *gen) getterStrategy(v *featureView, info *model.FeatureInfo, reflective bool) string {
	switch {
	case reflective:
		return "reflective"
	case info.IsContainer:
		return "container"
	case !info.IsVolatile:
		switch {
		case !info.IsList && info.IsProxy:
			return "proxy"
		case info.Lazy.IsLazy():
			return lazyStrategy(info.Lazy.Mode)
		case info.IsList:
			return "list"
		}
		return "field"
	case v.GetterBody != "":
		return "body"
	}
	return "panic"
}

// lazyStrategy is the field reading variant of a lazy mode: lazyfunction,
// lazyaction, or field for the lambda mode (lazy, no initializer).
func lazyStrategy(mode string) string {
	switch mode {
	case "function":
		return "lazyfunction"
	case "action":
		return "lazyaction"
	}
	return "field"
}

// basicGetterStrategy is the variant of the basic getter (v1
// generateBasicGetImplementation): container, lazyfunction, lazyaction,
// field, body or panic.
func (g *gen) basicGetterStrategy(v *featureView, info *model.FeatureInfo) string {
	switch {
	case info.IsContainer:
		return "container"
	case !info.IsVolatile:
		if info.Lazy.IsLazy() {
			return lazyStrategy(info.Lazy.Mode)
		}
		return "field"
	case v.GetterBody != "":
		return "body"
	}
	return "panic"
}

// setterStrategy is the variant of the setter (v1 generateSetImplementation):
// reflective, container, inverse (a bidirectional or containment reference
// notifies the other end), field, body or panic. A reference of a shared
// package checking its implementations gets the CheckSameImplementation call.
func (g *gen) setterStrategy(v *featureView, c ecore.EClass, info *model.FeatureInfo, reflective bool) string {
	switch {
	case reflective:
		return "reflective"
	case !info.IsVolatile:
		if g.isCheckImplementation(c.GetEPackage()) && info.Reference != nil && !info.IsContainer && !info.IsContainment {
			v.CheckImplementation = v.scopeless("CheckSameImplementation") + "(" + v.Receiver + ", " + v.NewVar + ")"
		}
		switch {
		case info.IsContainer:
			return "container"
		case info.IsBidirectional || info.IsContainment:
			return "inverse"
		}
		return "field"
	case v.SetterBody != "":
		return "body"
	}
	return "panic"
}

// scopeless returns an ecore identifier next to the already qualified ones:
// the package alias of EClass with the identifier (CheckSameImplementation is
// not an identifier of the ecoreNames table).
func (v *featureView) scopeless(identifier string) string {
	return identifierPackagePrefix(v.EClass) + identifier
}

// identifierPackagePrefix returns the alias prefix of a resolved identifier
// ("ecore." of "ecore.EClass", "" of a bare one).
func identifierPackagePrefix(resolved string) string {
	if i := len(resolved) - len(simpleIdentifier(resolved)); i > 0 {
		return resolved[:i]
	}
	return ""
}

// basicSetterStrategy is the variant of the basic setter (v1
// generateBasicSetImplementation): container, field, body or panic.
func (g *gen) basicSetterStrategy(v *featureView, info *model.FeatureInfo) string {
	switch {
	case info.IsContainer:
		return "container"
	case !info.IsVolatile:
		return "field"
	case v.SetterBody != "":
		return "body"
	}
	return "panic"
}

// unsetterStrategy is the variant of the unsetter (v1
// generateUnSetImplementation): reflective, listlazy (clear the list when
// initialized; v1 compared an attribute list with its default, an empty
// string that did not compile: nil is tested), list (clear), inverse (a
// bidirectional or containment reference notifies the other end; v1 declared
// the notification chain without var, which did not compile), field (reset to
// the default), body or panic.
func (g *gen) unsetterStrategy(v *featureView, info *model.FeatureInfo, reflective bool) string {
	switch {
	case reflective:
		return "reflective"
	case !info.IsVolatile:
		switch {
		case info.IsList && info.Lazy.Mode == "function":
			return "listlazy"
		case info.IsList:
			return "list"
		case info.IsBidirectional || info.IsContainment:
			return "inverse"
		}
		return "field"
	case v.UnsetterBody != "":
		return "body"
	}
	return "panic"
}

// basicUnsetterStrategy is the variant of the basic unsetter (v1
// generateBasicUnSetImplementation): field or panic.
func (g *gen) basicUnsetterStrategy(info *model.FeatureInfo) string {
	if !info.IsVolatile {
		return "field"
	}
	return "panic"
}

// listConstructor is the constructor of the list or map of a many feature
// (v1 getListConstructor, getMapConstructor): a data type list for an
// attribute (nil for a feature map), an object map for a map reference, an
// object list for a reference with the four flag tuples of v1.
func (g *gen) listConstructor(c ecore.EClass, f ecore.EStructuralFeature, scope *fileScope) string {
	info := g.m.Feature(f)
	ecore := func(name string) string { return scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), name)) }
	receiver := receiverName(c) + ".AsEObjectInternal()"
	id := g.featureID(c, f, scope)
	switch {
	case info.Attribute != nil && info.IsFeatureMap:
		return "nil"
	case info.Attribute != nil:
		return ecore("NewBasicEDataTypeList") + "(" + receiver + ", " + id + ", " + strconv.FormatBool(f.IsUnique()) + ")"
	case info.Reference == nil:
		return ""
	}
	reverse := "-1"
	if info.Opposite != nil {
		reverse = g.featureID(info.Opposite.GetEContainingClass(), info.Opposite, scope)
	}
	unsettable := strconv.FormatBool(info.IsUnsettable)
	if info.IsMap {
		entry := f.GetEType()
		entryClass := scope.qualify(g.packageInstance(entry.GetEPackage())) + "." + g.reflectiveGetterName(entry) + "()"
		return ecore("NewBasicEObjectMap") + "(" + entryClass + ", " + receiver + ", " + id + ", " + reverse + ", " + unsettable + ")"
	}
	proxy := strconv.FormatBool(info.IsProxy)
	var flags string
	switch {
	case info.IsContainment && info.Opposite == nil:
		flags = "true, true, false"
	case info.IsContainment:
		flags = "true, true, true"
	case info.Opposite == nil:
		// v1 took the id of the feature in its containing class here, not in
		// the implementing class: kept, both constants exist.
		id = g.featureID(f.GetEContainingClass(), f, scope)
		flags = "false, false, false"
	default:
		flags = "false, true, true"
	}
	return ecore("NewBasicEObjectList") + "(" + receiver + ", " + id + ", " + reverse + ", " + flags + ", " + proxy + ", " + unsettable + ")"
}

// initMethodDeclaration is the declaration of the lazy initialization method
// in the initializers interface: a function returns the type of the feature
// (v1 generateVariableInitializationMethodDeclaration).
func (v *featureView) initMethodDeclaration(info *model.FeatureInfo) string {
	if info.Lazy.Mode == "function" {
		return v.InitMethod + "() " + v.Type
	}
	return v.InitMethod + "()"
}

// basicsDeclarations lists the declarations of the basic accessors in the
// basics interface (v1 generateBasicsMethodDeclaration). v1 declared the
// basic unsetter with a value parameter its definition does not take: kept,
// it compiles.
func (v *featureView) basicsDeclarations(info *model.FeatureInfo) []string {
	var decls []string
	if info.BasicGet {
		decls = append(decls, v.BasicGetter+"() "+v.Type)
	}
	if info.BasicSet {
		decls = append(decls, v.BasicSetter+"("+v.Type+", "+v.ENotificationChain+") "+v.ENotificationChain)
	}
	if info.BasicUnset {
		decls = append(decls, v.BasicUnsetter+"("+v.Type+", "+v.ENotificationChain+") "+v.ENotificationChain)
	}
	return decls
}

// initMethodView builds the lazy initialization method of a feature (v1
// generateVariableInitializationMethodDefinition): a function returns the
// list, map or attribute default, an action assigns it, does nothing for a
// proxy; a reference without constructor is not implemented.
func (g *gen) initMethodView(v *featureView, info *model.FeatureInfo) initMethodView {
	m := initMethodView{Receiver: v.Receiver, Impl: v.Impl, Name: v.InitMethod, Var: v.Var, Mode: info.Lazy.Mode, Strategy: "panic"}
	switch {
	case info.IsList:
		m.Strategy, m.Value = "value", v.Constructor
	case info.Lazy.Mode == "action" && info.IsProxy:
		m.Strategy = "none"
	case info.Attribute != nil:
		m.Strategy, m.Value = "value", v.Default
	}
	if info.Lazy.Mode == "function" {
		m.Type = v.Type
	}
	return m
}

// featureDeclarations lists the accessor declarations of a feature in its
// interface (v1 generateDeclaration(EStructuralFeature)): the getter, the
// setter when changeable and single, the unsetter when changeable and unsettable.
func (g *gen) featureDeclarations(f ecore.EStructuralFeature, scope *fileScope) []string {
	info := g.m.Feature(f)
	t := scope.qualify(g.typeOf(f))
	decls := []string{g.getterName(f) + "() " + t}
	if info.Set {
		decls = append(decls, g.setterName(f)+"("+t+")")
	}
	if info.Unset {
		decls = append(decls, g.unsetterName(f)+"()")
	}
	return decls
}
