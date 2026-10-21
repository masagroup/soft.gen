package golang

// The shop model: a small general model (an interface class, a mixin class
// with an attribute and operations, containment with and without opposite, a
// bidirectional reference, an unsettable proxy, a non-proxy reference, an
// unsettable changeable container, an unsettable enum attribute, a date).
// Its cases generate it with the default templates and with the package
// properties: the shared package split, reflective delegation, split paths,
// root types, the Go version and the module.

// shopURI is the nsURI of the shop package.
const shopURI = "http://soft.gen/test/shop"

// shopExtURI is the nsURI of the shop package of shopext.ecore.
const shopExtURI = "http://soft.gen/test/shopext"

// shopModule is the module of the cases with a module.
const shopModule = "github.com/x"

// shopShared is the property of the shared package split, as soft.go's
// Makefile generates tournament: the interfaces and the package instance in
// <module>/shop/shop, the implementations in <module>/shop/<impl>/<impl>, the
// mocks with the interfaces.
func shopShared(impl string, extra map[string]string) map[string]string {
	p := map[string]string{
		"packageInstance": "interfaces",
		"module":          shopModule,
		"packages":        shopURI + "=" + shopModule + "/shop/shop#" + shopModule + "/shop/" + impl + "/" + impl + "#" + shopModule + "/shop/shop",
	}
	for k, v := range extra {
		p[k] = v
	}
	return p
}

