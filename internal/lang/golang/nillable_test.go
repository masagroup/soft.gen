package golang

import "testing"

func TestIsNillableType(t *testing.T) {
	for typ, want := range map[string]bool{
		"*time.Time":     true,
		"[]byte":         true,
		"map[string]int": true,
		"chan int":       true,
		"<-chan int":     true,
		"func(int) bool": true,
		"int":            false,
		"string":         false,
		"time.Duration":  false,
		"ecore.EList":    false,
		"any":            false,
	} {
		if got := isNillableType(typ); got != want {
			t.Errorf("isNillableType(%q) = %v, want %v", typ, got, want)
		}
	}
}
