//revive:disable:var-naming // allow 'utils' package name in tests for consistency and to avoid widespread refactors
package utils

import (
	"errors"
	"reflect"
	"testing"

	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/assert"
)

// TestModel is a test implementation of contract.Model
// AnotherTestModel is another test implementation
// NonModel is a struct that doesn't implement contract.Model
type (
	TestModel struct {
		ID   uint
		Name string
	}

	AnotherTestModel struct {
		ID    uint
		Value string
	}

	NonModel struct {
		ID   uint
		Name string
	}
)

func (t TestModel) PrimaryKey() string                              { return "id" }
func (t TestModel) TableName() string                               { return "test_models" }
func (t TestModel) GetID() any                                      { return t.ID }
func (t *TestModel) SetID(id any)                                   { t.ID = id.(uint) }
func (t TestModel) Relationships() map[string]contract.Relationship { return nil }

func (a AnotherTestModel) PrimaryKey() string                              { return "id" }
func (a AnotherTestModel) TableName() string                               { return "another_test_models" }
func (a AnotherTestModel) GetID() any                                      { return a.ID }
func (a *AnotherTestModel) SetID(id any)                                   { a.ID = id.(uint) }
func (a AnotherTestModel) Relationships() map[string]contract.Relationship { return nil }

func TestCreateEntityFromModel(t *testing.T) {
	tests := []struct {
		name        string
		model       contract.Model
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid model",
			model:       &TestModel{ID: 1, Name: "test"},
			expectError: false,
		},
		{
			name:        "another valid model",
			model:       &AnotherTestModel{ID: 2, Value: "test"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CreateEntityFromModel(tt.model)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.IsType(t, tt.model, result)
				// Verify it's a new instance (different pointer)
				assert.NotSame(t, tt.model, result)
				// Verify it has the same type
				assert.Equal(t, reflect.TypeOf(tt.model), reflect.TypeOf(result))
			}
		})
	}
}

func TestConvertModelsToSlice(t *testing.T) {
	tests := []struct {
		name        string
		models      []contract.Model
		modelType   reflect.Type
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid models",
			models: []contract.Model{
				&TestModel{ID: 1, Name: "test1"},
				&TestModel{ID: 2, Name: "test2"},
			},
			modelType:   reflect.TypeOf(TestModel{}),
			expectError: false,
		},
		{
			name: "valid models with pointer type",
			models: []contract.Model{
				&TestModel{ID: 1, Name: "test1"},
				&TestModel{ID: 2, Name: "test2"},
			},
			modelType:   reflect.TypeOf(&TestModel{}),
			expectError: false,
		},
		{
			name:        "empty models slice",
			models:      []contract.Model{},
			modelType:   reflect.TypeOf(TestModel{}),
			expectError: true,
			errorMsg:    "models slice cannot be empty",
		},
		{
			name: "nil model in slice",
			models: []contract.Model{
				&TestModel{ID: 1, Name: "test1"},
				nil,
			},
			modelType:   reflect.TypeOf(TestModel{}),
			expectError: true,
			errorMsg:    "model at index 1 cannot be nil",
		},
		{
			name: "incompatible model types",
			models: []contract.Model{
				&TestModel{ID: 1, Name: "test1"},
				&AnotherTestModel{ID: 2, Value: "test2"},
			},
			modelType:   reflect.TypeOf(TestModel{}),
			expectError: true,
			errorMsg:    "model at index 1 is not assignable to expected type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertModelsToSlice(tt.models, tt.modelType)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify the result is a slice
				resultValue := reflect.ValueOf(result)
				assert.Equal(t, reflect.Slice, resultValue.Kind())
				assert.Equal(t, len(tt.models), resultValue.Len())
			}
		})
	}
}

