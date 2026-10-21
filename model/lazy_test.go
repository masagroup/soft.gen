package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
)

// TestLazyEcore checks the lazy rules on the real Ecore features (package rule
// lazy=list:function, feature keys lazy=action).
func TestLazyEcore(t *testing.T) {
	m := ecoreMetamodel(t)
	// the package rule lazy=list:function applies to the lists
	eSuperTypes := m.Feature(featureNamed(t, m, "EClass", "eSuperTypes")).Lazy
	assert.Equal(t, LazyInfo{Mode: "function", HasInitMethod: true, InitMethodName: "initESuperTypes", Deferred: true}, eSuperTypes)
	assert.True(t, eSuperTypes.IsLazy())
	// a feature annotation lazy=action wins
	eAttributes := m.Feature(featureNamed(t, m, "EClass", "eAttributes")).Lazy
	assert.Equal(t, LazyInfo{Mode: "action", HasInitMethod: true, InitMethodName: "initEAttributes", Deferred: true}, eAttributes)
	// a single attribute and a single proxy have no rule
	abstract := m.Feature(featureNamed(t, m, "EClass", "abstract")).Lazy
	assert.Equal(t, LazyInfo{Mode: "no", InitMethodName: "initAbstract"}, abstract)
	assert.False(t, abstract.IsLazy())
	eOpposite := m.Feature(featureNamed(t, m, "EReference", "eOpposite")).Lazy
	assert.Equal(t, LazyInfo{Mode: "no", InitMethodName: "initEOpposite", Deferred: true}, eOpposite)
	// a derived non volatile list with lazy=action has an initialization method
	eAllAttributes := m.Feature(featureNamed(t, m, "EClass", "eAllAttributes"))
	assert.False(t, eAllAttributes.IsVolatile)
	assert.False(t, eAllAttributes.IsChangeable)
	assert.Equal(t, LazyInfo{Mode: "action", HasInitMethod: true, InitMethodName: "initEAllAttributes", Deferred: true}, eAllAttributes.Lazy)
}

// TestLazyWithoutAnnotation checks that without a lazy key nothing is lazy:
// every feature of references.ecore (lists, proxies, containments,
// containers) has the "no" mode and no initialization method.
func TestLazyWithoutAnnotation(t *testing.T) {
	m := loadModel(t, "references")
	for _, c := range m.Package(m.Root()).Classes {
		for _, f := range elements[ecore.EStructuralFeature](c.GetEStructuralFeatures()) {
			assert.Equal(t, "no", m.Feature(f).Lazy.Mode, "%s.%s", c.GetName(), f.GetName())
			assert.False(t, m.Feature(f).Lazy.HasInitMethod)
		}
	}
}

// TestLazyPackageRule checks the lazy tokens of lazy.ecore: the package rule
// "proxy:function list:action" applies to the proxies and the lists without
// a lazy key, a feature lazy key wins, a single non-proxy feature has no rule;
// a list or a proxy is deferred, an initializationName key names the method.
func TestLazyPackageRule(t *testing.T) {
	m := loadModel(t, "lazy")
	tests := []struct {
		class, feature string
		expected       LazyInfo
	}{
		{"Holder", "tags", LazyInfo{Mode: "action", HasInitMethod: true, InitMethodName: "initTags", Deferred: true}},
		{"Holder", "items", LazyInfo{Mode: "action", HasInitMethod: true, InitMethodName: "initItems", Deferred: true}},
		{"Holder", "links", LazyInfo{Mode: "lambda", InitMethodName: "initLinks", Deferred: true}},
		{"Holder", "cache", LazyInfo{Mode: "function", HasInitMethod: true, InitMethodName: "initCache", Deferred: true}},
		{"Holder", "favorite", LazyInfo{Mode: "function", HasInitMethod: true, InitMethodName: "initFavorite", Deferred: true}},
		{"Holder", "preferred", LazyInfo{Mode: "action", HasInitMethod: true, InitMethodName: "initPreferred", Deferred: true}},
		{"Holder", "computed", LazyInfo{Mode: "function", HasInitMethod: true, InitMethodName: "initComputed"}},
		{"Holder", "assigned", LazyInfo{Mode: "action", HasInitMethod: true, InitMethodName: "initAssigned"}},
		{"Holder", "title", LazyInfo{Mode: "action", HasInitMethod: true, InitMethodName: "computeTitle"}},
		{"Holder", "note", LazyInfo{Mode: "lambda", InitMethodName: "initNote"}},
		{"Item", "name", LazyInfo{Mode: "no", InitMethodName: "initName"}},
	}
	for _, test := range tests {
		t.Run(test.class+"."+test.feature, func(t *testing.T) {
			lazy := m.Feature(featureNamed(t, m, test.class, test.feature)).Lazy
			assert.Equal(t, test.expected, lazy)
			assert.Equal(t, test.expected.Mode != "no", lazy.IsLazy())
		})
	}
}

