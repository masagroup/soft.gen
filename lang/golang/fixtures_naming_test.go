package golang

// The naming model: every naming rule key set on the package (inherited by
// its classes, features and operations), in both forms, "pattern:replacement"
// and a bare replacement with $1, and overridden on classes, features and an
// operation; an exported classifier in an unexported package; extension
// classes, one extending the other, one renaming its extension.
func init() {
	register(
		// the default entry points: the rules in every file name and identifier
		fixture{Name: "naming", Model: "naming.ecore",
			V1Differences: map[string]string{"naming/namingbuilder_struct.go": v1FromContainer}},
		// the package exported (-p exported=true): every implementation name
		// upper-cased first
		fixture{Name: "naming-exported", Model: "naming.ecore",
			Props:         map[string]string{"exported": "true"},
			V1Differences: map[string]string{"naming/namingbuilder_struct.go": v1FromContainer}},
	)
}
