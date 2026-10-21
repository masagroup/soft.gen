package golang

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/model"
)

// reflectiveMethodView is the view model of one reflective method of an
// implementation (templates reflectives/<Kind>, v1 lib/generateReflectives).
type reflectiveMethodView struct {
	ecoreNames
	// Kind names the method: derivedfeatureid, derivedoperationid, eget, eset,
	// eunset, eisset, einvoke, einverseadd, einverseremove.
	Kind string
	// Receiver and Impl are the receiver and the struct; ExtendsSimple the
	// embedded struct the default case delegates to; Any the any type.
	Receiver, Impl, ExtendsSimple, Any string
	// Mixins lists the id mappings of the derived id methods.
	Mixins []mixinIDsView
	// Cases lists the feature cases of the switch, sorted by feature name.
	Cases []reflectiveCaseView
	// Operations lists the operation cases of EInvokeFromID, sorted by method name.
	Operations []*operationView
}

// mixinIDsView maps the ids of a mixin class to the ids of the class.
type mixinIDsView struct {
	// Class is the EClass expression of the mixin.
	Class string
	// IDs lists the mappings.
	IDs []idMappingView
}

// idMappingView is one id mapping: the constant of the mixin, the constant of the class.
type idMappingView struct {
	From, To string
}

// reflectiveCaseView is one feature case of a reflective method.
type reflectiveCaseView struct {
	*featureView
	// Strategy is the variant of the case.
	Strategy string
	// Value is the value the variant needs: the value set by EUnsetFromID, the
	// comparison of EIsSetFromID, the current value read by EBasicInverseAdd.
	Value string
	// RemoveID is the feature id the current other end of a single reference
	// is removed with by EBasicInverseAdd: EOPPOSITE_FEATURE_BASE - ID for a
	// containment (even bidirectional), else the id of the opposite.
	RemoveID string
}

// reflectiveMethods lists the reflective methods of a class (v1
// generateReflectivesDefinition): in reflective delegation only the derived
// id methods of a class with mixins; else the derived ids, EGetFromID,
// ESetFromID, EUnsetFromID, EIsSetFromID, EInvokeFromID, EBasicInverseAdd and
// EBasicInverseRemove, each when it has a case.
func (g *gen) reflectiveMethods(c ecore.EClass, views map[ecore.EStructuralFeature]*featureView, scope *fileScope) []*reflectiveMethodView {
	info := g.m.Class(c)
	var methods []*reflectiveMethodView
	if info.HasElementID {
		methods = append(methods, g.derivedIDMethods(c, scope)...)
	}
	if g.IsReflectiveDelegation() {
		return methods
	}
	method := func(kind string, features []ecore.EStructuralFeature, strategy func(reflectiveCaseView, *model.FeatureInfo) reflectiveCaseView) {
		if len(features) == 0 {
			return
		}
		m := g.reflectiveMethod(c, kind, scope)
		for _, f := range sortedByFeatureName(features) {
			m.Cases = append(m.Cases, strategy(reflectiveCaseView{featureView: views[f]}, g.m.Feature(f)))
		}
		methods = append(methods, m)
	}
	method("eget", info.EGetFeatures, egetCase)
	method("eset", info.ESetFeatures, esetCase)
	method("eunset", info.EUnsetFeatures, g.eunsetCase)
	method("eisset", info.EIsSetFeatures, g.eissetCase)
	if len(info.EInvokeOperations) > 0 {
		m := g.reflectiveMethod(c, "einvoke", scope)
		for _, op := range g.sortedByOperationName(info.EInvokeOperations) {
			m.Operations = append(m.Operations, g.operationView(c, op, scope))
		}
		methods = append(methods, m)
	}
	method("einverseadd", info.EInverseAddFeatures, einverseAddCase)
	method("einverseremove", info.EInverseRemoveFeatures, einverseRemoveCase)
	return methods
}

// reflectiveMethod starts the view of a reflective method of the class.
func (g *gen) reflectiveMethod(c ecore.EClass, kind string, scope *fileScope) *reflectiveMethodView {
	return &reflectiveMethodView{
		ecoreNames:    g.ecoreNames(scope),
		Kind:          kind,
		Receiver:      receiverName(c),
		Impl:          g.implementationName(c),
		ExtendsSimple: simpleIdentifier(g.classExtends(c)),
		Any:           g.anyType(),
	}
}

