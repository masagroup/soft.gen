package golang

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/lang/common"
	"github.com/masagroup/soft.gen/v2/internal/model"
	"github.com/masagroup/soft.gen/v2/internal/props"
)

// gen holds what the builders of a run share: the base of every language
// (properties, annotation keys, naming rules), the shorthands of its context,
// model and properties, and the memoized package paths.
type gen struct {
	common.Gen
	ctx   *engine.Context
	m     *model.Model
	props *props.Properties
	paths common.PackageSpecs
}

// newGen starts the builders of a run.
func newGen(ctx *engine.Context) *gen {
	return &gen{Gen: common.NewGen(ctx), ctx: ctx, m: ctx.Model, props: ctx.Props}
}

// isExported reports whether the implementation names of the element are
// exported: the classifier key exported, else the package property exported.
func (g *gen) isExported(el ecore.ENamedElement) bool {
	switch el := el.(type) {
	case ecore.EPackage:
		return g.PackagePropertyIs("exported", "true")
	case ecore.EClassifier:
		if v, ok := g.KeyValue(el, "exported"); ok {
			return v == "true"
		}
		return g.isExported(el.GetEPackage())
	}
	return false
}
