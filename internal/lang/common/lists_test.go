package common

import (
	"strings"
	"testing"

	"github.com/masagroup/soft.go/ecore"
	"github.com/stretchr/testify/assert"
)

func TestTokens(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, Tokens("a,b;c d\r\ne"))
	assert.Empty(t, Tokens(" ;, "))
}

func TestPackageSpecs(t *testing.T) {
	specs := ParsePackageSpecs("http://a=x/a#x/impl;y/b#y/c, z", map[string]string{"http://e": "ecore"})
	pkg := func(uri, name string) ecore.EPackage {
		p := ecore.GetFactory().CreateEPackage()
		p.SetNsURI(uri)
		p.SetName(name)
		return p
	}

	spec, ok := specs.Spec(pkg("http://a", "other"))
	assert.True(t, ok)
	assert.Equal(t, "x/a#x/impl", spec, "by nsURI")
	spec, _ = specs.Spec(pkg("http://unknown", "b"))
	assert.Equal(t, "y/b#y/c", spec, "by the last segment of the first part")
	spec, _ = specs.Spec(pkg("http://e", "ecore"))
	assert.Equal(t, "ecore", spec, "the defaults")
	_, ok = specs.Spec(pkg("http://none", "none"))
	assert.False(t, ok)

	parts := strings.Split("x/a#x/impl", "#")
	assert.Equal(t, "x/a", IndexPart(parts, 1))
	assert.Equal(t, "x/impl", IndexPart(parts, 2))
	assert.Equal(t, "x/impl", IndexPart(parts, 3), "a missing index gives the last part")
}

func TestLastSegment(t *testing.T) {
	assert.Equal(t, "c", LastSegment("a/b/c"))
	assert.Equal(t, "abc", LastSegment("abc"))
}
