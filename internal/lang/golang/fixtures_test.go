package golang

import (
	"testing"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/engine/goldentest"
)

// The fixtures are small purpose-built models (testdata/models), each about
// one topic, generated with entry points and properties into the committed
// expected files of testdata/golden/go/<case>. The expected files are the
// output of the v1 Go generator (1.13.1) once formatted, with the v2 header
// line and the v2 manifests, except the files of V1Differences, where v2
// differs from v1 on purpose:
//
//	go test ./lang/golang -run TestFixtures            compare v2 with the expected files
//	go test ./lang/golang -run TestFixtures -update    write v2's output as the expected files
//	SOFT_GEN_V1_IMAGE=... go test ./lang/golang -run TestV1  compare v1 with the expected files
//	SOFT_GEN_V1_IMAGE=... go test ./lang/golang -run TestV1 -v1update  write v1's output
//
// TestTemplateCoverage checks that the fixtures render every template.

// suite holds the Go fixtures; each topic file registers its own from init.
var suite = &goldentest.Suite{
	Lang:      New(),
	ModelsDir: "../../../testdata/models",
	GoldenDir: "../../../testdata/golden/go",
	// the property file of every case (the fileNotes header)
	PropFiles: []string{"../../../testdata/props/go.properties"},
	Unreachable: map[string]string{
		"feature/basicget.container": "a basic getter exists for a single proxy, and a proxy is never a container",
	},
}

// fixture is one golden case: a model generated with entry points and
// properties.
type fixture = goldentest.Fixture

// register adds cases to the suite.
func register(cases ...fixture) { suite.Register(cases...) }

// The reasons of the V1Differences: v2 differs from v1 on purpose, v1 being
// wrong.
const (
	v1FromContainer       = "v1 types the container parameter of Create<X>FromContainer with the bare class name: wrong outside the interface package or with an interfaceName rule"
	v1DerivedOperationID  = "v1 leaves the operation ids of the class unqualified in its EDerivedOperationID test: wrong outside the interface package"
	v1MockImport          = "v1 does not import ecore for the NewMockRun calls of a class with operations only"
	v1MockAlias           = "v1 imports testify's mock and a model package named mock under the same name"
	v1MixinOrder          = "v1 lists the features of a mixin extending a class in the order of an OCL set difference, which is not the declaration order; v2 keeps the declaration order"
	v1GoIntTestValue      = "v1 has no type table row for int32 and int8: its test values int32() and int8() do not compile"
	v1AnnotationContainer = "v1 creates the annotation of a parameter in p.Get<Parameter>() (no such getter, or another element) and an annotation nested in an annotation in nil, and drops the annotations of the contents of an annotation"
	v1MapContainerMock    = "v1 mocks the map opposite of a Create<X>FromContainer test with NewMockEList, which is no EMap: it does not compile"
	v1OppositeAccessor    = "v1 adds the element created by Create<X>FromContainer through Get/Set<Opposite>, without the getterName and setterName rules, and tests the factory through new<Package>FactoryImpl, without the factoryName and implementationName rules: neither compiles with those rules"
	v1LazyGetter          = "v1 leaves a lambda list (a C++ mode) nil and rebuilds an action list on every read, losing its elements; v2 builds a lambda list on first read and an action list once"
	v1LazyTests           = "v1 expects the getter of a lazy proxy to panic, expects the getter of an action attribute to return the value of its field, and reads back lazy features through accessors that rerun their initializer (a panic, or the default value)"
	v1EnumLiteralTests    = "v1 tests the conversions of an enum with the literal names, not the literal strings, and an empty string as the invalid literal even when it is one"
	v1BasicUnsetDecl      = "v1 declares the basic unsetter with the value parameter its implementation does not have: it does not compile"
	v1VolatileTests       = "v1 expects the basic setter of a volatile feature to notify, the getterBody of a volatile feature to panic, and EBasicInverseRemove of a volatile feature to return"
)

// TestFixtures compares the generation of every case with its expected files.
func TestFixtures(t *testing.T) { suite.TestFixtures(t) }

// TestTemplateCoverage checks that the fixtures, together, render every
// template define of the Go language but the unreachable ones.
func TestTemplateCoverage(t *testing.T) { suite.TestTemplateCoverage(t) }

// TestValidate checks the Go language: names, entry points, properties,
// templates and their conventions.
func TestValidate(t *testing.T) {
	if err := engine.Validate(New()); err != nil {
		t.Fatal(err)
	}
}
