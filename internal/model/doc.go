// Package model computes every language-independent fact about an Ecore model
// that the generators need: the loaded EPackage, the generator annotation of an
// element, the classification of a structural feature (container, containment,
// proxy, list, map, volatile...), the accessors a feature gets (get, set, unset,
// basicGet...), the class hierarchy (base class, extended class, mixins,
// implemented features and operations), the reflective feature sets (eGet, eSet,
// eInverseAdd...), the identifiers and counts of the package constants, the
// class ordering of the package constants, the lazy initialization mode and the
// annotation sources of a package.
//
// It is the port of the common part of the v1 generators (generateCommon.mtl,
// generateElement.mtl, generatePackage.mtl, generateReflectives.mtl,
// generateStructuralFeature.mtl) and of GenerateUtils.java. A language package
// never recomputes one of these facts: it reads them from Model.
//
// Facts are memoized per element: Model.Class(c), Model.Feature(f),
// Model.Operation(op) and Model.Package(p) return the same *Info for the same
// element.
package model
