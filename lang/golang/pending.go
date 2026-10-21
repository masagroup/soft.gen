package golang

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
)

// The file builders below are not ported yet: they generate nothing.

func (g *gen) implementationTestFile(ecore.EClass) []engine.File { return nil }
func (g *gen) factoryTestFile(ecore.EPackage) []engine.File      { return nil }
func (g *gen) packageTestFile(ecore.EPackage) []engine.File      { return nil }
