package golang

// The metamodel model: a small package named ecore (the IsEcore paths) whose
// GenGo packages key maps its nsURI to the soft.go ecore package, so its
// identifiers and those of Ecore are written bare: the package
// implementation names its factory parameter _ and does not initialize the
// annotation of another source; the tests and the mock tests use no ADD and
// no ecore prefix. Three classes (an abstract one, a subclass, a container)
// with an attribute, a reference and an operation each.
func init() {
	register(fixture{Name: "metamodel", Model: "metamodel.ecore"})
}
