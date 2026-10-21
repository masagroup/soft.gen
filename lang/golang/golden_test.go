package golang

import (
	"path/filepath"
	"testing"

	"github.com/masagroup/soft.gen/v2/engine"
	"github.com/masagroup/soft.gen/v2/engine/goldentest"
)

// The golden files are the outputs of the v1 generator in the soft.go
// repository (D:\dev\projects\sword\v2\soft\go): ecore/ from ecore.ecore,
// test/{empty,library,tournament} from the test models, generated as its
// Makefile does. The root is the soft.go checkout beside this repository
// (../../../go from this directory), or SOFT_GEN_GOLDEN_ROOT.

const (
	modelsDir = "../../testdata/models"
	goProps   = "../../testdata/props/go.properties"
)

func goldenRoot(t *testing.T) string {
	return goldentest.Root(t, filepath.Join("..", "..", "..", "go"))
}

func TestValidate(t *testing.T) {
	if err := engine.Validate(New()); err != nil {
		t.Fatal(err)
	}
}

// ported lists, for the output directory of a package, the file patterns of
// the checkpoints already ported (enums and package interface); the golden
// tests compare these files only.
func ported(dir, pkg string) []string {
	patterns := []string{pkg + "package.go", "bookcategory.go", "matchkind.go"}
	for i, p := range patterns {
		patterns[i] = dir + "/" + p
	}
	return patterns
}

func TestGoldenLibrary(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "library.ecore"), PropFiles: []string{goProps},
		Expected: filepath.Join(goldenRoot(t), "test"),
		Files:    ported("library", "library"),
	})
}

func TestGoldenEmpty(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "empty.ecore"), PropFiles: []string{goProps},
		Expected: filepath.Join(goldenRoot(t), "test"),
		Files:    ported("empty", "empty"),
	})
}

func TestGoldenEcore(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "ecore.ecore"), PropFiles: []string{goProps},
		Expected: goldenRoot(t),
		Files:    ported("ecore", "ecore"),
	})
}

const (
	tournamentModule     = "github.com/masagroup/soft.go"
	tournamentPackage    = tournamentModule + "/test/tournament"
	tournamentTyped      = "http://mattsch.com/emf/examples/tournament=" + tournamentPackage + "/tournament#" + tournamentPackage + "/typed/typed#" + tournamentPackage + "/tournament"
	tournamentReflective = "http://mattsch.com/emf/examples/tournament=" + tournamentPackage + "/tournament#" + tournamentPackage + "/reflective/reflective#" + tournamentPackage + "/tournament"
)

func TestGoldenTournamentInterfaces(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateInterfaces", "generateMocks", "generateMocksTests"},
		Props:       map[string]string{"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentTyped},
		Expected:    goldenRoot(t),
		Files:       ported("test/tournament", "tournament"),
	})
}
