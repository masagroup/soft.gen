package golang

// The maps model: a map of a map entry class whose key is a list and whose
// value is any-typed, a bidirectional map (the entry class has a container
// reference opposite to the map), and a list of a Map$Entry class that is
// not a map entry class (it has no key and value features).
func init() {
	register(fixture{Name: "maps", Model: "maps.ecore"})
}

// The maps model with synchronizedLists=true: the list and map getters build
// their field under the mutex of the object, and EIsSet reads them through
// their getter.
func init() {
	register(fixture{Name: "maps-synchronized", Model: "maps.ecore",
		Props:  map[string]string{"synchronizedLists": "true"},
		V1Skip: "v1 has no synchronizedLists property"})
}
