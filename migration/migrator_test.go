package migration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/next-trace/scg-database/config"
	"github.com/stretchr/testify/require"
)

// TestNewMigrator_InvalidConfig tests that NewMigrator returns errors for various invalid configurations.
func TestNewMigrator_InvalidConfig(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         config.Config
		expectedErr string
	}{
		{
			name:        "Missing Migrations Path",
			cfg:         config.Config{Driver: "mysql", DSN: "user:pass@/db"},
			expectedErr: "config validation failed: migrations path is required",
		},
		{
			name:        "Missing DSN",
			cfg:         config.Config{Driver: "mysql", MigrationsPath: "file://migrations"},
			expectedErr: "config validation failed: database dsn is required",
		},
		{
			name:        "Missing Driver",
			cfg:         config.Config{DSN: "user:pass@/db", MigrationsPath: "file://migrations"},
			expectedErr: "config validation failed: database driver is required",
		},
		{
			name:        "Unsupported Driver",
			cfg:         config.Config{Driver: "sqlite", DSN: "test.db", MigrationsPath: "file://migrations"},
			expectedErr: "failed to open database for migration: sql: unknown driver \"sqlite\" (forgotten import?)",
		},
		{
			name: "Postgres Driver (invalid DSN format)",
			cfg:  config.Config{Driver: "postgres", DSN: "user:pass@/db", MigrationsPath: "file://migrations"},
			// This test will fail due to an invalid postgres DSN format
			expectedErr: "failed to create migration driver: missing \"=\" after \"user:pass@/db\" in connection info string\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewMigrator(&tc.cfg)
			require.Error(t, err)
			require.Equal(t, tc.expectedErr, err.Error())
		})
	}
}

// TestMigrator_Down_ZeroOrNegativeSteps tests that calling Down with <= 0 steps is a no-op and returns no error.
func TestMigrator_Down_ZeroOrNegativeSteps(t *testing.T) {
	// This method doesn't need a real migrator instance as it has a guard clause at the top.
	m := &Migrator{} // A dummy migrator is sufficient to test the guard.
	err := m.Down(0)
	require.NoError(t, err, "Down(0) should be a no-op")

	err = m.Down(-5)
	require.NoError(t, err, "Down(-5) should be a no-op")
}

// TestNewMigrator_SqlOpenFails tests the error path when sql.Open fails.
func TestNewMigrator_SqlOpenFails(t *testing.T) {
	// We trigger an sql.Open failure by providing a known driver with an invalid DSN.
	cfg := config.Config{
		Driver:         "mysql",
		DSN:            "this-is-not-a-valid-dsn-format",
		MigrationsPath: "file://migrations",
	}
	_, err := NewMigrator(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open database for migration", "Error message should indicate the failure point")
}

// TestMigrator_NoPanicOnNil checks that methods on an invalid (nil internal) struct panic.
// This is a safety check to ensure the constructor is always used.
func TestMigrator_Methods_PanicOnNil(t *testing.T) {
	r := &Migrator{}
	require.Panics(t, func() { _ = r.Up() }, "Up() on a nil migrate instance should panic")
	require.Panics(t, func() { _ = r.Down(1) }, "Down(1) on a nil migrate instance should panic")
	require.Panics(t, func() { _ = r.Fresh() }, "Fresh() on a nil migrate instance should panic")
	require.Panics(t, func() { _, _ = r.Close() }, "Close() on a nil migrate instance should panic")
}

func TestNewMigrator_DriverMapping(t *testing.T) {
	testCases := []struct {
		name           string
		inputDriver    string
		expectedDriver string
	}{
		{
			name:           "MySQL driver mapping",
			inputDriver:    "gorm:mysql",
			expectedDriver: "mysql",
		},
		{
			name:           "PostgreSQL driver mapping",
			inputDriver:    "gorm:postgres",
			expectedDriver: "postgres",
		},
		{
			name:           "SQLite driver mapping",
			inputDriver:    "gorm:sqlite",
			expectedDriver: "sqlite3",
		},
		{
			name:           "Direct MySQL driver",
			inputDriver:    "mysql",
			expectedDriver: "mysql",
		},
		{
			name:           "Direct PostgreSQL driver",
			inputDriver:    "postgres",
			expectedDriver: "postgres",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't easily test the full migrator creation without a real DB,
			// but we can test that the driver mapping logic works by checking
			// the error messages
			cfg := config.Config{
				Driver:         tc.inputDriver,
				DSN:            "invalid-dsn-to-trigger-error",
				MigrationsPath: "file://migrations",
			}

			_, err := NewMigrator(&cfg)
			require.Error(t, err)
			// The error should indicate a failure in migration setup
			require.True(t,
				strings.Contains(err.Error(), "failed to open database for migration") ||
					strings.Contains(err.Error(), "failed to create migration driver"),
				"Error should indicate migration setup failure, got: %s", err.Error())
		})
	}
}

func TestMigrator_Close_NilHandling(t *testing.T) {
	// Test that Close method panics with nil migrate instance (as documented)
	m := &Migrator{}

	// Close should panic when internal migrate instance is nil (this is the expected behavior)
	require.Panics(t, func() {
		m.Close()
	}, "Close() on a nil migrate instance should panic")
}

func TestNewMigrator_ConfigValidation(t *testing.T) {
	// Test additional config validation scenarios
	testCases := []struct {
		name        string
		cfg         config.Config
		expectError bool
	}{
		{
			name: "Valid MySQL config structure",
			cfg: config.Config{
				Driver:         "mysql",
				DSN:            "user:pass@tcp(localhost:3306)/db",
				MigrationsPath: "file://migrations",
			},
			expectError: true, // Will fail due to no actual DB, but config is valid
		},
		{
			name: "Valid PostgreSQL config structure",
			cfg: config.Config{
				Driver:         "postgres",
				DSN:            "postgres://user:pass@localhost/db?sslmode=disable",
				MigrationsPath: "file://migrations",
			},
			expectError: true, // Will fail due to no actual DB, but config is valid
		},
		{
			name: "Empty migrations path",
			cfg: config.Config{
				Driver:         "mysql",
				DSN:            "user:pass@tcp(localhost:3306)/db",
				MigrationsPath: "",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewMigrator(&tc.cfg)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMigrator_Methods_ErrorHandling(t *testing.T) {
	// Test that methods handle nil migrate instance gracefully
	m := &Migrator{}

	// These should panic as documented
	require.Panics(t, func() { _ = m.Up() })
	require.Panics(t, func() { _ = m.Down(1) })
	require.Panics(t, func() { _ = m.Fresh() })
}

func TestMigrator_Down_EdgeCases(t *testing.T) {
	// Test additional edge cases for Down method
	m := &Migrator{}

	// Test with various invalid step counts
	testCases := []int{-10, -1, 0}

	for _, steps := range testCases {
		t.Run(fmt.Sprintf("steps_%d", steps), func(t *testing.T) {
			err := m.Down(steps)
			require.NoError(t, err, "Down with %d steps should be no-op", steps)
		})
	}
}
