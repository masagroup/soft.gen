package model

import (
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
)

func TestLazyEcore(t *testing.T) {
	m := loadModel(t, "ecore")
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

func TestLazyLibrary(t *testing.T) {
	m := loadModel(t, "library")
	for _, c := range m.Package(m.Root()).Classes {
		for _, f := range elements[ecore.EStructuralFeature](c.GetEStructuralFeatures()) {
			assert.Equal(t, "no", m.Feature(f).Lazy.Mode, "%s.%s", c.GetName(), f.GetName())
			assert.False(t, m.Feature(f).Lazy.HasInitMethod)
		}
	}
}

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
