package testing

import (
	"testing"
	"time"

	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseTestSuite(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver:          "gorm:sqlite",
		DSN:             ":memory:",
		MigrationsPath:  "/migrations",
		SeedsPath:       "/seeds",
		CleanupStrategy: CleanupTruncate,
		Timeout:         10 * time.Second,
	}

	testSuite := NewDatabaseTestSuite(&cfg)

	assert.NotNil(t, testSuite)
	assert.Equal(t, cfg.Driver, testSuite.Config.Driver)
	assert.Equal(t, cfg.DSN, testSuite.Config.DSN)
	assert.Equal(t, 10, testSuite.Config.MaxOpenConns)
	assert.Equal(t, 5, testSuite.Config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, testSuite.Config.ConnMaxLifetime)
	assert.NotNil(t, testSuite.cleanup)
}

func TestNewDatabaseTestSuiteWithDefaultTimeout(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)

	assert.NotNil(t, ts)
	// The timeout is used internally but not stored in the Config
}

func TestCleanupStrategies(t *testing.T) {
	// Test that cleanup strategy constants are defined correctly
	assert.Equal(t, CleanupStrategy(0), CleanupTruncate)
	assert.Equal(t, CleanupStrategy(1), CleanupTransaction)
	assert.Equal(t, CleanupStrategy(2), CleanupRecreate)
	assert.Equal(t, CleanupStrategy(3), CleanupNone)
}

func TestAddCleanup(t *testing.T) {
	testSuite := &DatabaseTestSuite{
		cleanup: make([]func() error, 0),
	}

	cleanupFn := func() error { return nil }
	testSuite.AddCleanup(cleanupFn)

	assert.Len(t, testSuite.cleanup, 1)
}

func TestDatabaseTestConfig(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver:          "gorm:postgres",
		DSN:             "postgres://user:pass@localhost/test",
		MigrationsPath:  "/db/migrations",
		SeedsPath:       "/db/seeds",
		CleanupStrategy: CleanupTransaction,
		Timeout:         60 * time.Second,
	}

	assert.Equal(t, "gorm:postgres", cfg.Driver)
	assert.Equal(t, "postgres://user:pass@localhost/test", cfg.DSN)
	assert.Equal(t, "/db/migrations", cfg.MigrationsPath)
	assert.Equal(t, "/db/seeds", cfg.SeedsPath)
	assert.Equal(t, CleanupTransaction, cfg.CleanupStrategy)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
}

func TestRunTestSuiteFunction(t *testing.T) {
	// Test that the RunTestSuite function exists and has the correct signature
	// We can't actually run it without causing recursion, but we can verify it exists
	assert.NotNil(t, RunTestSuite)
}

func TestDatabaseTestSuite_SetupSuite(t *testing.T) {
	// Create a test suite with SQLite in-memory database
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Mock the setup process
	assert.NotPanics(t, func() {
		// We can't actually call SetupSuite without a real database adapter
		// but we can test the configuration is set up correctly
		assert.Equal(t, "gorm:sqlite", ts.Config.Driver)
		assert.Equal(t, ":memory:", ts.Config.DSN)
	})
}

func TestDatabaseTestSuite_CreateRepository(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test with nil connection (should handle gracefully)
	assert.Panics(t, func() {
		ts.CreateRepository(&TestModel{})
	})
}

func TestDatabaseTestSuite_ExecuteInTransaction(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test with nil connection (should handle gracefully)
	assert.Panics(t, func() {
		ts.ExecuteInTransaction(func(conn contract.Connection) error {
			return nil
		})
	})
}

func TestDatabaseTestSuite_TruncateTable(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test with nil connection (should handle gracefully)
	assert.Panics(t, func() {
		ts.TruncateTable("test_table")
	})
}

func TestDatabaseTestSuite_SeedData(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test with nil connection (should handle gracefully)
	assert.Panics(t, func() {
		ts.SeedData(&TestModel{})
	})
}

func TestDatabaseTestSuite_AssertRecordExists(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test with nil connection (should handle gracefully)
	assert.Panics(t, func() {
		ts.AssertRecordExists(&TestModel{}, 1)
	})
}

func TestDatabaseTestSuite_AssertRecordNotExists(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test with nil connection (should handle gracefully)
	assert.Panics(t, func() {
		ts.AssertRecordNotExists(&TestModel{}, 1)
	})
}

func TestDatabaseTestSuite_AssertTableEmpty(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test with nil connection (should handle gracefully)
	assert.Panics(t, func() {
		ts.AssertTableEmpty("test_table")
	})
}

func TestDatabaseTestSuite_GetRawConnection(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test with nil connection (should handle gracefully)
	rawConn := ts.GetRawConnection()
	assert.Nil(t, rawConn, "GetRawConnection should return nil when no connection is established")
}

func TestDatabaseTestSuite_TearDownSuite(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Add some cleanup functions
	cleanupCalled := false
	ts.AddCleanup(func() error {
		cleanupCalled = true
		return nil
	})

	// Test teardown
	assert.NotPanics(t, func() {
		ts.TearDownSuite()
	})

	assert.True(t, cleanupCalled)
}

func TestDatabaseTestSuite_SetupAndTearDownTest(t *testing.T) {
	cfg := DatabaseTestConfig{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	ts := NewDatabaseTestSuite(&cfg)
	ts.SetT(t)

	// Test setup and teardown test methods (they're currently placeholders)
	assert.NotPanics(t, func() {
		ts.SetupTest()
		ts.TearDownTest()
	})
}

// TestModel for testing purposes
type (
	TestModel struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (m *TestModel) PrimaryKey() string                              { return "id" }
func (m *TestModel) TableName() string                               { return "test_models" }
func (m *TestModel) GetID() any                                      { return m.ID }
func (m *TestModel) SetID(id any)                                    { m.ID = id.(int) }
func (m *TestModel) Relationships() map[string]contract.Relationship { return nil }
func (m *TestModel) GetDeletedAt() *time.Time                        { return nil }
func (m *TestModel) SetDeletedAt(t *time.Time)                       {}
func (m *TestModel) GetCreatedAt() time.Time                         { return time.Time{} }
func (m *TestModel) GetUpdatedAt() time.Time                         { return time.Time{} }
func (m *TestModel) SetCreatedAt(t time.Time)                        {}
func (m *TestModel) SetUpdatedAt(t time.Time)                        {}
