package golang

import (
	"bytes"
	"fmt"

	"github.com/masagroup/soft.gen/v2/engine"
)

// goFile renders a Go source file: the package clause, the import block of
// the packages the body used, then the body template. The view model is built
// with the scope before rendering, so the body only holds placeholders for the
// import aliases, resolved once the set of used packages is complete.
type goFile struct {
	scope    *fileScope
	template string
	data     any
}

// goSource returns the Source of a Go file whose body is the named template
// rendered with data, in the package of the scope.
func goSource(scope *fileScope, template string, data any) engine.Source {
	return goFile{scope: scope, template: template, data: data}
}

// Render renders the file.
func (f goFile) Render(ctx *engine.Context) ([]byte, error) {
	body, err := ctx.Templates.Render(f.template, f.data)
	if err != nil {
		return nil, err
	}
	body, imports := f.scope.finish(body)
	var out bytes.Buffer
	fmt.Fprintf(&out, "package %s\n\n", packageNameFromPath(f.scope.packagePath))
	if len(imports) > 0 {
		out.WriteString("import (\n")
		for _, spec := range imports {
			if spec.Alias != "" {
				fmt.Fprintf(&out, "\t%s %q\n", spec.Alias, spec.Dir)
			} else {
				fmt.Fprintf(&out, "\t%q\n", spec.Dir)
			}
		}
		out.WriteString(")\n\n")
	}
	out.Write(body)
	out.WriteString("\n")
	return out.Bytes(), nil
}
