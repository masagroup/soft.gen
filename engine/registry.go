package engine

import (
	"fmt"
	"sort"
	"sync"
)

var (
	registryMu sync.Mutex
	registry   = map[string]Language{}
)

// Register adds a language, usually from the init function of its package.
// It panics on a duplicate name.
func Register(lang Language) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[lang.Name()]; exists {
		panic(fmt.Sprintf("engine: language %q registered twice", lang.Name()))
	}
	registry[lang.Name()] = lang
}

// Lookup returns the registered language of that name.
func Lookup(name string) (Language, bool) {
	registryMu.Lock()
	defer registryMu.Unlock()
	lang, ok := registry[name]
	return lang, ok
}

// Languages lists the registered languages, sorted by name.
func Languages() []Language {
	registryMu.Lock()
	defer registryMu.Unlock()
	langs := make([]Language, 0, len(registry))
	for _, lang := range registry {
		langs = append(langs, lang)
	}
	sort.Slice(langs, func(i, j int) bool { return langs[i].Name() < langs[j].Name() })
	return langs
}
