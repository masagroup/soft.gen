package model

import "github.com/masagroup/soft.go/ecore"

// orderedClasses returns the classes of p ordered so that the first super type
// chain of a class precedes it (v1 getOrderedClasses, GenerateUtils.java): for
// each class of the package in model order, its chain of first super types
// belonging to the package, from the root of the chain down to the class, the
// classes already placed being skipped.
func orderedClasses(p ecore.EPackage) []ecore.EClass {
	var result []ecore.EClass
	resultSet := map[ecore.EClass]struct{}{}
	for _, classifier := range p.GetEClassifiers().ToArray() {
		eClass, isClass := classifier.(ecore.EClass)
		if !isClass {
			continue
		}
		var extendChain []ecore.EClass
		visited := map[ecore.EClass]struct{}{}
		for eClass != nil {
			if _, seen := visited[eClass]; seen {
				break
			}
			visited[eClass] = struct{}{}
			if eClass.GetEPackage() == p {
				if _, placed := resultSet[eClass]; !placed {
					resultSet[eClass] = struct{}{}
					extendChain = append([]ecore.EClass{eClass}, extendChain...)
				}
			}
			eClass = firstSuperType(eClass)
		}
		result = append(result, extendChain...)
	}
	return result
}

// firstSuperType returns the first super type of c, nil when it has none (v1
// getSuperType, GenerateUtils.java; v1 getBaseClass).
func firstSuperType(c ecore.EClass) ecore.EClass {
	superTypes := c.GetESuperTypes()
	if superTypes.Empty() {
		return nil
	}
	superType, _ := superTypes.Get(0).(ecore.EClass)
	return superType
}
