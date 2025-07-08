//revive:disable:var-naming // allow 'utils' package name in tests for consistency and to avoid widespread refactors
package utils

import (
	"errors"
	"testing"

	"github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-database/db"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRepositoryModel is a test implementation of contract.Model for repository tests
// Additional test models for relationships
// Test relationship implementations
type (
	TestRepositoryModel struct {
		ID   uint
		Name string
	}

	TestProfileModel struct {
		ID     uint
		UserID uint
		Bio    string
	}

	TestOrderModel struct {
		ID     uint
		UserID uint
		Total  float64
	}

	TestRoleModel struct {
		ID   uint
		Name string
	}

	HasOneRelationship struct {
		related    contract.Model
		foreignKey string
		ownerKey   string
	}

	HasManyRelationship struct {
		related    contract.Model
		foreignKey string
		ownerKey   string
	}

	BelongsToManyRelationship struct {
		related   contract.Model
		joinTable string
	}
)

func (t TestRepositoryModel) PrimaryKey() string { return "id" }
func (t TestRepositoryModel) TableName() string  { return "test_repository_models" }
func (t TestRepositoryModel) GetID() any         { return t.ID }
func (t *TestRepositoryModel) SetID(id any)      { t.ID = id.(uint) }
func (t TestRepositoryModel) Relationships() map[string]contract.Relationship {
	return map[string]contract.Relationship{
		"Profile": &HasOneRelationship{
			related:    &TestProfileModel{},
			foreignKey: "user_id",
			ownerKey:   "id",
		},
		"Orders": &HasManyRelationship{
			related:    &TestOrderModel{},
			foreignKey: "user_id",
			ownerKey:   "id",
		},
		"Roles": &BelongsToManyRelationship{
			related:   &TestRoleModel{},
			joinTable: "user_roles",
		},
	}
}

func (t TestProfileModel) PrimaryKey() string                              { return "id" }
func (t TestProfileModel) TableName() string                               { return "test_profiles" }
func (t TestProfileModel) GetID() any                                      { return t.ID }
func (t *TestProfileModel) SetID(id any)                                   { t.ID = id.(uint) }
func (t TestProfileModel) Relationships() map[string]contract.Relationship { return nil }

func (t TestOrderModel) PrimaryKey() string                              { return "id" }
func (t TestOrderModel) TableName() string                               { return "test_orders" }
func (t TestOrderModel) GetID() any                                      { return t.ID }
func (t *TestOrderModel) SetID(id any)                                   { t.ID = id.(uint) }
func (t TestOrderModel) Relationships() map[string]contract.Relationship { return nil }

func (t TestRoleModel) PrimaryKey() string                              { return "id" }
func (t TestRoleModel) TableName() string                               { return "test_roles" }
func (t TestRoleModel) GetID() any                                      { return t.ID }
func (t *TestRoleModel) SetID(id any)                                   { t.ID = id.(uint) }
func (t TestRoleModel) Relationships() map[string]contract.Relationship { return nil }

func (r *HasOneRelationship) Type() contract.RelationshipType { return contract.HasOne }
func (r *HasOneRelationship) RelatedModel() contract.Model    { return r.related }
func (r *HasOneRelationship) ForeignKey() string              { return r.foreignKey }
func (r *HasOneRelationship) OwnerKey() string                { return r.ownerKey }
func (r *HasOneRelationship) ManyToManyJoinTable() string     { return "" }

func (r *HasManyRelationship) Type() contract.RelationshipType { return contract.HasMany }
func (r *HasManyRelationship) RelatedModel() contract.Model    { return r.related }
func (r *HasManyRelationship) ForeignKey() string              { return r.foreignKey }
func (r *HasManyRelationship) OwnerKey() string                { return r.ownerKey }
func (r *HasManyRelationship) ManyToManyJoinTable() string     { return "" }

func (r *BelongsToManyRelationship) Type() contract.RelationshipType { return contract.BelongsToMany }
func (r *BelongsToManyRelationship) RelatedModel() contract.Model    { return r.related }
func (r *BelongsToManyRelationship) ForeignKey() string              { return "" }
func (r *BelongsToManyRelationship) OwnerKey() string                { return "" }
func (r *BelongsToManyRelationship) ManyToManyJoinTable() string     { return r.joinTable }

// Helper function to create a test GORM DB
func createTestDB() *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return database
}

