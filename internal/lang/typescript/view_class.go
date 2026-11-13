package typescript

import (
	"fmt"
	"sort"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
)

// classView is the view model of a class made of members (fields, methods)
// separated by a blank line: each member is a template and its data.
type classView struct {
	Declaration string
	Members     []member
}

// member is a member of a class or a test of a test file: a template and its
// data; a tight member is not separated from the previous one by a blank line.
type member struct {
	Template string
	Data     any
	Tight    bool
}

// add appends a member.
func (v *classView) add(template string, data any) {
	v.Members = append(v.Members, member{Template: template, Data: data})
}

// names gives the templates the identifiers they write: each one is
// qualified, and imported, only when the template writes it.
type names struct {
	s *scope
}

// Ecore writes an identifier of the ecore module (ecore.Notification, or
// Notification in the ecore package).
func (n names) Ecore(identifier string) string {
	return n.s.qualify(n.s.g.ecoreIdentifier(identifier))
}

// Q writes a qualified identifier.
func (n names) Q(qid string) string { return n.s.qualify(qid) }

// implementationFile is the implementation of a class (v1
// generateImplementation): its fields, constructor and meta object, its
// accessors, operations, lazy initializers and reflective methods.
func (g *gen) implementationFile(c ecore.EClass) engine.File {
	s := g.newScope().declare(g.classImplementationIdentifier(c))
	n := names{s}
	p := c.GetEPackage()
	reflective := g.IsReflectiveDelegation()
	view := classView{Declaration: "export class " + g.implementationName(c) +
		" extends " + s.qualify(g.classExtends(c)) + " implements " + s.qualify(g.classInterfaceIdentifier(c))}

	features := g.implementedFeatures(c)
	var fields []*featureView
	for _, f := range features {
		if !reflective && g.m.Feature(f).IsField {
			fields = append(fields, g.featureView(c, f, s))
		}
	}
	var declarations, initializations []string
	for _, fv := range fields {
		declarations = append(declarations, "protected "+fv.Var+": "+fv.Type())
		if fv.ESetVar != "" {
			declarations = append(declarations, "protected "+fv.ESetVar+": boolean = false")
		}
		if value := fv.Default(); value != "" {
			initializations = append(initializations, "this."+fv.Var+" = "+value)
		}
	}
	hybrid := g.hybridFeatures(c)
	masks := map[string]bool{}
	for _, f := range hybrid {
		fv := g.featureView(c, f, s)
		if !masks[fv.HybridMask] {
			masks[fv.HybridMask] = true
			declarations = append(declarations, "protected "+fv.HybridMask+": number")
			initializations = append(initializations, "this."+fv.HybridMask+" = 0")
		}
		declarations = append(declarations, "protected "+fv.HybridVar+": "+fv.Type())
	}
	if len(declarations) > 0 {
		view.add("impl/members.fields", declarations)
	}
	view.add("impl/members.constructor", initializations)
	view.add("impl/members.staticclass", struct {
		names
		Getter, ClassGetter string
	}{n, s.qualify(g.packageGetterIdentifier(p)), g.ReflectiveGetterName(c)})
	if reflective {
		view.add("impl/members.staticfeaturecount", nil)
	}
	if len(hybrid) > 0 {
		var fvs []*featureView
		for _, f := range hybrid {
			fvs = append(fvs, g.featureView(c, f, s))
		}
		var maskList []string
		for _, fv := range fvs {
			if len(maskList) == 0 || maskList[len(maskList)-1] != fv.HybridMask {
				maskList = append(maskList, fv.HybridMask)
			}
		}
		view.add("impl/members.hybrid", struct {
			names
			Fields []*featureView
			Masks  []string
		}{n, fvs, maskList})
	}
	if g.isSharedPackage(p) {
		view.add("impl/members.efactory", struct {
			names
			GetFactory string
		}{n, s.qualify(g.implementationIdentifier(p, "getFactory"))})
	}
	for _, f := range features {
		info := g.m.Feature(f)
		if g.isPropertyAccessors(f) && (info.Get || info.Set) {
			fv := g.featureView(c, f, s)
			if info.Get {
				view.add("feature/property.get", fv)
			}
			if info.Set {
				view.add("feature/property.set", fv)
			}
			if fv.SimpleSet {
				view.add("feature/property.simpleset", fv)
			}
		}
	}
	if g.m.Class(c).IsMapEntry {
		g.addMapEntryAdapters(&view, c, s)
	}
	for _, f := range features {
		fv := g.featureView(c, f, s)
		info := g.m.Feature(f)
		if info.Get {
			view.add("feature/getter", fv)
		}
		if g.isAsyncGet(f) {
			view.add("feature/getterasync", fv)
		}
		if info.Set {
			view.add("feature/setter", fv)
		}
		if fv.SimpleSet {
			view.add("feature/simpleset", fv)
		}
		if info.Unset {
			view.add("feature/unsetter", fv)
		}
		if info.BasicGet && !reflective {
			view.add("feature/basicget", fv)
		}
		if info.BasicSet && !reflective {
			view.add("feature/basicset", fv)
		}
		if info.BasicUnset && !reflective {
			view.add("feature/basicunset", fv)
		}
		if info.IsSet {
			view.add("feature/issetter", fv)
		}
	}
	// v1 writes no blank line after the async variant of an operation
	afterAsync := false
	for _, op := range g.implementedOperations(c) {
		if g.isDefaultOperation(op) {
			view.Members = append(view.Members, member{Template: "operation/impl", Data: s.operationView(op), Tight: afterAsync})
			afterAsync = g.isAsync(op)
			if afterAsync {
				view.add("operation/impl.async", s.operationView(op))
			}
		}
	}
	if !reflective {
		// the initialization methods are one member, without blank lines (v1)
		var inits []*initMethodView
		for _, f := range features {
			if g.m.Feature(f).Lazy.HasInitMethod {
				inits = append(inits, g.initMethodView(c, f, s))
			}
		}
		if len(inits) > 0 {
			view.add("feature/init.all", inits)
		}
	}
	g.addReflectives(&view, c, s)
	return g.file(g.implementationFileName(c, c.GetName()), s, "class.ts", view)
}

