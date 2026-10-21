package golang

// The operations model: void and value operations; parameters of the any type
// (EJavaObject), list parameters (upperBound -1, of a data type, of the any
// type and of a class) and EEList parameters; operations returning a list,
// an EEList and an EMap; an operation with GenGo default=false; a class with
// only operations (Greeter) and a class extending it without adding anything
// (Polite); a mixin with operations (Circle extends Shape and the interface
// Measurable: EDerivedOperationID); void operations followed by others (the
// glued mock layout).
func init() {
	register(fixture{Name: "operations", Model: "operations.ecore",
		V1Differences: map[string]string{
			"operations/greeter_mock_test.go":    v1MockImport,
			"operations/measurable_mock_test.go": v1MockImport,
		}})
}
