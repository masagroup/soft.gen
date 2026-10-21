package typescript

import (
	"github.com/masagroup/soft.gen/v2/engine"
)

// tsFile renders a TypeScript source file: the imports of the identifiers its
// view model used, a blank line, then the body template. The view model is
// built with the scope before rendering, so the imports are complete.
type tsFile struct {
	scope    *scope
	template string
	data     any
}

// tsSource returns the Source of a TypeScript file whose body is the named
// template rendered with data.
func tsSource(s *scope, template string, data any) engine.Source {
	return tsFile{scope: s, template: template, data: data}
}

// Render renders the file.
func (f tsFile) Render(ctx *engine.Context) ([]byte, error) {
	body, err := ctx.Templates.Render(f.template, f.data)
	if err != nil {
		return nil, err
	}
	body = []byte(breakLongLines(string(body), f.scope.g.printWidth()))
	var out []byte
	imports := f.scope.scannedImports(string(body)) + f.scope.imports()
	if imports != "" {
		out = append(out, imports...)
		out = append(out, '\n')
	}
	out = append(out, body...)
	return append(out, '\n'), nil
}

// file returns a generated file of the package: its path in the directory of
// the package, its body, the header.
func (g *gen) file(name string, s *scope, template string, data any) engine.File {
	return engine.File{Path: g.filePath(g.pkg, name), Body: tsSource(s, template, data), Header: true}
}
