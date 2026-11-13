package golang

// The wide model: the hybrid classes of a class of 2 attributes, whose typed fields have the
// 8-byte ecore.FieldsHeader, and of a class of 33 attributes, more than a FieldsHeader holds,
// whose typed fields have the 16-byte ecore.WideFieldsHeader.
func init() {
	register(fixture{Name: "wide-hybrid", Model: "wide.ecore",
		EntryPoints: []string{"generateImplementations"},
		Props:       map[string]string{"storage": "hybrid"},
		V1Skip:      "v1 has no storage property"})
}
