package golang

// The annotations model: annotations of other sources than GenGo, with
// details, on the package, a class, an operation, an operation parameter, an
// attribute and an enum literal; annotations nested in an annotation of a
// class, in its eAnnotations and two levels of its contents, created with it
// as in EMF; an object of its contents of a package of the xsi:schemaLocation
// of the model (notes.ecore), loaded and left out as in EMF; and an
// annotation of a source the package suppresses (GenGo supressAnnotations),
// left out of the package initialization.
func init() {
	register(fixture{Name: "annotations", Model: "annotations.ecore",
		V1Differences: map[string]string{"annotations/annotationspackage_impl.go": v1AnnotationContainer}})
}
