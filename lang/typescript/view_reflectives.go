package typescript

import (
	"sort"
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/naming"
)

// operationView is the view model of an operation: its signature and the
// call of its default implementation.
type operationView struct {
	Name, Parameters, Type string
}

// operationView builds the view of an operation.
func (s *scope) operationView(op ecore.EOperation) operationView {
	return operationView{Name: s.g.operationName(op), Parameters: s.parameterList(op), Type: s.typedElementType(op)}
}

// isDefaultOperation reports whether an operation has a default
// implementation: no key default=false (v1 isDefault).
func (g *gen) isDefaultOperation(op ecore.EOperation) bool {
	value, ok := g.KeyValue(op, "default")
	return !ok || value == "true"
}

// reflectiveCase is a case of a reflective method: a feature or an operation
// id and the variant of its body.
type reflectiveCase struct {
	*featureView
	// CaseID is the id of the case.
	CaseID string
	// Strategy is the variant of the body.
	Strategy string
	// Value is a value of the body: the value of an unset attribute, the
	// comparison of an isset, the call of an invoke.
	Value string
	// SetterBasic is basic<Setter> (eBasicInverseAdd).
	SetterBasic string
	// BasicUnsetProperty is basicUnSet<Property> (eBasicInverseRemove).
	BasicUnsetProperty string
	// Local is the variable holding the current value (eBasicInverseAdd).
	Local string
	// ElementType is the type of the elements of a list.
	ElementType string
	// MapTypes is the key and value types of a map, "K, V".
	MapTypes string
	// containment is whether the feature is a containment.
	containment bool
}

// InverseRemoveOf is the id given to eInverseRemove on the current value of a
// single reference in eBasicInverseAdd: EOPPOSITE_FEATURE_BASE - ID for a
// containment, else the id of the opposite in the class of the value (v1).
func (rc reflectiveCase) InverseRemoveOf(target string) string {
	if rc.containment {
		return rc.Ecore("EOPPOSITE_FEATURE_BASE") + " - " + rc.ID()
	}
	return rc.InverseOf(target)
}

// reflectiveView is the view model of a reflective method: its cases.
type reflectiveView struct {
	names
	Cases []reflectiveCase
}

// derivedView is the eDerivedFeatureID or eDerivedOperationID method: the
// ids of the features or operations of each mixin class in the class.
type derivedView struct {
	names
	Mixins []derivedMixin
}

// derivedMixin is a mixin class and the mapping of its ids.
type derivedMixin struct {
	Class string
	IDs   [][2]string
}

// addReflectives adds the reflective methods of a class (v1
// generateReflectivesDefinition): with a reflective delegation only the
// mapping of the ids of the mixins.
func (g *gen) addReflectives(view *classView, c ecore.EClass, s *scope) {
	info := g.m.Class(c)
	if info.HasElementID {
		g.addDerivedIDs(view, c, s)
	}
	if g.IsReflectiveDelegation() {
		return
	}
	n := names{s}
	cases := func(features []ecore.EStructuralFeature, strategy func(f ecore.EStructuralFeature, rc *reflectiveCase)) reflectiveView {
		view := reflectiveView{names: n}
		for _, f := range sortedByName(features) {
			rc := reflectiveCase{featureView: g.featureView(c, f, s)}
			rc.CaseID = rc.ID()
			strategy(f, &rc)
			view.Cases = append(view.Cases, rc)
		}
		return view
	}
	if len(info.EGetFeatures) > 0 {
		view.add("reflectives/eget", cases(info.EGetFeatures, func(f ecore.EStructuralFeature, rc *reflectiveCase) {
			fi := g.m.Feature(f)
			switch {
			case fi.IsProxy && fi.IsList:
				rc.Strategy = "proxylist"
			case fi.IsProxy:
				rc.Strategy = "proxy"
			default:
				rc.Strategy = "get"
			}
		}))
		var proxies []ecore.EStructuralFeature
		for _, f := range info.EGetFeatures {
			if fi := g.m.Feature(f); fi.IsProxy && !fi.IsList {
				proxies = append(proxies, f)
			}
		}
		view.add("reflectives/egetasync", cases(proxies, func(ecore.EStructuralFeature, *reflectiveCase) {}))
	}
	if len(info.ESetFeatures) > 0 {
		view.add("reflectives/eset", cases(info.ESetFeatures, func(f ecore.EStructuralFeature, rc *reflectiveCase) {
			fi := g.m.Feature(f)
			switch {
			case fi.IsMap:
				key, value := g.mapTypeParameters(f)
				rc.Strategy, rc.MapTypes = "map", s.qualify(key)+", "+s.qualify(value)
			case fi.IsList:
				rc.Strategy = "list"
			default:
				rc.Strategy = "set"
			}
		}))
	}
	if len(info.EUnsetFeatures) > 0 {
		view.add("reflectives/eunset", cases(info.EUnsetFeatures, func(f ecore.EStructuralFeature, rc *reflectiveCase) {
			fi := g.m.Feature(f)
			switch {
			case fi.IsList && !fi.IsUnsettable:
				rc.Strategy = "clear"
			case fi.Unset && !g.isPropertyAccessors(f):
				rc.Strategy = "unset"
			case fi.Reference != nil:
				rc.Strategy, rc.Value = "set", "null"
			default:
				rc.Strategy, rc.Value = "set", s.defaultValue(f)
				if rc.Value == "" {
					rc.Value = "null"
				}
			}
		}))
	}
	if len(info.EIsSetFeatures) > 0 {
		view.add("reflectives/eisset", cases(info.EIsSetFeatures, func(f ecore.EStructuralFeature, rc *reflectiveCase) {
			rc.Strategy, rc.Value = "isset", s.comparison(f)
		}))
	}
	if len(info.EInvokeOperations) > 0 {
		invoke := reflectiveView{names: n}
		operations := append([]ecore.EOperation(nil), info.EInvokeOperations...)
		sort.SliceStable(operations, func(i, j int) bool { return g.operationName(operations[i]) < g.operationName(operations[j]) })
		for _, op := range operations {
			rc := reflectiveCase{featureView: &featureView{names: n}, CaseID: s.operationID(c, op), Value: s.operationInvoke(op), Strategy: "return"}
			if op.GetEType() == nil {
				rc.Strategy = "void"
			}
			invoke.Cases = append(invoke.Cases, rc)
		}
		view.add("reflectives/einvoke", invoke)
	}
	if len(info.EInverseAddFeatures) > 0 {
		view.add("reflectives/einverseadd", cases(info.EInverseAddFeatures, func(f ecore.EStructuralFeature, rc *reflectiveCase) {
			fi := g.m.Feature(f)
			rc.SetterBasic = "basic" + naming.UpperFirst(g.setterName(f))
			switch {
			case fi.IsMap:
				rc.Strategy = "map"
			case fi.IsList:
				rc.Strategy, rc.ElementType = "list", s.qualify(g.typeOf(f.GetEType()))
			case fi.IsContainer && fi.BasicSet:
				rc.Strategy = "container"
			case fi.IsContainer:
				rc.Strategy = "containerset"
			default:
				rc.Strategy, rc.Local, rc.containment = "inverse", f.GetName(), fi.IsContainment
				rc.Value = g.getterName(f) + "()"
				if fi.IsProxy {
					rc.Value = "basic" + naming.UpperFirst(g.getterName(f)) + "()"
				}
			}
		}))
	}
	if len(info.EInverseRemoveFeatures) > 0 {
		view.add("reflectives/einverseremove", cases(info.EInverseRemoveFeatures, func(f ecore.EStructuralFeature, rc *reflectiveCase) {
			fi := g.m.Feature(f)
			rc.BasicUnsetProperty = "basicUnSet" + naming.UpperFirst(propertyName(f))
			switch {
			case fi.IsMap:
				rc.Strategy = "map"
			case fi.IsList:
				rc.Strategy, rc.ElementType = "list", s.qualify(g.typeOf(f.GetEType()))
			case fi.IsContainer && !fi.BasicSet:
				rc.Strategy = "container"
			case fi.Unset:
				rc.Strategy = "unset"
			default:
				rc.Strategy = "set"
			}
		}))
	}
}

