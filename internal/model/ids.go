package model

import (
	"strconv"
	"strings"

	"github.com/masagroup/soft.go/ecore"
)

// Expr is the value of a package constant: a literal (3), a base constant
// alone (EMODEL_ELEMENT__EANNOTATIONS) or an offset from a base constant
// (EMODEL_ELEMENT_FEATURE_COUNT + 1). Value is its numeric value in every
// form. Package is the package of the base class, which declares the base
// constant: another package than the one of the constant when a class extends
// a class of another model, and the constant must then be qualified.
type Expr struct {
	Value     int            // the numeric value (v1 getFeatureIDInteger...)
	Base      string         // the base constant, when HasBase
	Package   ecore.EPackage // the package declaring the base constant, when HasBase
	Offset    int            // the offset from the base constant, when HasOffset
	HasBase   bool           // the value is expressed from a base constant
	HasOffset bool           // the value is the base constant plus Offset (printed even when 0)
}

// String returns the constant value as v1 printed it: "3",
// "EMODEL_ELEMENT__EANNOTATIONS" or "EMODEL_ELEMENT_FEATURE_COUNT + 1" (v1
// getFeatureIDValue, getFeatureCountIDValue, getOperationIDValue,
// getOperationCountIDValue).
func (e Expr) String() string {
	switch {
	case !e.HasBase:
		return strconv.Itoa(e.Value)
	case !e.HasOffset:
		return e.Base
	default:
		return e.Base + " + " + strconv.Itoa(e.Offset)
	}
}

// Int returns the numeric value (v1 getFeatureIDInteger,
// getFeatureCountIDInteger, getOperationIDInteger, getOperationCountIDInteger).
func (e Expr) Int() int {
	return e.Value
}

// literalExpr returns the literal form.
func literalExpr(value int) Expr {
	return Expr{Value: value}
}

// baseExpr returns the base constant alone form, base being a constant of the package pkg.
func baseExpr(pkg ecore.EPackage, base string, value int) Expr {
	return Expr{Value: value, Base: base, Package: pkg, HasBase: true}
}

// offsetExpr returns the base constant plus offset form, base being a constant of the package pkg.
func offsetExpr(pkg ecore.EPackage, base string, baseValue int, offset int) Expr {
	return Expr{Value: baseValue + offset, Base: base, Package: pkg, Offset: offset, HasBase: true, HasOffset: true}
}

// ClassIDs is the identifiers and values of the package constants of a class.
type ClassIDs struct {
	ClassifierIDName    string
	ClassifierIDValue   int
	FeatureIDs          []ID // one per eAllStructuralFeatures, in order
	FeatureCountIDName  string
	FeatureCountIDValue Expr
	OperationIDs        []ID // one per eAllOperations, in order
	OperationCountName  string
	OperationCountValue Expr
}

// ID is the identifier and the value of the package constant of a feature or
// an operation of a class.
type ID struct {
	Element ecore.ENamedElement // the feature or the operation
	Name    string
	Value   Expr
}

// classIDs computes the package constants of c.
func (m *Model) classIDs(c ecore.EClass) ClassIDs {
	ids := ClassIDs{
		ClassifierIDName:    m.ClassifierIDName(c),
		ClassifierIDValue:   m.ClassifierIDValue(c),
		FeatureCountIDName:  m.FeatureCountIDName(c),
		FeatureCountIDValue: m.FeatureCountIDValue(c),
		OperationCountName:  m.OperationCountIDName(c),
		OperationCountValue: m.OperationCountIDValue(c),
	}
	for _, f := range elements[ecore.EStructuralFeature](c.GetEAllStructuralFeatures()) {
		ids.FeatureIDs = append(ids.FeatureIDs, ID{Element: f, Name: m.FeatureIDName(c, f), Value: m.FeatureIDValue(c, f)})
	}
	for _, op := range elements[ecore.EOperation](c.GetEAllOperations()) {
		ids.OperationIDs = append(ids.OperationIDs, ID{Element: op, Name: m.OperationIDName(c, op), Value: m.OperationIDValue(c, op)})
	}
	return ids
}

