package golang

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/model"
	"github.com/masagroup/soft.gen/v2/naming"
	"github.com/masagroup/soft.gen/v2/props"
)

// gen holds what the builders of a run share: the context, the model, the
// properties and the memoized package paths.
type gen struct {
	ctx   *engine.Context
	m     *model.Model
	props *props.Properties
	paths map[string]string
}

// newGen starts the builders of a run.
func newGen(ctx *engine.Context) *gen {
	return &gen{ctx: ctx, m: ctx.Model, props: ctx.Props}
}

// packageProperty returns a property of the root package: a non-empty
// command-line or file property, else the key of its generator annotation
// (v1 hasPackageProperty / getPackageProperty).
func (g *gen) packageProperty(key string) (string, bool) {
	value, _, ok := g.props.Lookup(key)
	return value, ok
}

// explicitPackageProperty returns a package property given by the command
// line, a property file or the annotation: the declared default does not
// count (v1 hasPackageProperty knows no default), so that a fallback decided
// in Go (classExtends, interfaceExtends) applies.
func (g *gen) explicitPackageProperty(key string) (string, bool) {
	value, layer, ok := g.props.Lookup(key)
	if !ok || layer == props.LayerDefault {
		return "", false
	}
	return value, true
}

// packagePropertyIs reports whether the property has that value.
func (g *gen) packagePropertyIs(key, value string) bool {
	v, ok := g.packageProperty(key)
	return ok && v == value
}

// keyValue returns the value of a key of the generator annotation of the
// element itself (v1 hasKey / valueKey).
func (g *gen) keyValue(el ecore.EModelElement, key string) (string, bool) {
	a := g.m.Annotation(el)
	if !a.HasKey(key) {
		return "", false
	}
	return a.Value(key), true
}

// keyIs reports whether the generator annotation of the element has the key
// with that value.
func (g *gen) keyIs(el ecore.EModelElement, key, value string) bool {
	v, ok := g.keyValue(el, key)
	return ok && v == value
}

// elementName applies the naming rule found under key on the element or its
// ancestors to name, or the default rule (v1 getElementName).
func (g *gen) elementName(el ecore.EModelElement, name, key, defaultRule string) string {
	resolved, err := naming.Resolve(g.m.Annotation(el).Lookup, name, key, defaultRule, naming.Regex)
	if err != nil {
		g.ctx.Log.Warn("naming rule", "key", key, "element", name, "error", err)
		return name
	}
	return resolved
}

// isReflectiveDelegation reports whether the implementation delegates its
// features to the reflective store (property featureDelegation=Reflective).
func (g *gen) isReflectiveDelegation() bool {
	return g.packagePropertyIs("featureDelegation", "Reflective")
}

// hasExtension reports whether the element is extended by a hand-written type
// (annotation extension=true).
func (g *gen) hasExtension(el ecore.EModelElement) bool {
	return g.keyIs(el, "extension", "true")
}

// isExported reports whether the implementation names of the element are
// exported: the classifier key exported, else the package property exported.
func (g *gen) isExported(el ecore.ENamedElement) bool {
	switch el := el.(type) {
	case ecore.EPackage:
		return g.packagePropertyIs("exported", "true")
	case ecore.EClassifier:
		if v, ok := g.keyValue(el, "exported"); ok {
			return v == "true"
		}
		return g.isExported(el.GetEPackage())
	}
	return false
}
