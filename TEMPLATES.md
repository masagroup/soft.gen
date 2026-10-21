# How the generators are written

This document is for whoever changes a generator or adds a language. The
README is for whoever runs one.

## One rule: decisions in Go, text in templates

A generated file is produced in two steps:

1. Go code builds a **view model**: a plain struct of strings, booleans and
   slices holding everything the file shows, in the order it shows it. Every
   decision is taken here: which accessors a feature gets, which variant of a
   getter it needs, in which order the constants come, what the default value
   of an attribute is.
2. A **template** lays the view model out. It ranges over slices and tests
   booleans the view model already decided. It never asks the model anything:
   there is no `isProxy` or `getProperty` function in a template.

A variant of a piece of text is chosen by name:

```
{{ include (printf "feature/getter.%s" .GetterStrategy) . }}
```

where `GetterStrategy` is a string the Go code set (`plain`, `container`,
`proxy`, `lazyFunction`...). Adding a variant is adding a `define` and a case
in the Go function that chooses the strategy.

This split keeps the Acceleo features v1 relied on (overloaded templates
guarded by conditions, memoized queries) out of the templates and in Go
functions with explicit names, where they can be read and unit-tested.

## Where things are

| Package | Holds |
|---|---|
| `model` | Every fact about an Ecore model that does not depend on the language: feature classification, accessors, hierarchy, ids, ordering, lazy modes, annotation sources. The port of `soft.generator.common`. |
| `props` | The properties of a run: declarations, resolution (command line > files > package annotation > default), property file formats. |
| `naming` | Word splitting, case changes, naming rules (`pattern:replacement`, `{0}` formats), keyword avoidance. |
| `tmpl` | Loading of the template files, the shared template functions, the convention linter. |
| `writer` | Writing of the files: header, line endings, user code regions, formatter, dry run, diff. |
| `engine` | The `Language` contract, the run loop, the registry, `Validate`, the golden test harness. |
| `lang/<name>` | One language: its facts (paths, imports, names, types), its view models (`view_*.go`) and its templates (`templates/`). |

A language never recomputes a fact of `model`; if a fact is missing there, add
it there with the name of the v1 query it ports.

## Template conventions

The conventions are checked by `tmpl.Lint`, which `engine.Validate` runs; every
language runs `Validate` in its tests.

1. A template file holds only `define` blocks (and comments) at its top level.
2. A `define` is named after its file: `feature/getter.tmpl` defines
   `feature/getter`, `feature/getter.<variant>` or `feature/getter/<sub>`.
3. A block opens with `{{ define "name" -}}` and closes with `{{- end }}`: it
   renders complete lines and owns no trailing newline. The line that calls it
   (`{{ include "name" . }}` or `{{ template "name" . }}`) supplies the newline.
4. A line holding only a control action (`if`, `else`, `range`, `with`, `end`)
   is written `{{- if ... }}`: left trim only. The control line then leaves no
   trace in the output, and every content line keeps its own newline.
5. An intentional blank line is a literal empty line in the template. A
   `{{- if }}` line eats the blank line before it: put a `{{ blank }}` line
   between them to keep it (see `templates/package.go.tmpl`).
6. No `{{ "\n" }}`, no two consecutive blank lines.

Blank lines matter: `gofmt` collapses runs of blank lines but never adds or
removes one, and the expected files of the fixtures are the formatted v1
output. When v1 left a blank line somewhere odd, the template reproduces it.

## Qualified identifiers and imports (Go)

A type or function of another package is carried as a **qualified identifier**
`<import path>/<package name>.<Identifier>` (`github.com/masagroup/soft.go/ecore/ecore.EList`,
`*time/time.Time`). The view model builder passes it through
`scope.qualify(...)`, which returns the identifier as the file writes it
(bare in its own package, prefixed by an alias elsewhere) and records the
import. The import block is written from the recorded packages once the body
is rendered, with the v1 alias algorithm. A template never writes a package
prefix itself.

## Golden tests

The generated code is checked against **fixtures**: small purpose-built models
in `testdata/models`, each about one topic (attributes, references, maps,
volatile, lazy, operations, inheritance, enums, naming, annotations, a package
named `ecore`, and the package options on `shop`). A case is a model, entry
points and properties, registered from `init` in
`lang/golang/fixtures_<topic>_test.go`; its expected files are committed in
`testdata/golden/go/<case>`, laid out as the output folder.

- **`TestFixtures`** generates every case and compares it with its expected
  files (`engine/goldentest`). After an intended change of the output, run it
  with `-update` to write the new expected files, and review them with git.
- **`TestTemplateCoverage`** fails when a template define is rendered by no
  case: a new variant needs a fixture. The few defines no model can reach are
  listed in `unreachableTemplates`, each with its reason.
- **`TestV1`** compares the cases with the v1 generator, the reference of the
  expected files. It runs when `SOFT_GEN_V1_IMAGE` names the v1 image and
  docker is available:

  ```shell
  SOFT_GEN_V1_IMAGE=masagroup/soft.generator.go:v1.13.1 go test ./lang/golang -run TestV1
  ```

  The files where v2 differs from v1 on purpose (v1 being wrong) are listed in
  the `V1Differences` of their case with the reason; a case v1 cannot generate
  at all has a `V1Skip` reason. With `-v1update`, the test writes v1's
  formatted output as the expected files, keeping the listed differences: this
  is how the expected files of a new fixture are made.

Debugging a difference:

- `soft-gen go --diff -m model.ecore -o <dir>` prints the differences with the
  files in `<dir>` and writes nothing;
- `soft-gen go --dry-run ...` lists the files that would be written;
- `-p debug=true` raises the log level.

## Adding a language

1. Create `lang/<name>` with a `Language` type implementing `engine.Language`
   (name, generator name, annotation source, properties, entry points,
   templates, functions, regions, formatter, name style) and an `init`
   registering it.
2. Add `_ "github.com/masagroup/soft.gen/v2/lang/<name>"` to `cmd/soft-gen/main.go`.
3. Write the facts of the language (paths, names, types) as small documented
   functions, then one `view_*.go` per kind of file with its template.
4. Add a `TestValidate`, fixtures with their expected files (made from the v1
   generator when there is one) and a template coverage test, as
   `lang/golang` does.
5. Document its properties and entry points in the README.
