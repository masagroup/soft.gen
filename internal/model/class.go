package model

import (
	"sort"

	"github.com/masagroup/soft.go/ecore"
)

// ClassInfo is the generator facts of a class.
type ClassInfo struct {
	Class ecore.EClass
	// BaseClass is the first super type, nil when the class has none (v1 getBaseClass).
	BaseClass ecore.EClass
	// ExtendsClass is the first non-interface class along the first super type
	// chain, nil when there is none (v1 getClassExtendsClass).
	ExtendsClass ecore.EClass
	// Mixins is the classes the class implements besides ExtendsClass (v1
	// getMixinClasses): the mixins of an interface base class, then the all
	// super types after the base class.
	Mixins []ecore.EClass
	// ImplementedFeatures is eAllStructuralFeatures minus those of ExtendsClass,
	// in eAllStructuralFeatures order (v1 getImplementedFeatures).
	ImplementedFeatures []ecore.EStructuralFeature
	// ImplementedOperations is eAllOperations minus those of ExtendsClass, in
	// eAllOperations order (v1 getImplementedOperations).
	ImplementedOperations []ecore.EOperation
	// AllContainers is the container references of the class sorted by name,
	// followed by those of its super types (v1 getAllContainers).
	AllContainers []ecore.EReference
	// IsMapEntry is whether the class is a map entry: a Map$Entry instance
	// type with a key and a value feature (v1 isMapEntry).
	IsMapEntry bool
	// HasElementID is whether a mixin has an implemented feature or operation
	// (v1 hasElementID).
	HasElementID bool
	// RequiresList is whether an implemented feature or operation is a list or
	// returns an EEList (v1 requiresList).
	RequiresList bool
	// RequiresMap is whether an implemented feature or operation is a map or
	// returns an EMap (v1 requiresMap).
	RequiresMap bool
	// The reflective feature sets (v1 generateReflectives.mtl), in
	// eAllStructuralFeatures order.
	EGetFeatures           []ecore.EStructuralFeature // ImplementedFeatures (v1 getEGetFeatures)
	ESetFeatures           []ecore.EStructuralFeature // the changeable ones (v1 getESetFeatures)
	EIsSetFeatures         []ecore.EStructuralFeature // EGetFeatures (v1 getEIsSetFeatures)
	EUnsetFeatures         []ecore.EStructuralFeature // ESetFeatures (v1 getEUnSetFeatures)
	EInverseAddFeatures    []ecore.EStructuralFeature // the bidirectional non-volatile ones (v1 getEInverseAddFeatures)
	EInverseRemoveFeatures []ecore.EStructuralFeature // the containments and the bidirectional non-volatile ones (v1 getEInverseRemoveFeatures)
	EInvokeOperations      []ecore.EOperation         // ImplementedOperations (v1 getEInvokeOperations)
	// IDs is the identifiers and values of the package constants of the class.
	IDs ClassIDs
}

// Class returns the facts of c, computed once.
func (m *Model) Class(c ecore.EClass) *ClassInfo {
	if info := m.classes[c]; info != nil {
		return info
	}
	info := &ClassInfo{Class: c}
	m.classes[c] = info
	info.BaseClass = firstSuperType(c)
	info.ExtendsClass = m.classExtendsClass(c, info.BaseClass)
	info.Mixins = m.mixinClasses(c)
	info.ImplementedFeatures = m.implementedFeatures(c, info.ExtendsClass)
	info.ImplementedOperations = m.implementedOperations(c, info.ExtendsClass)
	info.AllContainers = m.allContainers(c)
	info.IsMapEntry = m.isMapEntry(c)
	info.HasElementID = m.hasElementID(info.Mixins)
	info.RequiresList = m.requiresList(info)
	info.RequiresMap = m.requiresMap(info)
	info.EGetFeatures = info.ImplementedFeatures
	info.ESetFeatures = m.selectFeatures(info.ImplementedFeatures, func(f *FeatureInfo) bool { return f.IsChangeable })
	info.EIsSetFeatures = info.EGetFeatures
	info.EUnsetFeatures = info.ESetFeatures
	info.EInverseAddFeatures = m.selectFeatures(info.ImplementedFeatures, func(f *FeatureInfo) bool { return f.IsBidirectional && !f.IsVolatile })
	info.EInverseRemoveFeatures = m.selectFeatures(info.ImplementedFeatures, func(f *FeatureInfo) bool {
		return f.IsContainment || (f.IsBidirectional && !f.IsVolatile)
	})
	info.EInvokeOperations = info.ImplementedOperations
	info.IDs = m.classIDs(c)
	return info
}

// classExtendsClass returns the first non-interface class along the first
// super type chain of c starting at base (v1 getClassExtendsClass).
func (m *Model) classExtendsClass(c ecore.EClass, base ecore.EClass) ecore.EClass {
	if c == base {
		return nil
	}
	if base == nil || !base.IsInterface() {
		return base
	}
	return m.classExtendsClass(c, firstSuperType(base))
}

