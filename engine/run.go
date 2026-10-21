package engine

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/masagroup/soft.gen/v2/model"
	"github.com/masagroup/soft.gen/v2/props"
	"github.com/masagroup/soft.gen/v2/tmpl"
	"github.com/masagroup/soft.gen/v2/writer"
)

// Run performs a generation: it loads the model, resolves the properties,
// loads the templates, then, for every selected entry point, lists, renders and
// writes the files. It stops at the first error.
func Run(ctx context.Context, req Request) (*Report, error) {
	if req.Lang == nil {
		return nil, fmt.Errorf("no language")
	}
	if req.Output == nil {
		req.Output = os.Stdout
	}
	if req.ErrOutput == nil {
		req.ErrOutput = os.Stderr
	}
	lang := req.Lang

	entryPoints, err := selectEntryPoints(lang, req.EntryPoints)
	if err != nil {
		return nil, err
	}

	m, err := model.Load(req.ModelPath, lang.AnnotationSource())
	if err != nil {
		return nil, fmt.Errorf("model %s: %w", req.ModelPath, err)
	}

	properties, err := resolveProperties(lang, req, m, entryPoints)
	if err != nil {
		return nil, err
	}
	warnUnknownProperties(req.ErrOutput, properties)

	templates, err := tmpl.Load(lang.Templates(), lang.Funcs())
	if err != nil {
		return nil, fmt.Errorf("templates of %s: %w", lang.Name(), err)
	}
	templates.OnRender(req.OnRender)

	level := slog.LevelInfo
	if properties.Bool("debug") {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(req.ErrOutput, &slog.HandlerOptions{Level: level}))

	out := writer.New(req.OutDir,
		writer.Silent(req.Silent),
		writer.DryRun(req.DryRun),
		writer.Diff(req.Diff),
		writer.Output(req.Output),
		writer.ErrOutput(req.ErrOutput),
		writer.Header(headerOf(lang, properties)),
		writer.Regions(lang.Regions()),
		writer.Format(formatterOf(lang, req.Format)),
	)

	run := &Context{Model: m, Props: properties, Lang: lang, Templates: templates, Out: out, Log: logger, Request: req}
	report := &Report{}
	for _, ep := range entryPoints {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		logger.Debug("entry point", "name", ep.Name)
		files, err := ep.Files(run, m.Root())
		if err != nil {
			return report, fmt.Errorf("%s: %w", ep.Name, err)
		}
		for _, file := range files {
			content, err := file.Body.Render(run)
			if err != nil {
				return report, fmt.Errorf("%s: %s: %w", ep.Name, file.Path, err)
			}
			var flags writer.Flags
			if file.Header {
				flags |= writer.WithHeader
			}
			if file.Protected {
				flags |= writer.WithRegions
			}
			status, err := out.Write(file.Path, content, flags)
			if err != nil {
				return report, fmt.Errorf("%s: %w", ep.Name, err)
			}
			report.Files = append(report.Files, FileResult{Path: file.Path, Status: status})
		}
	}
	if req.NoManifest {
		return report, nil
	}
	return report, writeManifests(out, manifestOf(req, m), report)
}

// writeManifests writes the manifest in every folder the run generated files
// in, through the writer: --dry-run lists them and --diff compares them as the
// other files.
func writeManifests(out *writer.Writer, manifest Manifest, report *Report) error {
	content, err := manifest.Content()
	if err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	for _, p := range manifestPaths(report.Files) {
		status, err := out.Write(p, content, 0)
		if err != nil {
			return err
		}
		report.Files = append(report.Files, FileResult{Path: p, Status: status})
	}
	return nil
}

// selectEntryPoints returns the entry points named by the request, or the
// default ones of the language when the request names none.
func selectEntryPoints(lang Language, names []string) ([]EntryPoint, error) {
	all := lang.EntryPoints()
	if len(names) == 0 {
		var defaults []EntryPoint
		for _, ep := range all {
			if ep.Default {
				defaults = append(defaults, ep)
			}
		}
		return defaults, nil
	}
	byName := map[string]EntryPoint{}
	var known []string
	for _, ep := range all {
		byName[ep.Name] = ep
		known = append(known, ep.Name)
	}
	var selected []EntryPoint
	for _, name := range names {
		ep, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("unknown template %q (templates: %s)", name, strings.Join(known, ", "))
		}
		selected = append(selected, ep)
	}
	return selected, nil
}

// resolveProperties builds the property store of the run: the reserved
// declarations, the declarations of the language, the command line, the
// property files and the generator annotation of the root package. nsURI and
// templates are set as in v1.
func resolveProperties(lang Language, req Request, m *model.Model, entryPoints []EntryPoint) (*props.Properties, error) {
	decls := append(props.Reserved(), lang.Properties()...)
	annotation := m.Annotation(m.Root())
	properties, err := props.New(decls, req.Props, req.PropFiles, annotation.Lookup)
	if err != nil {
		return nil, err
	}
	properties.Set("nsURI", lang.AnnotationSource())
	names := make([]string, len(entryPoints))
	for i, ep := range entryPoints {
		names[i] = ep.Name
	}
	properties.Set("templates", strings.Join(names, ","))
	return properties, nil
}

// warnUnknownProperties reports the properties nobody declared.
func warnUnknownProperties(out io.Writer, properties *props.Properties) {
	for _, key := range properties.Unknown() {
		fmt.Fprintf(out, "warning: property %q is not known\n", key)
	}
}

// headerOf returns the header function of a language: the v1 fileComment
// template ("// Code generated by soft-gen <language>. DO NOT EDIT.", a blank
// line, the fileNotes property and a newline).
func headerOf(lang Language, properties *props.Properties) func(path string) string {
	header := "// Code generated by " + GeneratorOf(lang) + ". DO NOT EDIT.\n\n" + properties.Get("fileNotes") + "\n"
	return func(string) string { return header }
}

// formatterOf returns the formatter to apply: an in-process formatter always
// applies; an external one only when the request asks for formatting and the
// command is available.
func formatterOf(lang Language, format bool) writer.Formatter {
	formatter := lang.Formatter()
	if external, ok := formatter.(writer.ExecFormatter); ok {
		if !format || !external.Available() {
			return nil
		}
	}
	return formatter
}
