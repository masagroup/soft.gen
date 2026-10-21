package typescript

import (
	"testing"

	"github.com/masagroup/soft.gen/v2/internal/engine/goldentest"
)

// The fixtures are small purpose-built models, each about one topic,
// generated with entry points and properties into the committed expected
// files of testdata/golden/ts/<case>. The models of testdata/models are
// shared with Go; the ones of testdata/models/ts carry GenTS annotations
// (adding them to the shared models would change the Go output, the Go
// package registering every annotation source). The expected files are the
// output of the v1 TypeScript generator once formatted as soft.ts does
// (remove-unused-imports, prettier), with the v2 header line and the v2
// manifests, except the files of V1Differences:
//
//	go test ./lang/typescript -run TestFixtures            compare v2 with the expected files
//	go test ./lang/typescript -run TestFixtures -update    write v2's output as the expected files
//	SOFT_GEN_V1_TS_IMAGE=... SOFT_GEN_TS_TOOLS=... go test ./lang/typescript -run TestV1
//	SOFT_GEN_TS_TOOLS=... go test ./lang/typescript -run TestPrettier
//
// TestTemplateCoverage checks that the fixtures render every template.

// suite holds the TypeScript fixtures.
var suite = &goldentest.Suite{
	Lang:      New(),
	ModelsDir: "../../../testdata/models",
	GoldenDir: "../../../testdata/golden/ts",
	// the property file of every case: the fileNotes header, packageDir=$1/src
	// and testFramework=vitest, as soft.ts
	PropFiles: []string{"../../../testdata/props/ts.properties"},
	Unreachable: map[string]string{
		"feature/basicget.container": "a basic getter exists for a single proxy, and a proxy is never a container",
	},
}

// fixture is one golden case.
type fixture = goldentest.Fixture

// v1Fix is the reason of the V1Differences where v2 creates an annotation in
// its element: v1 created the annotation of a parameter in the getter named
// after it, and a nested annotation in null.
const v1AnnotationContainer = "v1 creates the annotation of a parameter in this.get<Parameter>() (no such getter, or another element) and a nested annotation in null (lost)"

// The reasons of the other V1Differences, each a v1 bug v2 fixes.
const (
	v1DiamondImports       = "v1 imports the types a class with several super types uses from the files of its super types (Category and Shop missing)"
	v1EcoreNamespace       = "v1 writes ecore. instead of the namespace of @masagroup/ecore in some identifiers"
	v1ImplementationModule = "v1 writes the implementations of another module by their name, without importing them"
	v1PackageGetterName    = "v1 calls get<packageName>() instead of the getter named after the package interface, and writes a container type without its interfaceName rule"
	v1Unparsable           = "v1 writes a string default value unquoted and an empty test value: the file does not parse"
)

// shopV1Differences are the V1Differences of a shop case: Product.ts in dir,
// and the files of more.
func shopV1Differences(dir string, more map[string]string) map[string]string {
	differences := map[string]string{dir + "/Product.ts": v1DiamondImports}
	for file, reason := range more {
		differences[file] = reason
	}
	return differences
}

// splitV1Differences are the files of a package whose implementation is in another module.
var splitV1Differences = map[string]string{
	"shop/src/ShopFactoryImpl.ts": v1ImplementationModule,
	"shop/src/ShopPackage.ts":     v1ImplementationModule,
}

// sharedPackages is the packages property of the shared package cases: the
// interfaces in the module shop, the implementation in shop/<impl>.
func sharedPackages(impl string) string {
	return "http://soft.gen/test/shop=shop#shop/" + impl
}

