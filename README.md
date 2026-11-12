<!--
sword:
  schema: 1
  name: soft.gen
  family: ecore-esync
  kind: build-tooling
  generation: shared
  shipped: false  # UNVERIFIED
  languages: [Go, "Go text/template", Makefile, Bash]
  build: [go, make, docker]
  ng: [ecore-model]
  apps:
    - name: soft.gen
      kind: build-tool
      run: "docker run --rm -v \"$PWD\":/gen -w /gen masagroup/soft.gen go -m /gen/model.ecore -o /gen"
      ports: []
      launched_by: ["user (docker run, go install, go run)", "the Makefiles of soft.go and soft.ts (image masagroup/soft.gen:<SOFT_GEN_VERSION> since 2026-10-22; go run github.com/masagroup/soft.gen/v2/cmd/soft-gen@v2.0.0-alpha.2 before)", "the release/v2 branches of soft.sync.go (Makefile) and sw.api.ng.go (generate.sh)"]
      artifacts: "Docker image masagroup/soft.gen[:<version>] via make image / make publish; binary bin/soft.gen via make build; go install github.com/masagroup/soft.gen/v2/cmd/soft.gen@<version>"
      status: internal-tool
      # UNVERIFIED: release process (no CI publish job), internal-tool status inferred
-->

For the previous documentation of the project (v1), see [README.LEGACY.md](README.LEGACY.md).

# soft.gen

> Go (`text/template`) code generators, one command `soft.gen`, that turn an Ecore model (`.ecore`) into Go or TypeScript model code (C++ is being ported); the "model compiler" of the Soft (EMF-like) family used by SWORD next-gen models.

| | |
|---|---|
| **Family** | Ecore / esync |
| **Languages** | Go 1.27 (`go.mod`: `go 1.27.0`; image `golang:1.27.3-alpine`), Go `text/template` templates (92 `.tmpl`: go 61, ts 31), Makefile, Bash (`version.sh`) |
| **Build systems** | Go toolchain driven by a Makefile · Docker (one multi-stage `Dockerfile`, no compose profiles). No Ant, no Maven, no CMake build, no `package.json` |

