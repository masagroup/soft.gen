package model

import (
	"strings"

	"github.com/masagroup/soft.go/ecore"
)

// LazyInfo is the lazy initialization of a feature (v1 generateStructuralFeature.mtl).
type LazyInfo struct {
	// Mode is the lazy mode: "no", "lambda", "function" or "action" (v1
	// getLazy): the `lazy` key of the feature annotation, else the rule of the
	// `lazy` key of the package annotation for the feature (list, proxy or
	// other), else "no".
	Mode string
	// HasInitMethod is whether the feature is initialized by a method: not
	// volatile, with the "action" or "function" mode (v1 hasInitializationMethod).
	HasInitMethod bool
	// InitMethodName is the name of the initialization method: the
	// `initializationName` key, else "init" + the name upper first (v1
	// getInitializationMethodName).
	InitMethodName string
	// Deferred is whether the initialization is deferred: a list or a proxy
	// (v1 isInitializationDefered).
	Deferred bool
}

// IsLazy returns whether the mode is not "no" (v1 isLazy; v1 isNotLazy is its negation).
func (l LazyInfo) IsLazy() bool {
	return l.Mode != "no"
}

// lazyInfo computes the lazy initialization of a feature.
func (m *Model) lazyInfo(info *FeatureInfo) LazyInfo {
	mode := m.lazyMode(info)
	lazy := LazyInfo{Mode: mode}
	lazy.HasInitMethod = !info.IsVolatile && mode != "no" && (mode == "action" || mode == "function")
	lazy.InitMethodName = m.initMethodName(info.Feature)
	lazy.Deferred = info.IsList || info.IsProxy
	return lazy
}

// lazyMode returns the lazy mode of a feature (v1 getLazy on a feature): the
// `lazy` key of its annotation, else the package rule for "list" when the
// feature is a list, "proxy" when it is a proxy, "" otherwise.
func (m *Model) lazyMode(info *FeatureInfo) string {
	if annotation := m.Annotation(info.Feature); annotation.HasKey("lazy") {
		return annotation.Value("lazy")
	}
	var p ecore.EPackage
	if info.Class != nil {
		p = info.Class.GetEPackage()
	}
	switch {
	case info.IsList:
		return m.packageLazyMode(p, "list")
	case info.IsProxy:
		return m.packageLazyMode(p, "proxy")
	default:
		return m.packageLazyMode(p, "")
	}
}

// packageLazyMode returns the lazy mode of the `lazy` key of the package
// annotation for a feature type (v1 getLazy on a package). The key is a
// whitespace separated list of `[type:]mode` tokens, read in order: the first
// token without a type applies to any type, the first token whose type is the
// requested one applies to it; a type is matched verbatim, so the `*` the v1
// comment documents is not a wild card. The mode is "no" when no token
// applies, or the package has no `lazy` key.
func (m *Model) packageLazyMode(p ecore.EPackage, featureType string) string {
	if p == nil {
		return "no"
	}
	annotation := m.Annotation(p)
	if !annotation.HasKey("lazy") {
		return "no"
	}
	value := ""
	for _, token := range strings.Fields(annotation.Value("lazy")) {
		if index := strings.Index(token, ":"); index == -1 {
			value = token
		} else if token[:index] == featureType {
			value = token[index+1:]
		}
		if value != "" {
			break
		}
	}
	if value == "" {
		return "no"
	}
	return value
}

// initMethodName returns the name of the initialization method of f (v1
// getInitializationMethodName).
func (m *Model) initMethodName(f ecore.EStructuralFeature) string {
	if annotation := m.Annotation(f); annotation.HasKey("initializationName") {
		return annotation.Value("initializationName")
	}
	return "init" + upperFirst(f.GetName())
}
