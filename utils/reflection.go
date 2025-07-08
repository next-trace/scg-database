// Package utils provides utility functions for database configuration, connection pooling,
// and other common operations used throughout the SCG database toolkit.
//
//revive:disable:var-naming // allow package name 'utils' for this utilities module
package utils

import (
	"fmt"
	"reflect"

	"github.com/next-trace/scg-database/contract"
)

// CreateEntityFromModel creates a new entity instance from a model using reflection
func CreateEntityFromModel(model contract.Model) (contract.Model, error) {
	newInstance := reflect.New(reflect.TypeOf(model).Elem()).Interface()
	if result, ok := newInstance.(contract.Model); ok {
		return result, nil
	}
	return nil, fmt.Errorf("failed to assert created instance to contract.Model")
}

// ConvertModelsToSlice converts a slice of contract.Model to a concrete slice for GORM operations
func ConvertModelsToSlice(models []contract.Model, modelType reflect.Type) (interface{}, error) {
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

// ConvertSliceToModels converts a reflected slice back to []contract.Model
func ConvertSliceToModels(sliceValue reflect.Value) ([]contract.Model, error) {
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

// GetModelType returns the reflect.Type of a model, handling pointer types
func GetModelType(model contract.Model) reflect.Type {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		return modelType.Elem()
	}
	return modelType
}

// CreateSliceOfModelType creates a new slice of the given model type
func CreateSliceOfModelType(modelType reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(reflect.PointerTo(modelType))).Interface()
}

// IsNilOrEmpty checks if a reflected value is nil or represents an empty value
func IsNilOrEmpty(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Array:
		return v.Len() == 0
	case reflect.Slice:
		if v.IsNil() {
			return true
		}
		return v.Len() == 0
	default:
		return false
	}
}

// SafeTypeAssertion performs a safe type assertion with error handling
func SafeTypeAssertion[T any](value interface{}, typeName string) (T, error) {
	var zero T
	if value == nil {
		return zero, fmt.Errorf("cannot assert nil value to %s", typeName)
	}

	result, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("cannot assert value of type %T to %s", value, typeName)
	}

	return result, nil
}

// ExecuteQueryAndConvertToModels executes a query and converts the result to models
// This encapsulates the reflection logic for query execution
func ExecuteQueryAndConvertToModels(
	model contract.Model,
	queryExecutor func(interface{}) error,
) ([]contract.Model, error) {
	modelType := GetModelType(model)
	slice := CreateSliceOfModelType(modelType)

	if err := queryExecutor(slice); err != nil {
		return nil, err
	}

	val := reflect.Indirect(reflect.ValueOf(slice))
	return ConvertSliceToModels(val)
}

type (
	// QueryExecutor defines the interface for executing queries
	QueryExecutor interface {
		Execute(interface{}) error
	}

	// GormQueryExecutor implements QueryExecutor for GORM
	GormQueryExecutor struct {
		queryFunc func(interface{}) error
	}
)

// NewGormQueryExecutor creates a new GORM query executor
func NewGormQueryExecutor(queryFunc func(interface{}) error) QueryExecutor {
	return &GormQueryExecutor{queryFunc: queryFunc}
}

// Execute executes the query using GORM
func (g *GormQueryExecutor) Execute(dest interface{}) error {
	return g.queryFunc(dest)
}