func init() {
	register(
		// the default entry points, no property
		fixture{Name: "shop", Model: "shop.ecore"},
		// shared package, run 1 of soft.go's Makefile: the interfaces with the
		// shared package instance and the data type factory, the mocks and
		// their tests
		fixture{Name: "shop-shared-interfaces", Model: "shop.ecore",
			EntryPoints: []string{"generateInterfaces", "generateMocks", "generateMocksTests"},
			Props:       shopShared("typed", nil)},
		// shared package, run 2: the typed implementations (setters checking
		// the implementation, factory delegating the conversions, Install)
		// and their tests
		fixture{Name: "shop-shared-typed", Model: "shop.ecore",
			EntryPoints: []string{"generateImplementations", "generateTests"},
			Props:       shopShared("typed", nil),
			V1Differences: map[string]string{
				"shop/typed/shopfactory_impl.go":  v1FromContainer,
				"shop/typed/product_impl_test.go": v1DerivedOperationID,
			}},
		// shared package in one generateModel run: the data type factory and
		// the implementation factory together
		fixture{Name: "shop-shared-model", Model: "shop.ecore",
			EntryPoints:   []string{"generateModel"},
			Props:         shopShared("typed", nil),
			V1Differences: map[string]string{"shop/typed/shopfactory_impl.go": v1FromContainer}},
		// shared package with checkImplementation=false: no CheckSameImplementation
		fixture{Name: "shop-shared-nocheck", Model: "shop.ecore",
			EntryPoints:   []string{"generateImplementations"},
			Props:         shopShared("typed", map[string]string{"checkImplementation": "false"}),
			V1Differences: map[string]string{"shop/typed/shopfactory_impl.go": v1FromContainer}},
		// shared package with featureDelegation=Reflective: reflective
		// accessors, only the derived ids kept, reflective test kinds
		fixture{Name: "shop-reflective", Model: "shop.ecore",
			EntryPoints:   []string{"generateImplementations", "generateTests"},
			Props:         shopShared("reflective", map[string]string{"featureDelegation": "Reflective"}),
			V1Differences: map[string]string{"shop/reflective/shopfactory_impl.go": v1FromContainer}},
		// unshared split paths: the interfaces in a/shop, the implementations
		// (and mocks, the missing mock path being the last one) in b/impl
		fixture{Name: "shop-split", Model: "shop.ecore",
			Props: map[string]string{
				"module":   shopModule,
				"packages": shopURI + "=" + shopModule + "/a/shop#" + shopModule + "/b/impl",
			},
			V1Differences: map[string]string{
				"b/shopfactory_impl.go":  v1FromContainer,
				"b/product_impl_test.go": v1DerivedOperationID,
			}},
		// the root interface and struct given as qualified identifiers
		// (rootExtendsMock is not exercised: v1 1.13.1 fails generating the
		// mocks with it, "Undefined argument packages" in generateMock)
		fixture{Name: "shop-roots", Model: "shop.ecore", EntryPoints: []string{"generateModel"},
			Props: map[string]string{
				"rootExtendsInterface": "github.com/x/base/base.Base",
				"rootExtendsClass":     "github.com/x/base/base.BaseImpl",
			}},
		// goVersion=1.17: interface{} instead of any (in the model files)
		fixture{Name: "shop-go117", Model: "shop.ecore", EntryPoints: []string{"generateModel"},
			Props: map[string]string{"goVersion": "1.17"}},
		// the module equal to the package directory: output directory "" (the
		// interfaces suffice to show the paths)
		fixture{Name: "shop-module", Model: "shop.ecore", EntryPoints: []string{"generateInterfaces"},
			Props: map[string]string{"module": shopModule + "/shop", "packages": shopURI + "=" + shopModule + "/shop/shop"}},
		// a module outside the package path: output directory the package name
		fixture{Name: "shop-module-outside", Model: "shop.ecore", EntryPoints: []string{"generateInterfaces"},
			Props: map[string]string{"module": "github.com/y", "packages": shopURI + "=" + shopModule + "/somewhere/shop"}},
		// packages keyed by name (a bare path list, no uri=), the mocks in
		// their own package
		fixture{Name: "shop-packages", Model: "shop.ecore",
			Props: map[string]string{
				"module":   shopModule + "/p",
				"packages": shopModule + "/p/shop/shop#" + shopModule + "/p/impl/impl#" + shopModule + "/p/mock/mock",
			},
			V1Differences: map[string]string{
				"impl/shopfactory_impl.go":      v1FromContainer,
				"impl/shopfactory_impl_test.go": v1MockAlias,
				"impl/customer_impl_test.go":    v1MockAlias,
				"impl/order_impl_test.go":       v1MockAlias,
				"impl/product_impl_test.go":     v1MockAlias + "; " + v1DerivedOperationID,
				"impl/shop_impl_test.go":        v1MockAlias,
				"impl/tagged_impl_test.go":      v1MockAlias,
			}},
		// shopext.ecore carries the package keys: factory=extension, module
		// and packages keyed by name without the mock path (the mocks fall
		// back to the last path, the implementation's), all from the
		// annotation: GetFactory builds the factory extension
		fixture{Name: "shop-factory-ext", Model: "shopext.ecore",
			V1Differences: map[string]string{"impl/shopfactory_impl.go": v1FromContainer}},
		// factory=extension in a shared package (generateModel): GetFactory
		// ignores the extension, the shared package builds its data type
		// factory extension; the -p packages keyed by nsURI win over the
		// annotation's keyed by name
		fixture{Name: "shop-factory-ext-shared", Model: "shopext.ecore",
			EntryPoints: []string{"generateModel"},
			Props: map[string]string{
				"packageInstance": "interfaces",
				"packages":        shopExtURI + "=github.com/x/ext/shop/shop#github.com/x/ext/typed/typed#github.com/x/ext/shop/shop",
			},
			V1Differences: map[string]string{"typed/shopfactory_impl.go": v1FromContainer}},
		// packages merged from -p and the annotation, both keyed by the name
		// shop: the annotation, merged last, wins
		fixture{Name: "shop-packages-merge", Model: "shopext.ecore",
			Props:         map[string]string{"packages": "github.com/x/cli/shop/shop#github.com/x/cli/impl/impl#github.com/x/cli/mock/mock"},
			V1Differences: map[string]string{"impl/shopfactory_impl.go": v1FromContainer}},
	)
}
