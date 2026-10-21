package golang

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/masagroup/soft.gen/v2/engine/goldentest"
)

// The fixtures are small purpose-built models (testdata/models), each about
// one topic, generated with entry points and properties into the committed
// expected files of testdata/golden/go/<case>. The expected files are the
// output of the v1 generator (soft.generator.go 1.13.1) once formatted, except
// the files of V1Differences, where v2 differs from v1 on purpose:
//
//	go test ./lang/golang -run TestFixtures            compare v2 with the expected files
//	go test ./lang/golang -run TestFixtures -update    write v2's output as the expected files
//	SOFT_GEN_V1_IMAGE=... go test ./lang/golang -run TestV1  compare v1 with the expected files
//	SOFT_GEN_V1_IMAGE=... go test ./lang/golang -run TestV1 -v1update  write v1's output
//
// TestTemplateCoverage checks that the fixtures render every template.

// fixturesDir is the directory of the fixture models.
const fixturesDir = "../../testdata/models"

// goldenDir is the directory of the expected files, one sub-directory per case.
const goldenDir = "../../testdata/golden/go"

// fixtureProps is the property file of every case (the fileNotes header).
const fixtureProps = "../../testdata/props/go.properties"

// fixture is one golden case: a model generated with entry points and
// properties.
type fixture struct {
	// Name is the case, and the directory of its expected files.
	Name string
	// Model is the file name of the model in testdata/models.
	Model string
	// EntryPoints are the -t options; empty means the default ones.
	EntryPoints []string
	// Props are the -p options.
	Props map[string]string
	// V1Differences lists the expected files where v2 differs from v1 on
	// purpose (slash-separated path -> reason): the v1 cross-check leaves them
	// out and -v1update keeps them.
	V1Differences map[string]string
	// V1Skip, when set, is why the v1 cross-check leaves the case out: v1
	// cannot generate it (its output does not parse or does not compile, or
	// it fails); the expected files are v2's output, reviewed.
	V1Skip string
}

// The reasons of the V1Differences: v2 differs from v1 on purpose, v1 being
// wrong.
const (
	v1FromContainer       = "v1 types the container parameter of Create<X>FromContainer with the bare class name: wrong outside the interface package or with an interfaceName rule"
	v1DerivedOperationID  = "v1 leaves the operation ids of the class unqualified in its EDerivedOperationID test: wrong outside the interface package"
	v1MockImport          = "v1 does not import ecore for the NewMockRun calls of a class with operations only"
	v1MockAlias           = "v1 imports testify's mock and a model package named mock under the same name"
	v1MixinOrder          = "v1 lists the features of a mixin extending a class in the order of an OCL set difference, which is not the declaration order; v2 keeps the declaration order"
	v1GoIntTestValue      = "v1 has no type table row for int32 and int8: its test values int32() and int8() do not compile"
	v1AnnotationContainer = "v1 creates the annotation of a parameter in p.Get<Parameter>() (no such getter, or another element) and a nested annotation in nil (lost)"
)

// fixtures holds the registered cases.
var fixtures []fixture

// register adds cases; each topic file registers its own from init.
func register(cases ...fixture) {
	fixtures = append(fixtures, cases...)
}

// sortedFixtures returns the cases sorted by name.
func sortedFixtures() []fixture {
	sorted := append([]fixture(nil), fixtures...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	return sorted
}

// modelPath is the path of the model of the case.
func (f fixture) modelPath() string { return filepath.Join(fixturesDir, f.Model) }

// expectedDir is the directory of the expected files of the case.
func (f fixture) expectedDir() string { return filepath.Join(goldenDir, f.Name) }

// TestFixtures compares the generation of every case with its expected files.
func TestFixtures(t *testing.T) {
	if len(fixtures) == 0 {
		t.Skip("no fixture registered")
	}
	names := map[string]bool{}
	for _, f := range sortedFixtures() {
		if names[f.Name] {
			t.Fatalf("fixture %q registered twice", f.Name)
		}
		names[f.Name] = true
		goldentest.Run(t, goldentest.Case{
			Name: f.Name, Lang: New(), Model: f.modelPath(),
			Props: f.Props, PropFiles: []string{fixtureProps}, EntryPoints: f.EntryPoints,
			Expected: f.expectedDir(),
		})
	}
}
