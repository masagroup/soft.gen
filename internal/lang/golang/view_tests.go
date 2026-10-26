package golang

import (
	"sort"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/model"
	"github.com/masagroup/soft.gen/v2/internal/naming"
)

// implTestNames is the identifiers every part of the tests of an
// implementation uses.
type implTestNames struct {
	testNames
	// Ecore is the prefix of the ecore identifiers ("ecore.", "" in the ecore
	// package or when the file does not import ecore: v1 ecorePrefix).
	Ecore string
	// Class is the class name, upper-cased first, prefixing the test names.
	Class string
	// Constructor is the constructor of the implementation.
	Constructor string
}

// implTestView is the view model of the tests of an implementation
// (template impl_test.go, v1 generateImplementationTests).
type implTestView struct {
	*implTestNames
	// MockConstructor is the constructor of the mock of the class, referenced
	// by the discard function; EcoreADD is ecore.ADD when the file imports
	// ecore (v1 requiresEcore), "" otherwise.
	MockConstructor, EcoreADD string
	// Coverage is testing.Coverage, referenced by the discard function.
	Coverage string
	// AsInterface is the cast to the interface; StaticClass the expected EClass.
	AsInterface, StaticClass string
	// FeatureCount is the expected feature count: the constant, 0 in reflective delegation.
	FeatureCount string
	// MapEntry is the key and value getters of a map entry class, nil otherwise.
	MapEntry *implTestMapEntryView
	// Features lists the accessor tests of the implemented features, sorted by variable name.
	Features []*featureTestView
	// Operations lists the tests of the default operations, sorted by method name.
	Operations []operationTestView
	// Reflectives lists the tests of the reflective methods, in v1 order;
	// none in reflective delegation.
	Reflectives []*reflectiveTestView
}

// implTestMapEntryView is the typed key and value getters of a map entry class.
type implTestMapEntryView struct {
	KeyGetter, ValueGetter string
}

// operationTestView is the test of a default operation: the call panics.
type operationTestView struct {
	// Name is the method; Args the default values of its parameters.
	Name, Args string
}

// featureTestView is the view of a feature in the tests: its names, its
// values and the variant of each test.
type featureTestView struct {
	*implTestNames
	// Name is the feature name, upper-cased first, in the test names.
	Name string
	// The accessors and the field of the feature.
	Getter, Setter, Unsetter, IsSetter, BasicSetter, Var string
	// ESetVar is the ESet flag field of an unsettable single field.
	ESetVar string
	// ID is the feature id in the class; ReverseID the id of the opposite in
	// its class; ReverseGetter the meta object getter call of the opposite.
	ID, ReverseID, ReverseGetter string
	// DefaultNil is whether the default value is nil: asserted with Nil.
	DefaultNil bool
	// Any is the any type.
	Any string
	// values computes the values of the feature on demand (TestValue,
	// Type...): a value is qualified, and its package imported, only when a
	// template writes it.
	values featureTestValues
	// Reflective is whether the feature is set through its setter (reflective delegation).
	Reflective bool
	// IsProxy is whether the reference resolves proxies; IsContainer whether
	// its opposite is a containment.
	IsProxy, IsContainer bool
	// InverseAddID is the feature id an element added to a reference list is
	// inversely added with ("" when none, v1 generateFeatureMockAddedExpectations).
	InverseAddID string
	// RemoveID is the id the previous element of a single reference is
	// inversely removed with.
	RemoveID string
	// The variants of the accessor tests ("" when the test is not generated).
	GetKind, SetKind, UnsetKind, IsSetKind string
	// IsSetUnset is whether the issetter test also unsets the feature.
	IsSetUnset bool
	// ProxyInit is how the proxy getter test initializes the value: field,
	// set, or bidirectional (reflective delegation, v1 isProxy variant).
	ProxyInit string
	// BasicSetKind is the variant of the basic setter test (basicSetTestKind;
	// "" when it is not tested).
	BasicSetKind string
	// The variants of the feature in the reflective tests.
	EGetKind, ESetKind, EIsSetKind, EUnsetKind, EInverseAddKind, EInverseRemoveKind string
	// ReadBack is how the ESetFromID and EUnsetFromID tests read a single
	// value back (readBack).
	ReadBack string
}

