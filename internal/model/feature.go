package model

import "github.com/masagroup/soft.go/ecore"

// FeatureInfo is the generator facts of a structural feature.
type FeatureInfo struct {
	Feature   ecore.EStructuralFeature
	Class     ecore.EClass     // eContainingClass
	Reference ecore.EReference // nil for an attribute
	Attribute ecore.EAttribute // nil for a reference
	Opposite  ecore.EReference // eOpposite, nil when none

	IsContainer     bool // the opposite is a containment (v1 isContainer)
	IsContainment   bool // a containment reference (v1 isContains)
	IsBidirectional bool // a reference with an opposite (v1 isBidirectional)
	IsChangeable    bool // changeable (v1 isChangeable)
	IsUnsettable    bool // unsettable and not a container (v1 isUnSettable)
	IsVolatile      bool // volatile, or the opposite is (v1 isVolatile)
	IsList          bool // many or upperBound != 1 (v1 isListType)
	IsMap           bool // a list whose type is a Map$Entry (v1 isMapType)
	IsFeatureMap    bool // the type is a FeatureMap$Entry (v1 isFeatureMapType)
	IsProxy         bool // a resolving reference, neither container nor containment (v1 isProxy)
	IsField         bool // neither container nor volatile (v1 isField)
	IsBackReference bool // a container, or a reference annotated aggregation=shared (v1 isBackReference, go generateType.mtl)
	HasESet         bool // an unsettable single field: a flag tells whether it is set (v1 hasESetFlag, EMF xxxESet)

	// Kind is the kind of the feature: "attribute", "reference", "containment
	// reference", "container reference", "map", with " list" appended to a
	// many non-map feature (v1 getFeatureKind).
	Kind string

	// The accessors of the feature (v1 isGet, isSet, isUnSet, isIsSet,
	// isBasicGet, isBasicSet, isBasicUnSet).
	Get, Set, Unset, IsSet, BasicGet, BasicSet, BasicUnset bool

	// Lazy is the lazy initialization of the feature.
	Lazy LazyInfo

	// RequiresNotification is whether a set or an unset accessor exists (v1
	// requiresNotification).
	RequiresNotification bool
}

// Feature returns the facts of f, computed once.
func (m *Model) Feature(f ecore.EStructuralFeature) *FeatureInfo {
	if info := m.features[f]; info != nil {
		return info
	}
	info := &FeatureInfo{Feature: f, Class: f.GetEContainingClass()}
	m.features[f] = info
	info.Reference, _ = f.(ecore.EReference)
	info.Attribute, _ = f.(ecore.EAttribute)
	if info.Reference != nil {
		info.Opposite = info.Reference.GetEOpposite()
		info.IsContainer = info.Opposite != nil && info.Opposite.IsContainment()
		info.IsContainment = info.Reference.IsContainment()
		info.IsBidirectional = info.Opposite != nil
	}
	info.IsChangeable = f.IsChangeable()
	info.IsUnsettable = f.IsUnsettable() && !info.IsContainer
	info.IsVolatile = f.IsVolatile() || (info.Opposite != nil && info.Opposite.IsVolatile())
	info.IsList = isListType(f)
	info.IsMap = m.isMapType(f)
	info.IsFeatureMap = f.GetEType() != nil && m.IsFeatureMapEntryName(f.GetEType().GetInstanceTypeName())
	info.IsProxy = info.Reference != nil && !info.IsContainer && !info.IsContainment && info.Reference.IsResolveProxies()
	info.IsField = !info.IsContainer && !info.IsVolatile
	info.IsBackReference = info.Reference != nil && (info.IsContainer || m.Annotation(f).Value("aggregation") == "shared")
	info.HasESet = info.IsUnsettable && info.IsField && !info.IsList
	info.Kind = m.FeatureKind(f)

	info.Get = true
	info.Set = info.IsChangeable && f.GetUpperBound() == 1
	info.Unset = info.IsChangeable && info.IsUnsettable
	info.IsSet = info.IsUnsettable
	info.BasicSet = !info.IsList && (info.IsChangeable || !info.IsContainer) && ((info.IsBidirectional && !info.IsVolatile) || info.IsContainment)
	info.BasicGet = info.IsProxy && !info.IsList
	info.BasicUnset = info.IsUnsettable && info.IsChangeable && !info.IsList && info.Reference != nil && (info.IsBidirectional || info.IsContainment)
	info.RequiresNotification = info.Set || info.Unset

	info.Lazy = m.lazyInfo(info)
	return info
}

// isListType returns whether e is many valued: many or upperBound != 1 (v1 isListType).
func isListType(e ecore.ETypedElement) bool {
	return e != nil && (e.IsMany() || e.GetUpperBound() != 1)
}

// isMapType returns whether e is a list whose type is a map entry (v1 isMapType).
func (m *Model) isMapType(e ecore.ETypedElement) bool {
	return isListType(e) && e.GetEType() != nil && m.IsMapEntryName(e.GetEType().GetInstanceTypeName())
}

// IsListType reports whether a typed element (a feature, an operation, a
// parameter) is many valued (v1 isListType).
func IsListType(e ecore.ETypedElement) bool { return isListType(e) }

// IsMapType reports whether a typed element is a list of map entries (v1
// isMapType).
func (m *Model) IsMapType(e ecore.ETypedElement) bool { return m.isMapType(e) }

// FeatureKind returns the kind of f (v1 getFeatureKind): "attribute",
// "reference", "containment reference", "container reference" or "map", with
// " list" appended when f is many and not a map.
func (m *Model) FeatureKind(f ecore.EStructuralFeature) string {
	kind := "attribute"
	if ref, isRef := f.(ecore.EReference); isRef {
		switch {
		case ref.IsContainment():
			if m.isMapType(f) {
				kind = "map"
			} else {
				kind = "containment reference"
			}
		case ref.IsContainer():
			kind = "container reference"
		default:
			kind = "reference"
		}
	}
	if kind != "map" && f.IsMany() {
		kind += " list"
	}
	return kind
}