// ClassifierIDName returns the identifier of the classifier constant: the
// name in upper snake case (v1 getClassifierIDName).
func (m *Model) ClassifierIDName(c ecore.EClassifier) string {
	return strings.ToUpper(splitAndCombineWords(c.GetName(), "_"))
}

// ClassifierIDValue returns the value of the classifier constant: the
// classifier id, or the number of classifiers of the package when the id is 0
// (v1 getClassifierIDValue, a quirk kept).
func (m *Model) ClassifierIDValue(c ecore.EClassifier) int {
	if id := c.GetClassifierID(); id != 0 {
		return id
	}
	if p := c.GetEPackage(); p != nil {
		return p.GetEClassifiers().Size()
	}
	return 0
}

// FeatureIDName returns the identifier of the constant of f in c:
// CLASS__FEATURE (v1 getFeatureIDName).
func (m *Model) FeatureIDName(c ecore.EClass, f ecore.EStructuralFeature) string {
	return strings.ToUpper(m.ClassifierIDName(c) + "__" + splitAndCombineWords(f.GetName(), "_"))
}

// FeatureIDValue returns the value of the constant of f in c: the index of f
// in eAllStructuralFeatures, as a literal when c has no base class, as the
// constant of f in the base class when f belongs to it, as the feature count
// constant of the base class plus an offset otherwise (v1 getFeatureIDValue,
// getFeatureIDInteger).
func (m *Model) FeatureIDValue(c ecore.EClass, f ecore.EStructuralFeature) Expr {
	index := c.GetEAllStructuralFeatures().IndexOf(f)
	base := firstSuperType(c)
	if base == nil {
		return literalExpr(index)
	}
	baseCount := base.GetFeatureCount()
	if index < baseCount {
		return baseExpr(base.GetEPackage(), m.FeatureIDName(base, f), m.FeatureIDValue(base, f).Int())
	}
	return offsetExpr(base.GetEPackage(), m.FeatureCountIDName(base), m.FeatureCountIDValue(base).Int(), index-baseCount)
}

// FeatureCountIDName returns the identifier of the feature count constant of
// c: CLASS_FEATURE_COUNT (v1 getFeatureCountIDName).
func (m *Model) FeatureCountIDName(c ecore.EClass) string {
	return m.ClassifierIDName(c) + "_FEATURE_COUNT"
}

// FeatureCountIDValue returns the value of the feature count constant of c:
// the feature count as a literal when c has no base class, the feature count
// constant of the base class plus the number of added features otherwise (v1
// getFeatureCountIDValue, getFeatureCountIDInteger).
func (m *Model) FeatureCountIDValue(c ecore.EClass) Expr {
	count := c.GetFeatureCount()
	base := firstSuperType(c)
	if base == nil {
		return literalExpr(count)
	}
	return offsetExpr(base.GetEPackage(), m.FeatureCountIDName(base), m.FeatureCountIDValue(base).Int(), count-base.GetFeatureCount())
}

// OperationIDName returns the identifier of the constant of op in c:
// CLASS__OPERATION, suffixed by the parameter type names when op has
// parameters (v1 getOperationIDName).
func (m *Model) OperationIDName(c ecore.EClass, op ecore.EOperation) string {
	return strings.ToUpper(m.ClassifierIDName(c) + "__" + splitAndCombineWords(op.GetName(), "_") + m.parameterListName(op))
}

// OperationIDValue returns the value of the constant of op in c: the index of
// op in eAllOperations, as a literal when c has no base class, as the constant
// of op in the base class when op belongs to it, as the operation count
// constant of the base class plus an offset otherwise (v1 getOperationIDValue,
// getOperationIDInteger).
func (m *Model) OperationIDValue(c ecore.EClass, op ecore.EOperation) Expr {
	index := c.GetEAllOperations().IndexOf(op)
	base := firstSuperType(c)
	if base == nil {
		return literalExpr(index)
	}
	baseCount := base.GetOperationCount()
	if index < baseCount {
		return baseExpr(base.GetEPackage(), m.OperationIDName(base, op), m.OperationIDValue(base, op).Int())
	}
	return offsetExpr(base.GetEPackage(), m.OperationCountIDName(base), m.OperationCountIDValue(base).Int(), index-baseCount)
}