// reflectiveTestView is the test of a reflective method (template
// test/<Kind>): its feature cases, operation ids or derived id mappings.
type reflectiveTestView struct {
	*implTestNames
	// Kind names the test: derivedid, eget, eset, eisset, eunset, einvoke,
	// einverseadd, einverseremove.
	Kind string
	// Features lists the feature cases, sorted by feature name.
	Features []*featureTestView
	// IDs lists the operation ids of EInvokeFromID.
	IDs []string
	// Element is Feature or Operation (derivedid).
	Element string
	// Mixins lists the id mappings of the mixins (derivedid).
	Mixins []mixinIDsView
}

// implementationTestFile is the tests of the implementation of a class:
// <implementation dir>/<class>_impl_test.go (v1 generateImplementationTests).
func (g *gen) implementationTestFile(c ecore.EClass) []engine.File {
	pkg := c.GetEPackage()
	scope := newFileScope(g.implementationPackagePath(pkg))
	info := g.m.Class(c)
	reflective := g.IsReflectiveDelegation()
	names := &implTestNames{
		testNames:   testNamesOf(scope, true),
		Class:       naming.UpperFirst(c.GetName()),
		Constructor: constructorName(g.implementationName(c)),
	}
	view := &implTestView{
		implTestNames:   names,
		MockConstructor: scope.qualify(g.qualifiedMockIdentifier(pkg, "New"+g.mockName(c))),
		Coverage:        scope.qualify("testing/testing.Coverage"),
		AsInterface:     asInterfaceName(g.interfaceName(c)),
		StaticClass:     g.metaObject(c, scope),
		FeatureCount:    "0",
	}
	if !g.m.IsEcore(pkg) && (len(info.ImplementedFeatures) > 0 || info.RequiresList) {
		view.EcoreADD = g.ecoreIdentifier("ADD", scope)
		names.Ecore = identifierPackagePrefix(view.EcoreADD)
	}
	if !reflective {
		view.FeatureCount = scope.qualify(g.qualifiedInterfaceIdentifier(pkg, g.m.FeatureCountIDName(c)))
	}
	if info.IsMapEntry {
		view.MapEntry = &implTestMapEntryView{
			KeyGetter:   g.getterName(c.GetEStructuralFeatureFromName("key")),
			ValueGetter: g.getterName(c.GetEStructuralFeatureFromName("value")),
		}
	}
	features := map[ecore.EStructuralFeature]*featureTestView{}
	for _, f := range g.sortedByVariable(info.ImplementedFeatures) {
		features[f] = g.featureTestView(c, f, names, scope)
		view.Features = append(view.Features, features[f])
	}
	for _, op := range g.sortedByOperationName(info.ImplementedOperations) {
		if g.m.Operation(op).IsDefault {
			view.Operations = append(view.Operations, g.operationTestView(op, scope))
		}
	}
	if !reflective {
		view.Reflectives = g.reflectiveTests(c, features, names, scope)
	}
	return []engine.File{{
		Path:   filePath(g.implementationOutputDirectory(pkg), g.implementationTestFileName(c, c.GetName())),
		Body:   goSource(scope, "impl_test.go", view),
		Header: true,
	}}
}

// metaObject is the meta object getter call of a classifier through the
// instance of its package (v1 getPackageInstanceIdentifier + getReflectiveGetterName).
func (g *gen) metaObject(c ecore.EClassifier, scope *fileScope) string {
	return scope.qualify(g.packageInstance(c.GetEPackage())) + "." + g.reflectiveGetterName(c) + "()"
}

// operationTestView builds the test of an operation (v1
// generateTestOperation): the default values of the typed parameters.
func (g *gen) operationTestView(op ecore.EOperation, scope *fileScope) operationTestView {
	args := ""
	for _, p := range elementsOf[ecore.EParameter](op.GetEParameters()) {
		if p.GetEType() == nil {
			continue
		}
		if args != "" {
			args += ", "
		}
		args += scope.qualify(g.defaultValueOfClassifier(p.GetEType()))
	}
	return operationTestView{Name: g.operationName(op), Args: args}
}

