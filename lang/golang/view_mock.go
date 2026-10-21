package golang

import (
	"strconv"
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/naming"
)

// testifyMockPath is the package path of testify/mock.
const testifyMockPath = "github.com/stretchr/testify/mock/mock"

// testifyNames are the identifiers of testify/mock a mock uses, qualified for the file.
type testifyNames struct {
	// Mock is mock.Mock; Call mock.Call; Arguments mock.Arguments; TestingT mock.TestingT.
	Mock, Call, Arguments, TestingT string
}

// testifyNamesFor qualifies the testify/mock identifiers for the scope (v1 hard-coded mock.).
func testifyNamesFor(scope *fileScope) testifyNames {
	q := func(name string) string { return scope.qualify(qualifiedIdentifier(testifyMockPath, name)) }
	return testifyNames{Mock: q("Mock"), Call: q("Call"), Arguments: q("Arguments"), TestingT: q("TestingT")}
}

// mockView is the view model of the mock file of a class (template mock.go,
// v1 tests/generateMock).
type mockView struct {
	testifyNames
	// Name is the mock struct; Interface the interface it mocks; Receiver the
	// receiver of the EXPECT method.
	Name, Interface, Receiver string
	// Extends lists the embedded mocks.
	Extends []mockExtendView
	// Accessors lists the accessors of the implemented features, sorted by
	// feature name, each feature in getter, setter, unsetter order.
	Accessors []*mockAccessorView
	// Operations lists the implemented operations, sorted by method name.
	Operations []*mockOperationView
}

// mockExtendView is an embedded mock.
type mockExtendView struct {
	// Prototype and Expecter are the embedded prototype and expecter structs
	// as the file writes them; Simple is the mock name without package.
	Prototype, Expecter, Simple string
}

// mockMethodView holds what every mocked method uses.
type mockMethodView struct {
	testifyNames
	// Receiver is the receiver of the prototype method; Mock the mock struct.
	Receiver, Mock string
	// Method is the mocked method; CallType the struct of its expected call.
	Method, CallType string
}

// mockAccessorView is the mock of a feature accessor (templates
// mock/feature.<Kind>, v1 generateGetMock, generateSetMock, generateUnSetMock).
type mockAccessorView struct {
	mockMethodView
	// Kind is get, set or unset.
	Kind string
	// Var is the variable of the feature (v1 getVariableName); Type its Go type.
	Var, Type string
	// Cast is the assertion of a set argument to Type ("" for the any type).
	Cast string
}

// mockOperationView is the mock of an operation (templates
// mock/operation.<Kind>, v1 generateMock(EOperation)).
type mockOperationView struct {
	mockMethodView
	// Kind is void or value.
	Kind string
	// Params is the named parameter list; Types the parameter types.
	Params, Types string
	// Args is the comma separated parameter names; On the arguments of mock.On.
	Args, On string
	// ExpecterParams is the parameter list of the expecter method: every
	// parameter of the any type for a value operation, Params for a void one.
	ExpecterParams string
	// ParamDocs lists "name Type" per parameter, for the expecter comment.
	ParamDocs []string
	// RunArgs is the arguments given to the run function, asserted to their types.
	RunArgs string
	// Type is the return type; Cast the assertion of the returned value ("" for the any type).
	Type, Cast string
	// Lead prefixes the comment of the operation: "} " when the previous
	// operation is void, whose template does not close its last function
	// (v1 glued them on one line).
	Lead string
	// OpenEnd is whether the last function is left open, closed by the Lead
	// of the next operation.
	OpenEnd bool
}

