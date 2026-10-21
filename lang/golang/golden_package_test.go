package golang

import (
	"path/filepath"
	"testing"

	"github.com/masagroup/soft.gen/v2/engine/goldentest"
)

// The golden tests of the package implementation and the factories: the
// files generateModel writes for the library, empty and ecore models, and the
// files of the shared tournament package written by generateInterfaces (the
// package implementation, the factory interface and the data type factory)
// and by generateImplementations (the factory of the typed and the reflective
// implementations).

// packageFiles lists the package implementation and factory files of a package.
func packageFiles(dir, pkg string) []string {
	return []string{
		dir + "/" + pkg + "package_impl.go",
		dir + "/" + pkg + "factory.go",
		dir + "/" + pkg + "factory_impl.go",
	}
}

func TestGoldenPackageLibrary(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "library.ecore"), PropFiles: []string{goProps},
		Expected: filepath.Join(goldenRoot(t), "test"),
		Files:    packageFiles("library", "library"),
	})
}

func TestGoldenPackageEmpty(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "empty.ecore"), PropFiles: []string{goProps},
		Expected: filepath.Join(goldenRoot(t), "test"),
		Files:    packageFiles("empty", "empty"),
	})
}

func TestGoldenPackageEcore(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "ecore.ecore"), PropFiles: []string{goProps},
		Expected: goldenRoot(t),
		Files:    packageFiles("ecore", "ecore"),
	})
}

func TestGoldenPackageTournamentInterfaces(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateInterfaces"},
		Props:       map[string]string{"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentTyped},
		Expected:    goldenRoot(t),
		Files:       packageFiles("test/tournament", "tournament"),
	})
}

func TestGoldenPackageTournamentTyped(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateImplementations"},
		Props:       map[string]string{"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentTyped},
		Expected:    goldenRoot(t),
		Files:       []string{"test/tournament/typed/tournamentfactory_impl.go"},
	})
}

func TestGoldenPackageTournamentReflective(t *testing.T) {
	goldentest.Run(t, goldentest.Case{
		Lang: New(), Model: filepath.Join(modelsDir, "tournament.ecore"), PropFiles: []string{goProps},
		EntryPoints: []string{"generateImplementations"},
		Props: map[string]string{
			"packageInstance": "interfaces", "module": tournamentModule, "packages": tournamentReflective,
			"featureDelegation": "Reflective",
		},
		Expected: goldenRoot(t),
		Files:    []string{"test/tournament/reflective/tournamentfactory_impl.go"},
	})
}
