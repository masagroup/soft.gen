package golang

import (
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// enumView is the view model of the file of an enum (template enum.go).
type enumView struct {
	// Name is the enum type.
	Name string
	// Var is the receiver and map variable prefix: the name, lower-cased first.
	Var string
	// Literals lists the literals, sorted by name.
	Literals []enumLiteralView
}

// enumLiteralView is one literal of an enum.
type enumLiteralView struct {
	// Const is the constant of the literal.
	Const string
	// Literal is the string of the literal.
	Literal string
	// Value is the integer value of the literal.
	Value string
}

// enumFile is the file of an enum: <interface dir>/<enum>.go.
func (g *gen) enumFile(e ecore.EEnum) engine.File {
	pkg := e.GetEPackage()
	scope := newFileScope(g.interfacePackagePath(pkg))
	view := enumView{Name: e.GetName(), Var: naming.LowerFirst(e.GetName())}
	var literals []ecore.EEnumLiteral
	for it := e.GetELiterals().Iterator(); it.HasNext(); {
		literals = append(literals, it.Next().(ecore.EEnumLiteral))
	}
	for _, l := range sortedByName(literals) {
		view.Literals = append(view.Literals, enumLiteralView{
			Const:   g.constantsName(l),
			Literal: g.m.EnumLiteral(l),
			Value:   strconv.Itoa(l.GetValue()),
		})
	}
	return engine.File{
		Path:   filePath(g.interfaceOutputDirectory(pkg), g.interfaceFileName(e, e.GetName())),
		Body:   goSource(scope, "enum.go", view),
		Header: true,
	}
}
