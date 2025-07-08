// Package utils provides utility functions for database configuration, connection pooling,
// and other common operations used throughout the SCG database toolkit.
//
//revive:disable:var-naming // allow package name 'utils' for this utilities module
package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/next-trace/scg-database/config"
)

// validColumnNameRegex validates that a column name contains only safe characters
var (
	validColumnNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)
)

// ValidateConfigForMigration validates the configuration required for migrations
func ValidateConfigForMigration(cfg *config.Config) error {
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

// ValidateNonNegativeInt validates that an integer is non-negative
func ValidateNonNegativeInt(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
	return nil
}

// ValidatePositiveInt validates that an integer is positive
func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// ValidateColumnName validates that a column name contains only safe characters
func ValidateColumnName(column string) bool {
	return validColumnNameRegex.MatchString(column)
}

// ValidateOrderDirection validates and normalizes order direction
func ValidateOrderDirection(direction, defaultDirection string) string {
	direction = strings.ToUpper(strings.TrimSpace(direction))
	if direction != "ASC" && direction != "DESC" {
		return defaultDirection
	}
	return direction
}

// ValidateModelsSlice validates that a slice of models is not empty and contains no nil values
func ValidateModelsSlice(models []interface{}, operation string) error {
	if len(models) == 0 {
		return fmt.Errorf("models slice cannot be empty for %s operation", operation)
	}

	for i, model := range models {
		if model == nil {
			return fmt.Errorf("model at index %d cannot be nil for %s operation", i, operation)
		}
	}

	return nil
}

// ValidateDriverFormat validates GORM driver format
func ValidateDriverFormat(driver string) (string, error) {
	parts := strings.Split(driver, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid gorm driver format: %s (expected 'gorm:dialect')", driver)
	}
	return parts[1], nil
}
