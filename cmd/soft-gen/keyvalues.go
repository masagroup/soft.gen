package main

import (
	"fmt"
	"sort"
	"strings"
)

// keyValues is the value of the repeatable -p key=value option. Unlike a cli
// slice flag it never splits a value on commas, so that a property value may
// hold them (-p packages=a,b), as in v1.
type keyValues struct {
	values map[string]string
}

// Set adds one key=value; a later value of the same key wins.
func (k *keyValues) Set(value string) error {
	key, val, ok := strings.Cut(value, "=")
	if !ok || strings.TrimSpace(key) == "" {
		return fmt.Errorf("property %q is not key=value", value)
	}
	if k.values == nil {
		k.values = map[string]string{}
	}
	k.values[strings.TrimSpace(key)] = val
	return nil
}

// Get returns the properties, a map[string]string.
func (k *keyValues) Get() any {
	return k.values
}

// String lists the properties, sorted by key.
func (k *keyValues) String() string {
	keys := make([]string, 0, len(k.values))
	for key := range k.values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = key + "=" + k.values[key]
	}
	return strings.Join(parts, ", ")
}
