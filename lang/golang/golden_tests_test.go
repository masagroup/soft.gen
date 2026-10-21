package golang

import (
	"path/filepath"
	"testing"

	"github.com/masagroup/soft.gen/v2/engine/goldentest"
)

// The golden tests of the generated tests (generateTests): the tests of the
// implementations, the factory and the package.

// testFiles lists the generated test files of a package directory.
func testFiles(dir string) []string {
	return []string{dir + "/*_impl_test.go"}
}

func TestGoldenTestsLibrary(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "library.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateTests"},
		Expected:    filepath.Join(goldenRoot(t), "test"),
		Files:       testFiles("library"),
	})
}

func TestGoldenTestsEmpty(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "empty.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateTests"},
		Expected:    filepath.Join(goldenRoot(t), "test"),
		Files:       testFiles("empty"),
	})
}

func TestGoldenTestsEcore(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "ecore.ecore"), PropFiles: []string{goProps},
		EntryPoints:   []string{"generateTests"},
		Expected:      goldenRoot(t),
		Files:         testFiles("ecore"),
		OnlyGenerated: true,
	})
}

func TestGoldenTestsTournamentTyped(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateImplementations", "generateTests"},
		Props:       map[string]string{"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentTyped},
		Expected:    goldenRoot(t),
		Files:       []string{"test/tournament/typed/*_test.go", "test/tournament/tournamentpackage_impl_test.go"},
	})
}

func TestGoldenTestsTournamentReflective(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateImplementations", "generateTests"},
		Props: map[string]string{
			"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentReflective,
			"featureDelegation": "Reflective",
		},
		Expected: goldenRoot(t),
		Files:    []string{"test/tournament/reflective/*_test.go"},
	})
}