// featureTestView builds the view of a feature in the tests of the class.
func (g *gen) featureTestView(c ecore.EClass, f ecore.EStructuralFeature, names *implTestNames, scope *fileScope) *featureTestView {
	info := g.m.Feature(f)
	v := &featureTestView{
		implTestNames: names,
		Name:          naming.UpperFirst(f.GetName()),
		Getter:        g.getterName(f),
		Setter:        g.setterName(f),
		Unsetter:      g.unsetterName(f),
		IsSetter:      g.isSetterName(f),
		Var:           g.variableName(f),
		ID:            g.featureID(c, f, scope),
		Reflective:    g.IsReflectiveDelegation(),
		IsProxy:       info.IsProxy,
		IsContainer:   info.IsContainer,
		Any:           g.anyType(),
	}
	v.BasicSetter = "basic" + naming.UpperFirst(v.Setter)
	if info.Opposite != nil {
		v.ReverseID = g.featureID(info.Opposite.GetEContainingClass(), info.Opposite, scope)
		v.ReverseGetter = scope.qualify(g.packageInstance(c.GetEPackage())) + "." + g.reflectiveGetterName(info.Opposite) + "()"
	}
	v.RemoveID = v.ReverseID
	if info.IsContainment {
		v.RemoveID = names.Ecore + "EOPPOSITE_FEATURE_BASE - " + v.ID
	}
	v.InverseAddID = g.inverseAddID(v, info)
	g.featureTestKinds(v, info)
	v.values = featureTestValues{g: g, f: f, scope: scope}
	v.DefaultNil = g.defaultValueWithType(f) == "nil"
	return v
}

// featureTestValues computes the values of a feature the test variants write.
type featureTestValues struct {
	g     *gen
	f     ecore.EStructuralFeature
	scope *fileScope
}

// TestValue is the test value of the type of the feature (v1 getTestValue).
func (v *featureTestView) TestValue() string {
	return v.values.g.mockTestValue(v.values.f.GetEType(), v.values.scope)
}

// Type is the Go type of the type of the feature (v1 eType.getType()).
func (v *featureTestView) Type() string {
	return v.values.scope.qualify(v.values.g.typeOfClassifier(v.values.f.GetEType()))
}

// TypeClass is the meta object getter call of the type of the feature.
func (v *featureTestView) TypeClass() string {
	return v.values.g.metaObject(v.values.f.GetEType(), v.values.scope)
}

// Default is the default value converted to the type (v1 getDefaultValueWithType),
// or the zero value *new(T) of a single attribute without default (a data type the
// type table does not know): v1 wrote nothing there, which did not compile.
func (v *featureTestView) Default() string {
	value := v.values.g.defaultValueWithTypeIn(v.values.f, v.values.scope)
	if value == "" && !v.values.f.IsMany() {
		if _, attribute := v.values.f.(ecore.EAttribute); attribute {
			return zeroValueOf(v.values.scope.qualify(v.values.g.typeOfClassifier(v.values.f.GetEType())))
		}
	}
	return value
}

// KeyTestValue is the test value of the key of the map entry type of the feature.
func (v *featureTestView) KeyTestValue() string {
	return v.mapEntryTestValue("key")
}

// ValueTestValue is the test value of the value of the map entry type of the feature.
func (v *featureTestView) ValueTestValue() string {
	return v.mapEntryTestValue("value")
}

// mapEntryTestValue is the test value of a feature of the map entry type of the feature.
func (v *featureTestView) mapEntryTestValue(name string) string {
	entry := v.values.f.GetEType().(ecore.EClass)
	return v.values.g.mockTestValue(entry.GetEStructuralFeatureFromName(name).GetEType(), v.values.scope)
}

// inverseAddID is the id an element added to a reference list is inversely
// added with (v1 generateFeatureMockAddedExpectations): EOPPOSITE_FEATURE_BASE
// - ID for a containment without opposite, the opposite id for a reference
// with one, none otherwise.
func (g *gen) inverseAddID(v *featureTestView, info *model.FeatureInfo) string {
	switch {
	case info.Reference == nil:
		return ""
	case info.IsContainment && info.Opposite == nil:
		return v.Ecore + "EOPPOSITE_FEATURE_BASE - " + v.ID
	}
	return v.ReverseID
}

// defaultValueWithTypeIn is the default value of a feature converted to its
// type and written in the file (v1 getDefaultValueWithType): nil for a
// reference, "" for a many attribute.
func (g *gen) defaultValueWithTypeIn(f ecore.EStructuralFeature, scope *fileScope) string {
	a, ok := f.(ecore.EAttribute)
	if !ok {
		return "nil"
	}
	value := g.defaultValueOfAttribute(a)
	if value == "" || value == "nil" {
		return value
	}
	return convertedValue(scope.qualify(g.typeOfClassifier(f.GetEType())), scope.qualify(value))
}

