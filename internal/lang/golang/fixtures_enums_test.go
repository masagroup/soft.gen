package golang

// The enums model: two enums sharing a literal name (Red), so the constants
// are qualified by the enum name (ENUM__LITERAL; v1 qualifies them as soon
// as the package has another enum), an enum literal with its own literal
// string, an empty enum, a single-literal enum (its test value is the first
// literal), an attribute with an enum default value literal, and enums with
// an instance type: by the instanceTypeName annotation (Status) and by the
// instance class name of the model (Level).
func init() {
	register(fixture{Name: "enums", Model: "enums.ecore",
		V1Differences: map[string]string{"enums/enumsfactory_impl_test.go": v1EnumLiteralTests}})
}
