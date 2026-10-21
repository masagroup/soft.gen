package golang

// The inheritance model: an interface class (Identified) and an abstract
// class (Base); a diamond (Bottom extends Left and Right: its interface embeds
// EObject and redeclares everything; Right is a mixin with features:
// EDerivedFeatureID); a class extending an interface and an interface with
// a list attribute (Item); a class with GenGo extension=true (Ext) and a
// subclass of it (SubExt embeds the extension, the factory creates the
// extension); classes named like Ecore's: EObject (ENotifier,
// BasicEObjectImpl, MockENotifier, NewMockEObjectInternal test value),
// EClassifier (GetEClassifierClass getter), EClass, EAttribute and
// EOperation with a container reference (Create<X>FromContainerAndClassID
// with Set<Kind>ID).
func init() {
	register(fixture{Name: "inheritance", Model: "inheritance.ecore",
		V1Differences: map[string]string{
			"inheritance/bottom_impl.go":      v1MixinOrder,
			"inheritance/bottom_impl_test.go": v1MixinOrder,
		}})
}
