package golang

// The references model: containment lists with and without opposite, a
// single containment whose container reference is not changeable, a single
// containment without opposite, a changeable container, bidirectional single
// and list references, single and list proxies, non-proxy references
// (resolveProxies false), a reference annotated aggregation=shared, and an
// unsettable reference list. An unsettable single bidirectional or
// containment reference is left out: v1 declares the notification chain of
// its unsetter without var, which does not parse.
func init() {
	register(fixture{Name: "references", Model: "references.ecore"})
}
