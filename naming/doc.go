// Package naming holds the string transformations the generators apply to the
// names of the model: the EMF word splitting (eAllStructuralFeatures ->
// EALL_STRUCTURAL_FEATURES), the first-letter case changes, the naming rules
// of the generator annotations (pattern:replacement with $1 groups, or {0}
// message formats) and the keyword avoidance.
//
// It has no dependency on the model: a name rule is resolved from a lookup
// function that the caller binds to the annotation of an element.
package naming
