// Package props holds the properties of a generation run.
//
// A property is declared by a language (Decl: name, default value, documentation
// and scope) and resolved, in this order, from the command line (-p key=value),
// the property files (-P file, the first file that defines a key wins), the
// generator annotation of the root package (for properties with the Package
// scope) and the declared default. An empty command-line or file value does not
// hide the lower layers, as in the v1 generators.
//
// Property files are Java .properties, YAML or JSON files, chosen by extension.
package props
