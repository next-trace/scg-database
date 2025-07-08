package db_test

import (
	"testing"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/db"

	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/require"
)

type (
	dummyAdapter struct{}
)

func (d *dummyAdapter) Name() string                                            { return "dummy" }
func (d *dummyAdapter) Connect(cfg *config.Config) (contract.Connection, error) { return nil, nil }

func TestRegisterAdapter(t *testing.T) {
	// Note: Registry is global, but tests should be independent

	t.Run("PanicsOnNilAdapter", func(t *testing.T) {
		require.Panics(t, func() {
			db.RegisterAdapter(nil, "nil-adapter")
		})
	})

	t.Run("PanicsOnEmptyName", func(t *testing.T) {
		require.Panics(t, func() {
			db.RegisterAdapter(&dummyAdapter{}, "")
		})
	})

	t.Run("PanicsOnNoName", func(t *testing.T) {
		require.Panics(t, func() {
			db.RegisterAdapter(&dummyAdapter{})
		})
	})

	t.Run("Success", func(t *testing.T) {
		adapter := &dummyAdapter{}
		db.RegisterAdapter(adapter, "dummy", "dummy-alias")

		retrieved, err := db.GetAdapter("dummy")
		require.NoError(t, err)
		require.Same(t, adapter, retrieved)

		retrievedAlias, err := db.GetAdapter("dummy-alias")
		require.NoError(t, err)
		require.Same(t, adapter, retrievedAlias)
	})
}
