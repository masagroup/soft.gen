package common

import (
	"strings"

	"github.com/masagroup/soft.go/ecore"
)

// listSeparators separate the tokens of a list property (packages, paths,
// namespaces): ",", ";", spaces and newlines (v1 tokenize(',; \r\n')).
const listSeparators = ",; \r\n"

// Tokens splits a list property into its tokens.
func Tokens(list string) []string {
	return strings.FieldsFunc(list, func(r rune) bool { return strings.ContainsRune(listSeparators, r) })
}

// ListProperty returns a list property merged from its two sources, the
// command line (or a property file) then the package annotation, separated by
// ";": v1 concatenates them instead of letting the first win.
func (g Gen) ListProperty(key string) string {
	list := ""
	if cli, ok := g.Props.CLI(key); ok {
		list = cli
	}
	if annotated, ok := g.Props.Annotation(key); ok {
		list += ";" + annotated
	}
	return list
}

// PackageSpecs maps the packages of a model to their specification in the
// packages property: "<first>#<second>#...", the paths (Go) or the modules
// (TypeScript) of the package, by nsURI or by name (v1 getAllPackagePaths,
// getPackageToModuleMap).
type PackageSpecs map[string]string

// ParsePackageSpecs parses the tokens of the packages property: "uri=spec"
// keys the spec by nsURI, a bare "spec" by the last path segment of its
// first part. The defaults (the ecore package) come first.
func ParsePackageSpecs(list string, defaults map[string]string) PackageSpecs {
	specs := PackageSpecs{}
	for uri, spec := range defaults {
		specs["uri:"+uri] = spec
	}
	for _, token := range Tokens(list) {
		if uri, spec, ok := strings.Cut(token, "="); ok {
			specs["uri:"+uri] = spec
			continue
		}
		first, _, _ := strings.Cut(token, "#")
		specs["name:"+LastSegment(first)] = token
	}
	return specs
}

// Spec returns the spec of the package: the one found by its nsURI, else by
// its name. ok is false when the package has none.
func (s PackageSpecs) Spec(p ecore.EPackage) (spec string, ok bool) {
	spec, ok = s["uri:"+p.GetNsURI()]
	if !ok {
		spec, ok = s["name:"+p.GetName()]
	}
	return spec, ok
}

// IndexPart returns the part at a 1-based index of the parts of a spec (v1
// getIndexPath, getIndexModule): a missing index gives the last part. The
// parts are the spec split on "#", which each language does as its v1
// generator did.
func IndexPart(parts []string, index int) string {
	if index > len(parts) {
		return parts[len(parts)-1]
	}
	return parts[index-1]
}

// LastSegment returns what follows the last "/" of a path, or the path.
func LastSegment(path string) string {
	return path[strings.LastIndex(path, "/")+1:]
}
