package main

import (
	"fmt"
	"strings"

	"github.com/masagroup/soft.gen/v2/engine"
)

// developerLanguageVersion is the version of a language in a build that sets
// none (go build, go install): the Makefile and the Dockerfile set them all.
const developerLanguageVersion = "v0.0.0-developer-build"

// languageVersions maps a language name to its version: the version of the
// runtime (soft.go, soft.ts...) its templates target.
type languageVersions map[string]string

// parseLanguageVersions reads LanguageVersions, "go=v1.13.1,ts=v1.7.0": the
// pairs are separated by commas, a language from its version by "=". Every
// language must be registered and every version set.
func parseLanguageVersions(s string) (languageVersions, error) {
	versions := languageVersions{}
	if strings.TrimSpace(s) == "" {
		return versions, nil
	}
	for _, pair := range strings.Split(s, ",") {
		name, version, ok := strings.Cut(strings.TrimSpace(pair), "=")
		name, version = strings.TrimSpace(name), strings.TrimSpace(version)
		switch {
		case !ok || name == "":
			return nil, fmt.Errorf("%q is not language=version", pair)
		case version == "":
			return nil, fmt.Errorf("language %s: no version", name)
		case !registered(name):
			return nil, fmt.Errorf("unknown language %q (languages: %s)", name, languageNames())
		case versions[name] != "":
			return nil, fmt.Errorf("language %s: version set twice", name)
		}
		versions[name] = version
	}
	return versions, nil
}

// of returns the version of a language, or the developer build version when
// the build set none.
func (v languageVersions) of(lang engine.Language) string {
	if version := v[lang.Name()]; version != "" {
		return version
	}
	return developerLanguageVersion
}

// registered tells whether a language of this name is registered.
func registered(name string) bool {
	_, ok := engine.Lookup(name)
	return ok
}
