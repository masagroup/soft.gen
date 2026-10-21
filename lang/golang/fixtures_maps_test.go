package golang

// The maps model: a map of a map entry class whose key is a list and whose
// value is any-typed, a bidirectional map (the entry class has a container
// reference opposite to the map), and a list of a Map$Entry class that is
// not a map entry class (it has no key and value features).
func init() {
	register(fixture{Name: "maps", Model: "maps.ecore"})
}
