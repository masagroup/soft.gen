package typescript

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/masagroup/soft.gen/v2/internal/engine"
	"github.com/masagroup/soft.gen/v2/internal/props"
)

// TestValidate checks the TypeScript language: names, entry points,
// properties, templates and their conventions.
func TestValidate(t *testing.T) {
	if err := engine.Validate(New()); err != nil {
		t.Fatal(err)
	}
}

// TestEntryPoints_InvalidStorage checks that every entry point fails with a
// storage the generator does not know, instead of generating typed classes.
func TestEntryPoints_InvalidStorage(t *testing.T) {
	l := New()
	p, err := props.New(l.Properties(), map[string]string{"storage": "disk"}, nil, nil)
	require.NoError(t, err)
	ctx := &engine.Context{Props: p, Lang: l}
	for _, ep := range l.EntryPoints() {
		_, err := ep.Files(ctx, nil)
		assert.ErrorContains(t, err, "memory, store or hybrid", ep.Name)
	}
}
