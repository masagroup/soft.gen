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

// The annotations model with loadInitialization (EMF Initialize by Loading):
// the package loads its serialized .ecore, embedded beside it, whose
// annotations keep their contents, the Note of the notes package included;
// the annotations of the suppressed source are left out of it, and the notes
// package is registered for the loading.
func init() {
	register(fixture{Name: "annotations-loading", Model: "annotations.ecore",
		Props: map[string]string{
			"loadInitialization": "true",
			"packages":           "http://soft.gen/test/notes=github.com/x/notes/notes",
		},
		V1Skip: v1NoLoadInitialization})
}