// featureTestKinds chooses the variant of every test of the feature.
func (g *gen) featureTestKinds(v *featureTestView, info *model.FeatureInfo) {
	reflective := g.IsReflectiveDelegation()
	if info.Get {
		v.GetKind = getTestKind(info)
	}
	if info.Set {
		v.SetKind = setTestKind(info, reflective)
	}
	if info.BasicSet && !reflective {
		v.BasicSetKind = basicSetTestKind(info)
	}
	v.ProxyInit = proxyInitKind(info, reflective)
	if info.Unset {
		v.UnsetKind = unsetTestKind(info)
	}
	if info.IsSet {
		v.IsSetKind = isSetTestKind(info, reflective)
		if v.IsSetKind == "flag" {
			v.ESetVar = g.variableName(info.Feature) + "ESet"
			v.IsSetUnset = info.Unset
		}
	}
	v.EGetKind = egetTestKind(info)
	v.ESetKind = g.esetTestKind(info)
	v.EIsSetKind = "false"
	if info.IsVolatile {
		v.EIsSetKind = "volatile"
	}
	if info.IsVolatile && g.m.Annotation(info.Feature).Value("getterBody") != "" {
		// the getter is the getterBody, which v1 expected to panic: the
		// tests reading the feature only call it
		if v.GetKind == "volatile" {
			v.GetKind = "volatilebody"
		}
		v.EGetKind, v.EIsSetKind = "body", "body"
	}
	v.EUnsetKind = g.eunsetTestKind(info)
	v.EInverseAddKind = einverseAddTestKind(info)
	v.EInverseRemoveKind = einverseRemoveTestKind(info)
	v.ReadBack = readBack(v, info, reflective)
}

// readBack is how the reflective tests read a single value back: through
// EGetFromID, as v1 did, but the field when the read calls an initializer
// that would not return what was set (not in reflective delegation, which
// has no field): the generated initializer of a single reference panics (an
// unresolved proxy read too), the action one of an attribute resets its
// default value.
func readBack(v *featureTestView, info *model.FeatureInfo, reflective bool) string {
	if !reflective && info.Lazy.HasInitMethod && !info.IsList && !info.IsContainer {
		return "o." + v.Var
	}
	return "o.EGetFromID(" + v.ID + ", false)"
}

// isFeatureMapTest reports whether a feature is a feature map in the tests:
// a list of EFeatureMapEntry (v1 isFeatureMap of generateImplementationTests).
func isFeatureMapTest(info *model.FeatureInfo) bool {
	return info.IsList && info.Feature.GetEType() != nil && info.Feature.GetEType().GetName() == "EFeatureMapEntry"
}

// getTestKind is the variant of the getter test (v1
// generateGetTestImplementation, guards in declaration order): volatile,
// container, map, featuremap, list, lazy, proxy, reference, lazyattribute or
// attribute. Unlike v1, lazy is a single non-proxy reference with an
// initialization method only, whose generated initializer panics: the getter
// of a proxy resolves it without initializing it, and a lambda reference has
// no initializer. lazyattribute is an attribute of the action mode, whose
// getter calls the initializer, which resets the field to its default value;
// volatilebody a volatile feature with a getterBody, which is called.
func getTestKind(info *model.FeatureInfo) string {
	switch {
	case info.IsVolatile:
		return "volatile"
	case info.IsContainer:
		return "container"
	case info.IsMap:
		return "map"
	case isFeatureMapTest(info):
		return "featuremap"
	case info.IsList:
		return "list"
	case info.Reference != nil && info.Lazy.HasInitMethod && !info.IsProxy:
		return "lazy"
	case info.Reference != nil && info.IsProxy:
		return "proxy"
	case info.Reference != nil:
		return "reference"
	case info.Lazy.HasInitMethod && info.Lazy.Mode == "action":
		return "lazyattribute"
	}
	return "attribute"
}

// setTestKind is the variant of the setter test (v1
// generateSetTestImplementation): volatile, reflectivecontainer, container,
// reflectivebidirectional, bidirectional, containment or default.
func setTestKind(info *model.FeatureInfo, reflective bool) string {
	switch {
	case info.IsVolatile:
		return "volatile"
	case reflective && info.IsContainer:
		return "reflectivecontainer"
	case info.IsContainer:
		return "container"
	case reflective && info.IsBidirectional:
		return "reflectivebidirectional"
	case info.IsBidirectional:
		return "bidirectional"
	case info.IsContainment:
		return "containment"
	}
	return "default"
}

// basicSetTestKind is the variant of the basic setter test: volatile (the
// basic setter of a volatile feature but a container is its setterBody or
// panics, as for the setter and unsetter tests; v1 expected a SET
// notification) or default.
func basicSetTestKind(info *model.FeatureInfo) string {
	if info.IsVolatile && !info.IsContainer {
		return "volatile"
	}
	return "default"
}

