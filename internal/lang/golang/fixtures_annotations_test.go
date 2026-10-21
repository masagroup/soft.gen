package golang

// The annotations model: annotations of other sources than GenGo, with
// details, on the package, a class, an operation, an operation parameter, an
// attribute and an enum literal; an annotation nested in an annotation of a
// class; and an annotation of a source the package suppresses (GenGo
// supressAnnotations), left out of the package initialization.
func init() {
	register(fixture{Name: "annotations", Model: "annotations.ecore",
		V1Differences: map[string]string{"annotations/annotationspackage_impl.go": v1AnnotationContainer}})
}
