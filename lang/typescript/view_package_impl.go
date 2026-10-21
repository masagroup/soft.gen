package typescript

import (
	"strconv"
	"strings"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/engine"
)

// packageImplView is the data of the members of the package implementation.
type packageImplView struct {
	names
	// Name is the class; Interface the package interface.
	Name, Interface string
	// Factory is the type of the factory given to initialize; Instance the
	// factory created by getInstance; EcoreFactory the type of the ecore
	// factory and GetEcoreFactory its getter.
	Factory, Instance, EcoreFactory, GetEcoreFactory string
	// Constants is the class of the constants.
	Constants string
	// Lines are the statements of a method, built in Go ("" is a blank line).
	Lines []string
	// Getter is a getter method: its name, type and returned expression.
	Getter, Type, Value string
	// Source and Annotations are an annotation initializer.
	Source, Identifier string
	Annotations        []packageAnnotation
}

// packageAnnotation is an annotation created by an annotation initializer.
type packageAnnotation struct {
	Container string
	Details   [][2]string
}

// packageImplementationFile is the implementation of a package (v1
// generatePackageImplementation): the meta objects of its classes and data
// types, built by createPackageContents and initialized by
// initializePackageContents, their getters, and the annotations.
func (g *gen) packageImplementationFile(p ecore.EPackage) engine.File {
	s := g.newScope().declare(g.interfaceIdentifier(p, g.packageImplementationName(p)))
	n := names{s}
	name := g.packageImplementationName(p)
	shared := g.isSharedPackage(p)
	if !shared {
		s.declare(g.implementationIdentifier(p, name))
	}
	view := classView{Declaration: "export class " + name + " extends " + s.qualify(g.implementationIdentifier(g.m.Ecore(), "EPackageExt")) +
		" implements " + s.qualify(g.interfaceIdentifier(p, g.packageInterfaceName(p)))}
	data := func() packageImplView {
		return packageImplView{names: n, Name: name, Interface: g.packageInterfaceName(p),
			EcoreFactory: s.qualify(g.ecoreIdentifier("EcoreFactory")), Constants: s.qualify(g.interfaceIdentifier(p, g.constantsName(p)))}
	}
	classes := sortedByName(g.m.Package(p).AllClasses)
	dataTypes := sortedByName(g.allDataTypes(p))

	// the fields: the instance, the classes, the data types
	fields := []string{"private static _instance: " + name + " = null"}
	for _, c := range classes {
		fields = append(fields, "private "+g.elementInstance(c)+": "+s.metaType(c))
	}
	if len(dataTypes) > 0 {
		fields = append(fields, "")
		for _, dt := range dataTypes {
			fields = append(fields, "private "+g.elementInstance(dt)+": "+s.metaType(dt))
		}
	}
	view.add("impl/members.fields", fields)

	instance := data()
	if shared {
		instance.Instance = "new " + s.qualify(g.interfaceIdentifier(p, g.dataTypeFactoryName(p))) + "()"
		instance.Factory = s.qualify(g.ecoreIdentifier("EFactory"))
	} else {
		instance.Instance = s.qualify(g.interfaceIdentifier(p, "get"+g.factoryName(p))) + "()"
		instance.Factory = s.qualify(g.interfaceIdentifier(p, g.factoryName(p)))
	}
	instance.GetEcoreFactory = s.qualify(g.ecoreIdentifier("getEcoreFactory")) + "()"
	view.add("package/impl.instance", instance)
	view.add("package/impl.initialize", instance)

	// the getters, in groups without blank lines: the class, its attributes,
	// references and operations, the data types
	var group []packageImplView
	getter := func(el ecore.ENamedElement, value string) {
		v := data()
		v.Getter, v.Type, v.Value = g.ReflectiveGetterName(el), s.metaType(el), value
		group = append(group, v)
	}
	endGroup := func() {
		if len(group) > 0 {
			view.add("package/impl.getters", group)
			group = nil
		}
	}
	for _, c := range classes {
		getter(c, "this."+g.elementInstance(c))
		endGroup()
		for _, a := range sortedByName(namedElements(c.GetEAttributes())) {
			getter(a, "this."+g.elementInstance(c)+".getEStructuralFeatures().get("+strconv.Itoa(c.GetEStructuralFeatures().IndexOf(a))+") as "+s.metaType(a))
		}
		endGroup()
		for _, r := range sortedByName(namedElements(c.GetEReferences())) {
			getter(r, "this."+g.elementInstance(c)+".getEStructuralFeatures().get("+strconv.Itoa(c.GetEStructuralFeatures().IndexOf(r))+") as "+s.metaType(r))
		}
		endGroup()
		for _, op := range sortedByName(namedElements(c.GetEOperations())) {
			getter(op, "this."+g.elementInstance(c)+".getEOperations().get("+strconv.Itoa(c.GetEOperations().IndexOf(op))+") as "+s.metaType(op))
		}
		endGroup()
	}
	for _, dt := range dataTypes {
		getter(dt, "this."+g.elementInstance(dt))
	}
	endGroup()

	create := data()
	create.Lines = g.packageCreation(p, classes, dataTypes, s)
	view.add("package/impl.create", create)
	initialization := data()
	initialization.Lines = g.packageInitialization(p, classes, dataTypes, s)
	view.add("package/impl.init", initialization)
	if !g.m.IsEcore(p) {
		for _, source := range g.annotationInitializers(p, s) {
			view.add("package/impl.annotations", source)
		}
	}
	return g.file(g.implementationFileName(p, g.packageName(p)), s, "class.ts", view)
}

