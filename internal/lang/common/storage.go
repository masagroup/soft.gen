package common

import (
	"fmt"
	"slices"

	"github.com/masagroup/soft.go/ecore"

	"github.com/masagroup/soft.gen/v2/internal/engine"
)

// Storage is where the implementations keep the values of their features
// (property storage).
type Storage string

const (
	// StorageMemory keeps the values in the fields of the implementations.
	StorageMemory Storage = "memory"
	// StorageStore delegates the features to the reflective store.
	StorageStore Storage = "store"
	// StorageHybrid delegates the features to the reflective store, and keeps
	// the values read in typed fields behind the getters.
	StorageHybrid Storage = "hybrid"
)

// Storage returns the storage of the implementations: the property storage,
// else featureDelegation, its v1 and EMF name, Reflective being store and any
// other value memory. Both given must agree.
func (g Gen) Storage() (Storage, error) {
	delegated := StorageMemory
	fd, hasFD := g.ExplicitPackageProperty("featureDelegation")
	if hasFD && fd == "Reflective" {
		delegated = StorageStore
	}
	s, ok := g.ExplicitPackageProperty("storage")
	if !ok {
		return delegated, nil
	}
	storage := Storage(s)
	switch storage {
	case StorageMemory, StorageStore, StorageHybrid:
	default:
		return "", fmt.Errorf("storage=%s: memory, store or hybrid", s)
	}
	if hasFD && delegated != storage && (delegated != StorageStore || storage != StorageHybrid) {
		return "", fmt.Errorf("storage=%s and featureDelegation=%s disagree", s, fd)
	}
	return storage, nil
}

// storage returns the storage of the implementations, memory when the
// properties are invalid: CheckStorage reports them before any generation.
func (g Gen) storage() Storage {
	s, err := g.Storage()
	if err != nil {
		return StorageMemory
	}
	return s
}

// IsReflectiveDelegation reports whether the implementations delegate their
// features to the reflective store: storage store or hybrid
// (featureDelegation=Reflective).
func (g Gen) IsReflectiveDelegation() bool {
	s := g.storage()
	return s == StorageStore || s == StorageHybrid
}

// IsHybridStorage reports whether the implementations keep the values read
// from the store in typed fields (storage=hybrid).
func (g Gen) IsHybridStorage() bool {
	return g.storage() == StorageHybrid
}

// CheckStorage returns files, failing first when the storage properties are
// invalid or give a storage other than supported.
func CheckStorage(files engine.FileLister, supported ...Storage) engine.FileLister {
	return func(ctx *engine.Context, pkg ecore.EPackage) ([]engine.File, error) {
		s, err := NewGen(ctx).Storage()
		if err != nil {
			return nil, err
		}
		if !slices.Contains(supported, s) {
			return nil, fmt.Errorf("storage=%s is not supported by %s", s, ctx.Lang.Name())
		}
		return files(ctx, pkg)
	}
}
