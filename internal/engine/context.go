package engine

import (
	"io"
	"log/slog"

	"github.com/masagroup/soft.gen/v2/internal/model"
	"github.com/masagroup/soft.gen/v2/internal/props"
	"github.com/masagroup/soft.gen/v2/internal/tmpl"
	"github.com/masagroup/soft.gen/v2/internal/writer"
)

// Context is everything a file lister or a view-model builder needs during a
// run. It is built once per run.
type Context struct {
	// Model is the loaded model and its facts.
	Model *model.Model
	// Props is the resolved properties of the run.
	Props *props.Properties
	// Lang is the language being generated.
	Lang Language
	// Templates is the template set of the language.
	Templates *tmpl.Set
	// Out writes the generated files.
	Out *writer.Writer
	// Log receives the diagnostics; the "debug" property raises its level.
	Log *slog.Logger
	// Request is the request of the run.
	Request Request
}

// Request is a parsed command line.
type Request struct {
	// Lang is the language to generate.
	Lang Language
	// ModelPath is the .ecore file.
	ModelPath string
	// OutDir is the output directory.
	OutDir string
	// Version is the version of the language, written in the manifests; empty
	// leaves it out.
	Version string
	// Props holds the -p key=value entries.
	Props map[string]string
	// PropFiles holds the -P files, in order.
	PropFiles []string
	// EntryPoints holds the -t names; empty means the defaults of the language.
	EntryPoints []string
	// Silent turns the "file generated:" lines off.
	Silent bool
	// DryRun lists the files without writing them.
	DryRun bool
	// Diff prints the differences with the existing files and writes nothing.
	Diff bool
	// Format applies the formatter of the language (always on for in-process
	// formatters; external ones need it).
	Format bool
	// NoManifest leaves the manifests out: the folders the run generates files
	// in get no .soft.gen.yaml (an existing one is left as it is).
	NoManifest bool
	// OnRender, when set, is called with the name of every template rendered
	// (the coverage tests record them).
	OnRender func(template string)
	// Output receives the logs and diffs (default: stdout).
	Output io.Writer
	// ErrOutput receives the warnings (default: stderr).
	ErrOutput io.Writer
}

// Report is what a run did.
type Report struct {
	// Files lists every file with its status.
	Files []FileResult
}

// FileResult is the status of one generated file.
type FileResult struct {
	Path   string
	Status writer.Status
}

// Count returns how many files have the status.
func (r *Report) Count(status writer.Status) int {
	n := 0
	for _, f := range r.Files {
		if f.Status == status {
			n++
		}
	}
	return n
}
