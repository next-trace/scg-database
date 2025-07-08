// Package utils provides utility functions for database configuration, connection pooling,
// and other common operations used throughout the SCG database toolkit.
//
//revive:disable:var-naming // allow package name 'utils' for this utilities module
package utils

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/next-trace/scg-database/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConfigureConnectionPool applies connection pool settings to a sql.DB instance
func ConfigureConnectionPool(sqlDB *sql.DB, cfg *config.Config) {
	const defaultConnMaxLifetime = 10 * time.Second

	// Set connection max lifetime
	lifetime := defaultConnMaxLifetime
	if cfg.ConnMaxLifetime > 0 {
		lifetime = cfg.ConnMaxLifetime
	}
	sqlDB.SetConnMaxLifetime(lifetime)

	// Set max idle connections
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	// Set max open connections
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
}

// ExtractGormConfig extracts GORM configuration from settings
func ExtractGormConfig(cfg *config.Config) *gorm.Config {
	gormConfig := &gorm.Config{}

	// Extract gorm_config if present
	if c, ok := cfg.Settings["gorm_config"].(*gorm.Config); ok {
		gormConfig = c
	}

	// Extract gorm_logger if present
	if l, ok := cfg.Settings["gorm_logger"].(logger.Interface); ok {
		gormConfig.Logger = l
	}

	return gormConfig
}

type (
	// ConnectionPoolOptions represents connection pool configuration options
	ConnectionPoolOptions struct {
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
	}

	// ConnectionPoolOption is a functional option for connection pool configuration
	ConnectionPoolOption func(*ConnectionPoolOptions)
)

// ApplyConnectionPoolOptions applies connection pool options using functional options pattern
func ApplyConnectionPoolOptions(sqlDB *sql.DB, options ...ConnectionPoolOption) {
	opts := &ConnectionPoolOptions{
		ConnMaxLifetime: 10 * time.Second, // default
	}

	for _, option := range options {
		option(opts)
	}

	if opts.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}
	if opts.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	}
}

// WithMaxOpenConns sets the maximum number of open connections
func WithMaxOpenConns(maxConns int) ConnectionPoolOption {
	return func(opts *ConnectionPoolOptions) {
		opts.MaxOpenConns = maxConns
	}
}

// WithMaxIdleConns sets the maximum number of idle connections
func WithMaxIdleConns(maxConns int) ConnectionPoolOption {
	return func(opts *ConnectionPoolOptions) {
		opts.MaxIdleConns = maxConns
	}
}

// WithConnMaxLifetime sets the maximum connection lifetime
func WithConnMaxLifetime(lifetime time.Duration) ConnectionPoolOption {
	return func(opts *ConnectionPoolOptions) {
		opts.ConnMaxLifetime = lifetime
	}
}

// ConfigFromOptions creates connection pool options from a config
func ConfigFromOptions(cfg *config.Config) []ConnectionPoolOption {
	var options []ConnectionPoolOption

	if cfg.MaxOpenConns > 0 {
		options = append(options, WithMaxOpenConns(cfg.MaxOpenConns))
	}
	if cfg.MaxIdleConns > 0 {
		options = append(options, WithMaxIdleConns(cfg.MaxIdleConns))
	}
	if cfg.ConnMaxLifetime > 0 {
		options = append(options, WithConnMaxLifetime(cfg.ConnMaxLifetime))
	}

	return options
}

type (
	// ConnectionBuilder builds database connections using functional options
	ConnectionBuilder struct {
		config          config.Config
		dialectStrategy DialectStrategy
		poolOptions     []ConnectionPoolOption
	}

	// DialectStrategy defines the interface for database dialect strategies
	DialectStrategy interface {
		CreateDialector(string) (interface{}, error)
		ValidateDriver(string) error
		GetDriverName() string
	}

	// ConnectionBuilderOption defines functional options for ConnectionBuilder
	ConnectionBuilderOption func(*ConnectionBuilder)
)

// NewConnectionBuilder creates a new connection builder with functional options
func NewConnectionBuilder(cfg *config.Config, options ...ConnectionBuilderOption) *ConnectionBuilder {
	builder := &ConnectionBuilder{
		config:      *cfg,
		poolOptions: ConfigFromOptions(cfg),
	}

	for _, option := range options {
		option(builder)
	}

	return builder
}

// WithDialectStrategy sets the dialect strategy
func WithDialectStrategy(strategy DialectStrategy) ConnectionBuilderOption {
	return func(b *ConnectionBuilder) {
		b.dialectStrategy = strategy
	}
}

// WithPoolOptions sets custom pool options
func WithPoolOptions(options ...ConnectionPoolOption) ConnectionBuilderOption {
	return func(b *ConnectionBuilder) {
		b.poolOptions = append(b.poolOptions, options...)
	}
}

// Build creates the database connection using the configured options
func (b *ConnectionBuilder) Build() (interface{}, *sql.DB, error) {
	if b.dialectStrategy == nil {
		return nil, nil, fmt.Errorf("dialect strategy is required")
	}

	// Validate driver
	if err := b.dialectStrategy.ValidateDriver(b.config.Driver); err != nil {
		return nil, nil, fmt.Errorf("driver validation failed: %w", err)
	}

	// Create dialector using strategy
	dialector, err := b.dialectStrategy.CreateDialector(b.config.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dialector: %w", err)
	}

	return dialector, nil, nil
}

// ApplyPoolConfiguration applies connection pool configuration to sql.DB
func (b *ConnectionBuilder) ApplyPoolConfiguration(sqlDB *sql.DB) {
	ApplyConnectionPoolOptions(sqlDB, b.poolOptions...)
}
