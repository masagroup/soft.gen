// Command soft-gen generates code from an Ecore model.
//
//	soft-gen <language> -m <model> -o <folder> [-p key=value]... [-P file]... [-t template]... [-s] [--dry-run] [--diff] [--format] [--manifest=false] [--config file]
//	soft-gen <language> -v
//	soft-gen -h
//
// The command line is the one of the v1 generators, with the language as a
// subcommand; the docker image runs soft-gen, the language first. Built with
// urfave/cli: the options may also come from SOFT_GEN_* environment variables
// and from a configuration file (YAML, JSON or TOML) read with cli-altsrc.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v3"

	// The languages register themselves.
	_ "github.com/masagroup/soft.gen/v2/lang/golang"
)

// developerBuild is the version of a build that sets none.
const developerBuild = "v2.0.0-developer-build"

// Version is the version of the build, set by the Makefile and the Dockerfile
// from version.sh: -ldflags "-X main.Version=v2.1.0".
var Version = developerBuild

// LanguageVersions is the version of each language, set by the Makefile and
// the Dockerfile: -ldflags "-X main.LanguageVersions=go=v1.13.1,ts=v1.7.0".
// The version of a language is the version of the runtime (soft.go, soft.ts...)
// its templates target, printed by "soft-gen <language> -v".
var LanguageVersions string

// version returns the version of the build: Version when the build set it,
// else the module version Go stamped in the binary (go install ...@v2.1.0),
// else the developer build version.
func version() string {
	if Version != developerBuild {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return Version
}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the command line and returns the exit code: 0 on success, the
// code of a cli.ExitCoder error, 1 on any other error.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	versions, err := parseLanguageVersions(LanguageVersions)
	if err != nil {
		fmt.Fprintln(stderr, "invalid LanguageVersions:", err)
		return 1
	}
	app := newApp(stdout, stderr, versions)
	err = app.Run(ctx, append([]string{app.Name}, args...))
	return exitCode(err, stderr)
}

// exitCode prints the error, if any, and returns the exit code of the run.
func exitCode(err error, stderr io.Writer) int {
	if err == nil {
		return 0
	}
	var exit cli.ExitCoder
	if errors.As(err, &exit) {
		if message := exit.Error(); message != "" {
			fmt.Fprintln(stderr, message)
		}
		return exit.ExitCode()
	}
	fmt.Fprintln(stderr, err)
	return 1
}
