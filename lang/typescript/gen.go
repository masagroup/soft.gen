package typescript

import (
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/lang/internal/common"
	"github.com/masagroup/soft.gen/v2/model"
)

// gen holds what the builders of a run share: the base of every language
// (properties, annotation keys, naming rules), the model, the generated
// package and the memoized module maps.
type gen struct {
	common.Gen
	m *model.Model
	// pkg is the generated package, the root package of the model (v1
	// getCurrentPackage): its module is the current module.
	pkg ecore.EPackage

	modules    common.PackageSpecs
	paths      map[string]string
	namespaces map[string]string
}

// newGen starts the builders of a run.
func newGen(ctx *engine.Context) *gen {
	return &gen{Gen: common.NewGen(ctx), m: ctx.Model, pkg: ctx.Model.Root()}
}

// isEcore reports whether the generated package is the ecore package itself.
func (g *gen) isEcore() bool { return g.m.IsEcore(g.pkg) }

// propertyIsTrue reports whether a package property is "true".
func (g *gen) propertyIsTrue(key string) bool { return g.PackagePropertyIs(key, "true") }

// printWidth is the printWidth property, default 120.
func (g *gen) printWidth() int {
	if v, ok := g.PackageProperty("printWidth"); ok {
		if width, err := strconv.Atoi(v); err == nil && width > 0 {
			return width
		}
	}
	return 120
}

// isExported reports whether the element has the key exported=true.
func (g *gen) isExported(el ecore.EModelElement) bool { return g.KeyIs(el, "exported", "true") }
