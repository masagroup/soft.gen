package golang

import (
	"path/filepath"
	"testing"

	"github.com/masagroup/soft.gen/v2/engine/goldentest"
)

// The golden tests of the class interfaces and implementations: every
// generated Go file of a package except the files of the other checkpoints
// (tests, mocks, package, factory) and the hand-written files (extensions).

// classFiles returns the patterns of the interfaces and implementations of a
// package and the exclusions of the other checkpoints.
func classFiles(dir, pkg string, handWritten ...string) (files, exclude []string) {
	files = []string{dir + "/*.go"}
	for _, p := range []string{"*_impl_test.go", "*_mock*.go", pkg + "package*.go", pkg + "factory*.go"} {
		exclude = append(exclude, dir+"/"+p)
	}
	for _, p := range handWritten {
		exclude = append(exclude, dir+"/"+p)
	}
	return files, exclude
}

func TestGoldenLibraryClasses(t *testing.T) {
	files, exclude := classFiles("library", "library", "writer_ext.go", "bookcategory.go")
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "library.ecore"), PropFiles: []string{goProps},
		Expected: filepath.Join(goldenRoot(t), "test"),
		Files:    files, Exclude: exclude,
	})
}

// The empty model has no class: it generates no interface nor implementation
// (TestGoldenEmpty of golden_test.go covers its package files).

func TestGoldenEcoreClasses(t *testing.T) {
	files, exclude := classFiles("ecore", "ecore")
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "ecore.ecore"), PropFiles: []string{goProps},
		Expected: goldenRoot(t),
		Files:    files, Exclude: exclude,
		OnlyGenerated: true,
	})
}

func TestGoldenTournamentInterfaceClasses(t *testing.T) {
	files, exclude := classFiles("test/tournament", "tournament", "matchkind.go")
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateInterfaces"},
		Props:       map[string]string{"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentTyped},
		Expected:    goldenRoot(t),
		Files:       files, Exclude: exclude,
	})
}

func TestGoldenTournamentTypedClasses(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateImplementations"},
		Props:       map[string]string{"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentTyped},
		Expected:    goldenRoot(t),
		Files:       []string{"test/tournament/typed/*_impl.go"},
		Exclude:     []string{"test/tournament/typed/tournamentfactory*.go"},
	})
}

func TestGoldenTournamentReflectiveClasses(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateImplementations"},
		Props: map[string]string{
			"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentReflective,
			"featureDelegation": "Reflective",
		},
		Expected: goldenRoot(t),
		Files:    []string{"test/tournament/reflective/*_impl.go"},
		Exclude:  []string{"test/tournament/reflective/tournamentfactory*.go"},
	})
}