// mockFile is the mock of a class: <mock dir>/<class>_mock.go (v1 generateMock).
func (g *gen) mockFile(c ecore.EClass) []engine.File {
	pkg := c.GetEPackage()
	scope := newFileScope(g.mockPackagePath(pkg))
	info := g.m.Class(c)
	view := &mockView{
		testifyNames: testifyNamesFor(scope),
		Name:         g.mockName(c),
		Interface:    g.interfaceName(c),
		Receiver:     receiverName(c),
	}
	for _, extends := range g.mockExtends(c) {
		view.Extends = append(view.Extends, mockExtendView{
			Prototype: scope.qualify(extends + "_Prototype"),
			Expecter:  scope.qualify(extends + "_Expecter"),
			Simple:    simpleIdentifier(extends),
		})
	}
	for _, f := range sortedByFeatureName(info.ImplementedFeatures) {
		view.Accessors = append(view.Accessors, g.mockAccessors(c, f, scope)...)
	}
	for _, op := range g.sortedByOperationName(info.ImplementedOperations) {
		view.Operations = append(view.Operations, g.mockOperation(c, op, scope))
	}
	glueVoidOperations(view.Operations)
	return []engine.File{{
		Path:   filePath(g.mockOutputDirectory(pkg), g.mockFileName(c, c.GetName())),
		Body:   goSource(scope, "mock.go", view),
		Header: true,
	}}
}

// mockExtends lists the qualified mocks the mock of a class embeds (v1
// getMockExtends): the mock of the extended class, else MockENotifier for
// EObject, the rootExtendsMock property, MockEObjectInternal and
// MockEMapEntry for a map entry, or MockEObjectInternal.
func (g *gen) mockExtends(c ecore.EClass) []string {
	e := g.m.Ecore()
	info := g.m.Class(c)
	switch {
	case info.ExtendsClass != nil:
		return []string{g.classMock(info.ExtendsClass)}
	case c.GetName() == "EObject":
		return []string{g.qualifiedInterfaceIdentifier(e, "MockENotifier")}
	}
	if root, ok := g.explicitPackageProperty("rootExtendsMock"); ok {
		return []string{root}
	}
	if info.IsMapEntry {
		return []string{g.qualifiedMockIdentifier(e, "MockEObjectInternal"), g.qualifiedInterfaceIdentifier(e, "MockEMapEntry")}
	}
	return []string{g.qualifiedMockIdentifier(e, "MockEObjectInternal")}
}

// mockMethod builds the names shared by the mocked method of a class.
func (g *gen) mockMethod(c ecore.EClass, method string, scope *fileScope) mockMethodView {
	mock := g.mockName(c)
	return mockMethodView{
		testifyNames: testifyNamesFor(scope),
		Receiver:     receiverName(c),
		Mock:         mock,
		Method:       method,
		CallType:     mock + "_" + method + "_Call",
	}
}

// mockAccessors lists the mocked accessors of a feature (v1
// generateMock(EStructuralFeature)): getter, setter, unsetter.
func (g *gen) mockAccessors(c ecore.EClass, f ecore.EStructuralFeature, scope *fileScope) []*mockAccessorView {
	info := g.m.Feature(f)
	t := scope.qualify(g.typeOf(f))
	accessor := func(kind, method string) *mockAccessorView {
		return &mockAccessorView{mockMethodView: g.mockMethod(c, method, scope), Kind: kind, Var: g.variableName(f), Type: t}
	}
	var accessors []*mockAccessorView
	if info.Get {
		accessors = append(accessors, accessor("get", g.getterName(f)))
	}
	if info.Set {
		set := accessor("set", g.setterName(f))
		if !g.isAnyType(t) {
			set.Cast = ".(" + t + ")"
		}
		accessors = append(accessors, set)
	}
	if info.Unset {
		accessors = append(accessors, accessor("unset", g.unsetterName(f)))
	}
	return accessors
}

// mockOperation builds the mock of an operation (v1 generateMock(EOperation)).
func (g *gen) mockOperation(c ecore.EClass, op ecore.EOperation, scope *fileScope) *mockOperationView {
	v := &mockOperationView{mockMethodView: g.mockMethod(c, g.operationName(op), scope), Kind: "value"}
	var params, types, names, anys, runArgs []string
	for i, p := range elementsOf[ecore.EParameter](op.GetEParameters()) {
		t := scope.qualify(g.typeOf(p))
		params = append(params, p.GetName()+" "+t)
		types = append(types, t)
		names = append(names, p.GetName())
		anys = append(anys, p.GetName()+" any")
		arg := "_args[" + strconv.Itoa(i) + "]"
		if !(p.GetUpperBound() == 1 && g.m.IsAny(p.GetEType())) {
			arg += ".(" + t + ")"
		}
		runArgs = append(runArgs, arg)
	}
	v.Params, v.Types, v.Args = strings.Join(params, ", "), strings.Join(types, ", "), strings.Join(names, ", ")
	v.On = strings.Join(append([]string{strconv.Quote(v.Method)}, names...), ", ")
	v.ParamDocs = params
	v.RunArgs = strings.Join(runArgs, ", ")
	v.ExpecterParams = strings.Join(anys, ", ")
	if g.m.Operation(op).IsVoid {
		v.Kind = "void"
		v.ExpecterParams = v.Params
		return v
	}
	v.Type = scope.qualify(g.typeOf(op))
	if !g.m.IsAny(op.GetEType()) {
		v.Cast = ".(" + v.Type + ")"
	}
	return v
}

