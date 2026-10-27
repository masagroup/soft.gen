# soft.gen: contributing

> How to set up a development environment, the Git and Jira workflow and the conventions of `soft.gen`. Back to the [README](README.md) · [Architecture](ARCHITECTURE.md)

## Development environment

The build of the repository is described in the [README](README.md#build); the tooling of the ecosystem is on the transverse pages:

- Windows, Ant and sw.poney: Apache Ant, JDK 8, Visual Studio 2022 Professional with the MSVC v140 toolset and the Windows SDK 10.0.19041.0, `gh` authenticated, access to `\\file1` or a local `sw.poney` clone in `PONEY_HOME`; initialise the MSVC environment before any `ant` (see [Build on Windows](https://github.com/masagroup/wiki/wiki/build-windows)).
- Linux and Docker Compose: Docker with BuildKit, the GitHub PAT file named by `SECRET_GITHUB_PAT` (`~/.masabot/github.pat`) or the SSH key of `SECRET_GITHUB_SSH` for the private Go modules and npm packages, the profiles `prod`, `test`, `pack` and `dev` (see [Build on Linux](https://github.com/masagroup/wiki/wiki/build-linux)); never push an image, the CI does it.
- Go: the version of `go.mod`, `GOPRIVATE=github.com/masagroup/*` with `gh auth setup-git` or an SSH `insteadOf` rewrite for the private modules; `staticcheck` installed by the build.
- Editor rules for Claude Code: the shared `claude/` folder of [cicd](https://github.com/masagroup/cicd) installs the same `CLAUDE.md` for every developer (rules, roles, build pages, conventions).

## Git and Jira workflow

- The organisation repository `masagroup/soft.gen` is the reference (CI, releases); work on a fork or on a branch of it, never directly on `master` nor on a release branch (`6.28`...).
- Branch name: `<TICKET>-<short-description>` (for example `SWNG-421-tests-and-fixes`), created from the target branch of the ticket.
- Commit message: `<type>(<scope>): <TICKET-ID>: <imperative description in English>`, with `type` in `feat`, `fix`, `refactor`, `test`, `build`, `ci`, `perf`, `docs`, `chore`; the body says what and why, and the CI impact when there is one. One commit is one coherent change; no `Co-Authored-By` nor other automatic signature.
- Pull request: masabot reads the Jira keys (`SW`, `SWNG`, `IO`, `AI`, `GEFORA`, `COA`) in the title, the branch, the body and the commits, comments the Jira issue and moves it (*In Progress*, *In Review* with `[DONE]`, *In QA* or *In Validation* after the merge). The PR is built rebased on its base branch; the comment commands `/rebase`, `/rebase-all` and `/automerge` are available. See [CI/CD pipelines](https://github.com/masagroup/wiki/wiki/cicd-pipelines).
- Never push to `master` by hand, never create a release or push an image: the CI does it.

## Conventions

- Code, comments, commit messages and file names in English; documentation of the repository in English.
- Formatting of the repository (`clang-format`, `gofmt`/`goimports`, `prettier`/`eslint`) on the lines you touch; no reformatting of a whole file for a one-line change; no dead code, no `TODO` without a ticket; logs at the right level, never a secret.
- Go: the version of `go.mod` and of the Docker image, `GOPRIVATE=github.com/masagroup/*`; `go vet`, `staticcheck`, `gofmt` and `go test -race ./...` green before a commit; wrapped errors (`fmt.Errorf("...: %w", err)`), no `log.Fatal` outside `main`, clean stop on `SIGINT` and `SIGTERM`, timeouts on every network I/O; REST APIs from an OpenAPI contract (`oapi-codegen`).
- Shared contracts (`.proto`, `.ecore`, OpenAPI, exported Go modules, file formats) are changed with the consumers in mind: grep the workspace, state the compatibility strategy, update the dependants.
- Documentation: this README set (README, ARCHITECTURE, CONTRIBUTING, README.LEGACY) is the source of the [SWORD wiki](https://github.com/masagroup/wiki/wiki/soft.gen); keep the metadata block at the top of the README up to date (the `apps` list feeds the Interfaces section).

## CI/CD

| Workflow | Triggers | What it does |
|---|---|---|
| `.github/workflows/build_and_test.yml` ("build and test") | `push` on `master`; `pull_request` (any branch); `workflow_dispatch` | job `test` on `ubuntu-latest`: `actions/checkout@v4`, `actions/setup-go@v5` (Go from `go.mod`), `go vet ./...`, `go test ./...`, `go build ./cmd/soft.gen`; job `image` (needs `test`): checkout with `fetch-depth: 0` (tags for `version.sh`), `make image`, `make image.test` |

- It does not use any `masagroup/cicd` reusable workflow (no shared `uses:`, no repository/organization variable override, no secrets). This deviates from the ecosystem convention; see [cicd](https://github.com/masagroup/cicd) and the ecosystem wiki page [CI/CD pipelines](https://github.com/masagroup/wiki/wiki/cicd-pipelines).
- No release job. `make publish` (Docker Hub `masagroup/soft.gen:<VERSION>` and `:latest`) exists, but no workflow runs it. How v2 tags, GitHub releases and Docker Hub pushes are produced is UNVERIFIED (probably manual).
- Artifacts: none produced by CI (the image is built then discarded). Locally: `bin/soft.gen` and the image `masagroup/soft.gen[:<VERSION>]`; the `org/repo:component-version` convention is not used.
- Not run by CI: `make lint`, and the TypeScript formatting tests (`SOFT_GEN_TS_TOOLS` is not set, so they are skipped).
- CI impact: a Go build and test plus one `docker build` of a static binary; no credentials needed (v1 needed GitHub Packages credentials for Maven).

## Known issues and open questions

They are not kept in the repository: see the [private artifact](https://claude.ai/artifact/UfvKmky7e8LwEXf192hQjQ#soft.gen) that holds them for every page of the wiki.
