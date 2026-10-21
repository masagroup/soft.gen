package typescript

import (
	"testing"

	"github.com/masagroup/soft.gen/v2/engine"
)

// TestValidate checks the TypeScript language: names, entry points,
// properties, templates and their conventions.
func TestValidate(t *testing.T) {
	if err := engine.Validate(New()); err != nil {
		t.Fatal(err)
	}
}