// glueVoidOperations reproduces the v1 layout of a void operation followed by
// another operation: the void template ends without a newline, so the closing
// brace of its last function and the comment of the next operation share a line.
func glueVoidOperations(operations []*mockOperationView) {
	for i := 1; i < len(operations); i++ {
		if operations[i-1].Kind == "void" {
			operations[i-1].OpenEnd = true
			operations[i].Lead = "} "
		}
	}
}

// mockTestNames are the identifiers a mock test uses, qualified for the file.
type mockTestNames struct {
	// T is *testing.T; Equal assert.Equal.
	T, Equal string
}

// mockTestView is the view model of the tests of the mock of a class
// (template mock_test.go, v1 tests/generateMockTests).
type mockTestView struct {
	// Discard is the function keeping the imports used.
	Discard string
	// Equal is assert.Equal; Coverage testing.Coverage; ADD ecore.ADD, ""
	// when the file does not discard the ecore import.
	Equal, Coverage, ADD string
	// Tests lists the tests of the accessors (sorted by feature name), then of
	// the operations (sorted by method name).
	Tests []*mockTestCaseView
}

// mockTestCaseView is the test of a mocked method (templates
// mocktest/feature.<Kind> and mocktest/operation.<Kind>, v1
// generateMockTests).
type mockTestCaseView struct {
	mockTestNames
	// Kind is feature.get, feature.set, feature.unset, operation.void or operation.value.
	Kind string
	// Test is the test function; New the mock constructor; Method the tested method.
	Test, New, Method string
	// Var is the variable of the returned value of a getter: l for a list, else r.
	Var string
	// Value is the test value of the feature or of the returned value.
	Value string
	// Type is the type of the feature or of the returned value.
	Type string
	// Run is the NewMockRun call.
	Run string
	// Locals lists the parameters of an operation with their test value.
	Locals []mockLocalView
	// Args is the comma separated parameter names; Params the named parameter list.
	Args, Params string
}

// mockLocalView is a local variable of a test.
type mockLocalView struct {
	Name, Value string
}

// mockTestFile is the tests of the mock of a class: <mock dir>/<class>_mock_test.go (v1 generateMockTests).
func (g *gen) mockTestFile(c ecore.EClass) []engine.File {
	pkg := c.GetEPackage()
	scope := newFileScope(g.mockPackagePath(pkg))
	info := g.m.Class(c)
	mock := g.mockName(c)
	view := &mockTestView{
		Discard:  "discard" + naming.UpperFirst(mock),
		Equal:    scope.qualify("github.com/stretchr/testify/assert/assert.Equal"),
		Coverage: scope.qualify("testing/testing.Coverage"),
	}
	if !g.m.IsEcore(pkg) && (len(info.ImplementedFeatures) > 0 || info.RequiresList) {
		view.ADD = g.ecoreIdentifier("ADD", scope)
	}
	for _, f := range sortedByFeatureName(info.ImplementedFeatures) {
		view.Tests = append(view.Tests, g.mockAccessorTests(c, f, scope)...)
	}
	for _, op := range g.sortedByOperationName(info.ImplementedOperations) {
		view.Tests = append(view.Tests, g.mockOperationTest(c, op, scope))
	}
	return []engine.File{{
		Path:   filePath(g.mockOutputDirectory(pkg), g.mockTestFileName(c, c.GetName())),
		Body:   goSource(scope, "mock_test.go", view),
		Header: true,
	}}
}

// ecoreIdentifier is an identifier of the ecore package as the file writes it.
func (g *gen) ecoreIdentifier(name string, scope *fileScope) string {
	return scope.qualify(g.qualifiedInterfaceIdentifier(g.m.Ecore(), name))
}

