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
func (g *gen) factoryInterfaceFile(p ecore.EPackage) engine.File {
	return g.pending(g.interfaceFileName(p, g.factoryName(p)))
}
func (g *gen) factoryImplementationFile(p ecore.EPackage) engine.File {
	return g.pending(g.implementationFileName(p, g.factoryName(p)))
}
func (g *gen) dataTypeFactoryFile(p ecore.EPackage) engine.File {
	return g.pending(g.implementationFileName(p, g.factoryName(p)))
}
func (g *gen) packageImplementationFile(p ecore.EPackage) engine.File {
	return g.pending(g.implementationFileName(p, g.packageName(p)))
}
func (g *gen) factoryTestFile(p ecore.EPackage) engine.File {
	return g.pending(g.factoryTestFileName(p))
}
func (g *gen) internalFile(p ecore.EPackage) engine.File { return g.pending(g.internalFileName(p)) }
func (g *gen) indexFile(p ecore.EPackage) engine.File    { return g.pending(g.indexFileName(p)) }
func (g *gen) interfacesInternalFile(p ecore.EPackage) engine.File {
	return g.pending(g.internalFileName(p))
}
func (g *gen) interfacesIndexFile(p ecore.EPackage) engine.File { return g.pending(g.indexFileName(p)) }
func (g *gen) implementationsInternalFile(p ecore.EPackage) engine.File {
	return g.pending(g.internalFileName(p))
}
func (g *gen) implementationsIndexFile(p ecore.EPackage) engine.File {
	return g.pending(g.indexFileName(p))
}