// elementInstance is the field of the meta object of a classifier: _<member>Class
// for a class, _<member>Type for a data type (v1 getElementInstance).
func (g *gen) elementInstance(el ecore.ENamedElement) string {
	suffix := ""
	switch el.(type) {
	case ecore.EClass:
		suffix = "Class"
	case ecore.EDataType:
		suffix = "Type"
	}
	return "_" + g.m.MemberName(el) + suffix
}

// metaType is the type of the meta object of an element: ecore.EClass,
// ecore.EAttribute...
func (s *scope) metaType(el ecore.ENamedElement) string {
	return s.qualify(s.g.ecoreIdentifier(el.EClass().GetName()))
}

// dataTypeFactoryName is the factory of the data types of a shared package:
// its extension with factory=extension, else the generated one.
func (g *gen) dataTypeFactoryName(p ecore.EPackage) string {
	if g.KeyIs(p, "factory", "extension") {
		return g.factoryExtensionName(p)
	}
	return g.factoryImplementationName(p)
}

// packageCreation is the body of createPackageContents: per class its meta
// object, its features and operations, then the data types.
func (g *gen) packageCreation(p ecore.EPackage, classes []ecore.EClass, dataTypes []ecore.EDataType, s *scope) []string {
	var lines []string
	for _, c := range classes {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "this."+g.elementInstance(c)+" = ecoreFactory.createEClassFromContainerAndClassID(this, "+s.classifierID(c)+")")
		for it := c.GetEStructuralFeatures().Iterator(); it.HasNext(); {
			f := it.Next().(ecore.EStructuralFeature)
			kind := "EReference"
			if _, ok := f.(ecore.EAttribute); ok {
				kind = "EAttribute"
			}
			lines = append(lines, "ecoreFactory.create"+kind+"FromContainerAndClassID(this."+g.elementInstance(c)+", "+s.featureID(c, f)+")")
		}
		for it := c.GetEOperations().Iterator(); it.HasNext(); {
			op := it.Next().(ecore.EOperation)
			lines = append(lines, "ecoreFactory.createEOperationFromContainerAndClassID(this."+g.elementInstance(c)+", "+s.operationID(c, op)+")")
		}
	}
	if len(dataTypes) > 0 && len(lines) > 0 {
		lines = append(lines, "")
	}
	for _, dt := range dataTypes {
		kind := "EDataType"
		if _, ok := dt.(ecore.EEnum); ok {
			kind = "EEnum"
		}
		lines = append(lines, "this."+g.elementInstance(dt)+" = ecoreFactory.create"+kind+"FromContainerAndClassID(this, "+s.classifierID(dt)+")")
	}
	return lines
}