// addDerivedIDs adds the mapping of the ids of the features and operations of
// the mixin classes to their ids in the class (v1 generateElementIDDefinition).
func (g *gen) addDerivedIDs(view *classView, c ecore.EClass, s *scope) {
	p := c.GetEPackage()
	getter := s.qualify(g.packageGetterIdentifier(p)) + "()"
	features := derivedView{names: names{s}}
	operations := derivedView{names: names{s}}
	hasFeatures, hasOperations := false, false
	for _, mixin := range g.m.Class(c).Mixins {
		class := getter + ".get" + g.m.ElementAccessorName(mixin) + "()"
		fm := derivedMixin{Class: class}
		for _, f := range g.m.Class(mixin).ImplementedFeatures {
			fm.IDs = append(fm.IDs, [2]string{s.featureID(mixin, f), s.featureID(c, f)})
			hasFeatures = true
		}
		features.Mixins = append(features.Mixins, fm)
		om := derivedMixin{Class: class}
		for _, op := range g.m.Class(mixin).ImplementedOperations {
			om.IDs = append(om.IDs, [2]string{s.operationID(mixin, op), s.operationID(c, op)})
			hasOperations = true
		}
		operations.Mixins = append(operations.Mixins, om)
	}
	if hasFeatures {
		view.add("reflectives/derived.feature", features)
	}
	if hasOperations {
		view.add("reflectives/derived.operation", operations)
	}
}

// comparison is the isset test of a feature (v1 generateComparison).
func (s *scope) comparison(f ecore.EStructuralFeature) string {
	g := s.g
	info := g.m.Feature(f)
	value := "this." + g.getterName(f) + "()"
	if info.IsField {
		value = "this." + variableName(f)
	}
	switch {
	case g.m.IsAny(f.GetEType()):
		return value + " != null"
	case info.IsList:
		if info.IsField {
			return value + " && !" + value + ".isEmpty()"
		}
		return "!" + value + ".isEmpty()"
	case info.Reference != nil && info.IsBackReference:
		return value + " != null"
	}
	return value + " != " + s.defaultValue(f)
}

// operationInvoke is the call of an operation with the arguments of eInvoke:
// args.get(i), cast to the type of the parameter but a single any.
func (s *scope) operationInvoke(op ecore.EOperation) string {
	g := s.g
	var args []string
	for i := 0; i < op.GetEParameters().Size(); i++ {
		p := op.GetEParameters().Get(i).(ecore.EParameter)
		arg := "args.get(" + strconv.Itoa(i) + ")"
		if p.GetUpperBound() != 1 || p.GetEType() == nil || !g.m.IsAny(p.GetEType()) {
			arg += " as " + s.typedElementType(p)
		}
		args = append(args, arg)
	}
	call := g.operationName(op) + "("
	for i, arg := range args {
		if i > 0 {
			call += ", "
		}
		call += arg
	}
	return call + ")"
}
