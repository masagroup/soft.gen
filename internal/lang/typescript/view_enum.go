package typescript

import (
	"sort"
	"strconv"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// enumView is the view model of an enum (v1 generateEnum): the enum, its
// literals sorted by name, the conversions to and from their literal.
type enumView struct {
	Name     string
	Var      string
	Literals []literalView
}

// literalView is a literal of an enum: its constant, value and literal.
type literalView struct {
	Const, Value, Literal string
}

// enumFile is the file of an enum.
func (g *gen) enumFile(e ecore.EEnum) engine.File {
	var literals []ecore.EEnumLiteral
	for it := e.GetELiterals().Iterator(); it.HasNext(); {
		literals = append(literals, it.Next().(ecore.EEnumLiteral))
	}
	sort.SliceStable(literals, func(i, j int) bool { return literals[i].GetName() < literals[j].GetName() })
	view := enumView{Name: e.GetName(), Var: naming.LowerFirst(e.GetName())}
	for _, l := range literals {
		view.Literals = append(view.Literals, literalView{
			Const: literalConstantName(l), Value: strconv.Itoa(l.GetValue()), Literal: quote(l.GetLiteral()),
		})
	}
	return g.file(g.interfaceFileName(e, e.GetName()), g.newScope(), "enum.ts", view)
}
