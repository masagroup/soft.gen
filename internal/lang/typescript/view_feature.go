package typescript

import (
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// featureView is the view model of the accessors of a feature in the
// implementation of a class (v1 generateStructuralFeature.mtl). Its strategies
// pick the variant of each accessor template.
type featureView struct {
	names
	c ecore.EClass
	f ecore.EStructuralFeature
	// Name is the model name; Var the field; Property the property name.
	Name, Var, Property string
	// The accessor names.
	Getter, Setter, Unsetter, BasicGetter, BasicSetter, BasicUnsetter, InitMethod string
	// NewVar and OldVar are the local variables of the setters.
	NewVar, OldVar string
	// HasCheckImplementation is whether the setter checks that the new value
	// shares the implementation (a reference of a shared package).
	HasCheckImplementation bool
	// GetterBody, SetterBody and UnsetterBody are the hand-written bodies.
	GetterBody, SetterBody, UnsetterBody string
	// SimpleSet is whether the feature has a setter of its own, not declared
	// by the interface (v1 isSimpleSet): readable, not settable, not volatile,
	// not a container.
	SimpleSet bool
	// IsList is whether the feature is many valued.
	IsList bool
	// The strategy of each accessor: the variant of its template.
	GetterStrategy, GetterAsyncStrategy, SetterStrategy, SimpleSetStrategy, UnsetterStrategy,
	BasicGetterStrategy, BasicSetterStrategy, BasicUnsetterStrategy string
}

// featureView builds the view of a feature implemented by the class.
func (g *gen) featureView(c ecore.EClass, f ecore.EStructuralFeature, s *scope) *featureView {
	info := g.m.Feature(f)
	upper := naming.UpperFirst(f.GetName())
	v := &featureView{
		names:         names{s},
		c:             c,
		f:             f,
		Name:          f.GetName(),
		Var:           variableName(f),
		Property:      propertyName(f),
		Getter:        g.getterName(f),
		Setter:        g.setterName(f),
		Unsetter:      g.unsetterName(f),
		BasicGetter:   "basic" + naming.UpperFirst(g.getterName(f)),
		BasicSetter:   "basicSet" + naming.UpperFirst(propertyName(f)),
		BasicUnsetter: "basic" + naming.UpperFirst(g.unsetterName(f)),
		InitMethod:    info.Lazy.InitMethodName,
		NewVar:        "new" + upper,
		OldVar:        "old" + upper,
		SimpleSet:     info.Get && !info.Set && !info.IsVolatile && !info.IsContainer,
		IsList:        info.IsList,
	}
	v.HasCheckImplementation = g.isCheckImplementation(c.GetEPackage()) && info.Reference != nil && !info.IsContainer && !info.IsContainment
	v.GetterBody, _ = g.KeyValue(f, "getterBody")
	v.SetterBody, _ = g.KeyValue(f, "setterBody")
	v.UnsetterBody, _ = g.KeyValue(f, "unsetterBody")
	v.GetterStrategy = g.getterStrategy(f)
	v.GetterAsyncStrategy = g.getterAsyncStrategy(f)
	v.SetterStrategy = g.setterStrategy(f)
	v.SimpleSetStrategy = g.simpleSetStrategy(f)
	v.UnsetterStrategy = g.unsetterStrategy(f)
	v.BasicGetterStrategy = g.basicGetterStrategy(f)
	v.BasicSetterStrategy = g.basicSetterStrategy(f)
	v.BasicUnsetterStrategy = g.basicUnsetterStrategy(f)
	return v
}

// featureID writes the id of a feature in a class (v1 getImportedIDName):
// <Constants>.<CLASS__FEATURE> of the package of the class.
func (s *scope) featureID(c ecore.EClass, f ecore.EStructuralFeature) string {
	return s.qualify(s.g.constantsIdentifier(c.GetEPackage(), s.g.m.FeatureIDName(c, f)))
}

// operationID writes the id of an operation in a class.
func (s *scope) operationID(c ecore.EClass, op ecore.EOperation) string {
	return s.qualify(s.g.constantsIdentifier(c.GetEPackage(), s.g.m.OperationIDName(c, op)))
}

// classifierID writes the id of a classifier.
func (s *scope) classifierID(c ecore.EClassifier) string {
	return s.qualify(s.g.constantsIdentifier(c.GetEPackage(), s.g.m.ClassifierIDName(c)))
}

// oppositeID writes the id of an opposite feature in the class of the object
// target (v1 getImportedOppositeIDName): its id in the ecore package, else
// target.eClass().getFeatureID(<package>.<getter>()), the id in the class of
// the target differing from the id in the class declaring the opposite with
// multiple inheritance.
func (s *scope) oppositeID(opposite ecore.EStructuralFeature, target string) string {
	g := s.g
	class := opposite.GetEContainingClass()
	if g.m.IsEcore(class.GetEPackage()) {
		return s.featureID(class, opposite)
	}
	return target + ".eClass().getFeatureID(" + s.qualify(g.packageGetterIdentifier(class.GetEPackage())) + "()." +
		g.ReflectiveGetterName(opposite) + "())"
}

// constructor writes the constructor of the list or map of a feature (v1
// getListConstructor, getMapConstructor).
func (s *scope) constructor(c ecore.EClass, f ecore.EStructuralFeature) string {
	g := s.g
	info := g.m.Feature(f)
	if info.IsMap {
		reference, ok := f.(ecore.EReference)
		if !ok {
			return ""
		}
		id := s.featureID(c, f)
		if opposite := reference.GetEOpposite(); opposite != nil {
			id = s.featureID(opposite.GetEContainingClass(), f)
		}
		key, value := g.mapTypeParameters(f)
		classifier := f.GetEType()
		return "new " + s.qualify(g.ecoreIdentifier("BasicEObjectMap")) + "<" + s.qualify(key) + ", " + s.qualify(value) + ">(" +
			s.qualify(g.packageGetterIdentifier(classifier.GetEPackage())) + "()." + g.ReflectiveGetterName(classifier) +
			"(), this, " + id + ", -1, " + strconv.FormatBool(info.IsUnsettable) + ")"
	}
	elementType := s.qualify(g.typeOf(f.GetEType()))
	if info.Attribute != nil {
		if info.IsFeatureMap {
			return "null"
		}
		return "new " + s.qualify(g.ecoreIdentifier("BasicEDataTypeList")) + "<" + elementType + ">(this, " + s.featureID(c, f) + ")"
	}
	list := "new " + s.qualify(g.ecoreIdentifier("BasicEObjectList")) + "<" + elementType + ">(this, "
	proxyUnsettable := strconv.FormatBool(info.IsProxy) + ", " + strconv.FormatBool(info.IsUnsettable) + ")"
	opposite := info.Opposite
	switch {
	case info.IsContainment && opposite == nil:
		return list + s.featureID(c, f) + ", -1, true, true, false, " + proxyUnsettable
	case info.IsContainment:
		return list + s.featureID(c, f) + ", " + s.featureID(opposite.GetEContainingClass(), opposite) + ", true, true, true, " + proxyUnsettable
	case opposite == nil:
		return list + s.featureID(f.GetEContainingClass(), f) + ", -1, false, false, false, " + proxyUnsettable
	}
	return list + s.featureID(c, f) + ", " + s.featureID(opposite.GetEContainingClass(), opposite) + ", false, true, true, " + proxyUnsettable
}

// The strategies of the accessors, in the order of the v1 tests.

// getterStrategy is the variant of the getter: reflective, container, proxy,
// lazyfunction, lazyaction, lazy (another mode), list (a list or a map built
// on first read), field, body, panic.
func (g *gen) getterStrategy(f ecore.EStructuralFeature) string {
	info := g.m.Feature(f)
	switch {
	case g.IsReflectiveDelegation():
		return "reflective"
	case info.IsContainer:
		return "container"
	case !info.IsVolatile:
		switch {
		case !info.IsList && info.IsProxy:
			return "proxy"
		case info.Lazy.IsLazy():
			switch info.Lazy.Mode {
			case "function":
				return "lazyfunction"
			case "action":
				return "lazyaction"
			}
			return "field"
		case info.IsList:
			return "list"
		}
		return "field"
	case g.hasKey(f, "getterBody"):
		return "body"
	}
	return "panic"
}

// getterAsyncStrategy is the variant of the async getter: reflective, proxy, panic.
func (g *gen) getterAsyncStrategy(f ecore.EStructuralFeature) string {
	switch {
	case g.IsReflectiveDelegation():
		return "reflective"
	case !g.m.Feature(f).IsVolatile:
		return "proxy"
	}
	return "panic"
}

// setterStrategy is the variant of the setter: reflective, container,
// inverse (bidirectional or containment), field, body, panic.
func (g *gen) setterStrategy(f ecore.EStructuralFeature) string {
	info := g.m.Feature(f)
	switch {
	case g.IsReflectiveDelegation():
		return "reflective"
	case !info.IsVolatile:
		switch {
		case info.IsContainer:
			return "container"
		case info.IsBidirectional || info.IsContainment:
			return "inverse"
		}
		return "field"
	case g.hasKey(f, "setterBody"):
		return "body"
	}
	return "panic"
}

// simpleSetStrategy is the variant of the setter of a feature the interface
// does not declare: list, reflective, field.
func (g *gen) simpleSetStrategy(f ecore.EStructuralFeature) string {
	switch {
	case g.m.Feature(f).IsList:
		return "list"
	case g.IsReflectiveDelegation():
		return "reflective"
	}
	return "field"
}

// unsetterStrategy is the variant of the unsetter: reflective, lazylist (a
// list initialized by a lazy function), list, inverse, field, body, panic.
func (g *gen) unsetterStrategy(f ecore.EStructuralFeature) string {
	info := g.m.Feature(f)
	switch {
	case g.IsReflectiveDelegation():
		return "reflective"
	case !info.IsVolatile:
		switch {
		case info.IsList && info.Lazy.Mode == "function":
			return "lazylist"
		case info.IsList:
			return "list"
		case info.IsBidirectional || info.IsContainment:
			return "inverse"
		}
		return "field"
	case g.hasKey(f, "unsetterBody"):
		return "body"
	}
	return "panic"
}

// basicGetterStrategy is the variant of the basic getter: container,
// lazyfunction, lazyaction, field, body, panic.
func (g *gen) basicGetterStrategy(f ecore.EStructuralFeature) string {
	info := g.m.Feature(f)
	switch {
	case info.IsContainer:
		return "container"
	case !info.IsVolatile:
		switch info.Lazy.Mode {
		case "function":
			return "lazyfunction"
		case "action":
			return "lazyaction"
		}
		return "field"
	case g.hasKey(f, "getterBody"):
		return "body"
	}
	return "panic"
}

// basicSetterStrategy is the variant of the basic setter: container, field,
// body, panic.
func (g *gen) basicSetterStrategy(f ecore.EStructuralFeature) string {
	info := g.m.Feature(f)
	switch {
	case info.IsContainer:
		return "container"
	case !info.IsVolatile:
		return "field"
	case g.hasKey(f, "setterBody"):
		return "body"
	}
	return "panic"
}

// basicUnsetterStrategy is the variant of the basic unsetter: field, panic.
func (g *gen) basicUnsetterStrategy(f ecore.EStructuralFeature) string {
	if !g.m.Feature(f).IsVolatile {
		return "field"
	}
	return "panic"
}

// hasKey reports whether the generator annotation of the element has the key.
func (g *gen) hasKey(el ecore.EModelElement, key string) bool {
	_, ok := g.KeyValue(el, key)
	return ok
}

// initMethodView is the lazy initialization method of a feature (v1
// generateVariableInitializationMethodDefinition).
type initMethodView struct {
	Name, Var, Type string
	// Mode is function (returns the value) or action (sets the field).
	Mode string
	// Strategy is value (Value is returned or assigned), none (an action
	// initializing a proxy does nothing) or panic (not implemented).
	Strategy string
	Value    string
}

// initMethodView builds the initialization method of a lazy feature.
func (g *gen) initMethodView(c ecore.EClass, f ecore.EStructuralFeature, s *scope) *initMethodView {
	info := g.m.Feature(f)
	v := &initMethodView{Name: info.Lazy.InitMethodName, Var: variableName(f), Type: s.typedElementType(f), Mode: info.Lazy.Mode, Strategy: "panic"}
	switch {
	case info.IsList:
		v.Strategy, v.Value = "value", s.constructor(c, f)
	case v.Mode == "action" && info.IsProxy:
		v.Strategy = "none"
	case info.Attribute != nil:
		v.Strategy, v.Value = "value", s.defaultValue(f)
	}
	return v
}

// The identifiers of a feature view are written, and imported, only when a
// template uses them.

// Type is the type of the feature.
func (v *featureView) Type() string { return v.s.typedElementType(v.f) }

// ID is the id of the feature in the class.
func (v *featureView) ID() string { return v.s.featureID(v.c, v.f) }

// Default is the default value of the field: the attribute default, null for
// a reference, "" for a many attribute.
func (v *featureView) Default() string { return v.s.defaultValue(v.f) }

// Guard is the value a lazily initialized field is compared with: the
// attribute default, else null.
func (v *featureView) Guard() string {
	if _, ok := v.f.(ecore.EAttribute); ok {
		return v.s.defaultValue(v.f)
	}
	return "null"
}

// Constructor is the list or map constructor of a many feature.
func (v *featureView) Constructor() string { return v.s.constructor(v.c, v.f) }

// inverse is the id given to eInverseAdd and eInverseRemove on target: the id
// of the opposite in the class of target, else EOPPOSITE_FEATURE_BASE - ID.
func (v *featureView) inverse(target string) string {
	if opposite := v.s.g.m.Feature(v.f).Opposite; opposite != nil {
		return v.s.oppositeID(opposite, target)
	}
	return v.Ecore("EOPPOSITE_FEATURE_BASE") + " - " + v.ID()
}

// InverseOld is the inverse id on the old value.
func (v *featureView) InverseOld() string { return v.inverse(v.OldVar) }

// InverseNew is the inverse id on the new value.
func (v *featureView) InverseNew() string { return v.inverse(v.NewVar) }

// InverseThis is the inverse id on the value of the field.
func (v *featureView) InverseThis() string { return v.inverse("this." + v.Var) }

// InverseOf is the inverse id on a local variable.
func (v *featureView) InverseOf(target string) string { return v.inverse(target) }

// CheckImplementation is the line of the setter of a reference of a shared
// package checking that the new value shares the implementation, followed by
// its line break; "" otherwise.
func (v *featureView) CheckImplementation() string {
	if !v.HasCheckImplementation {
		return ""
	}
	return v.Ecore("checkSameImplementation") + "(this, " + v.NewVar + ")\n"
}