// derivedIDMethods builds EDerivedFeatureID and EDerivedOperationID (v1
// generateElementIDDefinition): each exists when a mixin implements a feature
// (an operation), and maps the ids of every mixin, even one without any. The
// mixin EClass is read with Get<accessor name>, not the reflective getter
// name, as v1 did.
func (g *gen) derivedIDMethods(c ecore.EClass, scope *fileScope) []*reflectiveMethodView {
	info := g.m.Class(c)
	var methods []*reflectiveMethodView
	features, operations := false, false
	for _, mixin := range info.Mixins {
		mixinInfo := g.m.Class(mixin)
		features = features || len(mixinInfo.ImplementedFeatures) > 0
		operations = operations || len(mixinInfo.ImplementedOperations) > 0
	}
	mixinClass := func(mixin ecore.EClass) string {
		return scope.qualify(g.packageInstance(mixin.GetEPackage())) + ".Get" + g.m.ElementAccessorName(mixin) + "()"
	}
	if features {
		m := g.reflectiveMethod(c, "derivedfeatureid", scope)
		for _, mixin := range info.Mixins {
			ids := mixinIDsView{Class: mixinClass(mixin)}
			for _, f := range g.m.Class(mixin).ImplementedFeatures {
				ids.IDs = append(ids.IDs, idMappingView{From: g.featureID(mixin, f, scope), To: g.featureID(c, f, scope)})
			}
			m.Mixins = append(m.Mixins, ids)
		}
		methods = append(methods, m)
	}
	if operations {
		m := g.reflectiveMethod(c, "derivedoperationid", scope)
		for _, mixin := range info.Mixins {
			ids := mixinIDsView{Class: mixinClass(mixin)}
			for _, op := range g.m.Class(mixin).ImplementedOperations {
				ids.IDs = append(ids.IDs, idMappingView{From: g.operationID(mixin, op, scope), To: g.operationID(c, op, scope)})
			}
			m.Mixins = append(m.Mixins, ids)
		}
		methods = append(methods, m)
	}
	return methods
}

// operationID is the constant of an operation in a class, qualified and
// resolved for the file (v1 getImportedIDName).
func (g *gen) operationID(c ecore.EClass, op ecore.EOperation, scope *fileScope) string {
	return scope.qualify(g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.m.OperationIDName(c, op)))
}

// egetCase is the variant of an EGetFromID case (v1 generateEGetDefinition):
// proxylist (the unresolved list unless resolving), proxy (the basic getter
// unless resolving) or get.
func egetCase(v reflectiveCaseView, info *model.FeatureInfo) reflectiveCaseView {
	switch {
	case info.IsProxy && info.IsList:
		v.Strategy = "proxylist"
	case info.IsProxy:
		v.Strategy = "proxy"
	default:
		v.Strategy = "get"
	}
	return v
}

// esetCase is the variant of an ESetFromID case (v1 generateESetDefinition):
// map, list, reference (a nil value is allowed) or attribute.
func esetCase(v reflectiveCaseView, info *model.FeatureInfo) reflectiveCaseView {
	switch {
	case info.IsMap:
		v.Strategy = "map"
	case info.IsList:
		v.Strategy = "list"
	case info.Reference != nil:
		v.Strategy = "reference"
	default:
		v.Strategy = "attribute"
	}
	return v
}

// eunsetCase is the variant of an EUnsetFromID case (v1
// generateEUnSetDefinition): clear (a list that is not unsettable), unset
// (the unsetter) or set with Value: nil for a reference, the default of an
// attribute, or its zero value *new(T) without default.
func (g *gen) eunsetCase(v reflectiveCaseView, info *model.FeatureInfo) reflectiveCaseView {
	switch {
	case info.IsList && !info.IsUnsettable:
		v.Strategy = "clear"
	case info.Unset:
		v.Strategy = "unset"
	case info.Reference != nil:
		v.Strategy, v.Value = "set", "nil"
	default:
		v.Strategy, v.Value = "set", zeroValue(v.featureView)
	}
	return v
}

