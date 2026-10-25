package golang

// The attributes model: a data type for every factory conversion kind
// (string, bool, int, int64, int32 and int8 through the instanceTypeName
// annotation, int16, byte, []byte, float64, float32, reflect.Type, time.Time,
// unsupported for java.lang.Object and an unknown class), data types with the
// defaultValue, testValue and instanceTypeName annotations, an EChar data
// type (no Java primitive default) next to the same type under another name,
// a non serializable data type; and a class with a string attribute with a
// default value literal, a boolean attribute, attributes named as Go keywords
// (type, func, range), an unsettable single attribute, an unsettable and a
// plain attribute list, a feature map list, an EJavaObject attribute.
// The quoted default literal and the attribute of an unknown Go type, which v1
// cannot generate, are in the v1bugs model.
func init() {
	register(fixture{Name: "attributes", Model: "attributes.ecore",
		V1Differences: map[string]string{
			"attributes/holder_impl_test.go": v1GoIntTestValue,
			"attributes/holder_mock_test.go": v1GoIntTestValue,
		}})
}

// The attributes model with isSetFlags=true: every changeable single attribute
// held in a field records that it was set; EIsSet reads the flag, the setter
// raises it, the unsetter and EUnset lower it.
func init() {
	register(fixture{Name: "attributes-isset", Model: "attributes.ecore",
		Props:  map[string]string{"isSetFlags": "true"},
		V1Skip: "v1 has no isSetFlags property"})
}
