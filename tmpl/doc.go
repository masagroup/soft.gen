// Package tmpl loads the text/template files of a language from an embedded file
// system into one template set, adds the presentational functions every
// language shares (include, indent, join, quote...) and checks the naming and
// whitespace conventions of the template files (Lint).
//
// Templates only lay text out: every decision is taken in Go, in the view models
// the language builds. A variant of a template is chosen by name with
// include (printf "feature/getter.%s" .Strategy).
package tmpl
