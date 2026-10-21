package golang

// The v1bugs model holds the constructs v1 cannot generate, so its expected
// files are v2's output, reviewed (the v1 cross-check skips its cases):
//
//   - an unsettable single containment and an unsettable bidirectional single
//     reference: v1's unsetter and basic unsetter declare the notification
//     chain without var, which does not parse;
//   - an attribute of a Go type the type table does not know, without
//     default value: v1 writes "x != " in EIsSetFromID and an empty expected
//     value in the getter test; v2 uses the zero value *new(T);
//   - a string default literal with quotes and a backslash: v1 does not
//     escape it in the package initialization.
//
// The v1bugs-rootmock case adds the rootExtendsMock property, with which v1
// fails generating the mocks ("Undefined argument packages").
func init() {
	register(
		fixture{Name: "v1bugs", Model: "v1bugs.ecore",
			V1Skip: "v1's output does not parse: the unsetters of unsettable single inverse references, the unknown-type attribute, the quoted default literal"},
		fixture{Name: "v1bugs-rootmock", Model: "v1bugs.ecore",
			EntryPoints: []string{"generateMocks", "generateMocksTests"},
			Props:       map[string]string{"rootExtendsMock": "github.com/x/base/base.MockBase"},
			V1Skip:      "v1 fails generating the mocks with rootExtendsMock (Undefined argument packages)"},
	)
}