// packageInitialization is the body of initializePackageContents: the super
// types, per class its meta object, features and operations, the data types
// and the enums, then the calls of the annotation initializers.
func (g *gen) packageInitialization(p ecore.EPackage, classes []ecore.EClass, dataTypes []ecore.EDataType, s *scope) []string {
	var lines []string
	blank := func() {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
	}
	for _, c := range classes {
		for it := c.GetESuperTypes().Iterator(); it.HasNext(); {
			lines = append(lines, "this."+g.elementInstance(c)+".getESuperTypes().add("+s.classGetter(it.Next().(ecore.EClass), p)+")")
		}
	}
	for _, c := range classes {
		blank()
		lines = append(lines, "this.initEClass(this."+g.elementInstance(c)+", "+quote(c.GetName())+", "+quote(g.instanceTypeName(c))+", "+
			strconv.FormatBool(c.IsAbstract())+", "+strconv.FormatBool(c.IsInterface())+")")
		for it := c.GetEAttributes().Iterator(); it.HasNext(); {
			a := it.Next().(ecore.EAttribute)
			lines = append(lines, "this.initEAttribute("+strings.Join(append([]string{
				g.elementGetter(a), s.typeGetter(a.GetEType(), c), quote(a.GetName()), quote(a.GetDefaultValueLiteral()),
				strconv.Itoa(a.GetLowerBound()), strconv.Itoa(a.GetUpperBound())},
				bools(a.IsTransient(), a.IsVolatile(), a.IsChangeable(), a.IsUnsettable(), a.IsUnique(), a.IsDerived(), a.IsOrdered(), a.IsID())...), ", ")+")")
		}
		for it := c.GetEReferences().Iterator(); it.HasNext(); {
			r := it.Next().(ecore.EReference)
			opposite := "null"
			if r.GetEOpposite() != nil {
				opposite = g.elementGetter(r.GetEOpposite())
			}
			lines = append(lines, "this.initEReference("+strings.Join(append([]string{
				g.elementGetter(r), s.typeGetter(r.GetEType(), c), opposite, quote(r.GetName()), quote(r.GetDefaultValueLiteral()),
				strconv.Itoa(r.GetLowerBound()), strconv.Itoa(r.GetUpperBound())},
				bools(r.IsTransient(), r.IsVolatile(), r.IsChangeable(), r.IsContainment(), r.IsResolveProxies(), r.IsUnsettable(), r.IsUnique(), r.IsDerived(), r.IsOrdered())...), ", ")+")")
		}
		for it := c.GetEOperations().Iterator(); it.HasNext(); {
			lines = append(lines, g.operationInitialization(it.Next().(ecore.EOperation), c, s)...)
		}
	}
	if len(dataTypes) > 0 {
		blank()
	}
	for _, dt := range dataTypes {
		if e, ok := dt.(ecore.EEnum); ok {
			instance := "number"
			if e.GetInstanceTypeName() != "" {
				instance = g.instanceTypeName(e)
			}
			lines = append(lines, "this.initEEnum("+g.elementGetter(e)+", "+quote(e.GetName())+", "+quote(instance)+")")
			for it := e.GetELiterals().Iterator(); it.HasNext(); {
				l := it.Next().(ecore.EEnumLiteral)
				lines = append(lines, "this.addEEnumLiteral("+g.elementGetter(e)+", "+quote(l.GetName())+", "+quote(l.GetLiteral())+", "+strconv.Itoa(l.GetValue())+")")
			}
			lines = append(lines, "")
			continue
		}
		defaultValue := strings.NewReplacer(`'`, "", `"`, "").Replace(s.classifierDefaultValue(dt))
		if defaultValue == "null" {
			defaultValue = ""
		}
		lines = append(lines, "this.initEDataType("+g.elementGetter(dt)+", "+quote(dt.GetName())+", "+quote(g.instanceTypeName(dt))+", "+
			quote(defaultValue)+", "+strconv.FormatBool(dt.IsSerializable())+")")
	}
	if !g.m.IsEcore(p) {
		if sources := g.m.Package(p).AnnotationSources; len(sources) > 0 {
			blank()
			for _, source := range sources {
				lines = append(lines, "this.initialize"+source.Identifier+"Annotations(ecoreFactory)")
			}
		}
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// operationInitialization initializes an operation and its parameters.
func (g *gen) operationInitialization(op ecore.EOperation, c ecore.EClass, s *scope) []string {
	typeGetter := s.operationTypeGetter(op.GetEType(), op)
	tail := quote(op.GetName()) + ", " + strconv.Itoa(op.GetLowerBound()) + ", " + strconv.Itoa(op.GetUpperBound()) + ", " +
		strconv.FormatBool(op.IsUnique()) + ", " + strconv.FormatBool(op.IsOrdered()) + ")"
	if op.GetEParameters().Empty() {
		return []string{"this.initEOperation(" + g.elementGetter(op) + ", " + typeGetter + ", " + tail}
	}
	lines := []string{"{",
		indentUnit + "const operation = " + g.elementGetter(op),
		indentUnit + "this.initEOperation(operation, " + typeGetter + ", " + tail}
	for it := op.GetEParameters().Iterator(); it.HasNext(); {
		parameter := it.Next().(ecore.EParameter)
		lines = append(lines, indentUnit+"this.addEParameter(operation, "+s.operationTypeGetter(parameter.GetEType(), op)+", "+quote(parameter.GetName())+
			", 0, 1, "+strconv.FormatBool(parameter.IsUnique())+", "+strconv.FormatBool(parameter.IsOrdered())+")")
	}
	return append(lines, "}")
}

// elementGetter is the getter of an element in the package implementation:
// this.<getter>() (v1 getPackageGetter(ENamedElement)).
func (g *gen) elementGetter(el ecore.ENamedElement) string {
	return "this." + g.ReflectiveGetterName(el) + "()"
}

// typeGetter is the getter of the type of a feature: in this package
// (this.getX()), else through the package of the type (ecore.getEcorePackage().getEString())
// (v1 getPackageGetter(EClassifier, EStructuralFeature)).
func (s *scope) typeGetter(t ecore.EClassifier, owner ecore.EClass) string {
	if t == nil {
		return "null"
	}
	return s.packageInstance(t, owner) + "." + s.g.ReflectiveGetterName(t) + "()"
}

// operationTypeGetter is the getter of the type of an operation or of one of
// its parameters: null when the operation has no type, as v1 (whose
// parameters of a void operation are typed null).
func (s *scope) operationTypeGetter(t ecore.EClassifier, op ecore.EOperation) string {
	if op.GetEType() == nil || t == nil {
		return "null"
	}
	return s.packageInstance(t, op.GetEContainingClass()) + "." + s.g.ReflectiveGetterName(t) + "()"
}

// packageInstance is the package of a classifier seen from a class: this in
// the same package (by name), else get<Package>() of its package.
func (s *scope) packageInstance(t ecore.EClassifier, owner ecore.EClass) string {
	if t.GetEPackage().GetName() == owner.GetEPackage().GetName() {
		return "this"
	}
	return s.qualify(s.g.packageGetterIdentifier(t.GetEPackage())) + "()"
}

// classGetter is the meta object of a super class: the field of this package,
// else the getter of its package (v1 getGetter).
func (s *scope) classGetter(c ecore.EClass, p ecore.EPackage) string {
	if c.GetEPackage() == p || c.GetEPackage().GetName() == p.GetName() {
		return "this." + s.g.elementInstance(c)
	}
	return s.qualify(s.g.packageGetterIdentifier(c.GetEPackage())) + "()." + s.g.ReflectiveGetterName(c) + "()"
}

// annotationInitializers lists, per annotation source of the package, the
// annotations of the source in model order (v1 initialize<Source>Annotations).
func (g *gen) annotationInitializers(p ecore.EPackage, s *scope) []packageImplView {
	info := g.m.Package(p)
	var views []packageImplView
	for _, source := range info.AnnotationSources {
		view := packageImplView{names: names{s}, EcoreFactory: s.qualify(g.ecoreIdentifier("EcoreFactory")),
			Source: quote(source.URI), Identifier: source.Identifier}
		for _, a := range info.Annotations {
			if a.GetSource() != source.URI {
				continue
			}
			annotation := packageAnnotation{Container: g.annotatedElementGetter(a.GetEModelElement(), s)}
			for _, entry := range a.GetDetails().ToArray() {
				entry := entry.(ecore.EMapEntry)
				key, _ := entry.GetKey().(string)
				value, _ := entry.GetValue().(string)
				annotation.Details = append(annotation.Details, [2]string{quote(key), quote(value)})
			}
			view.Annotations = append(view.Annotations, annotation)
		}
		views = append(views, view)
	}
	return views
}

// annotatedElementGetter is the expression of the element an annotation is
// created in (v1 getPackageGetter(EModelElement)): this for the package, the
// literal found by name in its enum, the getter of a named element, null
// otherwise. As in the Go language, two cases differ from v1, which got them
// wrong: a parameter is reached through its operation, a nested annotation
// through its parent annotation.
func (g *gen) annotatedElementGetter(el ecore.EModelElement, s *scope) string {
	switch el := el.(type) {
	case ecore.EPackage:
		return "this"
	case ecore.EEnumLiteral:
		return g.elementGetter(el.GetEEnum()) + `.getEEnumLiteralByName("` + el.GetName() + `")`
	case ecore.EParameter:
		op := el.GetEOperation()
		return g.elementGetter(op) + ".getEParameters().get(" + strconv.Itoa(op.GetEParameters().IndexOf(el)) + ") as " +
			s.qualify(g.ecoreIdentifier("EParameter"))
	case ecore.EAnnotation:
		parent := g.annotatedElementGetter(el.GetEModelElement(), s)
		if parent == "null" {
			return parent
		}
		return parent + ".getEAnnotation(" + quote(el.GetSource()) + ")"
	case ecore.ENamedElement:
		return g.elementGetter(el)
	}
	return "null"
}

// bools writes booleans.
func bools(values ...bool) []string {
	flags := make([]string, len(values))
	for i, v := range values {
		flags[i] = strconv.FormatBool(v)
	}
	return flags
}