// proxyInitKind is the initialization of the value of the proxy getter
// test: the field, else the setter, with the expectations of the opposite of
// a bidirectional reference (v1 also had container and containment cases,
// dead code: a proxy is neither).
func proxyInitKind(info *model.FeatureInfo, reflective bool) string {
	switch {
	case !reflective:
		return "field"
	case info.IsBidirectional:
		return "bidirectional"
	}
	return "set"
}

// unsetTestKind is the variant of the unsetter test (v1
// generateUnSetTestImplementation): volatile, list or default.
func unsetTestKind(info *model.FeatureInfo) string {
	switch {
	case info.IsVolatile:
		return "volatile"
	case info.IsList:
		return "list"
	}
	return "default"
}

// isSetTestKind is the variant of the issetter test: volatile, flag (the ESet
// flag of a single field drives the issetter and EIsSetFromID, and the
// unsetter clears it) or false (a new object has it unset).
func isSetTestKind(info *model.FeatureInfo, reflective bool) string {
	switch {
	case info.IsVolatile:
		return "volatile"
	case info.HasESet && !reflective:
		return "flag"
	}
	return "false"
}

// egetTestKind is the variant of an EGetFromID case (v1
// generateEGetFromIDTestImplementation): panics (volatile, or a single
// non-proxy reference with an initialization method), list (a reference
// list: the unresolved list too) or get.
func egetTestKind(info *model.FeatureInfo) string {
	switch {
	case info.IsVolatile || (info.Lazy.HasInitMethod && info.Reference != nil && !info.IsList && !info.IsProxy):
		return "panics"
	case !info.IsMap && info.IsList && info.Reference != nil:
		return "list"
	}
	return "get"
}

// esetTestKind is the variant of an ESetFromID case (v1
// generateFeatureESetFromIDTestImplementation): volatile, container, map (a
// map of a map entry class), none (another map), referencelist, list,
// bidirectional, containment, reference or attribute.
func (g *gen) esetTestKind(info *model.FeatureInfo) string {
	switch {
	case info.IsVolatile:
		return "volatile"
	case info.IsContainer:
		return "container"
	case info.IsMap:
		if entry, ok := info.Feature.GetEType().(ecore.EClass); ok && g.m.Class(entry).IsMapEntry {
			return "map"
		}
		return "none"
	case info.IsList && info.Reference != nil:
		return "referencelist"
	case info.IsList:
		return "list"
	case info.IsBidirectional:
		return "bidirectional"
	case info.IsContainment:
		return "containment"
	case info.Reference != nil:
		return "reference"
	}
	return "attribute"
}

// eunsetTestKind is the variant of an EUnsetFromID case (v1
// generateEUnSetFromIDTestImplementation): volatile, list, reference, nil
// (an attribute without default value), value (with one) or none.
func (g *gen) eunsetTestKind(info *model.FeatureInfo) string {
	switch {
	case info.IsVolatile:
		return "volatile"
	case info.IsList:
		return "list"
	case info.Reference != nil:
		return "reference"
	}
	switch g.defaultValueOfAttribute(info.Attribute) {
	case "":
		return "none"
	case "nil":
		return "nil"
	}
	return "value"
}

// einverseAddTestKind is the variant of an EBasicInverseAdd case (v1
// generateFeatureEInverseAddTestImplementation): none (map), list,
// container or single.
func einverseAddTestKind(info *model.FeatureInfo) string {
	switch {
	case info.IsMap:
		return "none"
	case info.IsList:
		return "list"
	case info.IsContainer:
		return "container"
	}
	return "single"
}

// einverseRemoveTestKind is the variant of an EBasicInverseRemove case (v1
// generateFeatureEInverseRemoveTestImplementation): map, list, volatile (the
// basic unsetter or setter of a volatile feature panics; v1 called it as a
// single) or single.
func einverseRemoveTestKind(info *model.FeatureInfo) string {
	switch {
	case info.IsMap:
		return "map"
	case info.IsList:
		return "list"
	case info.IsVolatile:
		return "volatile"
	}
	return "single"
}

