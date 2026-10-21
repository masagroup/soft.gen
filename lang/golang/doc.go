// Package golang generates Go code from an Ecore model: the interfaces and
// the implementations of the classes, the enums, the package and the factory,
// the mocks and the tests, as the v1 Go generator did.
//
// The package is organized in three layers:
//
//   - the facts of the language: paths.go (package paths and output
//     directories), imports.go (qualified identifiers and import aliases),
//     names.go (the names of the generated types, files and accessors) and
//     types.go (the Go type, default value and test value of a classifier);
//   - the view models: one view_*.go file per kind of generated file builds a
//     plain struct holding every string the file needs, taking every decision
//     in Go;
//   - the templates: templates/*.tmpl lay the view models out; a template
//     never asks the model anything.
//
// A qualified identifier is "<import path>/<package name>.<Identifier>", with
// an optional "*" prefix; a fileScope turns it into the identifier the file
// uses and records the import.
package golang
