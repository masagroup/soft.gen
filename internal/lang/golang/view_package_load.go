package golang

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/model"
)

// xmiPackagePath is the package of the XMI codec of soft.go ("<path>/<package>"), whose EcoreCodec
// reads the serialized package.
const xmiPackagePath = "github.com/masagroup/soft.go/codec/xmi/xmi"

// packageLoadView is the loading of the serialized package of a package with
// loadInitialization (template package/load, EMF Initialize by Loading): the
// package loads its .ecore, embedded beside it, takes its classifiers and its
// annotations, then fixes them as its creation in code would have made them.
type packageLoadView struct {
	// FileName is the serialized package, in the directory of the package implementation.
	FileName string
	// Var is the variable embedding it, of type EmbedFS.
	Var, EmbedFS string
	// NewResourceSet, NewURI, ResolveAll and EcoreCodec are the qualified
	// functions and codec of the loading.
	NewResourceSet, NewURI, ResolveAll, EcoreCodec string
	// EPackage, EClassifier, EAnnotation, EDiagnostic and EDataTypeInternal are
	// the qualified ecore types of the loading.
	EPackage, EClassifier, EAnnotation, EDiagnostic, EDataTypeInternal string
	// Packages lists the instances of the packages the serialized package refers
	// to, registered in the package registry of the loading.
	Packages []string
	// Classifiers lists the fixes of the loaded classifiers, in model order.
	Classifiers []classifierFixView
}

// classifierFixView gives a loaded classifier its member, its id and the Go
// types its creation in code gives it (EMF fixEClassifiers).
type classifierFixView struct {
	// Member is the member receiving the classifier, Type its Ecore interface.
	Member, Type string
	// Index is the index of the classifier in the classifiers of the package.
	Index int
	// ID is the id constant of the classifier.
	ID string
	// InstanceTypeName is the Go type name of the classifier.
	InstanceTypeName string
	// Default is the default value literal of a data type, "" for none.
	Default string
	// Literals lists the literals of an enum, in model order.
	Literals []enumLiteralInitView
}

// isLoadInitialization reports whether the package is loaded from its
// serialized .ecore (property loadInitialization); never the ecore package.
func (g *gen) isLoadInitialization(p ecore.EPackage) bool {
	return !g.m.IsEcore(p) && g.PackagePropertyIs("loadInitialization", "true")
}

// packageLoad is the view of the loading of the serialized package.
func (g *gen) packageLoad(p ecore.EPackage, scope *fileScope) *packageLoadView {
	e := g.m.Ecore()
	ecoreIdentifier := func(name string) string { return scope.qualify(g.qualifiedInterfaceIdentifier(e, name)) }
	view := &packageLoadView{
		FileName:          g.serializedPackageFileName(p),
		Var:               "serializedPackage",
		EmbedFS:           scope.qualify("embed/embed.FS"),
		NewResourceSet:    ecoreIdentifier("NewEResourceSetImpl"),
		NewURI:            ecoreIdentifier("NewURI"),
		ResolveAll:        ecoreIdentifier("ResolveAllInResource"),
		EcoreCodec:        scope.qualify(qualifiedIdentifier(xmiPackagePath, "EcoreCodec")),
		EPackage:          ecoreIdentifier("EPackage"),
		EClassifier:       ecoreIdentifier("EClassifier"),
		EAnnotation:       ecoreIdentifier("EAnnotation"),
		EDiagnostic:       ecoreIdentifier("EDiagnostic"),
		EDataTypeInternal: ecoreIdentifier("EDataTypeInternal"),
	}
	for _, other := range referredPackages(p) {
		view.Packages = append(view.Packages, scope.qualify(g.packageInstance(other)))
	}
	for i, el := range p.GetEClassifiers().ToArray() {
		classifier := el.(ecore.EClassifier)
		fix := classifierFixView{
			Member: g.m.MemberName(classifier),
			Type:   scope.qualify(g.metaType(classifier)),
			Index:  i,
			ID:     scope.qualify(g.qualifiedInterfaceIdentifier(p, g.m.ClassifierIDName(classifier))),
		}
		switch classifier := classifier.(type) {
		case ecore.EClass:
			fix.InstanceTypeName = g.instanceTypeName(classifier)
		case ecore.EDataType:
			// the Go types, default and literals of dataTypeInitialization
			init := g.dataTypeInitialization(classifier, scope)
			fix.InstanceTypeName = init.InstanceTypeName
			fix.Default = init.Default
			fix.Literals = init.Literals
		}
		view.Classifiers = append(view.Classifiers, fix)
	}
	return view
}

// referredPackages returns the packages the serialized package refers to,
// sorted by namespace URI: the packages of the classes of its objects, such as
// the objects of the contents of its annotations, and the packages of the
// elements its references reach; neither itself nor the ecore package.
func referredPackages(p ecore.EPackage) []ecore.EPackage {
	found := map[string]ecore.EPackage{}
	add := func(other ecore.EPackage) {
		if other != nil && other != p && other.GetNsURI() != model.EcoreNsURI {
			found[other.GetNsURI()] = other
		}
	}
	for it := p.EAllContents(); it.HasNext(); {
		object := it.Next().(ecore.EObject)
		add(object.EClass().GetEPackage())
		for _, el := range object.EClass().GetEAllReferences().ToArray() {
			reference := el.(ecore.EReference)
			if reference.IsContainment() || reference.IsContainer() || reference.IsDerived() ||
				reference.IsTransient() || !object.EIsSet(reference) {
				continue
			}
			switch value := object.EGet(reference).(type) {
			case ecore.EObject:
				add(rootPackage(value))
			case ecore.EList:
				for _, target := range value.ToArray() {
					if target, ok := target.(ecore.EObject); ok {
						add(rootPackage(target))
					}
				}
			}
		}
	}
	uris := make([]string, 0, len(found))
	for uri := range found {
		uris = append(uris, uri)
	}
	sort.Strings(uris)
	packages := make([]ecore.EPackage, len(uris))
	for i, uri := range uris {
		packages[i] = found[uri]
	}
	return packages
}

