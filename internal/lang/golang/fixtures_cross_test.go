package golang

// The cross model: classes extending a class of another model (crossbase.ecore,
// referenced by href), as an IFML model split in one model per package.
// Component extends base.NamedElement, whose package constants are qualified
// with the base package (base.NAMED_ELEMENT__NAME,
// base.NAMED_ELEMENT_FEATURE_COUNT + 0); Form extends Component of its own
// package, whose constants are not. The implementation names are exported, so
// that ComponentImpl embeds base.NamedElementImpl.

// crossURI and crossBaseURI are the nsURIs of the cross and base packages.
const (
	crossURI     = "http://soft.gen/test/cross"
	crossBaseURI = "http://soft.gen/test/crossbase"
)

// v1CrossConstants is why the package of the cross model differs from v1.
const v1CrossConstants = "v1 leaves the base constants of a class extending a class of another package unqualified: they do not compile"

func init() {
	register(fixture{Name: "cross", Model: "crossmodel.ecore",
		Props: map[string]string{
			"module":   "github.com/x",
			"packages": crossURI + "=github.com/x/cross/cross " + crossBaseURI + "=github.com/x/base/base",
			"exported": "true",
		},
		V1Differences: map[string]string{"cross/crosspackage.go": v1CrossConstants},
	})
}
