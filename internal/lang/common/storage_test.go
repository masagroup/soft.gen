package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/masagroup/soft.gen/v2/internal/props"
)

func storageGen(t *testing.T, cli map[string]string) Gen {
	t.Helper()
	p, err := props.New([]props.Decl{
		{Name: "storage", Default: "memory", Scope: props.ScopePackage},
		{Name: "featureDelegation", Scope: props.ScopePackage},
	}, cli, nil, nil)
	require.NoError(t, err)
	return Gen{Props: p}
}

func TestStorage(t *testing.T) {
	for _, c := range []struct {
		name       string
		cli        map[string]string
		storage    Storage
		reflective bool
		hybrid     bool
	}{
		{"default", nil, StorageMemory, false, false},
		{"memory", map[string]string{"storage": "memory"}, StorageMemory, false, false},
		{"store", map[string]string{"storage": "store"}, StorageStore, true, false},
		{"hybrid", map[string]string{"storage": "hybrid"}, StorageHybrid, true, true},
		{"featureDelegation alias", map[string]string{"featureDelegation": "Reflective"}, StorageStore, true, false},
		{"featureDelegation other value", map[string]string{"featureDelegation": "None"}, StorageMemory, false, false},
		{"both agree", map[string]string{"storage": "store", "featureDelegation": "Reflective"}, StorageStore, true, false},
		{"hybrid over Reflective", map[string]string{"storage": "hybrid", "featureDelegation": "Reflective"}, StorageHybrid, true, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := storageGen(t, c.cli)
			s, err := g.Storage()
			require.NoError(t, err)
			assert.Equal(t, c.storage, s)
			assert.Equal(t, c.reflective, g.IsReflectiveDelegation())
			assert.Equal(t, c.hybrid, g.IsHybridStorage())
		})
	}
}

func TestStorage_Invalid(t *testing.T) {
	for name, cli := range map[string]map[string]string{
		"unknown value":      {"storage": "disk"},
		"memory, Reflective": {"storage": "memory", "featureDelegation": "Reflective"},
		"store, None":        {"storage": "store", "featureDelegation": "None"},
	} {
		t.Run(name, func(t *testing.T) {
			g := storageGen(t, cli)
			_, err := g.Storage()
			assert.Error(t, err)
			assert.False(t, g.IsReflectiveDelegation(), "memory when invalid")
		})
	}
}