// TestPackageLazyMode checks the parsing of the [type:]mode tokens of the package lazy key.
func TestPackageLazyMode(t *testing.T) {
	p := ecore.GetFactory().CreateEPackage()
	p.SetName("p")
	m := FromPackage(p, genGoSource)
	assert.Equal(t, "no", m.packageLazyMode(p, "list"))
	assert.Equal(t, "no", m.packageLazyMode(nil, "list"))

	annotation := ecore.GetFactory().CreateEAnnotationFromContainer(p)
	annotation.SetSource(genGoSource)
	set := func(value string) {
		annotation.GetDetails().Put("lazy", value)
		m = FromPackage(p, genGoSource)
	}
	set("list:function")
	assert.Equal(t, "function", m.packageLazyMode(p, "list"))
	assert.Equal(t, "no", m.packageLazyMode(p, "proxy"))
	assert.Equal(t, "no", m.packageLazyMode(p, ""))

	set("list:function proxy:lambda :action")
	assert.Equal(t, "function", m.packageLazyMode(p, "list"))
	assert.Equal(t, "lambda", m.packageLazyMode(p, "proxy"))
	assert.Equal(t, "action", m.packageLazyMode(p, ""))

	// a token without type applies to any type, the first applying token wins
	set("proxy:lambda function list:action")
	assert.Equal(t, "lambda", m.packageLazyMode(p, "proxy"))
	assert.Equal(t, "function", m.packageLazyMode(p, "list"))
	assert.Equal(t, "function", m.packageLazyMode(p, ""))

	// '*' is matched verbatim, not as a wild card (v1)
	set("*:function")
	assert.Equal(t, "no", m.packageLazyMode(p, "list"))
	assert.Equal(t, "no", m.packageLazyMode(p, ""))
	assert.Equal(t, "function", m.packageLazyMode(p, "*"))

	set("")
	assert.Equal(t, "no", m.packageLazyMode(p, "list"))
}

// TestLazyFeatureAnnotation checks the feature lazy and initializationName keys.
func TestLazyFeatureAnnotation(t *testing.T) {
	p := ecore.GetFactory().CreateEPackage()
	p.SetName("p")
	c := ecore.GetFactory().CreateEClassFromContainer(p)
	c.SetName("C")
	f := ecore.GetFactory().CreateEAttributeFromContainer(c)
	f.SetName("value")
	annotation := ecore.GetFactory().CreateEAnnotationFromContainer(f)
	annotation.SetSource(genGoSource)
	annotation.GetDetails().Put("lazy", "lambda")
	annotation.GetDetails().Put("initializationName", "computeValue")
	m := FromPackage(p, genGoSource)
	lazy := m.Feature(f).Lazy
	assert.Equal(t, LazyInfo{Mode: "lambda", HasInitMethod: false, InitMethodName: "computeValue", Deferred: false}, lazy)
	assert.True(t, lazy.IsLazy())

	annotation.GetDetails().Put("lazy", "function")
	m = FromPackage(p, genGoSource)
	assert.True(t, m.Feature(f).Lazy.HasInitMethod)

	// a volatile lazy feature has no initialization method
	f.SetVolatile(true)
	m = FromPackage(p, genGoSource)
	assert.Equal(t, "function", m.Feature(f).Lazy.Mode)
	assert.False(t, m.Feature(f).Lazy.HasInitMethod)
}