// mixinClasses returns the mixins of c (v1 getMixinClasses): none when c has
// no super type or a single non-interface one; else the mixins of the base
// class when it is an interface, followed by the all super types after the
// base class, without duplicate.
func (m *Model) mixinClasses(c ecore.EClass) []ecore.EClass {
	superTypes := c.GetESuperTypes()
	if superTypes.Empty() || (superTypes.Size() == 1 && !superTypes.Get(0).(ecore.EClass).IsInterface()) {
		return nil
	}
	base := firstSuperType(c)
	var result []ecore.EClass
	if base.IsInterface() {
		result = appendUnique(result, m.Class(base).Mixins...)
	}
	allBases := elements[ecore.EClass](c.GetEAllSuperTypes())
	for i, b := range allBases {
		if b == base {
			result = appendUnique(result, allBases[i+1:]...)
			break
		}
	}
	return result
}

// implementedFeatures returns eAllStructuralFeatures minus those of extends
// (v1 getImplementedFeatures).
func (m *Model) implementedFeatures(c ecore.EClass, extends ecore.EClass) []ecore.EStructuralFeature {
	all := elements[ecore.EStructuralFeature](c.GetEAllStructuralFeatures())
	if extends == nil {
		return all
	}
	inherited := extends.GetEAllStructuralFeatures()
	result := make([]ecore.EStructuralFeature, 0, len(all))
	for _, f := range all {
		if !inherited.Contains(f) {
			result = append(result, f)
		}
	}
	return result
}

// implementedOperations returns eAllOperations minus those of extends (v1
// getImplementedOperations).
func (m *Model) implementedOperations(c ecore.EClass, extends ecore.EClass) []ecore.EOperation {
	all := elements[ecore.EOperation](c.GetEAllOperations())
	if extends == nil {
		return all
	}
	inherited := extends.GetEAllOperations()
	result := make([]ecore.EOperation, 0, len(all))
	for _, op := range all {
		if !inherited.Contains(op) {
			result = append(result, op)
		}
	}
	return result
}

// allContainers returns the container references of c sorted by name followed
// by those of its super types, without duplicate (v1 getAllContainers,
// getContainerAttribute, getSuperTypeContainers).
func (m *Model) allContainers(c ecore.EClass) []ecore.EReference {
	var own []ecore.EReference
	for _, ref := range elements[ecore.EReference](c.GetEReferences()) {
		if ref.IsContainer() {
			own = append(own, ref)
		}
	}
	sort.SliceStable(own, func(i, j int) bool { return own[i].GetName() < own[j].GetName() })
	var inherited []ecore.EReference
	for _, superType := range elements[ecore.EClass](c.GetESuperTypes()) {
		inherited = appendUnique(inherited, m.Class(superType).AllContainers...)
	}
	if len(own) == 0 {
		return inherited
	}
	return appendUnique(own, inherited...)
}

// isMapEntry returns whether c is a map entry class (v1 isMapEntry).
func (m *Model) isMapEntry(c ecore.EClass) bool {
	return m.IsMapEntryName(c.GetInstanceTypeName()) &&
		c.GetEStructuralFeatureFromName("key") != nil &&
		c.GetEStructuralFeatureFromName("value") != nil
}

// hasElementID returns whether a mixin has an implemented feature or
// operation (v1 hasElementID).
func (m *Model) hasElementID(mixins []ecore.EClass) bool {
	for _, mixin := range mixins {
		if info := m.Class(mixin); len(info.ImplementedFeatures) > 0 || len(info.ImplementedOperations) > 0 {
			return true
		}
	}
	return false
}

// requiresList returns whether an implemented feature is a list or an
// implemented operation is a list or returns an EEList (v1 requiresList).
func (m *Model) requiresList(info *ClassInfo) bool {
	for _, f := range info.ImplementedFeatures {
		if m.Feature(f).IsList {
			return true
		}
	}
	for _, op := range info.ImplementedOperations {
		if opInfo := m.Operation(op); opInfo.IsList || classifierNamed(op.GetEType(), "EEList") {
			return true
		}
	}
	return false
}

// requiresMap returns whether an implemented feature is a map or an
// implemented operation is a map or returns an EMap (v1 requiresMap).
func (m *Model) requiresMap(info *ClassInfo) bool {
	for _, f := range info.ImplementedFeatures {
		if m.Feature(f).IsMap {
			return true
		}
	}
	for _, op := range info.ImplementedOperations {
		if opInfo := m.Operation(op); opInfo.IsMap || classifierNamed(op.GetEType(), "EMap") {
			return true
		}
	}
	return false
}

// classifierNamed returns whether c is defined and named name (v1
// requiresList/requiresMap on a classifier).
func classifierNamed(c ecore.EClassifier, name string) bool {
	return c != nil && c.GetName() == name
}

// selectFeatures returns the features whose facts satisfy predicate, in order.
func (m *Model) selectFeatures(features []ecore.EStructuralFeature, predicate func(*FeatureInfo) bool) []ecore.EStructuralFeature {
	result := make([]ecore.EStructuralFeature, 0, len(features))
	for _, f := range features {
		if predicate(m.Feature(f)) {
			result = append(result, f)
		}
	}
	return result
}
