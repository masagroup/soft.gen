package props

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// format parses one property file format. Adding a format is one file
// implementing it and one register call in that file's init.
type format interface {
	// name is the format name given to ParsePropertiesReader.
	name() string
	// extensions lists the file extensions, with the dot, mapped to the format.
	extensions() []string
	// parse reads a whole file and returns its flat key/value mapping.
	parse(r io.Reader) (map[string]string, error)
}

var (
	formatsByName = map[string]format{}
	formatsByExt  = map[string]format{}
)

// register adds f to the formats known by name and by extension.
func register(f format) {
	formatsByName[f.name()] = f
	for _, ext := range f.extensions() {
		formatsByExt[ext] = f
	}
}

// Formats lists the supported property file extensions, sorted.
func Formats() []string {
	exts := make([]string, 0, len(formatsByExt))
	for ext := range formatsByExt {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return exts
}

// ParseProperties loads a property file; the format is chosen by extension:
// .properties (Java), .yaml/.yml, .json. An unknown extension is an error
// listing Formats.
func ParseProperties(path string) (map[string]string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	f, known := formatsByExt[ext]
	if !known {
		return nil, fmt.Errorf("props: %s: unknown property file extension %q (expected one of %s)", path, ext, strings.Join(Formats(), ", "))
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("props: %w", err)
	}
	defer file.Close()
	values, err := f.parse(file)
	if err != nil {
		return nil, fmt.Errorf("props: %s: %w", path, err)
	}
	return values, nil
}

// ParsePropertiesReader parses r in the given format ("properties", "yaml",
// "json").
func ParsePropertiesReader(r io.Reader, format string) (map[string]string, error) {
	f, known := formatsByName[strings.ToLower(format)]
	if !known {
		names := make([]string, 0, len(formatsByName))
		for name := range formatsByName {
			names = append(names, name)
		}
		sort.Strings(names)
		return nil, fmt.Errorf("props: unknown property format %q (expected one of %s)", format, strings.Join(names, ", "))
	}
	values, err := f.parse(r)
	if err != nil {
		return nil, fmt.Errorf("props: %w", err)
	}
	return values, nil
}

// ParseKeyValues parses the -p key=value command-line entries: each argument is
// split on its first "=", the key is trimmed, the value kept as is and a later
// entry overrides an earlier one with the same key. An argument without "=" or
// with an empty key is an error.
func ParseKeyValues(args []string) (map[string]string, error) {
	values := make(map[string]string, len(args))
	for _, arg := range args {
		key, value, found := strings.Cut(arg, "=")
		if !found {
			return nil, fmt.Errorf("props: property %q: expected key=value", arg)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("props: property %q: empty key", arg)
		}
		values[key] = value
	}
	return values, nil
}
