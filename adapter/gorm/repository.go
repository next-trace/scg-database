package gorm

import (
	"context"
	"errors"
	"fmt"

	"github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-database/db"
	"gorm.io/gorm"
)

type (
	repository struct {
		db  *gorm.DB
		mdl contract.Model
	}
)

//nolint:grouper // Only One Global Variable
var _ contract.Repository = (*repository)(nil)

func newGormRepository(database *gorm.DB, mdl contract.Model) contract.Repository {
	return &repository{db: database.Model(mdl), mdl: mdl}
}

// --- Query Building ---
func (r *repository) With(relations ...string) contract.Repository {
	tx := handleRelationshipPreload(r.db, r.mdl, relations)
	return &repository{db: tx, mdl: r.mdl}
}

func (r *repository) Where(query any, args ...any) contract.Repository {
	return &repository{db: r.db.Where(query, args...), mdl: r.mdl}
}

func (r *repository) Unscoped() contract.Repository {
	return &repository{db: r.db.Unscoped(), mdl: r.mdl}
}

func (r *repository) Limit(limit int) contract.Repository {
	tx := validateAndApplyLimit(r.db, limit)
	return &repository{db: tx, mdl: r.mdl}
}

func (r *repository) Offset(offset int) contract.Repository {
	tx := validateAndApplyOffset(r.db, offset)
	return &repository{db: tx, mdl: r.mdl}
}

func (r *repository) OrderBy(column, direction string) contract.Repository {
	tx := applyOrderBy(r.db, column, direction)
	return &repository{db: tx, mdl: r.mdl}
}

// --- Helper Functions ---

// convertModelsToSlice converts a slice of contract.Model to a concrete slice for GORM operations
func (r *repository) convertModelsToSlice(models []contract.Model) (interface{}, error) {
	modelType := getModelType(r.mdl)
	return convertModelsToSlice(models, modelType)
}

// --- Read Operations ---
func (r *repository) Find(ctx context.Context, id any) (contract.Model, error) {
	entity, err := createEntityFromModel(r.mdl)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity from model: %w", err)
	}
	err = r.db.WithContext(ctx).First(entity, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return entity, err
}

func (r *repository) FindOrFail(ctx context.Context, id any) (contract.Model, error) {
	model, err := r.Find(ctx, id)
	return handleFindOrFailError(model, err)
}

func (r *repository) First(ctx context.Context) (contract.Model, error) {
	entity, err := createEntityFromModel(r.mdl)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity from model: %w", err)
	}
	err = r.db.WithContext(ctx).First(entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return entity, err
}

func (r *repository) FirstOrFail(ctx context.Context) (contract.Model, error) {
	model, err := r.First(ctx)
	return handleFindOrFailError(model, err)
}

func (r *repository) Get(ctx context.Context) ([]contract.Model, error) {
	// Use reflection helper with dependency injection
	queryExecutor := func(dest interface{}) error {
		return r.db.WithContext(ctx).Find(dest).Error
	}

	return executeQueryAndConvertToModels(r.mdl, queryExecutor)
}

func (r *repository) Pluck(ctx context.Context, column string, dest any) error {
	return r.db.WithContext(ctx).Pluck(column, dest).Error
}

// --- Write Operations ---
func (r *repository) Create(ctx context.Context, models ...contract.Model) error {
	// Use helper to optimize create operation
	useSingleOptimization, singleModel, err := optimizeCreateOperation(models)
	if err != nil {
		return err
	}

	// Handle empty slice case
	if len(models) == 0 {
		return nil
	}

	// Single model optimization
	if useSingleOptimization {
		return r.db.WithContext(ctx).Create(singleModel).Error
	}

	// Multiple models - use helper function
	slice, err := r.convertModelsToSlice(models)
	if err != nil {
		return fmt.Errorf("failed to convert models: %w", err)
	}

	return r.db.WithContext(ctx).Create(slice).Error
}

func (r *repository) CreateInBatches(ctx context.Context, models []contract.Model, batchSize int) error {
	if len(models) == 0 {
		return nil
	}
	if batchSize <= 0 {
		return errors.New("batch size must be positive")
	}

	// Convert interface slice to concrete slice for GORM
	slice, err := r.convertModelsToSlice(models)
	if err != nil {
		return fmt.Errorf("failed to convert models: %w", err)
	}

	return r.db.WithContext(ctx).CreateInBatches(slice, batchSize).Error
}

func (r *repository) Update(ctx context.Context, models ...contract.Model) error {
	if len(models) == 0 {
		return nil
	}

	// Single model optimization
	if len(models) == 1 {
		if models[0] == nil {
			return errors.New("model cannot be nil")
		}
		model := models[0]
		// Use Updates with WHERE condition based on primary key
		return r.db.WithContext(ctx).Model(model).Where(model.PrimaryKey()+" = ?", model.GetID()).Updates(model).Error
	}

	// For multiple models, update each one individually
	for _, model := range models {
		if model == nil {
			return errors.New("model cannot be nil")
		}
		err := r.db.WithContext(ctx).Model(model).
			Where(model.PrimaryKey()+" = ?", model.GetID()).
			Updates(model).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, models ...contract.Model) error {
	if len(models) == 0 {
		return nil
	}

	// Single model optimization
	if len(models) == 1 {
		if models[0] == nil {
			return errors.New("model cannot be nil")
		}
		return r.db.WithContext(ctx).Delete(models[0]).Error
	}

	// Multiple models - use helper function
	slice, err := r.convertModelsToSlice(models)
	if err != nil {
		return fmt.Errorf("failed to convert models: %w", err)
	}

	return r.db.WithContext(ctx).Delete(slice).Error
}

func (r *repository) ForceDelete(ctx context.Context, models ...contract.Model) error {
	if len(models) == 0 {
		return nil
	}

	// Single model optimization
	if len(models) == 1 {
		if models[0] == nil {
			return errors.New("model cannot be nil")
		}
		return r.db.WithContext(ctx).Unscoped().Delete(models[0]).Error
	}

	// Multiple models - use helper function
	slice, err := r.convertModelsToSlice(models)
	if err != nil {
		return fmt.Errorf("failed to convert models: %w", err)
	}

	return r.db.WithContext(ctx).Unscoped().Delete(slice).Error
}

// --- Upsert Operations ---
func (r *repository) FirstOrCreate(
	ctx context.Context,
	condition contract.Model,
	create ...contract.Model,
) (contract.Model, error) {
	var toCreate contract.Model
	if len(create) > 0 {
		toCreate = create[0]
	} else {
		toCreate = condition
	}
	// GORM mutates the first argument, which must be the condition
	err := r.db.WithContext(ctx).Where(condition).FirstOrCreate(toCreate).Error
	return toCreate, err
}

func (r *repository) UpdateOrCreate(ctx context.Context, condition contract.Model, values any) (contract.Model, error) {
	// GORM mutates the condition model in this case
	err := r.db.WithContext(ctx).Where(condition).Assign(values).FirstOrCreate(condition).Error
	return condition, err
}

// QueryBuilder provides access to the fluent query builder interface
func (r *repository) QueryBuilder() contract.QueryBuilder {
	factory, err := db.GetQueryBuilderFactory("gorm")
	if err != nil {
		// Fallback to creating a basic query builder if factory is not found
		// This should not happen in normal circumstances since GORM registers its factory
		panic(fmt.Sprintf("GORM query builder factory not found: %v", err))
	}

	return factory.NewQueryBuilder(r.mdl, r.db)
}
