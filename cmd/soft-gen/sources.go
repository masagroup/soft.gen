package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcjson "github.com/urfave/cli-altsrc/v3/json"
	altsrctoml "github.com/urfave/cli-altsrc/v3/toml"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

// envPrefix is the prefix of the environment variables of the options.
const envPrefix = "SOFT_GEN_"

// configFormats lists the extensions of the configuration files.
var configFormats = []string{".yaml", ".yml", ".json", ".toml"}

// envVar is the environment variable of an option: SOFT_GEN_ and the option
// name upper-cased, dashes as underscores (dry-run -> SOFT_GEN_DRY_RUN).
func envVar(option string) string {
	return envPrefix + strings.ToUpper(strings.ReplaceAll(option, "-", "_"))
}

// configFile is the configuration file of a run, given by --config: a YAML,
// JSON or TOML file whose top-level keys are the long option names
// (model, output, properties, template, silent, dry-run, diff, format); a
// list value (template, properties) is a list. Its path is known only once
// the options are parsed, so the sources read it lazily, each reading the
// file only when its extension is the one of its format.
type configFile struct {
	path string
}

// sources is the value source chain of an option: its environment variable,
// then its key in the configuration file.
func (c *configFile) sources(option string) cli.ValueSourceChain {
	return cli.NewValueSourceChain(
		cli.EnvVar(envVar(option)),
		altsrcyaml.YAML(option, c.format(".yaml", ".yml")),
		altsrcjson.JSON(option, c.format(".json")),
		altsrctoml.TOML(option, c.format(".toml")),
	)
}

// format returns the source of the configuration file for the readers of the
// given extensions: the path when it has one of them, else nothing.
func (c *configFile) format(extensions ...string) altsrc.Sourcer {
	return configFormat{file: c, extensions: extensions}
}

// check fails when --config names a file that does not exist or whose
// extension is not a configuration format: cli-altsrc would ignore it.
func (c *configFile) check(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if c.path == "" {
		return ctx, nil
	}
	if _, err := os.Stat(c.path); err != nil {
		return ctx, cli.Exit(fmt.Sprintf("configuration file %s: %v", c.path, err), 1)
	}
	ext := strings.ToLower(filepath.Ext(c.path))
	for _, format := range configFormats {
		if ext == format {
			return ctx, nil
		}
	}
	return ctx, cli.Exit(fmt.Sprintf("configuration file %s: unsupported extension %q (supported: %s)", c.path, ext, strings.Join(configFormats, ", ")), 1)
}

// configFormat is the configuration file seen by the reader of one format.
type configFormat struct {
	file       *configFile
	extensions []string
}

// SourceURI is the path of the configuration file when its extension is one
// of the format, else "" (the reader then finds nothing).
func (f configFormat) SourceURI() string {
	ext := strings.ToLower(filepath.Ext(f.file.path))
	for _, e := range f.extensions {
		if ext == e {
			return f.file.path
		}
	}
	return ""
}
