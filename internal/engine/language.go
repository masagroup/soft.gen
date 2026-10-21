package engine

import (
	"io/fs"
	"text/template"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/naming"
	"github.com/masagroup/soft.gen/v2/internal/props"
	"github.com/masagroup/soft.gen/v2/internal/writer"
)

// Language is a code generator for one target language.
//
// A language contributes the facts-to-text part of a generation only: which
// files an entry point produces, the view model of each file and the template
// that lays it out. Every fact about the model comes from the model package,
// every property from the props package, every name rule from naming.
type Language interface {
	// Name is the command-line subcommand: "go", "ts", "cpp". The generator of
	// the language is named after it in the header of the generated files and
	// in the manifests: "soft-gen go".
	Name() string
	// Version is the version of the language: the version of the runtime its
	// templates target (soft.go, soft.ts...), "v1.13.1". It is written in the
	// manifests and printed by "soft-gen <language> -v". It is a constant of
	// the language, so that every build has it (go install, go run ...@v2.x).
	Version() string
	// AnnotationSource is the source of the EAnnotations the language reads:
	// "http://net.masagroup/soft/2019/GenGo".
	AnnotationSource() string
	// Properties declares the properties the language understands.
	Properties() []props.Decl
	// EntryPoints lists the templates of the -t option, in the order they run
	// by default.
	EntryPoints() []EntryPoint
	// Templates is the file system holding the template files (*.tmpl).
	Templates() fs.FS
	// Funcs adds presentational template functions to the base ones; may be nil.
	Funcs() template.FuncMap
	// Regions is the syntax of the user code regions; nil when the language has none.
	Regions() *writer.RegionSyntax
	// Formatter formats the generated files; nil when the language has none.
	Formatter() writer.Formatter
	// NameStyle is the syntax of the naming rules of the annotations.
	NameStyle() naming.Style
}

// EntryPoint is a named template of the -t option: it lists the files it
// produces for a package.
type EntryPoint struct {
	// Name is the name given to -t, as in v1: generateModel, generateTests...
	Name string
	// Doc is one sentence for the usage text.
	Doc string
	// Default marks the entry points that run when -t is absent.
	Default bool
	// Files lists the files to generate for the root package.
	Files FileLister
}

// FileLister lists the files an entry point generates for a package. It
// decides which files exist; it writes no text.
type FileLister func(ctx *Context, pkg ecore.EPackage) ([]File, error)

// File is one generated file.
type File struct {
	// Path is the slash-separated path of the file, relative to the output directory.
	Path string
	// Body renders the content of the file (without the header).
	Body Source
	// Header prepends the generated-code header (the "DO NOT EDIT" line and the fileNotes).
	Header bool
	// Protected preserves the user code regions of the previous version of the file.
	Protected bool
}

// Source renders the text of a file.
type Source interface {
	Render(ctx *Context) ([]byte, error)
}

// Templated renders the named template with data.
func Templated(name string, data any) Source {
	return templated{name: name, data: data}
}

type templated struct {
	name string
	data any
}

func (t templated) Render(ctx *Context) ([]byte, error) {
	return ctx.Templates.Render(t.name, t.data)
}

// Bytes is a fixed content, such as a bundled file copied to the output.
func Bytes(content []byte) Source {
	return fixed(content)
}

type fixed []byte

func (f fixed) Render(*Context) ([]byte, error) { return []byte(f), nil }

// Func renders with a function, for files needing more than one template
// (for example a body rendered first to know the imports of the header).
func Func(render func(ctx *Context) ([]byte, error)) Source {
	return funcSource(render)
}

type funcSource func(ctx *Context) ([]byte, error)

func (f funcSource) Render(ctx *Context) ([]byte, error) { return f(ctx) }