// OperationCountIDName returns the identifier of the operation count constant
// of c: CLASS_OPERATION_COUNT (v1 getOperationCountIDName).
func (m *Model) OperationCountIDName(c ecore.EClass) string {
	return m.ClassifierIDName(c) + "_OPERATION_COUNT"
}

// OperationCountIDValue returns the value of the operation count constant of
// c: the operation count as a literal when c has no base class, the operation
// count constant of the base class plus the number of added operations
// otherwise (v1 getOperationCountIDValue, getOperationCountIDInteger).
func (m *Model) OperationCountIDValue(c ecore.EClass) Expr {
	count := c.GetOperationCount()
	base := firstSuperType(c)
	if base == nil {
		return literalExpr(count)
	}
	return offsetExpr(base.GetEPackage(), m.OperationCountIDName(base), m.OperationCountIDValue(base).Int(), count-base.GetOperationCount())
}

// MemberName returns the name of the member holding el in the package (v1
// getMemberName): the name lower first for a classifier, class_Feature for a
// feature, class_Operation with the parameter list name for an operation, the
// name lower first otherwise.
func (m *Model) MemberName(el ecore.ENamedElement) string {
	switch el := el.(type) {
	case ecore.EClassifier:
		return lowerFirst(el.GetName())
	case ecore.EStructuralFeature:
		return lowerFirst(containingClassName(el.GetEContainingClass())) + "_" + upperFirst(el.GetName())
	case ecore.EOperation:
		return lowerFirst(containingClassName(el.GetEContainingClass())) + "_" + upperFirst(el.GetName()) + m.parameterListName(el)
	default:
		return lowerFirst(el.GetName())
	}
}

// containingClassName returns the name of c, "" when nil.
func containingClassName(c ecore.EClass) string {
	if c == nil {
		return ""
	}
	return c.GetName()
}

// parameterListName returns the suffix distinguishing an operation by its
// parameters: "_" + the parameter list details name when it has parameters,
// "" otherwise (v1 getParameterListName).
func (m *Model) parameterListName(op ecore.EOperation) string {
	if op.GetEParameters().Empty() {
		return ""
	}
	return "_" + m.parameterListDetailsName(op)
}

// parameterListDetailsName returns the type name of the single parameter, or
// the type names of the first and the last parameters joined by "_" (v1
// getParameterListDetailsName).
func (m *Model) parameterListDetailsName(op ecore.EOperation) string {
	parameters := op.GetEParameters()
	first := parameters.Get(0).(ecore.EParameter)
	if parameters.Size() > 1 {
		last := parameters.Get(parameters.Size() - 1).(ecore.EParameter)
		return typeName(first.GetEType()) + "_" + typeName(last.GetEType())
	}
	return typeName(first.GetEType())
}

// typeName returns the name of c, "" when nil.
func typeName(c ecore.EClassifier) string {
	if c == nil {
		return ""
	}
	return c.GetName()
}

// ElementIdentifier returns the identifier of el: its name, with the
// parameter list name for an operation (v1 getElementIdentifier).
func (m *Model) ElementIdentifier(el ecore.ENamedElement) string {
	if op, isOp := el.(ecore.EOperation); isOp {
		return el.GetName() + m.parameterListName(op)
	}
	return el.GetName()
}

// ElementAccessorName returns the name of the package accessor of el: its
// member name upper first (v1 getElementAccessorName).
func (m *Model) ElementAccessorName(el ecore.ENamedElement) string {
	return upperFirst(m.MemberName(el))
}

// ElementIDAccessorName returns the kind of id the instances of an ecore
// class have: "Classifier" for EClass, EDataType and EEnum, "Feature" for
// EAttribute and EReference, "Operation" for EOperation, "" otherwise (v1
// getElementIDAccessorName).
func (m *Model) ElementIDAccessorName(c ecore.EClass) string {
	switch c.GetName() {
	case "EClass", "EDataType", "EEnum":
		return "Classifier"
	case "EAttribute", "EReference":
		return "Feature"
	case "EOperation":
		return "Operation"
	default:
		return ""
	}
}