// mockTestCase builds the names shared by the tests of the mock of a class.
func (g *gen) mockTestCase(c ecore.EClass, kind, method string, scope *fileScope) *mockTestCaseView {
	mock := g.mockName(c)
	return &mockTestCaseView{
		mockTestNames: mockTestNames{T: "*" + scope.qualify("testing/testing.T"), Equal: scope.qualify("github.com/stretchr/testify/assert/assert.Equal")},
		Kind:          kind,
		Test:          "Test" + mock + naming.UpperFirst(method),
		New:           "New" + mock,
		Method:        method,
	}
}

// mockRun is the NewMockRun call with the arguments (v1 ecorePrefix + NewMockRun).
func (g *gen) mockRun(args string, scope *fileScope) string {
	if args != "" {
		args = ", " + args
	}
	return g.ecoreIdentifier("NewMockRun", scope) + "(t" + args + ")"
}

// mockAccessorTests lists the tests of the mocked accessors of a feature (v1
// generateGetMockTest, generateSetMockTest, generateUnSetMockTest).
func (g *gen) mockAccessorTests(c ecore.EClass, f ecore.EStructuralFeature, scope *fileScope) []*mockTestCaseView {
	info := g.m.Feature(f)
	t := scope.qualify(g.typeOf(f))
	var tests []*mockTestCaseView
	if info.Get {
		get := g.mockTestCase(c, "feature.get", g.getterName(f), scope)
		get.Type, get.Run, get.Var = t, g.mockRun("", scope), "r"
		switch {
		case info.IsMap:
			get.Var, get.Value = "l", g.ecoreIdentifier("NewMockEMap", scope)+"(t)"
		case info.IsList:
			get.Var, get.Value = "l", g.ecoreIdentifier("NewMockEList", scope)+"(t)"
		default:
			get.Value = g.mockTestValue(f.GetEType(), scope)
		}
		tests = append(tests, get)
	}
	if info.Set {
		set := g.mockTestCase(c, "feature.set", g.setterName(f), scope)
		set.Type, set.Run, set.Value = t, g.mockRun("v", scope), g.mockTestValue(f.GetEType(), scope)
		tests = append(tests, set)
	}
	if info.Unset {
		unset := g.mockTestCase(c, "feature.unset", g.unsetterName(f), scope)
		unset.Run = g.mockRun("", scope)
		tests = append(tests, unset)
	}
	return tests
}

// mockOperationTest builds the test of a mocked operation (v1 generateMockTests(EOperation)).
func (g *gen) mockOperationTest(c ecore.EClass, op ecore.EOperation, scope *fileScope) *mockTestCaseView {
	test := g.mockTestCase(c, "operation.value", g.operationName(op), scope)
	var names, params []string
	for _, p := range elementsOf[ecore.EParameter](op.GetEParameters()) {
		test.Locals = append(test.Locals, mockLocalView{Name: p.GetName(), Value: g.mockTestValue(p.GetEType(), scope)})
		names = append(names, p.GetName())
		params = append(params, p.GetName()+" "+scope.qualify(g.typeOf(p)))
	}
	test.Args, test.Params = strings.Join(names, ", "), strings.Join(params, ", ")
	test.Run = g.mockRun(test.Args, scope)
	if g.m.Operation(op).IsVoid {
		test.Kind = "operation.void"
		return test
	}
	test.Value = g.mockTestValue(op.GetEType(), scope)
	test.Type = scope.qualify(g.typeOf(op))
	return test
}

// mockTestValue is the test value of a classifier as the file writes it (v1
// getTestValue): the type and the value of a data type are qualified apart,
// the conversion of a pointer type between parentheses.
func (g *gen) mockTestValue(c ecore.EClassifier, scope *fileScope) string {
	dt, ok := c.(ecore.EDataType)
	if _, isEnum := c.(ecore.EEnum); !ok || isEnum {
		return scope.qualify(g.testValue(c))
	}
	t := g.typeOfClassifier(dt)
	value := "(" + scope.qualify(g.testValueOfDataType(dt)) + ")"
	if strings.HasPrefix(t, "*") {
		return "(" + scope.qualify(t) + ")" + value
	}
	return scope.qualify(t) + value
}
