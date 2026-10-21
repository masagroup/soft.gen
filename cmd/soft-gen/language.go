package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/props"
	"github.com/masagroup/soft.gen/v2/writer"
)

// The names of the options of a language command; the long name is also the
// key of the option in a configuration file and, upper-cased, the suffix of
// its environment variable (SOFT_GEN_MODEL, SOFT_GEN_DRY_RUN...).
const (
	optionModel      = "model"
	optionOutput     = "output"
	optionProperty   = "property"
	optionProperties = "properties"
	optionTemplate   = "template"
	optionSilent     = "silent"
	optionDryRun     = "dry-run"
	optionDiff       = "diff"
	optionFormat     = "format"
	optionConfig     = "config"
)

// options holds the values of a language command that its flags fill at run
// time: the configuration file named by --config and the -p properties.
type options struct {
	config     *configFile
	properties *keyValues
}

// languageCommand builds the command of a language: its options, its help
// (templates and properties) and its action, which runs the generation;
// version is the version of the language, printed by -v.
func languageCommand(lang engine.Language, version string) *cli.Command {
	opts := &options{config: &configFile{}, properties: &keyValues{}}
	return &cli.Command{
		Name:        lang.Name(),
		Usage:       "generates " + lang.Name() + " code (" + lang.GeneratorName() + " " + version + ")",
		UsageText:   "soft-gen " + lang.Name() + " -m <model> -o <folder> [-p key=value]... [-P file]... [-t template]... [-s] [--dry-run] [--diff] [--format] [--config file]",
		Description: languageDescription(lang),
		Flags:       languageFlags(opts),
		Before:      opts.config.check,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return generate(ctx, cmd, lang, version, opts)
		},
	}
}

// languageFlags declares the options of a language command. Every option but
// -p also reads its SOFT_GEN_* environment variable, then the configuration
// file given by --config. --config comes first: the sources of the flags are
// applied in their order, and the others read the file it names (from
// SOFT_GEN_CONFIG too).
func languageFlags(opts *options) []cli.Flag {
	config := opts.config
	return []cli.Flag{
		&cli.StringFlag{Name: optionConfig, Aliases: []string{"c"}, Usage: "read the options from a configuration `file` (" + strings.Join(configFormats, ", ") + ")", Sources: cli.EnvVars(envVar(optionConfig)), Destination: &config.path},
		&cli.StringFlag{Name: optionModel, Aliases: []string{"m"}, Usage: "the input model", Sources: config.sources(optionModel)},
		&cli.StringFlag{Name: optionOutput, Aliases: []string{"o"}, Usage: "the output folder", Sources: config.sources(optionOutput)},
		&cli.GenericFlag{Name: optionProperty, Aliases: []string{"p"}, Usage: "set value for given property (`key=value`, repeatable)", Value: opts.properties},
		&cli.StringSliceFlag{Name: optionProperties, Aliases: []string{"P"}, Usage: "load properties from a property `file` (" + strings.Join(props.Formats(), ", ") + "), repeatable", Sources: config.sources(optionProperties)},
		&cli.StringSliceFlag{Name: optionTemplate, Aliases: []string{"t"}, Usage: "the `template` to be executed, repeatable (default: the templates marked *)", Sources: config.sources(optionTemplate)},
		&cli.BoolFlag{Name: optionSilent, Aliases: []string{"s"}, Usage: "print nothing but failures", Sources: config.sources(optionSilent)},
		&cli.BoolFlag{Name: optionDryRun, Usage: "list the files without writing them", Sources: config.sources(optionDryRun)},
		&cli.BoolFlag{Name: optionDiff, Usage: "print the differences with the existing files, write nothing, exit 1 on any difference", Sources: config.sources(optionDiff)},
		&cli.BoolFlag{Name: optionFormat, Usage: "run the external formatter of the language on the generated files", Sources: config.sources(optionFormat)},
		versionFlag(),
	}
}

// languageDescription is the help text of a language: its templates, the
// default ones marked *, and its properties.
func languageDescription(lang engine.Language) string {
	var b strings.Builder
	b.WriteString("Templates:\n")
	for _, ep := range lang.EntryPoints() {
		mark := " "
		if ep.Default {
			mark = "*"
		}
		fmt.Fprintf(&b, "  %s %-28s %s\n", mark, ep.Name, ep.Doc)
	}
	b.WriteString("\nProperties:\n")
	b.WriteString(props.FormatHelp(append(props.Reserved(), lang.Properties()...)))
	return strings.TrimRight(b.String(), "\n")
}

// generate is the action of a language command: it prints the version of the
// language with -v, builds the request from the options and runs the
// generation; with --diff, any difference exits with code 1.
func generate(ctx context.Context, cmd *cli.Command, lang engine.Language, version string, opts *options) error {
	if cmd.Bool("version") {
		fmt.Fprintf(cmd.Writer, "%s version: %s\n", lang.GeneratorName(), version)
		return nil
	}
	req, err := request(cmd, lang, opts)
	if err != nil {
		return err
	}
	report, err := engine.Run(ctx, req)
	if err != nil {
		return err
	}
	if req.Diff && report.Count(writer.Different) > 0 {
		return cli.Exit("", 1)
	}
	return nil
}

// request builds the engine request from the options of a language command;
// the model and the output folder are required.
func request(cmd *cli.Command, lang engine.Language, opts *options) (engine.Request, error) {
	model, output := cmd.String(optionModel), cmd.String(optionOutput)
	if model == "" || output == "" {
		_ = cli.ShowSubcommandHelp(cmd)
		return engine.Request{}, cli.Exit("the model (-m) and the output folder (-o) are required", 1)
	}
	properties := opts.properties.values
	if properties == nil {
		properties = map[string]string{}
	}
	return engine.Request{
		Lang:        lang,
		ModelPath:   model,
		OutDir:      output,
		Props:       properties,
		PropFiles:   cmd.StringSlice(optionProperties),
		EntryPoints: cmd.StringSlice(optionTemplate),
		Silent:      cmd.Bool(optionSilent),
		DryRun:      cmd.Bool(optionDryRun),
		Diff:        cmd.Bool(optionDiff),
		Format:      cmd.Bool(optionFormat),
		Output:      cmd.Writer,
		ErrOutput:   errWriter(cmd),
	}, nil
}

// errWriter is the error output of a command (stderr when unset).
func errWriter(cmd *cli.Command) io.Writer {
	if w := cmd.Root().ErrWriter; w != nil {
		return w
	}
	return os.Stderr
}
