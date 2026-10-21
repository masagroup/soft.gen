package golang

// The lazy model: the package lazy key "proxy:function list:action" gives
// the lists the action mode (an attribute list, a containment list) and the
// proxies the function mode (a single proxy); feature lazy keys give the
// lambda mode to a list and an attribute, the function mode to an unsettable
// reference list and a single non-proxy reference, the action mode to a
// single proxy, a single non-proxy reference and an attribute renamed by an
// initializationName key.
func init() {
	register(fixture{Name: "lazy", Model: "lazy.ecore"})
}
