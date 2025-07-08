// Package utils provides utility functions for database configuration, connection pooling,
// and other common operations used throughout the SCG database toolkit.
//
//revive:disable:var-naming // allow package name 'utils' for this utilities module
package utils

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

// Database driver constants for migration
const (
	MigrationDriverMySQL    = "mysql"
	MigrationDriverPostgres = "postgres"
	MigrationDriverSQLite   = "sqlite"
)

// MapDriverName maps composite driver names to SQL driver names
func MapDriverName(driver string) string {
	switch driver {
	case "gorm:mysql", MigrationDriverMySQL:
		return MigrationDriverMySQL
	case "gorm:postgres", MigrationDriverPostgres:
		return MigrationDriverPostgres
	default:
		// Assume the driver name is compatible if not a known composite
		return driver
	}
}

// CreateDatabaseDriver creates a database driver instance for migrations
func CreateDatabaseDriver(sqlDriverName string, sqlDB *sql.DB) (database.Driver, error) {
	switch sqlDriverName {
	case MigrationDriverMySQL:
		return mysql.WithInstance(sqlDB, &mysql.Config{})
	case MigrationDriverPostgres:
		return postgres.WithInstance(sqlDB, &postgres.Config{})
	default:
		return nil, fmt.Errorf("unsupported migration driver: %s", sqlDriverName)
	}
}

// HandleMigrationError handles common migration errors, particularly ErrNoChange
func HandleMigrationError(err error) error {
	if errors.Is(err, migrate.ErrNoChange) {
		return nil // Not an error
	}
	return err
}

// SafeCloseSQLDB safely closes a SQL database connection with error handling
func SafeCloseSQLDB(sqlDB *sql.DB) error {
	if sqlDB == nil {
		return nil
	}
	return sqlDB.Close()
}

type (
	// MigrationDriverFactory creates database drivers using a factory pattern
	MigrationDriverFactory struct{}

	// MigrationResult represents the result of a migration operation
	MigrationResult struct {
		Success bool
		Error   error
		Message string
	}

	// MigrationConfig holds configuration for migration operations
	MigrationConfig struct {
		DriverName     string
		DSN            string
		MigrationsPath string
		Steps          int
	}
)

// CreateDriver creates a database driver for the given driver name and connection
func (f *MigrationDriverFactory) CreateDriver(driverName string, sqlDB *sql.DB) (database.Driver, error) {
	return CreateDatabaseDriver(driverName, sqlDB)
}

// SupportedDrivers returns a list of supported migration drivers
func (f *MigrationDriverFactory) SupportedDrivers() []string {
	return []string{MigrationDriverMySQL, MigrationDriverPostgres}
}

// IsDriverSupported checks if a driver is supported for migrations
func (f *MigrationDriverFactory) IsDriverSupported(driverName string) bool {
	supported := f.SupportedDrivers()
	for _, driver := range supported {
		if driver == driverName {
			return true
		}
	}
	return false
}

// NewMigrationResult creates a new migration result
func NewMigrationResult(success bool, err error, message string) *MigrationResult {
	return &MigrationResult{
		Success: success,
		Error:   err,
		Message: message,
	}
}

// ExecuteMigrationWithCleanup executes a migration operation with automatic cleanup
func ExecuteMigrationWithCleanup(sqlDB *sql.DB, operation func() error) *MigrationResult {
	err := operation()

	// Handle migration-specific errors
	err = HandleMigrationError(err)
	if err != nil {
		// Ensure cleanup on error
		if closeErr := SafeCloseSQLDB(sqlDB); closeErr != nil {
			// Log the close error but return the original error
			return NewMigrationResult(false, err, fmt.Sprintf("Migration failed: %v (cleanup error: %v)", err, closeErr))
		}
		return NewMigrationResult(false, err, fmt.Sprintf("Migration failed: %v", err))
	}

	return NewMigrationResult(true, nil, "Migration completed successfully")
}

// Validate validates the migration configuration
func (c *MigrationConfig) Validate() error {
	if c.DriverName == "" {
		return errors.New("driver name is required")
	}
	if c.DSN == "" {
		return errors.New("DSN is required")
	}
	if c.MigrationsPath == "" {
		return errors.New("migrations path is required")
	}
	return nil
}
