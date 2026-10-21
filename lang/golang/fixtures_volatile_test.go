package golang

// The volatile model: volatile unsettable attributes with getterBody,
// setterBody and unsetterBody, and without bodies; volatile single proxy
// references with and without bodies (basic getter body and panic); volatile
// unsettable single containments with and without bodies (basic setter body
// and panic, basic unsetter panic, inverse remove through the unsetter).
func init() {
	register(fixture{Name: "volatile", Model: "volatile.ecore"})
}
