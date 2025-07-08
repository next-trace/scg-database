package gorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-database/db"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Repository utility functions

// handleRelationshipPreload handles relationship preloading logic
func handleRelationshipPreload(tx *gorm.DB, model contract.Model, relations []string) *gorm.DB {
	if len(relations) == 0 {
		return tx
	}

	modelRelations := model.Relationships()

	for _, rel := range relations {
		if relationship, exists := modelRelations[rel]; exists {
			// Use the relationship metadata to build more intelligent queries
			switch relationship.Type() {
			case contract.HasOne, contract.BelongsTo:
				tx = tx.Preload(rel)
			case contract.HasMany, contract.BelongsToMany:
				tx = tx.Preload(rel)
			case contract.Many2Many:
				// For many-to-many, we might need special handling
				tx = tx.Preload(rel)
			default:
				// Fallback to simple preload
				tx = tx.Preload(rel)
			}
		} else {
			// If relationship is not defined in the model, still allow it
			// but this could be logged as a warning in the future
			tx = tx.Preload(rel)
		}
	}

	return tx
}

// validateAndApplyLimit validates and applies limit with bounds checking
func validateAndApplyLimit(tx *gorm.DB, limit int) *gorm.DB {
	if limit < 0 {
		// Return repository without limit for negative values
		return tx
	}
	return tx.Limit(limit)
}

// validateAndApplyOffset validates and applies offset with bounds checking
func validateAndApplyOffset(tx *gorm.DB, offset int) *gorm.DB {
	if offset < 0 {
		// Return repository without offset for negative values
		return tx
	}
	return tx.Offset(offset)
}

// handleFindOrFailError handles the common pattern of Find followed by error checking
func handleFindOrFailError(model contract.Model, err error) (contract.Model, error) {
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, db.ErrRecordNotFound
	}
	return model, nil
}

// optimizeCreateOperation optimizes create operations based on the number of models
func optimizeCreateOperation(models []contract.Model) (bool, contract.Model, error) {
	if len(models) == 0 {
		return false, nil, nil // No operation needed
	}

	// Single model optimization
	if len(models) == 1 {
		if models[0] == nil {
			return false, nil, errors.New("model cannot be nil")
		}
		return true, models[0], nil // Use single model optimization
	}

	return false, nil, nil // Use batch operation
}

// applyOrderBy applies ordering with validation
func applyOrderBy(tx *gorm.DB, column, direction string) *gorm.DB {
	// Validate column name to prevent SQL injection
	if !validateColumnName(column) {
		// For invalid columns, just return the repository without ordering
		return tx
	}

	// Validate and normalize direction
	normalizedDirection := validateOrderDirection(direction, "ASC")

	return tx.Order(fmt.Sprintf("%s %s", column, normalizedDirection))
}

// Reflection utility functions

// createEntityFromModel creates a new entity instance from a model using reflection
func createEntityFromModel(model contract.Model) (contract.Model, error) {
	newInstance := reflect.New(reflect.TypeOf(model).Elem()).Interface()
	if result, ok := newInstance.(contract.Model); ok {
		return result, nil
	}
	return nil, fmt.Errorf("failed to assert created instance to contract.Model")
}

// convertModelsToSlice converts a slice of contract.Model to a concrete slice for GORM operations
func convertModelsToSlice(models []contract.Model, modelType reflect.Type) (interface{}, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("models slice cannot be empty")
	}

	// Ensure we have the element type, not pointer type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	slice := reflect.MakeSlice(reflect.SliceOf(modelType), len(models), len(models))

	for i, model := range models {
		if model == nil {
			return nil, fmt.Errorf("model at index %d cannot be nil", i)
		}

		modelValue := reflect.ValueOf(model)
		if modelValue.Kind() == reflect.Ptr {
			modelValue = modelValue.Elem()
		}

		if !modelValue.Type().AssignableTo(modelType) {
			return nil, fmt.Errorf("model at index %d is not assignable to expected type %s", i, modelType)
		}

		slice.Index(i).Set(modelValue)
	}

	return slice.Interface(), nil
}

// convertSliceToModels converts a reflected slice back to []contract.Model
func convertSliceToModels(sliceValue reflect.Value) ([]contract.Model, error) {
	models := make([]contract.Model, sliceValue.Len())
	for i := range sliceValue.Len() {
		item := sliceValue.Index(i).Interface()
		if model, ok := item.(contract.Model); ok {
			models[i] = model
		} else {
			return nil, fmt.Errorf("failed to assert item at index %d to contract.Model, got type %T", i, item)
		}
	}
	return models, nil
}

// getModelType returns the reflect.Type of a model, handling pointer types
func getModelType(model contract.Model) reflect.Type {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		return modelType.Elem()
	}
	return modelType
}

// createSliceOfModelType creates a new slice of the given model type
func createSliceOfModelType(modelType reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(reflect.PointerTo(modelType))).Interface()
}

// executeQueryAndConvertToModels executes a query and converts the result to models
func executeQueryAndConvertToModels(
	model contract.Model,
	queryExecutor func(interface{}) error,
) ([]contract.Model, error) {
	modelType := getModelType(model)
	slice := createSliceOfModelType(modelType)

	if err := queryExecutor(slice); err != nil {
		return nil, err
	}

	val := reflect.Indirect(reflect.ValueOf(slice))
	return convertSliceToModels(val)
}

// Config utility functions

// extractGormConfig extracts GORM configuration from settings
func extractGormConfig(cfg *config.Config) *gorm.Config {
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

// applyConnectionPoolOptions applies connection pool options using functional options pattern
func applyConnectionPoolOptions(sqlDB *sql.DB, options ...ConnectionPoolOption) {
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

// withMaxOpenConns sets the maximum number of open connections
func withMaxOpenConns(maxConns int) ConnectionPoolOption {
	return func(opts *ConnectionPoolOptions) {
		opts.MaxOpenConns = maxConns
	}
}

// withMaxIdleConns sets the maximum number of idle connections
func withMaxIdleConns(maxConns int) ConnectionPoolOption {
	return func(opts *ConnectionPoolOptions) {
		opts.MaxIdleConns = maxConns
	}
}

// withConnMaxLifetime sets the maximum connection lifetime
func withConnMaxLifetime(lifetime time.Duration) ConnectionPoolOption {
	return func(opts *ConnectionPoolOptions) {
		opts.ConnMaxLifetime = lifetime
	}
}

// configFromOptions creates connection pool options from a config
func configFromOptions(cfg *config.Config) []ConnectionPoolOption {
	var options []ConnectionPoolOption

	if cfg.MaxOpenConns > 0 {
		options = append(options, withMaxOpenConns(cfg.MaxOpenConns))
	}
	if cfg.MaxIdleConns > 0 {
		options = append(options, withMaxIdleConns(cfg.MaxIdleConns))
	}
	if cfg.ConnMaxLifetime > 0 {
		options = append(options, withConnMaxLifetime(cfg.ConnMaxLifetime))
	}

	return options
}

// ConnectionBuilder and related functions are handled directly in adapter.go
// to avoid type conflicts with DialectStrategy interface

// Validation utility functions

// validateColumnName validates that a column name contains only safe characters
var (
	validColumnNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)
)

func validateColumnName(column string) bool {
	return validColumnNameRegex.MatchString(column)
}

// validateOrderDirection validates and normalizes order direction
func validateOrderDirection(direction, defaultDirection string) string {
	direction = strings.ToUpper(strings.TrimSpace(direction))
	if direction != "ASC" && direction != "DESC" {
		return defaultDirection
	}
	return direction
}
