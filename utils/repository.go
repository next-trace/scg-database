// Package utils provides utility functions for database configuration, connection pooling,
// and other common operations used throughout the SCG database toolkit.
//
//revive:disable:var-naming // allow package name 'utils' for this utilities module
package utils

import (
	"errors"
	"fmt"

	"github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-database/db"
	"gorm.io/gorm"
)

type (
	// RepositoryBuilder helps create repository instances with common patterns
	RepositoryBuilder struct {
		db  *gorm.DB
		mdl contract.Model
	}

	// BatchOperationResult represents the result of a batch operation
	BatchOperationResult struct {
		ProcessedCount int
		Errors         []error
		Success        bool
	}
)

// NewRepositoryBuilder creates a new repository builder
func NewRepositoryBuilder(gdb *gorm.DB, model contract.Model) *RepositoryBuilder {
	return &RepositoryBuilder{
		db:  gdb,
		mdl: model,
	}
}

// BuildRepository creates a new repository instance
func (rb *RepositoryBuilder) BuildRepository(database *gorm.DB, model contract.Model) interface{} {
	// This would return a repository struct, but we need to define the interface
	// For now, return a generic interface
	return struct {
		db  *gorm.DB
		mdl contract.Model
	}{db: database, mdl: model}
}

// HandleRelationshipPreload handles relationship preloading logic
func HandleRelationshipPreload(tx *gorm.DB, model contract.Model, relations []string) *gorm.DB {
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

// ValidateAndApplyLimit validates and applies limit with bounds checking
func ValidateAndApplyLimit(tx *gorm.DB, limit int) *gorm.DB {
	if limit < 0 {
		// Return repository without limit for negative values
		return tx
	}
	return tx.Limit(limit)
}

// ValidateAndApplyOffset validates and applies offset with bounds checking
func ValidateAndApplyOffset(tx *gorm.DB, offset int) *gorm.DB {
	if offset < 0 {
		// Return repository without offset for negative values
		return tx
	}
	return tx.Offset(offset)
}

// HandleFindOrFailError handles the common pattern of Find followed by error checking
func HandleFindOrFailError(model contract.Model, err error) (contract.Model, error) {
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, db.ErrRecordNotFound
	}
	return model, nil
}

// ValidateModelForOperation validates a model before performing operations
func ValidateModelForOperation(model contract.Model, operation string) error {
	if model == nil {
		return fmt.Errorf("model cannot be nil for %s operation", operation)
	}
	return nil
}

// ValidateModelsSliceForOperation validates a slice of models before operations
func ValidateModelsSliceForOperation(models []contract.Model, operation string) error {
	if len(models) == 0 {
		return nil // Empty slice is valid for some operations
	}

	for i, model := range models {
		if model == nil {
			return fmt.Errorf("model at index %d cannot be nil for %s operation", i, operation)
		}
	}

	return nil
}

// OptimizeCreateOperation optimizes create operations based on the number of models
func OptimizeCreateOperation(models []contract.Model) (bool, contract.Model, error) {
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

// CreateRepositoryInstance creates a repository instance with the given database and model
func CreateRepositoryInstance(database *gorm.DB, model contract.Model) interface{} {
	return struct {
		DB    *gorm.DB
		Model contract.Model
	}{
		DB:    database.Model(model),
		Model: model,
	}
}

// ApplyOrderBy applies ordering with validation
func ApplyOrderBy(tx *gorm.DB, column, direction string) *gorm.DB {
	// Validate column name to prevent SQL injection
	if !ValidateColumnName(column) {
		// For invalid columns, just return the repository without ordering
		return tx
	}

	// Validate and normalize direction
	normalizedDirection := ValidateOrderDirection(direction, "ASC")

	return tx.Order(fmt.Sprintf("%s %s", column, normalizedDirection))
}

// NewBatchOperationResult creates a new batch operation result
func NewBatchOperationResult() *BatchOperationResult {
	return &BatchOperationResult{
		Errors: make([]error, 0),
	}
}

// AddError adds an error to the batch result
func (r *BatchOperationResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
	r.Success = false
}

// IncrementProcessed increments the processed count
func (r *BatchOperationResult) IncrementProcessed() {
	r.ProcessedCount++
}

// HasErrors checks if there are any errors
func (r *BatchOperationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// GetFirstError returns the first error if any
func (r *BatchOperationResult) GetFirstError() error {
	if len(r.Errors) > 0 {
		return r.Errors[0]
	}
	return nil
}