func init() {
	suite.Register(
		// the models shared with Go
		fixture{Name: "maps", Model: "maps.ecore"},
		fixture{Name: "shop", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil)},
		fixture{Name: "shop-property", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil), Props: map[string]string{"accessors": "property"}},
		fixture{Name: "shop-both", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil), Props: map[string]string{"accessors": "method,property"}},
		fixture{Name: "shop-async", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil), Props: map[string]string{"async": "true"}},
		fixture{Name: "shop-reflective", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil), Props: map[string]string{"featureDelegation": "Reflective"}},
		fixture{Name: "shop-jest", Model: "shop.ecore", EntryPoints: []string{"generateTests"}, Props: map[string]string{"testFramework": "jest"}},
		fixture{Name: "shop-node", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil), Props: map[string]string{"moduleResolution": "node"}},
		fixture{Name: "shop-synthetic", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil), EntryPoints: []string{"generateModel"},
			Props: map[string]string{"allowSyntheticDefaultImports": "true"}},
		fixture{Name: "shop-namespaces", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", map[string]string{"shop/src/CustomerImpl.ts": v1EcoreNamespace, "shop/src/OrderImpl.ts": v1EcoreNamespace, "shop/src/ProductImpl.ts": v1EcoreNamespace, "shop/src/ShopImpl.ts": v1EcoreNamespace, "shop/src/TaggedImpl.ts": v1EcoreNamespace, "shop/src/ShopFactoryImpl.ts": v1EcoreNamespace}), EntryPoints: []string{"generateModel"},
			Props: map[string]string{"namespaces": "@masagroup/ecore=core"}},
		fixture{Name: "shop-roots", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil), EntryPoints: []string{"generateModel"},
			Props: map[string]string{"rootExtendsClass": "base/BaseImpl", "rootExtendsInterface": "base/Base"}},
		fixture{Name: "shop-packagedir", Model: "shop.ecore", EntryPoints: []string{"generateModule"}, Props: map[string]string{"packageDir": "lib/$1"}},
		fixture{Name: "shop-width", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", nil), EntryPoints: []string{"generateModel"}, Props: map[string]string{"printWidth": "80"}},
		fixture{Name: "shop-split", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src", splitV1Differences), EntryPoints: []string{"generateModel"},
			Props: map[string]string{"packages": sharedPackages("impl")}},
		fixture{Name: "shop-shared-interfaces", Model: "shop.ecore", V1Differences: shopV1Differences("shop/src/shop", nil), EntryPoints: []string{"generateInterfaces"},
			Props: map[string]string{"packageInstance": "interfaces", "packages": sharedPackages("typed"), "packageDir": "$1/src/shop"}},
		fixture{Name: "shop-shared-typed", Model: "shop.ecore", EntryPoints: []string{"generateImplementations", "generateTests"},
			Props: map[string]string{"packageInstance": "interfaces", "packages": sharedPackages("typed"), "packageDir": "$1/src/typed",
				"paths": "shop=../shop/internal.js"}},
		fixture{Name: "shop-shared-reflective", Model: "shop.ecore", EntryPoints: []string{"generateImplementations"},
			Props: map[string]string{"packageInstance": "interfaces", "packages": sharedPackages("reflective"), "packageDir": "$1/src/reflective",
				"paths": "shop=../shop/internal.js", "featureDelegation": "Reflective"}},
		fixture{Name: "shop-shared-nocheck", Model: "shop.ecore", EntryPoints: []string{"generateImplementations"},
			Props: map[string]string{"packageInstance": "interfaces", "packages": sharedPackages("typed"), "packageDir": "$1/src/typed",
				"paths": "shop=../shop/internal.js", "checkImplementation": "false"}},
		// the models with GenTS annotations
		fixture{Name: "accessors", Model: "ts/accessors.ecore"},
		fixture{Name: "accessors-reflective", Model: "ts/accessors.ecore", EntryPoints: []string{"generateModel"},
			Props: map[string]string{"featureDelegation": "Reflective"}},
		fixture{Name: "annotations", Model: "ts/annotations.ecore",
			V1Differences: map[string]string{"annotations/src/AnnotationsPackageImpl.ts": v1AnnotationContainer}},
		fixture{Name: "attributes", Model: "ts/attributes.ecore", V1Differences: map[string]string{"attributes/src/HolderImpl.ts": v1Unparsable, "attributes/src/HolderImpl.test.ts": v1Unparsable}},
		fixture{Name: "enums", Model: "ts/enums.ecore"},
		fixture{Name: "inheritance", Model: "ts/inheritance.ecore"},
		fixture{Name: "lazy", Model: "ts/lazy.ecore"},
		fixture{Name: "metamodel", Model: "ts/metamodel.ecore"},
		fixture{Name: "naming", Model: "ts/naming.ecore", V1Differences: map[string]string{"naming/src/Custom.struct.ts": v1PackageGetterName, "naming/src/ItemBody.ts": v1PackageGetterName, "naming/src/NamingBuilder.struct.ts": v1PackageGetterName, "naming/src/Part.check.test.ts": v1PackageGetterName, "naming/src/Part.struct.ts": v1PackageGetterName, "naming/src/Public.struct.ts": v1PackageGetterName, "naming/src/Special.struct.ts": v1PackageGetterName}},
		fixture{Name: "operations", Model: "ts/operations.ecore"},
		fixture{Name: "references", Model: "ts/references.ecore"},
		fixture{Name: "shopext", Model: "ts/shopext.ecore", V1Differences: splitV1Differences},
		fixture{Name: "shopext-shared", Model: "ts/shopext.ecore", EntryPoints: []string{"generateInterfaces"},
			Props: map[string]string{"packageInstance": "interfaces"}},
		fixture{Name: "volatile", Model: "ts/volatile.ecore"},
		// the unsettable single containment and bidirectional references
		fixture{Name: "unsettable", Model: "v1bugs.ecore", V1Differences: map[string]string{"v1bugs/src/OwnerImpl.ts": v1Unparsable, "v1bugs/src/OwnerImpl.test.ts": v1Unparsable, "v1bugs/src/V1bugsPackageImpl.ts": v1Unparsable}},
	)
}

// TestFixtures compares the generation of every case with its expected files.
func TestFixtures(t *testing.T) { suite.TestFixtures(t) }

// TestTemplateCoverage checks that the fixtures, together, render every
// template define of the TypeScript language but the unreachable ones.
func TestTemplateCoverage(t *testing.T) { suite.TestTemplateCoverage(t) }
