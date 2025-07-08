package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3" // driver
	_ "github.com/golang-migrate/migrate/v4/source/file"      // driver
	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
)

// Database driver constants
const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"
)

// validateConfigForMigration validates the configuration required for migrations
func validateConfigForMigration(cfg *config.Config) error {
	if strings.TrimSpace(cfg.MigrationsPath) == "" {
		return errors.New("migrations path is required")
	}
	if strings.TrimSpace(cfg.DSN) == "" {
		return errors.New("database dsn is required")
	}
	if strings.TrimSpace(cfg.Driver) == "" {
		return errors.New("database driver is required")
	}
	return nil
}

// mapDriverName maps composite driver names to SQL driver names
func mapDriverName(driver string) string {
	switch driver {
	case "gorm:mysql", DriverMySQL:
		return DriverMySQL
	case "gorm:postgres", DriverPostgres:
		return DriverPostgres
	default:
		// Assume the driver name is compatible if not a known composite
		return driver
	}
}

// safeCloseSQLDB safely closes a SQL database connection with error handling
func safeCloseSQLDB(sqlDB *sql.DB) error {
	if sqlDB == nil {
		return nil
	}
	return sqlDB.Close()
}

type (
	// migrationDriverFactory creates database drivers using a factory pattern
	migrationDriverFactory struct{}

	// Migrator is the main migration runner.
	Migrator struct {
		migrate *migrate.Migrate
	}
)

// createDriver creates a database driver for the given driver name and connection
func (f *migrationDriverFactory) createDriver(driverName string, sqlDB *sql.DB) (database.Driver, error) {
	switch driverName {
	case DriverMySQL:
		return mysql.WithInstance(sqlDB, &mysql.Config{})
	case DriverPostgres:
		return postgres.WithInstance(sqlDB, &postgres.Config{})
	default:
		return nil, fmt.Errorf("unsupported migration driver: %s", driverName)
	}
}

// NewMigrator creates a new Migrator using config.Config with validation and migration helpers.
func NewMigrator(cfg *config.Config) (contract.Migrator, error) {
	// Use validation helper for input checking
	if err := validateConfigForMigration(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Map driver name using migration helper
	sqlDriverName := mapDriverName(cfg.Driver)

	// Open database connection with error handling
	sqlDB, err := sql.Open(sqlDriverName, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database for migration: %w", err)
	}

	// Create migration driver factory for strategy pattern
	factory := &migrationDriverFactory{}

	// Create database driver using strategy pattern
	driver, err := factory.createDriver(sqlDriverName, sqlDB)
	if err != nil {
		// Use helper for safe cleanup
		_ = safeCloseSQLDB(sqlDB)
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance with proper error handling
	m, err := migrate.NewWithDatabaseInstance(cfg.MigrationsPath, sqlDriverName, driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

// Up applies all available up migrations.
func (m *Migrator) Up() error {
	err := m.migrate.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil // Not an error
	}
	return err
}

// Down rolls back a specific number of migrations.
func (m *Migrator) Down(steps int) error {
	if steps <= 0 {
		return nil // Nothing to do
	}
	err := m.migrate.Steps(-steps)
	if errors.Is(err, migrate.ErrNoChange) {
		return nil // Not an error
	}
	return err
}

// Fresh drops all tables and re-runs all migrations from scratch.
func (m *Migrator) Fresh() error {
	// Drop all tables
	err := m.migrate.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	// Run all migrations
	err = m.migrate.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil // Not an error, although unusual after a Down()
	}
	return err
}

// Close closes the underlying source and database connections.
func (m *Migrator) Close() (sourceErr, dbErr error) {
	return m.migrate.Close()
}
