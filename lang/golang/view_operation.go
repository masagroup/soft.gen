package golang

import (
	"strconv"
	"strings"

	"github.com/masagroup/soft.go/ecore"
)

// operationView is the view model of an operation of an implementation
// (template operation and reflectives/einvoke, v1 lib/generateOperation).
type operationView struct {
	// Receiver and Impl are the receiver and the struct of the method;
	// AsInterface the accessor casting the object to its interface.
	Receiver, Impl, AsInterface string
	// Name is the method; Params its parameter types, comma separated; Type
	// its return type ("" when void).
	Name, Params, Type string
	// Void is whether the operation returns nothing.
	Void bool
	// ID is the operation id constant in the class.
	ID string
	// Call is the invocation from the EInvokeFromID arguments (v1 generateOperationInvoke).
	Call string
}

// operationView builds the view of an operation of the class.
func (g *gen) operationView(c ecore.EClass, op ecore.EOperation, scope *fileScope) *operationView {
	info := g.m.Operation(op)
	return &operationView{
		Receiver:    receiverName(c),
		Impl:        g.implementationName(c),
		AsInterface: asInterfaceName(g.interfaceName(c)),
		Name:        g.operationName(op),
		Params:      g.parameterTypes(op, scope),
		Type:        scope.qualify(g.typeOf(op)),
		Void:        info.IsVoid,
		ID:          scope.qualify(g.qualifiedInterfaceIdentifier(c.GetEPackage(), g.m.OperationIDName(c, op))),
		Call:        g.operationCall(op, scope),
	}
}

// operationSignature is the declaration of an operation in its interface:
// Name(ParamTypes) Type (v1 generateDeclaration(EOperation)).
func (g *gen) operationSignature(op ecore.EOperation, scope *fileScope) string {
	return strings.TrimSpace(g.operationName(op) + "(" + g.parameterTypes(op, scope) + ") " + scope.qualify(g.typeOf(op)))
}

// parameterTypes lists the parameter types of an operation, comma separated
// (v1 generateParameterList without names).
func (g *gen) parameterTypes(op ecore.EOperation, scope *fileScope) string {
	var types []string
	for it := op.GetEParameters().Iterator(); it.HasNext(); {
		types = append(types, scope.qualify(g.typeOf(it.Next().(ecore.EParameter))))
	}
	return strings.Join(types, ", ")
}

// operationCall is the call of the operation from the EInvokeFromID
// arguments: Name(arguments.Get(0).(T), ...), an argument being asserted to
// its type unless the parameter is single and of the any type (v1
// generateOperationInvoke).
func (g *gen) operationCall(op ecore.EOperation, scope *fileScope) string {
	var args []string
	for i, p := range elementsOf[ecore.EParameter](op.GetEParameters()) {
		arg := "arguments.Get(" + strconv.Itoa(i) + ")"
		if !(p.GetUpperBound() == 1 && g.m.IsAny(p.GetEType())) {
			arg += ".(" + scope.qualify(g.typeOf(p)) + ")"
		}
		args = append(args, arg)
	}
	return g.operationName(op) + "(" + strings.Join(args, ", ") + ")"
}

// elementsOf converts an EList to a slice of T.
func elementsOf[T any](list ecore.EList) []T {
	var elements []T
	for it := list.Iterator(); it.HasNext(); {
		elements = append(elements, it.Next().(T))
	}
	return elements
}
