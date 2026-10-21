// Command soft-gen generates code from an Ecore model.
//
//	soft-gen <language> -m <model> -o <folder> [-p key=value]... [-P file]... [-t template]... [-s] [--dry-run] [--diff] [--format]
//	soft-gen <language> -v
//	soft-gen -h
//
// The command line is the one of the v1 generators, with the language as a
// subcommand; the docker image runs soft-gen, the language first.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/props"
	"github.com/masagroup/soft.gen/v2/writer"

	// The languages register themselves.
	_ "github.com/masagroup/soft.gen/v2/lang/golang"
)

// developerBuild is the version of a build that sets none.
const developerBuild = "v2.0.0-developer-build"

// Version is the version of the build, set by the Makefile and the Dockerfile
// from version.sh: -ldflags "-X main.Version=v2.1.0".
var Version = developerBuild

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
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the command line and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printLanguages(stdout)
		return 0
	}
	if args[0] == "-v" || args[0] == "--version" {
		fmt.Fprintf(stdout, "soft-gen version: %s\n", version())
		return 0
	}
	lang, ok := engine.Lookup(args[0])
	if !ok {
		fmt.Fprintf(stderr, "unknown language %q\n", args[0])
		printLanguages(stderr)
		return 1
	}
	req, code, done := parse(lang, args[1:], stdout, stderr)
	if done {
		return code
	}
	report, err := engine.Run(context.Background(), req)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if req.Diff && report.Count(writer.Different) > 0 {
		return 1
	}
	return 0
}

// repeatable collects a repeatable flag.
type repeatable []string

func (r *repeatable) String() string     { return strings.Join(*r, ",") }
func (r *repeatable) Set(v string) error { *r = append(*r, v); return nil }

// parse reads the flags of a language; done is true when the command printed
// its help or version and must exit with code.
func parse(lang engine.Language, args []string, stdout, stderr io.Writer) (req engine.Request, code int, done bool) {
	fs := flag.NewFlagSet(lang.GeneratorName(), flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		model, output          string
		propertyArgs           repeatable
		propertyFiles          repeatable
		templates              repeatable
		silent, showVersion    bool
		dryRun, diff, doFormat bool
	)
	for _, name := range []string{"m", "model"} {
		fs.StringVar(&model, name, "", "the input model")
	}
	for _, name := range []string{"o", "output"} {
		fs.StringVar(&output, name, "", "the output folder")
	}
	for _, name := range []string{"p", "property"} {
		fs.Var(&propertyArgs, name, "set value for given property")
	}
	for _, name := range []string{"P", "properties"} {
		fs.Var(&propertyFiles, name, "load properties from a property file")
	}
	for _, name := range []string{"t", "template"} {
		fs.Var(&templates, name, "the template to be executed")
	}
	for _, name := range []string{"s", "silent"} {
		fs.BoolVar(&silent, name, false, "print nothing but failures")
	}
	for _, name := range []string{"v", "version"} {
		fs.BoolVar(&showVersion, name, false, "print version information and exit")
	}
	fs.BoolVar(&dryRun, "dry-run", false, "list the files without writing them")
	fs.BoolVar(&diff, "diff", false, "print the differences with the existing files, write nothing, exit 1 on any difference")
	fs.BoolVar(&doFormat, "format", false, "run the external formatter of the language on the generated files")
	fs.Usage = func() { printUsage(lang, stderr) }
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return req, 0, true
		}
		return req, 1, true
	}
	if showVersion {
		fmt.Fprintf(stdout, "%s version: %s\n", lang.GeneratorName(), version())
		return req, 0, true
	}
	if model == "" || output == "" {
		fmt.Fprintln(stderr, "the model (-m) and the output folder (-o) are required")
		printUsage(lang, stderr)
		return req, 1, true
	}
	properties, err := props.ParseKeyValues(propertyArgs)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return req, 1, true
	}
	return engine.Request{
		Lang:        lang,
		ModelPath:   model,
		OutDir:      output,
		Props:       properties,
		PropFiles:   propertyFiles,
		EntryPoints: templates,
		Silent:      silent,
		DryRun:      dryRun,
		Diff:        diff,
		Format:      doFormat,
		Output:      stdout,
		ErrOutput:   stderr,
	}, 0, false
}

// printLanguages prints the global usage.
func printLanguages(out io.Writer) {
	fmt.Fprintln(out, "Usage: soft-gen <language> [-hsv] -m=<model> -o=<folder> [-p=<property=value>]... [-P=<propertyfile>]... [-t=<template>]... [--dry-run] [--diff] [--format]")
	fmt.Fprintln(out, "Languages:")
	for _, lang := range engine.Languages() {
		fmt.Fprintf(out, "  %-6s %s\n", lang.Name(), lang.GeneratorName())
	}
	fmt.Fprintln(out, "soft-gen <language> -h prints the templates and properties of a language.")
}

// printUsage prints the usage of a language: the v1 usage, the templates and
// the properties.
func printUsage(lang engine.Language, out io.Writer) {
	fmt.Fprintf(out, "Usage: soft-gen %s [-hsv] -m=<model> -o=<folder> [-p=<property=value>]... [-P=<propertyfile>]... [-t=<template>]... [--dry-run] [--diff] [--format]\n", lang.Name())
	fmt.Fprintln(out, "  -h, --help                        print this help and exit")
	fmt.Fprintln(out, "  -m, --model=<model>               the input model")
	fmt.Fprintln(out, "  -o, --output=<folder>             the output folder")
	fmt.Fprintln(out, "  -p, --property=<property=value>   set value for given property")
	fmt.Fprintf(out, "  -P, --properties=<propertyfile>   load properties from a property file (%s)\n", strings.Join(props.Formats(), ", "))
	fmt.Fprintln(out, "  -s, --silent                      print nothing but failures")
	fmt.Fprintln(out, "  -t, --template=<template>         the template to be executed (default: the ones marked *)")
	fmt.Fprintln(out, "  -v, --version                     print version information and exit")
	fmt.Fprintln(out, "      --dry-run                     list the files without writing them")
	fmt.Fprintln(out, "      --diff                        print the differences with the existing files, write nothing, exit 1 on any difference")
	fmt.Fprintln(out, "      --format                      run the external formatter of the language on the generated files")
	fmt.Fprintln(out, "Templates:")
	for _, ep := range lang.EntryPoints() {
		mark := " "
		if ep.Default {
			mark = "*"
		}
		fmt.Fprintf(out, "  %s %-28s %s\n", mark, ep.Name, ep.Doc)
	}
	fmt.Fprintln(out, "Properties:")
	fmt.Fprint(out, props.FormatHelp(append(props.Reserved(), lang.Properties()...)))
}
