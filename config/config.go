package config

import (
	"errors"
	"time"
)

type (
	// Config is the single, unified configuration struct for the entire database package.
	Config struct {
		// Driver is the name of the database driver to use (e.g., "gorm:mysql").
		Driver string
		// DSN is the connection string for the database.
		DSN string
		// MigrationsPath is the source URL for migrations ("file://path/to/migrations").
		MigrationsPath string
		// ModelsPath is the directory for generated models.
		ModelsPath string

		// connection pool settings
		MaxIdleConns    int
		MaxOpenConns    int
		ConnMaxLifetime time.Duration

		// Adapter is an optional field to explicitly inject a database adapter, bypassing the registry.
		Adapter any

		// Settings is a generic map to hold adapter-specific configurations
		// that are applied via functional options.
		Settings map[string]any
	}

	// Option defines a functional option for modifying a Config.
	Option func(*Config)
)

// New creates a new Config with defaults and initializes the Settings map.
func New() *Config {
	return &Config{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ModelsPath:      "domain",
		Settings:        make(map[string]any),
	}
}

// Validate checks if the essential configuration fields are set.
func (c *Config) Validate() error {
	if c.Driver == "" {
		return errors.New("database driver is required")
	}
	if c.DSN == "" {
		return errors.New("database dsn is required")
	}
	return nil
}
