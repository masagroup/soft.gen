package golang

import (
	"bytes"
	"sort"
	"strings"
)

// Placeholders delimiting a package path inside a rendered body; they are
// replaced by the alias of the package once every used package is known.
const (
	placeholderStart = "\x01"
	placeholderEnd   = "\x02"
)

// fileScope resolves the qualified identifiers of one generated file: an
// identifier of the file's own package is bare, an identifier of another
// package is prefixed by the alias of that package, and every package used
// is imported. The aliases follow the v1 algorithm (the last segment of the
// path, extended with the previous segments on conflicts), which needs the
// whole set of used packages: qualify returns the identifier with a
// placeholder for the alias, and finish replaces the placeholders once the
// body is rendered.
type fileScope struct {
	packagePath string
	used        map[string]bool
}

// newFileScope starts the scope of a file of the package at packagePath.
func newFileScope(packagePath string) *fileScope {
	return &fileScope{packagePath: packagePath, used: map[string]bool{}}
}

// qualify turns a qualified identifier ("<path>/<pkg>.<Ident>", "*" prefix
// allowed) into the identifier the file uses, recording the import (v1
// getImportedIdentifier). A string without a package path (a literal such as
// "0.0" or `"x"`) is returned unchanged.
func (s *fileScope) qualify(qualified string) string {
	dot := strings.LastIndex(qualified, ".")
	if dot == -1 {
		return qualified
	}
	prefix, ident := qualified[:dot], qualified[dot+1:]
	ptr := ""
	if strings.HasPrefix(prefix, "*") {
		ptr, prefix = "*", prefix[1:]
	}
	if !strings.Contains(prefix, "/") {
		return qualified
	}
	if prefix == s.packagePath {
		return ptr + ident
	}
	s.used[prefix] = true
	return ptr + placeholderStart + prefix + placeholderEnd + "." + ident
}

// importSpec is one line of an import block.
type importSpec struct {
	// Alias is the import alias; "" when the package name of the path is used.
	Alias string
	// Dir is the import path.
	Dir string
}

// finish replaces the placeholders of a rendered body by the aliases of the
// used packages and returns the import block to write before it.
func (s *fileScope) finish(body []byte) ([]byte, []importSpec) {
	paths := make([]string, 0, len(s.used))
	for p := range s.used {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	aliases := computeImportAliases(paths)
	for _, p := range paths {
		body = bytes.ReplaceAll(body, []byte(placeholderStart+p+placeholderEnd), []byte(aliases[p]))
	}
	var specs []importSpec
	for _, p := range paths {
		dir := packageDirectoryFromPath(p)
		spec := importSpec{Dir: dir}
		if aliases[p] != packageNameFromPath(dir) {
			spec.Alias = aliases[p]
		}
		specs = append(specs, spec)
	}
	return body, specs
}

// computeImportAliases returns the alias of each package path (v1
// GeneratePath.getPackageImports): the last segment of the path, extended
// with the previous segments, joined by "_", while another path has the same
// segments; "." becomes "_".
func computeImportAliases(paths []string) map[string]string {
	split := map[string][]string{}
	for _, p := range paths {
		split[p] = splitPath(p)
	}
	names := map[string][]string{}
	remaining := append([]string(nil), paths...)
	for index := 0; len(remaining) > 0; index++ {
		segment := map[string]string{}
		for _, p := range paths {
			if index < len(split[p]) {
				segment[p] = split[p][index]
			}
		}
		for i := len(remaining) - 1; i >= 0; i-- {
			p := remaining[i]
			name := segment[p]
			if name == "" {
				remaining = append(remaining[:i], remaining[i+1:]...)
				continue
			}
			other := false
			for q, n := range segment {
				if n == name && q != p {
					other = true
					break
				}
			}
			names[p] = append(names[p], name)
			if !other {
				remaining = append(remaining[:i], remaining[i+1:]...)
			}
		}
	}
	aliases := map[string]string{}
	for p, segments := range names {
		reversed := make([]string, len(segments))
		for i, s := range segments {
			reversed[len(segments)-1-i] = s
		}
		aliases[p] = strings.ReplaceAll(strings.Join(reversed, "_"), ".", "_")
	}
	return aliases
}

// splitPath splits a package path on "/", drops the consecutive duplicates
// ("ecore/ecore" is one segment) and reverses the segments.
func splitPath(path string) []string {
	var segments []string
	last := ""
	for i, s := range strings.Split(path, "/") {
		if i == 0 || s != last {
			segments = append(segments, s)
		}
		last = s
	}
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}
	return segments
}

// simpleIdentifier returns the identifier of a qualified identifier (v1 getSimpleIdentifier).
func simpleIdentifier(qualified string) string {
	if i := strings.LastIndex(qualified, "."); i != -1 {
		return qualified[i+1:]
	}
	return qualified
}

// identifierPackagePath returns the package path of a qualified identifier
// ("" without one), ignoring a "*" prefix (v1 getIdentifierPackagePath).
func identifierPackagePath(qualified string) string {
	qualified = strings.TrimPrefix(qualified, "*")
	if i := strings.LastIndex(qualified, "."); i != -1 {
		return qualified[:i]
	}
	return ""
}

// qualifiedIdentifier builds "<package path>.<identifier>".
func qualifiedIdentifier(packagePath, identifier string) string {
	return packagePath + "." + identifier
}
