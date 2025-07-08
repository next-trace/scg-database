package testing

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-database/db"
	"github.com/stretchr/testify/suite"
)

type (
	// DatabaseTestSuite provides a comprehensive testing framework for microservices
	// that need to test against real databases running in Docker containers.
	DatabaseTestSuite struct {
		suite.Suite
		Connection contract.Connection
		Config     config.Config
		cleanup    []func() error
	}

	// DatabaseTestConfig holds configuration for database testing
	DatabaseTestConfig struct {
		Driver          string
		DSN             string
		MigrationsPath  string
		SeedsPath       string
		CleanupStrategy CleanupStrategy
		Timeout         time.Duration
	}

	// CleanupStrategy defines how to clean up test data
	CleanupStrategy int
)

const (
	// CleanupTruncate truncates all tables after each test
	CleanupTruncate CleanupStrategy = iota
	// CleanupTransaction wraps each test in a transaction and rolls back
	CleanupTransaction
	// CleanupRecreate drops and recreates the database
	CleanupRecreate
	// CleanupNone does no cleanup (useful for debugging)
	CleanupNone
)

// NewDatabaseTestSuite creates a new database test suite with the given configuration
func NewDatabaseTestSuite(cfg *DatabaseTestConfig) *DatabaseTestSuite {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &DatabaseTestSuite{
		Config: config.Config{
			Driver:          cfg.Driver,
			DSN:             cfg.DSN,
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		cleanup: make([]func() error, 0),
	}
}

// SetupSuite runs once before all tests in the suite
func (s *DatabaseTestSuite) SetupSuite() {
	s.T().Log("Setting up database test suite...")

	// Wait for database to be ready (useful for Docker containers)
	s.waitForDatabase()

	// Get database adapter
	adapter, err := db.GetAdapter(s.Config.Driver)
	s.Require().NoError(err, "Failed to get database adapter")

	// Create connection
	conn, err := adapter.Connect(&s.Config)
	s.Require().NoError(err, "Failed to connect to database")

	s.Connection = conn

	// Run migrations if specified
	if migrationsPath := os.Getenv("TEST_MIGRATIONS_PATH"); migrationsPath != "" {
		s.runMigrations(migrationsPath)
	}

	s.T().Log("Database test suite setup completed")
}

// TearDownSuite runs once after all tests in the suite
func (s *DatabaseTestSuite) TearDownSuite() {
	s.T().Log("Tearing down database test suite...")

	// Run cleanup functions in reverse order
	for i := len(s.cleanup) - 1; i >= 0; i-- {
		if err := s.cleanup[i](); err != nil {
			s.T().Logf("Cleanup function failed: %v", err)
		}
	}

	// Close database connection
	if s.Connection != nil {
		if err := s.Connection.Close(); err != nil {
			s.T().Logf("Failed to close database connection: %v", err)
		}
	}

	s.T().Log("Database test suite teardown completed")
}

// SetupTest runs before each test
func (s *DatabaseTestSuite) SetupTest() {
	// Implementation depends on cleanup strategy
	// This is a placeholder - specific strategies will be implemented
}

// TearDownTest runs after each test
func (s *DatabaseTestSuite) TearDownTest() {
	// Implementation depends on cleanup strategy
	// This is a placeholder - specific strategies will be implemented
}

// waitForDatabase waits for the database to be ready
func (s *DatabaseTestSuite) waitForDatabase() {
	timeout := 30 * time.Second
	if envTimeout := os.Getenv("TEST_DB_TIMEOUT"); envTimeout != "" {
		if d, err := time.ParseDuration(envTimeout); err == nil {
			timeout = d
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	s.T().Log("Waiting for database to be ready...")

	for {
		select {
		case <-ctx.Done():
			s.T().Fatal("Database did not become ready within timeout")
		case <-ticker.C:
			if s.isDatabaseReady() {
				s.T().Log("Database is ready")
				return
			}
		}
	}
}

// isDatabaseReady checks if the database is ready to accept connections
func (s *DatabaseTestSuite) isDatabaseReady() bool {
	adapter, err := db.GetAdapter(s.Config.Driver)
	if err != nil {
		return false
	}

	conn, err := adapter.Connect(&s.Config)
	if err != nil {
		return false
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Failed to close test connection: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return conn.Ping(ctx) == nil
}

// runMigrations runs database migrations
func (s *DatabaseTestSuite) runMigrations(migrationsPath string) {
	s.T().Logf("Running migrations from: %s", migrationsPath)
	// Migration implementation would go here
	// This is a placeholder for now
}

// CreateRepository creates a new repository for the given model
func (s *DatabaseTestSuite) CreateRepository(model contract.Model) contract.Repository {
	repo, err := s.Connection.NewRepository(model)
	s.Require().NoError(err, "Failed to create repository")
	return repo
}

// ExecuteInTransaction executes a function within a database transaction
func (s *DatabaseTestSuite) ExecuteInTransaction(fn func(conn contract.Connection) error) error {
	return s.Connection.Transaction(context.Background(), fn)
}

// TruncateTable truncates the specified table
func (s *DatabaseTestSuite) TruncateTable(tableName string) {
	ctx := context.Background()
	query := "TRUNCATE TABLE " + tableName

	// Handle different database dialects
	switch s.Config.Driver {
	case "gorm:sqlite":
		query = "DELETE FROM " + tableName
	case "gorm:postgres":
		query = "TRUNCATE TABLE " + tableName + " RESTART IDENTITY CASCADE"
	}

	_, err := s.Connection.Statement(ctx, query)
	s.Require().NoError(err, fmt.Sprintf("Failed to truncate table: %s", tableName))
}

// SeedData provides utilities for seeding test data
func (s *DatabaseTestSuite) SeedData(models ...contract.Model) {
	for _, model := range models {
		repo := s.CreateRepository(model)
		err := repo.Create(context.Background(), model)
		s.Require().NoError(err, fmt.Sprintf("Failed to seed data for model: %T", model))
	}
}

// AssertRecordExists asserts that a record exists in the database
func (s *DatabaseTestSuite) AssertRecordExists(model contract.Model, id any) {
	repo := s.CreateRepository(model)
	found, err := repo.Find(context.Background(), id)
	s.Require().NoError(err, "Failed to find record")
	s.Require().NotNil(found, "Record should exist")
}

// AssertRecordNotExists asserts that a record does not exist in the database
func (s *DatabaseTestSuite) AssertRecordNotExists(model contract.Model, id any) {
	repo := s.CreateRepository(model)
	found, err := repo.Find(context.Background(), id)
	s.Require().NoError(err, "Failed to find record")
	s.Require().Nil(found, "Record should not exist")
}

// AssertTableEmpty asserts that a table is empty
func (s *DatabaseTestSuite) AssertTableEmpty(tableName string) {
	ctx := context.Background()
	query := "SELECT COUNT(*) FROM " + tableName

	rows, err := s.Connection.Select(ctx, query)
	s.Require().NoError(err, fmt.Sprintf("Failed to count records in table: %s", tableName))
	s.Require().Len(rows, 1, "Expected one row from count query")

	count, ok := rows[0]["COUNT(*)"]
	if !ok {
		count, ok = rows[0]["count"]
		if !ok {
			count, ok = rows[0]["count(*)"]
		}
	}
	s.Require().True(ok, "Could not find count column in result")
	s.Require().Equal(int64(0), count, fmt.Sprintf("Table should be empty: %s", tableName))
}

// GetRawConnection returns the underlying database connection for advanced operations
func (s *DatabaseTestSuite) GetRawConnection() *sql.DB {
	if s.Connection == nil {
		return nil
	}

	if gormConn := s.Connection.GetConnection(); gormConn != nil {
		if gormDB, ok := gormConn.(interface{ DB() (*sql.DB, error) }); ok {
			if sqlDB, err := gormDB.DB(); err == nil {
				return sqlDB
			}
		}
	}
	return nil
}

// AddCleanup adds a cleanup function to be executed during teardown
func (s *DatabaseTestSuite) AddCleanup(fn func() error) {
	s.cleanup = append(s.cleanup, fn)
}

// RunTestSuite runs the test suite with the given testing.T
func RunTestSuite(t *testing.T, testSuite suite.TestingSuite) {
	suite.Run(t, testSuite)
}
