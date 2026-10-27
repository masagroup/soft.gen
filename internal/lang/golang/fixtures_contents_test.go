package golang

// The contents model: with the property annotationContents, the objects of
// the contents of an annotation of a class are created with it. A Memo of a
// package of the xsi:schemaLocation of the model (memo.ecore), created by the
// factory of its package, with attributes of every literal kind (a string
// with quotes and a newline, an int, a double, a boolean, an enum, lists of
// strings, shorts and enums) and objects contained by a single and a
// many-valued containment, two levels deep; then an annotation of the
// contents, with an object of its own contents.

// The nsURIs of the contents and memo packages.
const (
	contentsURI = "http://soft.gen/test/contents"
	memoURI     = "http://soft.gen/test/memo"
)

// v1AnnotationContents is why the package of the contents model differs from v1.
const v1AnnotationContents = "v1 drops the objects of the contents of an annotation: it has no annotationContents"

func init() {
	register(fixture{Name: "contents", Model: "contents.ecore",
		Props: map[string]string{
			"module":   "github.com/x",
			"packages": contentsURI + "=github.com/x/contents/contents " + memoURI + "=github.com/x/memo/memo",
		},
		V1Differences: map[string]string{"contents/contentspackage_impl.go": v1AnnotationContents},
	})
}