// classExtends is the class the implementation of a class extends (v1
// getClassExtends): the implementation, or the extension, of its first
// concrete super class; for a root class BasicEObject (the class EObject),
// the rootExtendsClass property, HybridEObjectImpl for a hybrid class
// (storage=hybrid), EStoreEObjectImpl with a reflective delegation, else
// EObjectImpl.
func (g *gen) classExtends(c ecore.EClass) string {
	extends := g.m.Class(c).ExtendsClass
	if extends == nil {
		switch {
		case c.GetName() == "EObject":
			return g.implementationIdentifier(g.m.Ecore(), "BasicEObject")
		}
		if root, ok := g.ExplicitPackageProperty("rootExtendsClass"); ok {
			return root
		}
		if g.IsHybridStorage() {
			return g.implementationIdentifier(g.m.Ecore(), "HybridEObjectImpl")
		}
		if g.IsReflectiveDelegation() {
			return g.implementationIdentifier(g.m.Ecore(), "EStoreEObjectImpl")
		}
		return g.implementationIdentifier(g.m.Ecore(), "EObjectImpl")
	}
	if g.HasExtension(extends) {
		return g.classExtensionIdentifier(extends)
	}
	return g.classImplementationIdentifier(extends)
}

// implementedFeatures are the features the implementation of a class holds,
// sorted by their field name (v1 getImplementedFeatures).
func (g *gen) implementedFeatures(c ecore.EClass) []ecore.EStructuralFeature {
	features := append([]ecore.EStructuralFeature(nil), g.m.Class(c).ImplementedFeatures...)
	sort.SliceStable(features, func(i, j int) bool { return variableName(features[i]) < variableName(features[j]) })
	return features
}

// implementedOperations are the operations the implementation of a class
// holds, sorted by their name (v1 getImplementedOperations).
func (g *gen) implementedOperations(c ecore.EClass) []ecore.EOperation {
	operations := append([]ecore.EOperation(nil), g.m.Class(c).ImplementedOperations...)
	sort.SliceStable(operations, func(i, j int) bool { return g.operationName(operations[i]) < g.operationName(operations[j]) })
	return operations
}

// mapEntryAdapterView is a getKey/setKey/getValue/setValue adapter of a map
// entry class whose key or value accessors have other names.
type mapEntryAdapterView struct {
	names
	// Name is key or value; Upper Key or Value.
	Name, Upper          string
	Type, Getter, Setter string
	IsList               bool
}

// addMapEntryAdapters adds the generic map entry accessors of a map entry
// class whose key or value accessors are renamed.
func (g *gen) addMapEntryAdapters(view *classView, c ecore.EClass, s *scope) {
	// v1 writes the adapters without blank lines between them
	first := len(view.Members)
	defer func() {
		for i := first + 1; i < len(view.Members); i++ {
			view.Members[i].Tight = true
		}
	}()
	for _, name := range []string{"key", "value"} {
		f := c.GetEStructuralFeatureFromName(name)
		if f == nil {
			continue
		}
		upper := "Key"
		if name == "value" {
			upper = "Value"
		}
		adapter := mapEntryAdapterView{names: names{s}, Name: name, Upper: upper, Type: s.typedElementType(f),
			Getter: g.getterName(f), Setter: g.setterName(f), IsList: g.m.Feature(f).IsList}
		if adapter.Getter != "get"+upper {
			view.add("impl/members.mapentryget", adapter)
		}
		if adapter.Setter != "set"+upper {
			view.add("impl/members.mapentryset", adapter)
		}
	}
}

// isHybridFeature reports whether a feature of a hybrid class (storage=hybrid) is
// held by a typed field: stored (neither a container, volatile nor transient), and
// not a map.
func (g *gen) isHybridFeature(f ecore.EStructuralFeature) bool {
	info := g.m.Feature(f)
	return info.IsField && !f.IsTransient() && !info.IsMap && !info.IsFeatureMap
}

// hybridFeatures are the features the implementation of a hybrid class holds in
// typed fields, in the order of their fields; none for another class.
func (g *gen) hybridFeatures(c ecore.EClass) []ecore.EStructuralFeature {
	if !g.IsHybridStorage() {
		return nil
	}
	var features []ecore.EStructuralFeature
	for _, f := range g.implementedFeatures(c) {
		if g.isHybridFeature(f) {
			features = append(features, f)
		}
	}
	return features
}

// hybridIndex is the index of the typed field of f in the implementation of c, -1
// when f has none.
func (g *gen) hybridIndex(c ecore.EClass, f ecore.EStructuralFeature) int {
	for i, h := range g.hybridFeatures(c) {
		if h == f {
			return i
		}
	}
	return -1
}

// hybridMask returns the mask field of the typed field of index i of the
// implementation of c, one per 32 fields and per class level, and its bit there.
func (g *gen) hybridMask(c ecore.EClass, i int) (string, string) {
	return fmt.Sprintf("_hm%s%d", c.GetName(), i/32), fmt.Sprintf("0x%x", uint32(1)<<(i%32))
}