func TestConvertSliceToModels(t *testing.T) {
	tests := []struct {
		name        string
		setupSlice  func() reflect.Value
		expectError bool
		errorMsg    string
		expectedLen int
	}{
		{
			name: "valid slice of models",
			setupSlice: func() reflect.Value {
				models := []*TestModel{
					{ID: 1, Name: "test1"},
					{ID: 2, Name: "test2"},
				}
				return reflect.ValueOf(models)
			},
			expectError: false,
			expectedLen: 2,
		},
		{
			name: "empty slice",
			setupSlice: func() reflect.Value {
				models := []*TestModel{}
				return reflect.ValueOf(models)
			},
			expectError: false,
			expectedLen: 0,
		},
		{
			name: "slice with non-model items",
			setupSlice: func() reflect.Value {
				nonModels := []NonModel{
					{ID: 1, Name: "test1"},
				}
				return reflect.ValueOf(nonModels)
			},
			expectError: true,
			errorMsg:    "failed to assert item at index 0 to contract.Model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sliceValue := tt.setupSlice()
			result, err := ConvertSliceToModels(sliceValue)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedLen)
				for _, model := range result {
					assert.Implements(t, (*contract.Model)(nil), model)
				}
			}
		})
	}
}

func TestGetModelType(t *testing.T) {
	tests := []struct {
		name     string
		model    contract.Model
		expected reflect.Type
	}{
		{
			name:     "pointer model",
			model:    &TestModel{},
			expected: reflect.TypeOf(TestModel{}),
		},
		{
			name:     "another pointer model",
			model:    &AnotherTestModel{},
			expected: reflect.TypeOf(AnotherTestModel{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetModelType(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateSliceOfModelType(t *testing.T) {
	tests := []struct {
		name      string
		modelType reflect.Type
	}{
		{
			name:      "TestModel type",
			modelType: reflect.TypeOf(TestModel{}),
		},
		{
			name:      "AnotherTestModel type",
			modelType: reflect.TypeOf(AnotherTestModel{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateSliceOfModelType(tt.modelType)
			assert.NotNil(t, result)

			// Verify it's a pointer to a slice
			resultValue := reflect.ValueOf(result)
			assert.Equal(t, reflect.Ptr, resultValue.Kind())

			// Verify the element type is a slice of pointers to the model type
			sliceType := resultValue.Type().Elem()
			assert.Equal(t, reflect.Slice, sliceType.Kind())
			assert.Equal(t, reflect.Ptr, sliceType.Elem().Kind())
			assert.Equal(t, tt.modelType, sliceType.Elem().Elem())
		})
	}
}

func TestIsNilOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "nil pointer",
			value:    (*TestModel)(nil),
			expected: true,
		},
		{
			name:     "valid pointer",
			value:    &TestModel{},
			expected: false,
		},
		{
			name:     "nil interface",
			value:    interface{}(nil),
			expected: true,
		},
		{
			name:     "empty string",
			value:    "",
			expected: true,
		},
		{
			name:     "non-empty string",
			value:    "test",
			expected: false,
		},
		{
			name:     "nil slice",
			value:    []string(nil),
			expected: true,
		},
		{
			name:     "empty slice",
			value:    []string{},
			expected: true,
		},
		{
			name:     "non-empty slice",
			value:    []string{"test"},
			expected: false,
		},
		{
			name:     "empty array",
			value:    [0]string{},
			expected: true,
		},
		{
			name:     "non-empty array",
			value:    [1]string{"test"},
			expected: false,
		},
		{
			name:     "nil map",
			value:    map[string]string(nil),
			expected: true,
		},
		{
			name:     "empty map",
			value:    map[string]string{},
			expected: false,
		},
		{
			name:     "non-empty map",
			value:    map[string]string{"key": "value"},
			expected: false,
		},
		{
			name:     "zero int",
			value:    0,
			expected: false,
		},
		{
			name:     "non-zero int",
			value:    42,
			expected: false,
		},
		{
			name:     "zero bool",
			value:    false,
			expected: false,
		},
		{
			name:     "true bool",
			value:    true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v reflect.Value
			if tt.value == nil {
				v = reflect.Value{}
			} else {
				v = reflect.ValueOf(tt.value)
			}
			result := IsNilOrEmpty(v)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test invalid value
	t.Run("invalid value", func(t *testing.T) {
		var v reflect.Value // zero value, invalid
		result := IsNilOrEmpty(v)
		assert.True(t, result)
	})
}

func TestSafeTypeAssertion(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		typeName    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid string assertion",
			value:       "test",
			typeName:    "string",
			expectError: false,
		},
		{
			name:        "valid int assertion",
			value:       42,
			typeName:    "int",
			expectError: false,
		},
		{
			name:        "nil value",
			value:       nil,
			typeName:    "string",
			expectError: true,
			errorMsg:    "cannot assert nil value to string",
		},
		{
			name:        "invalid type assertion",
			value:       "test",
			typeName:    "int",
			expectError: true,
			errorMsg:    "cannot assert value of type string to int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.typeName {
			case "string":
				result, err := SafeTypeAssertion[string](tt.value, tt.typeName)
				if tt.expectError {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorMsg)
					assert.Equal(t, "", result) // zero value for string
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.value, result)
				}
			case "int":
				result, err := SafeTypeAssertion[int](tt.value, tt.typeName)
				if tt.expectError {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorMsg)
					assert.Equal(t, 0, result) // zero value for int
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.value, result)
				}
			}
		})
	}
}