**[Architecture](ARCHITECTURE.md)** · [Contributing](CONTRIBUTING.md) · [previous README (v1)](README.LEGACY.md) · [known issues and open questions](https://github.com/masagroup/wiki/wiki/known-issues#soft.gen) · [repository on GitHub](https://github.com/masagroup/soft.gen)

## Overview

soft.gen is the code generator of the [Soft](https://github.com/masagroup/soft) family: one command, `soft.gen`, with a subcommand per language (`go`, `ts`), written in Go with `text/template`. It reads an Ecore model (`.ecore`) and writes the model code for the Soft runtime of that language ([soft.go](https://github.com/masagroup/soft.go), [soft.ts](https://github.com/masagroup/soft.ts)). C++ is not ported yet: it still uses the v1 image `masagroup/soft.generator.cpp`, as the v1 generators (Java and Acceleo, up to the `v1.x` tags) do.

## Where to start

| You want to... | Read |
|---|---|
| build with Docker | [Build on Linux](https://github.com/masagroup/wiki/wiki/build-linux) |
| run the generator on a model | [Run](#run) below |
| change a generator or add a language | [How the generators are written](TEMPLATES.md) |
| see what changed since v1 | [Changelog](CHANGELOG.md) |
| understand how it is made | [ARCHITECTURE.md](ARCHITECTURE.md) |
| contribute | [CONTRIBUTING.md](CONTRIBUTING.md) |

## Build

The Makefile needs GNU make (Git Bash on Windows). `build`, `test` and `versions` use the `go` on the PATH, or a `golang` container when Go is missing; `lint`, `image`, `image.test` and `publish` need Docker.

### Go (Windows and Linux)

```
make build                   # bin/soft.gen
make versions                # the version of each language
make version.go TO=v1.14.0   # set the version of a language (then commit its VERSION file)
```

```
go install github.com/masagroup/soft.gen/v2/cmd/soft.gen@latest   # install without cloning
```

### Docker

```
make image             # image masagroup/soft.gen, VERSION computed by version.sh
make image.test        # checks the versions reported by the image
docker build --tag masagroup/soft.gen .   # without make (PowerShell), VERSION left to its default
```

`make publish` pushes the image to Docker Hub: it is a human action, the CI does not run it.

## Test

```
make test              # go vet ./... then go test ./...
make lint              # golangci-lint in Docker
```

```
go test ./internal/lang/golang -run TestFixtures           # compare the generation with the expected files in testdata/golden
go test ./internal/lang/golang -run TestFixtures -update   # rewrite the expected files, then review them with git
```

The TypeScript formatting tests (`TestPrettier`, `TestV1`) are skipped unless `SOFT_GEN_TS_TOOLS` names a directory that resolves prettier and oxlint, with `node` on the PATH. The v1 cross-check is described in [TEMPLATES.md](TEMPLATES.md#golden-tests).

## Run

### soft.gen

No service and no port: a command-line tool, run from its Docker image (the language is the first argument), from the binary, or without installing.

```
docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.gen go -m /gen/model.ecore -o /gen   # or: ts
docker run --rm masagroup/soft.gen -v          # version of soft.gen
docker run --rm masagroup/soft.gen go -v       # version of the Go generator (the soft.go version it targets)
soft.gen go -h                                 # templates and properties of a language
go run github.com/masagroup/soft.gen/v2/cmd/soft.gen@v2.0.0-alpha.4 go -m model.ecore -o .
```

Pin a tag of the image in builds (`masagroup/soft.gen:v2.0.0-alpha.5`), as the Makefiles of soft.go and soft.ts do.

```
soft.gen <language> -m <model> -o <folder> [-p key=value]... [-P file]... [-t template]... [-s] [--dry-run] [--diff] [--format] [--manifest=false] [--config file]
  --config, -c file       read the options from a configuration file (.yaml, .yml, .json, .toml)   $SOFT_GEN_CONFIG
  --model, -m string      the input model                                                          $SOFT_GEN_MODEL
  --output, -o string     the output folder                                                        $SOFT_GEN_OUTPUT
  --property, -p k=v      set value for given property, repeatable
  --properties, -P file   load properties from a property file (.json, .properties, .yaml, .yml)  $SOFT_GEN_PROPERTIES
  --template, -t name     the template to be executed, repeatable (default: the ones marked *)     $SOFT_GEN_TEMPLATE
  --silent, -s            print nothing but failures                                               $SOFT_GEN_SILENT
  --dry-run               list the files without writing them                                      $SOFT_GEN_DRY_RUN
  --diff                  print the differences with the existing files, write nothing, exit 1 on any difference   $SOFT_GEN_DIFF
  --format                run the external formatter of the language on the generated files        $SOFT_GEN_FORMAT
  --manifest              write the manifest (.soft.gen.yaml) in every folder of generated files (default: true)   $SOFT_GEN_MANIFEST
  --version, -v           print version information and exit
  --help, -h              show help
```

`-p`, `-P` and `-t` may be repeated, and a `-p` value keeps its commas (`-p packages=a,b`). `soft.gen` exits 0 on success and 1 on any failure.

The options can also come from a configuration file (`--config`), the command line winning over the environment, which wins over the file:

```yaml
model: models/shop.ecore
output: gen
template: [generateModel, generateMocks]
properties: [generator.yaml]
silent: true
```

The formats of the property files and of the generated manifest are in [ARCHITECTURE.md](ARCHITECTURE.md#interfaces).

## Documentation and guides

- [Architecture](ARCHITECTURE.md): position in the ecosystem, repository layout, internal structure, how it works (diagrams), dependencies.
- [Contributing](CONTRIBUTING.md): development environment, Git and Jira workflow, conventions, CI/CD.
- [How the generators are written](TEMPLATES.md) and the [Changelog](CHANGELOG.md).
- [Previous README (v1)](README.LEGACY.md): the documentation of the repository before the wiki rewrite.
- [Known issues and open questions](https://github.com/masagroup/wiki/wiki/known-issues#soft.gen): the known defects and open points of the repository, kept in the SWORD wiki.
- Related repositories: [soft](https://github.com/masagroup/soft), [soft.go](https://github.com/masagroup/soft.go), [soft.ts](https://github.com/masagroup/soft.ts).
- Transverse pages: [Build on Linux](https://github.com/masagroup/wiki/wiki/build-linux), [CI/CD pipelines](https://github.com/masagroup/wiki/wiki/cicd-pipelines), [esync and ecore](https://github.com/masagroup/wiki/wiki/esync-ecore), [Ecosystem next gen](https://github.com/masagroup/wiki/wiki/ecosystem-nextgen), [Glossary](https://github.com/masagroup/wiki/wiki/glossary).
- Repository on GitHub: [masagroup/soft.gen](https://github.com/masagroup/soft.gen).
- The same documentation in the SWORD wiki: [soft.gen](https://github.com/masagroup/wiki/wiki/soft.gen), [architecture](https://github.com/masagroup/wiki/wiki/soft.gen-architecture), [contributing](https://github.com/masagroup/wiki/wiki/soft.gen-contributing), [previous README](https://github.com/masagroup/wiki/wiki/soft.gen-legacy).
