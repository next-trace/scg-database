// Package gorm provides a GORM-based implementation of the SCG database toolkit interfaces.
// It includes adapters, connections, repositories, and query builders for GORM ORM.
package gorm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-database/db"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database driver constants
const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"

	GormDriverMySQL    = "gorm:mysql"
	GormDriverPostgres = "gorm:postgres"
	GormDriverSQLite   = "gorm:sqlite"
)

//nolint:grouper // Only One Global Variable
var registerOnce sync.Once

type (
	// Adapter is the GORM implementation of the contract.DBAdapter interface.
	Adapter struct{}

	// DialectStrategy implements utils.DialectStrategy for GORM
	DialectStrategy struct {
		dialectName string
	}
)

// Register registers the GORM adapter with the central registry.
// This function is safe to call multiple times and will only register once.
func Register() {
	registerOnce.Do(func() {
		// Register this adapter with the central registry for CLI use.
		db.RegisterAdapter(&Adapter{}, "gorm", GormDriverMySQL, GormDriverPostgres, GormDriverSQLite)

		// Register the GORM query builder factory
		db.RegisterQueryBuilderFactory("gorm", &GormQueryBuilderFactory{})
	})
}

// Name returns the name of this database adapter
func (a *Adapter) Name() string { return "gorm" }

// NewDialectStrategy creates a new GORM dialect strategy
func NewDialectStrategy(driver string) (*DialectStrategy, error) {
	driverParts := strings.Split(driver, ":")
	if len(driverParts) != 2 {
		return nil, fmt.Errorf("invalid gorm driver format: %s (expected 'gorm:dialect')", driver)
	}

	return &DialectStrategy{dialectName: driverParts[1]}, nil
}

// CreateDialector creates a GORM dialector for the specific database type
func (g *DialectStrategy) CreateDialector(dsn string) (interface{}, error) {
	switch g.dialectName {
	case DriverMySQL:
		return mysql.Open(dsn), nil
	case DriverPostgres:
		return postgres.Open(dsn), nil
	case DriverSQLite:
		return sqlite.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported gorm dialect: %s", g.dialectName)
	}
}

// ValidateDriver validates the driver format
func (g *DialectStrategy) ValidateDriver(driver string) error {
	driverParts := strings.Split(driver, ":")
	if len(driverParts) != 2 {
		return fmt.Errorf("invalid gorm driver format: %s (expected 'gorm:dialect')", driver)
	}

	dialectName := driverParts[1]
	switch dialectName {
	case DriverMySQL, DriverPostgres, DriverSQLite:
		return nil
	default:
		return fmt.Errorf("unsupported gorm dialect: %s", dialectName)
	}
}

// GetDriverName returns the driver name
func (g *DialectStrategy) GetDriverName() string {
	return g.dialectName
}

// New creates a new GORM database connection (*gorm.DB) and optionally registers
// GORM plugins (e.g., OpenTelemetry tracing) on the returned instance.
//
// This function enables observability by allowing callers to pass any number of
// GORM plugins that implement gorm.Plugin. Each plugin is registered via db.Use
// after the base connection is established.
//
// Example (OpenTelemetry tracing):
//
//	package main
//
//	import (
//	  "log"
//
//	  gormadapter "github.com/next-trace/scg-database/adapter/gorm"
//	  "github.com/next-trace/scg-database/config"
//	  oteltracing "gorm.io/plugin/opentelemetry/tracing"
//	)
//
//	func main() {
//	  cfg := &config.Config{
//	    Driver: gormadapter.GormDriverPostgres,
//	    DSN:    "postgres://user:pass@localhost:5432/app?sslmode=disable",
//	  }
//
//	  db, err := gormadapter.New(cfg, oteltracing.New())
//	  if err != nil {
//	    log.Fatal(err)
//	  }
//
//	  // use db as *gorm.DB ...
//	  _ = db
//	}
//
// Parameters:
//   - cfg: Standard scg-database Config containing driver and DSN.
//   - plugins: Optional list of gorm.Plugin to register (e.g., tracing, metrics).
//
// Returns:
//   - *gorm.DB with all plugins registered, or an error.
func New(cfg *config.Config, plugins ...gorm.Plugin) (*gorm.DB, error) {
	// Create dialect strategy
	dialectStrategy, err := NewDialectStrategy(cfg.Driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create dialect strategy: %w", err)
	}

	// Create dialector using strategy
	dialectorInterface, err := dialectStrategy.CreateDialector(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create dialector: %w", err)
	}

	// Type assert to GORM dialector
	dialector, ok := dialectorInterface.(gorm.Dialector)
	if !ok {
		return nil, fmt.Errorf("invalid dialector type")
	}

	// Extract GORM config using helper
	gormConfig := extractGormConfig(cfg)

	// Create GORM database connection
	gdb, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("gorm connection failed: %w", err)
	}

	// Register provided plugins
	for _, plugin := range plugins {
		if plugin == nil {
			continue
		}
		if err := gdb.Use(plugin); err != nil {
			return nil, err
		}
	}

	return gdb, nil
}

// Connect establishes a new database connection using GORM and wraps it into the
// library's contract.Connection. This maintains backward compatibility for
// existing consumers of the adapter while internally using New for creation.
func (a *Adapter) Connect(cfg *config.Config) (contract.Connection, error) {
	// Create GORM database connection (no plugins by default)
	gormDB, err := New(cfg)
	if err != nil {
		return nil, err
	}

	// Apply connection pool settings
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Apply connection pool options
	poolOptions := configFromOptions(cfg)
	applyConnectionPoolOptions(sqlDB, poolOptions...)

	return &connection{db: gormDB, config: *cfg}, nil
}