// rootPackage returns the outermost package containing an object, nil when none.
func rootPackage(object ecore.EObject) ecore.EPackage {
	var root ecore.EPackage
	for o := object; o != nil; o = o.EContainer() {
		if p, ok := o.(ecore.EPackage); ok {
			root = p
		}
	}
	return root
}

// serializedPackageFileName is the file of the serialized package: the name of
// the package with the .ecore extension (EMF getSerializedPackageFilename).
func (g *gen) serializedPackageFileName(p ecore.EPackage) string {
	return p.GetName() + ".ecore"
}

// serializedPackageFile is the serialized package of a package with
// loadInitialization, beside its implementation: <package instance dir>/<name>.ecore.
func (g *gen) serializedPackageFile(p ecore.EPackage) engine.File {
	return engine.File{
		Path: filePath(g.packageInstanceOutputDirectory(p), g.serializedPackageFileName(p)),
		Body: serializedPackage{g: g, p: p},
	}
}

// serializedPackage renders the serialized package of a package (see serialize).
type serializedPackage struct {
	g *gen
	p ecore.EPackage
}

// Render serializes the package.
func (s serializedPackage) Render(*engine.Context) ([]byte, error) {
	return s.g.serialize(s.p)
}

// serialize returns the .ecore document the generated package loads, as EMF
// serializes a package for Initialize by Loading: a copy of the package without
// the annotations of its suppressed sources, saved as an .ecore resource, whose
// references to the other packages, the ecore package included, are written
// with their namespace URI so that the package registry of the loading
// resolves them.
func (g *gen) serialize(p ecore.EPackage) ([]byte, error) {
	if !p.GetESubPackages().Empty() {
		return nil, fmt.Errorf("loadInitialization: package %s has sub-packages, which are not loaded yet", p.GetName())
	}
	// the other resources are named after the namespace URI of their package
	// while the copy is saved
	if own := p.EResource(); own != nil && own.GetResourceSet() != nil {
		for _, el := range own.GetResourceSet().GetResources().ToArray() {
			resource := el.(ecore.EResource)
			if resource == own || resource.GetContents().Empty() {
				continue
			}
			if other, ok := resource.GetContents().Get(0).(ecore.EPackage); ok {
				uri := resource.GetURI()
				resource.SetURI(ecore.NewURI(other.GetNsURI()))
				defer resource.SetURI(uri)
			}
		}
	}
	copied := ecore.Copy(p).(ecore.EPackage)
	removeAnnotations(copied, g.m.SuppressedAnnotations(p))
	// the ids of the operations are computed by their class, as EMF does
	for it := copied.EAllContents(); it.HasNext(); {
		if operation, ok := it.Next().(ecore.EOperation); ok {
			operation.SetOperationID(-1)
		}
	}
	resource := ecore.NewEResourceImpl()
	resource.SetURI(ecore.NewURI(g.serializedPackageFileName(p)))
	resource.GetContents().Add(copied)
	var buffer bytes.Buffer
	resource.SaveWithWriter(&buffer, nil)
	if errors := resource.GetErrors(); !errors.Empty() {
		var messages []string
		for _, el := range errors.ToArray() {
			messages = append(messages, el.(ecore.EDiagnostic).GetMessage())
		}
		return nil, fmt.Errorf("loadInitialization: saving package %s: %s", p.GetName(), strings.Join(messages, "; "))
	}
	return buffer.Bytes(), nil
}

// removeAnnotations removes the annotations of the sources from the package and
// its elements, as their creation in code leaves them out; the annotations
// nested in an annotation stay with it.
func removeAnnotations(p ecore.EPackage, sources []string) {
	if len(sources) == 0 {
		return
	}
	suppressed := map[string]bool{}
	for _, source := range sources {
		suppressed[source] = true
	}
	elements := []ecore.EModelElement{p}
	for it := p.EAllContents(); it.HasNext(); {
		if el, ok := it.Next().(ecore.EModelElement); ok && !isInAnnotation(el) {
			elements = append(elements, el)
		}
	}
	for _, el := range elements {
		// a copy of the list: removing an annotation changes the list
		for _, a := range slices.Clone(el.GetEAnnotations().ToArray()) {
			if annotation := a.(ecore.EAnnotation); suppressed[annotation.GetSource()] {
				el.GetEAnnotations().Remove(annotation)
			}
		}
	}
}

// isInAnnotation reports whether an element is an annotation or is contained in one.
func isInAnnotation(el ecore.EObject) bool {
	for o := el; o != nil; o = o.EContainer() {
		if _, ok := o.(ecore.EAnnotation); ok {
			return true
		}
	}
	return false
}
