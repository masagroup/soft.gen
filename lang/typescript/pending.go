package typescript

import (
	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
)

// pending is a file not ported yet: its path, an empty body.
func (g *gen) pending(name string) engine.File {
	return engine.File{Path: g.filePath(g.pkg, name), Body: engine.Bytes([]byte("// pending\n")), Header: true}
}

func (g *gen) implementationTestFile(c ecore.EClass) engine.File {
	return g.pending(g.implementationTestFileName(c, c.GetName()))
}
func (g *gen) factoryTestFile(p ecore.EPackage) engine.File {
	return g.pending(g.factoryTestFileName(p))
}