// reflectiveTests lists the tests of the reflective methods (v1
// generateImplementationTests, not in reflective delegation): the derived
// feature and operation ids, EGetFromID, ESetFromID, EIsSetFromID,
// EUnsetFromID, EInvokeFromID, EBasicInverseAdd and EBasicInverseRemove,
// each when it has a case.
func (g *gen) reflectiveTests(c ecore.EClass, features map[ecore.EStructuralFeature]*featureTestView, names *implTestNames, scope *fileScope) []*reflectiveTestView {
	info := g.m.Class(c)
	var tests []*reflectiveTestView
	tests = append(tests, g.derivedIDTests(c, names, scope)...)
	method := func(kind string, list []ecore.EStructuralFeature, skipFeatureMaps bool) {
		test := &reflectiveTestView{implTestNames: names, Kind: kind}
		cases := 0
		for _, f := range sortedByFeatureName(list) {
			if skipFeatureMaps && isFeatureMapTest(g.m.Feature(f)) {
				continue
			}
			cases++
			if !features[f].emptyCase(kind) {
				test.Features = append(test.Features, features[f])
			}
		}
		if cases > 0 {
			tests = append(tests, test)
		}
	}
	method("eget", info.EGetFeatures, false)
	method("eset", info.ESetFeatures, true)
	method("eisset", info.EIsSetFeatures, true)
	method("eunset", info.EUnsetFeatures, true)
	if c.GetName() != "EObject" && len(info.EInvokeOperations) > 0 {
		test := &reflectiveTestView{implTestNames: names, Kind: "einvoke"}
		operations := append([]ecore.EOperation(nil), info.EInvokeOperations...)
		idName := func(op ecore.EOperation) string { return g.m.OperationIDName(op.GetEContainingClass(), op) }
		sort.SliceStable(operations, func(i, j int) bool { return idName(operations[i]) < idName(operations[j]) })
		for _, op := range operations {
			test.IDs = append(test.IDs, g.operationID(c, op, scope))
		}
		tests = append(tests, test)
	}
	method("einverseadd", info.EInverseAddFeatures, false)
	method("einverseremove", info.EInverseRemoveFeatures, false)
	return tests
}

// derivedIDTests builds the tests of EDerivedFeatureID and
// EDerivedOperationID (v1 generateElementIDTestImplementation): each when a
// mixin implements a feature (an operation), mapping the ids of every mixin.
// v1 wrote the operation id of the class unqualified; it is qualified here,
// identical within one package.
func (g *gen) derivedIDTests(c ecore.EClass, names *implTestNames, scope *fileScope) []*reflectiveTestView {
	info := g.m.Class(c)
	features, operations := false, false
	for _, mixin := range info.Mixins {
		features = features || len(g.m.Class(mixin).ImplementedFeatures) > 0
		operations = operations || len(g.m.Class(mixin).ImplementedOperations) > 0
	}
	mixinClass := func(mixin ecore.EClass) string {
		return scope.qualify(g.packageInstance(mixin.GetEPackage())) + ".Get" + g.m.ElementAccessorName(mixin) + "()"
	}
	var tests []*reflectiveTestView
	if features {
		test := &reflectiveTestView{implTestNames: names, Kind: "derivedid", Element: "Feature"}
		for _, mixin := range info.Mixins {
			ids := mixinIDsView{Class: mixinClass(mixin)}
			for _, f := range g.m.Class(mixin).ImplementedFeatures {
				ids.IDs = append(ids.IDs, idMappingView{From: g.featureID(mixin, f, scope), To: g.featureID(c, f, scope)})
			}
			test.Mixins = append(test.Mixins, ids)
		}
		tests = append(tests, test)
	}
	if operations {
		test := &reflectiveTestView{implTestNames: names, Kind: "derivedid", Element: "Operation"}
		for _, mixin := range info.Mixins {
			ids := mixinIDsView{Class: mixinClass(mixin)}
			for _, op := range g.m.Class(mixin).ImplementedOperations {
				ids.IDs = append(ids.IDs, idMappingView{From: g.operationID(mixin, op, scope), To: g.operationID(c, op, scope)})
			}
			test.Mixins = append(test.Mixins, ids)
		}
		tests = append(tests, test)
	}
	return tests
}

// emptyCase reports whether the case of the feature in the test of a
// reflective method writes nothing (v1 empty variants: a map that is not of a
// map entry class in ESetFromID, a map in EBasicInverseAdd); the test still
// exists when its only cases are empty.
func (v *featureTestView) emptyCase(kind string) bool {
	switch kind {
	case "eset":
		return v.ESetKind == "none"
	case "einverseadd":
		return v.EInverseAddKind == "none"
	}
	return false
}