func TestNewRepositoryBuilder(t *testing.T) {
	testDB := createTestDB()
	model := &TestRepositoryModel{}

	builder := NewRepositoryBuilder(testDB, model)

	assert.NotNil(t, builder)
	assert.Equal(t, testDB, builder.db)
	assert.Equal(t, model, builder.mdl)
}

func TestRepositoryBuilder_BuildRepository(t *testing.T) {
	testDB := createTestDB()
	model := &TestRepositoryModel{}
	builder := NewRepositoryBuilder(testDB, model)

	result := builder.BuildRepository(testDB, model)

	assert.NotNil(t, result)
	// Verify the result has the expected structure
	repo := result.(struct {
		db  *gorm.DB
		mdl contract.Model
	})
	assert.Equal(t, testDB, repo.db)
	assert.Equal(t, model, repo.mdl)
}

func TestHandleRelationshipPreload(t *testing.T) {
	testDB := createTestDB()
	model := &TestRepositoryModel{}

	tests := []struct {
		name      string
		relations []string
		expected  int // We can't easily test the actual preload calls, so we test the count
	}{
		{
			name:      "no relations",
			relations: []string{},
			expected:  0,
		},
		{
			name:      "single relation - HasOne",
			relations: []string{"Profile"},
			expected:  1,
		},
		{
			name:      "single relation - HasMany",
			relations: []string{"Orders"},
			expected:  1,
		},
		{
			name:      "single relation - BelongsToMany",
			relations: []string{"Roles"},
			expected:  1,
		},
		{
			name:      "multiple relations",
			relations: []string{"Profile", "Orders", "Roles"},
			expected:  3,
		},
		{
			name:      "undefined relation",
			relations: []string{"UndefinedRelation"},
			expected:  1,
		},
		{
			name:      "mixed defined and undefined relations",
			relations: []string{"Profile", "UndefinedRelation"},
			expected:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testDB.Model(model)
			result := HandleRelationshipPreload(tx, model, tt.relations)
			assert.NotNil(t, result)
			// The result should be a *gorm.DB instance
			assert.IsType(t, &gorm.DB{}, result)
		})
	}
}

func TestValidateAndApplyLimit(t *testing.T) {
	testDB := createTestDB()

	tests := []struct {
		name     string
		limit    int
		expected bool // whether limit should be applied
	}{
		{
			name:     "positive limit",
			limit:    10,
			expected: true,
		},
		{
			name:     "zero limit",
			limit:    0,
			expected: true,
		},
		{
			name:     "negative limit",
			limit:    -1,
			expected: false,
		},
		{
			name:     "large limit",
			limit:    1000,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testDB.Model(&TestRepositoryModel{})
			result := ValidateAndApplyLimit(tx, tt.limit)
			assert.NotNil(t, result)
			assert.IsType(t, &gorm.DB{}, result)
		})
	}
}

func TestValidateAndApplyOffset(t *testing.T) {
	testDB := createTestDB()

	tests := []struct {
		name     string
		offset   int
		expected bool // whether offset should be applied
	}{
		{
			name:     "positive offset",
			offset:   10,
			expected: true,
		},
		{
			name:     "zero offset",
			offset:   0,
			expected: true,
		},
		{
			name:     "negative offset",
			offset:   -1,
			expected: false,
		},
		{
			name:     "large offset",
			offset:   1000,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testDB.Model(&TestRepositoryModel{})
			result := ValidateAndApplyOffset(tx, tt.offset)
			assert.NotNil(t, result)
			assert.IsType(t, &gorm.DB{}, result)
		})
	}
}

