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

The command is `cmd/soft-gen`; every other package is under `internal/`: the
module has no public Go API, and a language is added in this repository.

| Package | Holds |
|---|---|
| `internal/model` | Every fact about an Ecore model that does not depend on the language: feature classification, accessors, hierarchy, ids, ordering, lazy modes, annotation sources. The port of the common part of the v1 generators. |
| `internal/props` | The properties of a run: declarations, resolution (command line > files > package annotation > default), property file formats. |
| `internal/naming` | Word splitting, case changes, naming rules (`pattern:replacement`, `{0}` formats), keyword avoidance. |
| `internal/tmpl` | Loading of the template files, the shared template functions, the convention linter. |
| `internal/writer` | Writing of the files: header, line endings, user code regions, formatter, dry run, diff. |
| `internal/engine` | The `Language` contract, the run loop, the manifests, the registry, `Validate`, the golden tests (`goldentest`: the fixture suite, the template coverage, the v1 cross-check). |
| `internal/lang/common` | What the builders of every language share: package properties, generator annotation keys, naming rules with the name style of the language, list properties (`packages`, `paths`...). |
| `internal/lang/<name>` | One language: its facts (paths, imports, names, types), its view models (`view_*.go`) and its templates (`templates/`). |

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

## Qualified identifiers and imports (TypeScript)

A name of another file is carried as `<module>/<name>`, where the module may
hold `/` and the name may be followed by members and type arguments
(`@masagroup/ecore/EList<shop/Product>`, `shop/ShopConstants.PRODUCT__NAME`).
`scope.qualify(...)` returns it as the file writes it (the bare name for a
module imported by name, `<namespace>.<name>` for a module imported as a
namespace) and records the import; `scope.declare(...)` marks the names the
file defines, which are never imported. The view models hand the templates
lazy methods (`names.Ecore`, `names.Q`, the `featureView` methods), so a name is
imported only when a template writes it. The functions of the test frameworks
(vitest, ts-mockito, deep-equal) are found in the rendered body instead.

## Layout (TypeScript)

The expected files are the v1 output formatted by prettier with the soft.ts
options, so the generator writes the layout prettier would: the lists it builds
are wrapped by `wrapList`, and `breakLongLines` (`layout.go`) breaks the
rendered lines longer than `printWidth` with the subset of prettier's rules the
generated code needs (call arguments, member chains, conditionals,
assignments, logical expressions). Templates write each statement on one
line. `TestPrettier` checks that prettier rewrites none of the expected files;
it runs when `SOFT_GEN_TS_TOOLS` names a folder whose `node_modules` hold
prettier and oxlint (soft.ts's `ecore`, after `npm install`):

```shell
SOFT_GEN_TS_TOOLS=../soft.ts/ecore go test ./internal/lang/typescript -run TestPrettier
# TestV1: v1 and v2 formatted as soft.ts does (unused imports removed, prettier); the shared
# package cases need an image of the v1 branch head (masa/v1), built from Dockerfile-ts
SOFT_GEN_V1_TS_IMAGE=masagroup/soft.generator.ts:v1.7.0 SOFT_GEN_TS_TOOLS=../soft.ts/ecore go test ./internal/lang/typescript -run TestV1
```

## Golden tests

The generated code is checked against **fixtures**: small purpose-built models
in `testdata/models`, each about one topic (attributes, references, maps,
volatile, lazy, operations, inheritance, enums, naming, annotations, a package
named `ecore`, and the package options on `shop`). A case is a model, entry
points and properties, registered from `init` in
`internal/lang/golang/fixtures_<topic>_test.go` into the `goldentest.Suite` of the
language; its expected files are committed in `testdata/golden/go/<case>`,
laid out as the output folder (TypeScript: `internal/lang/typescript/fixtures_test.go`,
`testdata/golden/ts`, its own models in `testdata/models/ts`). The suite runs
the three tests below for any language.

- **`TestFixtures`** generates every case and compares it with its expected
  files (`internal/engine/goldentest`). After an intended change of the output, run it
  with `-update` to write the new expected files, and review them with git.
- **`TestTemplateCoverage`** fails when a template define is rendered by no
  case: a new variant needs a fixture. The few defines no model can reach are
  listed in the `Unreachable` map of the suite, each with its reason.
- **`TestV1`** compares the cases with the v1 generator, the reference of the
  expected files. It runs when `SOFT_GEN_V1_IMAGE` names the v1 image and
  docker is available:

  ```shell
  SOFT_GEN_V1_IMAGE=masagroup/soft.generator.go:v1.13.1 go test ./internal/lang/golang -run TestV1
  ```

  The first line of every file, which names the generator, is the v2 one in
  the comparison, and the manifests (`.soft-gen.yaml`) are v2 only: they are
  compared by `TestFixtures`, whose cases write them with the version
  `v0.0.0-test`.
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

1. Create `internal/lang/<name>` with a `Language` type implementing `engine.Language`
   (name, annotation source, properties, entry points, templates, functions,
   regions, formatter, name style) and an `init` registering it. Its builder
   embeds `internal/lang/common.Gen`: package properties, annotation keys,
   naming rules and list properties (`packages`...) are the same for every
   language.
2. Add `_ "github.com/masagroup/soft.gen/v2/internal/lang/<name>"` to `cmd/soft-gen/main.go`.
   Give it a `VERSION` file (the version of the runtime its templates
   target, embedded and returned by `Version()`), and add it to `LANGUAGES`
   and `lang.package.<name>` in the Makefile.
3. Write the facts of the language (paths, names, types) as small documented
   functions, then one `view_*.go` per kind of file with its template.
4. Add a `TestValidate` and a `goldentest.Suite` with fixtures and their
   expected files (made from the v1 generator when there is one), whose
   `TestFixtures`, `TestTemplateCoverage` and `TestV1` the language tests
   call, as `internal/lang/golang` does.
5. Document its properties and entry points in the README.