func TestExecuteQueryAndConvertToModels(t *testing.T) {
	tests := []struct {
		name          string
		model         contract.Model
		queryExecutor func(interface{}) error
		expectError   bool
		errorMsg      string
		expectedLen   int
	}{
		{
			name:  "successful query execution",
			model: &TestModel{},
			queryExecutor: func(dest interface{}) error {
				// Simulate filling the slice with data
				slice := dest.(*[]*TestModel)
				*slice = []*TestModel{
					{ID: 1, Name: "test1"},
					{ID: 2, Name: "test2"},
				}
				return nil
			},
			expectError: false,
			expectedLen: 2,
		},
		{
			name:  "query execution error",
			model: &TestModel{},
			queryExecutor: func(dest interface{}) error {
				return errors.New("query failed")
			},
			expectError: true,
			errorMsg:    "query failed",
		},
		{
			name:  "empty result",
			model: &TestModel{},
			queryExecutor: func(dest interface{}) error {
				// Don't fill the slice, leave it empty
				return nil
			},
			expectError: false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteQueryAndConvertToModels(tt.model, tt.queryExecutor)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedLen)
				for _, model := range result {
					assert.Implements(t, (*contract.Model)(nil), model)
				}
			}
		})
	}
}

func TestNewGormQueryExecutor(t *testing.T) {
	queryFunc := func(dest interface{}) error {
		return nil
	}

	executor := NewGormQueryExecutor(queryFunc)
	assert.NotNil(t, executor)
	assert.IsType(t, &GormQueryExecutor{}, executor)
}

func TestGormQueryExecutor_Execute(t *testing.T) {
	tests := []struct {
		name        string
		queryFunc   func(interface{}) error
		dest        interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful execution",
			queryFunc: func(dest interface{}) error {
				return nil
			},
			dest:        &[]TestModel{},
			expectError: false,
		},
		{
			name: "execution error",
			queryFunc: func(dest interface{}) error {
				return errors.New("execution failed")
			},
			dest:        &[]TestModel{},
			expectError: true,
			errorMsg:    "execution failed",
		},
		{
			name: "execution with data modification",
			queryFunc: func(dest interface{}) error {
				if slice, ok := dest.(*[]*TestModel); ok {
					*slice = []*TestModel{
						{ID: 1, Name: "test"},
					}
				}
				return nil
			},
			dest:        &[]*TestModel{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewGormQueryExecutor(tt.queryFunc)
			err := executor.Execute(tt.dest)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