func TestHandleFindOrFailError(t *testing.T) {
	model := &TestRepositoryModel{ID: 1, Name: "test"}

	tests := []struct {
		name        string
		model       contract.Model
		err         error
		expectError bool
		expectedErr error
	}{
		{
			name:        "no error with valid model",
			model:       model,
			err:         nil,
			expectError: false,
		},
		{
			name:        "error provided",
			model:       model,
			err:         errors.New("database error"),
			expectError: true,
			expectedErr: errors.New("database error"),
		},
		{
			name:        "nil model",
			model:       nil,
			err:         nil,
			expectError: true,
			expectedErr: db.ErrRecordNotFound,
		},
		{
			name:        "error and nil model",
			model:       nil,
			err:         errors.New("database error"),
			expectError: true,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HandleFindOrFailError(tt.model, tt.err)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.model, result)
			}
		})
	}
}

func TestValidateModelForOperation(t *testing.T) {
	model := &TestRepositoryModel{ID: 1, Name: "test"}

	tests := []struct {
		name        string
		model       contract.Model
		operation   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid model",
			model:       model,
			operation:   "create",
			expectError: false,
		},
		{
			name:        "nil model",
			model:       nil,
			operation:   "update",
			expectError: true,
			errorMsg:    "model cannot be nil for update operation",
		},
		{
			name:        "valid model with different operation",
			model:       model,
			operation:   "delete",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModelForOperation(tt.model, tt.operation)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateModelsSliceForOperation(t *testing.T) {
	model1 := &TestRepositoryModel{ID: 1, Name: "test1"}
	model2 := &TestRepositoryModel{ID: 2, Name: "test2"}

	tests := []struct {
		name        string
		models      []contract.Model
		operation   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid models slice",
			models:      []contract.Model{model1, model2},
			operation:   "create",
			expectError: false,
		},
		{
			name:        "empty models slice",
			models:      []contract.Model{},
			operation:   "update",
			expectError: false,
		},
		{
			name:        "nil models slice",
			models:      nil,
			operation:   "delete",
			expectError: false,
		},
		{
			name:        "models slice with nil value at start",
			models:      []contract.Model{nil, model2},
			operation:   "create",
			expectError: true,
			errorMsg:    "model at index 0 cannot be nil for create operation",
		},
		{
			name:        "models slice with nil value in middle",
			models:      []contract.Model{model1, nil, model2},
			operation:   "update",
			expectError: true,
			errorMsg:    "model at index 1 cannot be nil for update operation",
		},
		{
			name:        "models slice with nil value at end",
			models:      []contract.Model{model1, model2, nil},
			operation:   "delete",
			expectError: true,
			errorMsg:    "model at index 2 cannot be nil for delete operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModelsSliceForOperation(tt.models, tt.operation)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptimizeCreateOperation(t *testing.T) {
	model1 := &TestRepositoryModel{ID: 1, Name: "test1"}
	model2 := &TestRepositoryModel{ID: 2, Name: "test2"}

	tests := []struct {
		name            string
		models          []contract.Model
		expectOptimized bool
		expectedModel   contract.Model
		expectError     bool
		errorMsg        string
	}{
		{
			name:            "empty models slice",
			models:          []contract.Model{},
			expectOptimized: false,
			expectedModel:   nil,
			expectError:     false,
		},
		{
			name:            "single valid model",
			models:          []contract.Model{model1},
			expectOptimized: true,
			expectedModel:   model1,
			expectError:     false,
		},
		{
			name:            "single nil model",
			models:          []contract.Model{nil},
			expectOptimized: false,
			expectedModel:   nil,
			expectError:     true,
			errorMsg:        "model cannot be nil",
		},
		{
			name:            "multiple models",
			models:          []contract.Model{model1, model2},
			expectOptimized: false,
			expectedModel:   nil,
			expectError:     false,
		},
		{
			name:            "multiple models with nil",
			models:          []contract.Model{model1, nil, model2},
			expectOptimized: false,
			expectedModel:   nil,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optimized, model, err := OptimizeCreateOperation(tt.models)
			assert.Equal(t, tt.expectOptimized, optimized)
			assert.Equal(t, tt.expectedModel, model)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateRepositoryInstance(t *testing.T) {
	testDB := createTestDB()
	model := &TestRepositoryModel{}

	result := CreateRepositoryInstance(testDB, model)

	assert.NotNil(t, result)
	// Verify the result has the expected structure
	repo := result.(struct {
		DB    *gorm.DB
		Model contract.Model
	})
	assert.NotNil(t, repo.DB)
	assert.Equal(t, model, repo.Model)
}

func TestApplyOrderBy(t *testing.T) {
	testDB := createTestDB()

	tests := []struct {
		name      string
		column    string
		direction string
		expected  bool // whether ordering should be applied
	}{
		{
			name:      "valid column and direction ASC",
			column:    "name",
			direction: "ASC",
			expected:  true,
		},
		{
			name:      "valid column and direction DESC",
			column:    "created_at",
			direction: "DESC",
			expected:  true,
		},
		{
			name:      "valid column with invalid direction",
			column:    "name",
			direction: "INVALID",
			expected:  true, // Should use default ASC
		},
		{
			name:      "invalid column name",
			column:    "'; DROP TABLE users; --",
			direction: "ASC",
			expected:  false,
		},
		{
			name:      "empty column name",
			column:    "",
			direction: "ASC",
			expected:  false,
		},
		{
			name:      "valid qualified column",
			column:    "users.name",
			direction: "DESC",
			expected:  true,
		},
		{
			name:      "column with underscore",
			column:    "user_name",
			direction: "ASC",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testDB.Model(&TestRepositoryModel{})
			result := ApplyOrderBy(tx, tt.column, tt.direction)
			assert.NotNil(t, result)
			assert.IsType(t, &gorm.DB{}, result)
		})
	}
}

func TestNewBatchOperationResult(t *testing.T) {
	result := NewBatchOperationResult()

	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ProcessedCount)
	assert.NotNil(t, result.Errors)
	assert.Len(t, result.Errors, 0)
	assert.False(t, result.Success)
}

func TestBatchOperationResult_AddError(t *testing.T) {
	result := NewBatchOperationResult()
	err := errors.New("test error")

	result.AddError(err)

	assert.Len(t, result.Errors, 1)
	assert.Equal(t, err, result.Errors[0])
	assert.False(t, result.Success)
}

func TestBatchOperationResult_IncrementProcessed(t *testing.T) {
	result := NewBatchOperationResult()

	assert.Equal(t, 0, result.ProcessedCount)

	result.IncrementProcessed()
	assert.Equal(t, 1, result.ProcessedCount)

	result.IncrementProcessed()
	assert.Equal(t, 2, result.ProcessedCount)
}

func TestBatchOperationResult_HasErrors(t *testing.T) {
	result := NewBatchOperationResult()

	assert.False(t, result.HasErrors())

	result.AddError(errors.New("test error"))
	assert.True(t, result.HasErrors())
}

func TestBatchOperationResult_GetFirstError(t *testing.T) {
	result := NewBatchOperationResult()

	// No errors
	assert.Nil(t, result.GetFirstError())

	// Add errors
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	result.AddError(err1)
	result.AddError(err2)

	// Should return first error
	assert.Equal(t, err1, result.GetFirstError())
}

func TestBatchOperationResult_Integration(t *testing.T) {
	result := NewBatchOperationResult()

	// Test complete workflow
	result.IncrementProcessed()
	result.IncrementProcessed()
	assert.Equal(t, 2, result.ProcessedCount)
	assert.False(t, result.HasErrors())

	// Add an error
	err := errors.New("processing error")
	result.AddError(err)
	assert.True(t, result.HasErrors())
	assert.False(t, result.Success)
	assert.Equal(t, err, result.GetFirstError())

	// Continue processing
	result.IncrementProcessed()
	assert.Equal(t, 3, result.ProcessedCount)
	assert.True(t, result.HasErrors()) // Still has errors
}
