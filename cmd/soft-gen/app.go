package main

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/masagroup/soft.gen/v2/internal/engine"
)

// newApp builds the root command: one subcommand per registered language,
// each with its version.
// The root only prints the help (no argument), the version (-v) or rejects an
// unknown language.
func newApp(stdout, stderr io.Writer, versions languageVersions) *cli.Command {
	app := &cli.Command{
		Name:        "soft-gen",
		Usage:       "generates code from an Ecore model",
		UsageText:   "soft-gen <language> [options]\nsoft-gen <language> -h    the templates, properties and options of a language",
		Writer:      stdout,
		ErrWriter:   stderr,
		HideVersion: true,
		Flags:       []cli.Flag{versionFlag()},
		// the exit code is computed by run: never exit from inside the command
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Action:         rootAction,
	}
	for _, lang := range engine.Languages() {
		app.Commands = append(app.Commands, languageCommand(lang, versions.of(lang)))
	}
	return app
}

// versionFlag is the -v, --version flag of the root and language commands.
func versionFlag() cli.Flag {
	return &cli.BoolFlag{Name: "version", Aliases: []string{"v"}, Usage: "print version information and exit"}
}

// rootAction prints the version with -v, rejects an argument that is not a
// language, and prints the help otherwise.
func rootAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool("version") {
		fmt.Fprintf(cmd.Writer, "soft-gen version: %s\n", version())
		return nil
	}
	if cmd.Args().Present() {
		return cli.Exit(fmt.Sprintf("unknown language %q (languages: %s)", cmd.Args().First(), languageNames()), 1)
	}
	return cli.ShowRootCommandHelp(cmd)
}

// languageNames lists the registered languages, comma separated.
func languageNames() string {
	names := ""
	for i, lang := range engine.Languages() {
		if i > 0 {
			names += ", "
		}
		names += lang.Name()
	}
	return names
}
