package common

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/model"
	"github.com/masagroup/soft.gen/v2/naming"
	"github.com/masagroup/soft.gen/v2/props"
)

// Gen is the base of the builders of a run, embedded by the builder of each
// language: the context, the model and the properties of the run.
type Gen struct {
	Context *engine.Context
	Model   *model.Model
	Props   *props.Properties
}

// NewGen starts the builders of a run.
func NewGen(ctx *engine.Context) Gen {
	return Gen{Context: ctx, Model: ctx.Model, Props: ctx.Props}
}

// PackageProperty returns a property of the root package: a non-empty
// command-line or file property, else the key of its generator annotation,
// else the declared default (v1 hasPackageProperty / getPackageProperty).
func (g Gen) PackageProperty(key string) (string, bool) {
	value, _, ok := g.Props.Lookup(key)
	return value, ok
}

// ExplicitPackageProperty returns a package property given by the command
// line, a property file or the annotation: the declared default does not
// count (v1 hasPackageProperty knows no default), so that a fallback decided
// in Go applies.
func (g Gen) ExplicitPackageProperty(key string) (string, bool) {
	value, layer, ok := g.Props.Lookup(key)
	if !ok || layer == props.LayerDefault {
		return "", false
	}
	return value, true
}

// PackagePropertyIs reports whether the property has that value.
func (g Gen) PackagePropertyIs(key, value string) bool {
	v, ok := g.PackageProperty(key)
	return ok && v == value
}

// KeyValue returns the value of a key of the generator annotation of the
// element itself (v1 hasKey / valueKey).
func (g Gen) KeyValue(el ecore.EModelElement, key string) (string, bool) {
	a := g.Model.Annotation(el)
	if !a.HasKey(key) {
		return "", false
	}
	return a.Value(key), true
}

// KeyIs reports whether the generator annotation of the element has the key
// with that value.
func (g Gen) KeyIs(el ecore.EModelElement, key, value string) bool {
	v, ok := g.KeyValue(el, key)
	return ok && v == value
}

// ElementName applies the naming rule found under key on the element or its
// ancestors to name, else the default rule, with the name style of the
// language (v1 getElementName). An invalid rule is reported and leaves the
// name unchanged.
func (g Gen) ElementName(el ecore.EModelElement, name, key, defaultRule string) string {
	resolved, err := naming.Resolve(g.Model.Annotation(el).Lookup, name, key, defaultRule, g.Context.Lang.NameStyle())
	if err != nil {
		g.Context.Log.Warn("naming rule", "key", key, "element", name, "error", err)
		return name
	}
	return resolved
}

// IsReflectiveDelegation reports whether the implementations delegate their
// features to the reflective store (property featureDelegation=Reflective).
func (g Gen) IsReflectiveDelegation() bool {
	return g.PackagePropertyIs("featureDelegation", "Reflective")
}

// HasExtension reports whether the element is extended by a hand-written type
// (annotation extension=true).
func (g Gen) HasExtension(el ecore.EModelElement) bool {
	return g.KeyIs(el, "extension", "true")
}