// zeroValue is the value an attribute has when it is not set: its default, or
// the zero value *new(T) of its type when it has none (a data type the type table
// does not know, without defaultValue).
func zeroValue(v *featureView) string {
	if v.Default != "" {
		return v.Default
	}
	return zeroValueOf(v.Type)
}

// eissetCase gives an EIsSetFromID case its comparison (v1 generateComparison).
func (g *gen) eissetCase(v reflectiveCaseView, info *model.FeatureInfo) reflectiveCaseView {
	v.Value = g.comparison(v.featureView, info)
	return v
}

// comparison is the expression telling whether a feature is set (v1
// generateComparison): the value is read from the field when the feature has
// one, else from the getter. A list is set when non-empty (and non-nil when
// a field); a back reference (container or shared) when non-nil (v1 also
// tested an EBoolean back reference, impossible); a data type with a
// comparison function through it (none in the table); anything else when
// different from its default: the attribute default, the default of the type
// of a reference or of the any type.
func (g *gen) comparison(v *featureView, info *model.FeatureInfo) string {
	value := v.Receiver + "." + v.Var
	if !info.IsField {
		value = v.Receiver + "." + v.AsInterface + "()." + v.Getter + "()"
	}
	any := g.m.IsAny(info.Feature.GetEType())
	switch {
	case !any && info.IsList:
		if info.IsField {
			return value + " != nil && " + value + ".Size() != 0"
		}
		return value + ".Size() != 0"
	case !any && info.Attribute != nil:
		if dt, ok := info.Feature.GetEType().(ecore.EDataType); ok {
			if _, isEnum := dt.(ecore.EEnum); !isEnum {
				if fn := g.comparisonFunction(dt); fn != "" {
					if fn[0] == '!' {
						return "!" + value + "." + fn[1:] + "()"
					}
					return value + "." + fn + "()"
				}
			}
		}
		return value + " != " + zeroValue(v)
	case !any && info.IsBackReference:
		return value + " != nil"
	}
	return value + " != " + g.defaultValueOfClassifier(info.Feature.GetEType())
}

// einverseAddCase is the variant of an EBasicInverseAdd case (v1
// generateEInverseAddDefinition): map (nothing), list (add to the notifying
// list), containerbasic and containerset (set the container through the
// basic setter or directly) or single (remove the current other end, set the
// new one through the basic setter): Value reads the current value, from the
// field (named after the feature, as v1 did) or from the (basic) getter of a
// volatile feature; RemoveID is the id the current other end is removed with.
func einverseAddCase(v reflectiveCaseView, info *model.FeatureInfo) reflectiveCaseView {
	switch {
	case info.IsMap:
		v.Strategy = "map"
	case info.IsList:
		v.Strategy = "list"
	case info.IsContainer && info.BasicSet:
		v.Strategy = "containerbasic"
	case info.IsContainer:
		v.Strategy = "containerset"
	default:
		v.Strategy = "single"
		v.RemoveID = v.ReverseID
		if info.IsContainment {
			v.RemoveID = v.EOPPOSITE_FEATURE_BASE + " - " + v.ID
		}
		switch {
		case info.IsVolatile && info.IsProxy:
			v.Value = v.Receiver + ".asBasics()." + v.BasicGetter + "()"
		case info.IsVolatile:
			v.Value = v.Receiver + "." + v.Getter + "()"
		default:
			v.Value = v.Receiver + "." + v.Name
		}
	}
	return v
}

// einverseRemoveCase is the variant of an EBasicInverseRemove case (v1
// generateEInverseRemoveDefinition): map (nothing), list (remove from the
// notifying list), container (reset the container), unset (the basic
// unsetter) or set (the basic setter with nil).
func einverseRemoveCase(v reflectiveCaseView, info *model.FeatureInfo) reflectiveCaseView {
	switch {
	case info.IsMap:
		v.Strategy = "map"
	case info.IsList:
		v.Strategy = "list"
	case info.IsContainer && !info.BasicSet:
		v.Strategy = "container"
	case info.Unset:
		v.Strategy = "unset"
	default:
		v.Strategy = "set"
	}
	return v
}
