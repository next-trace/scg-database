package db

import (
	"testing"

	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/assert"
)

// MockQueryBuilderFactory is a mock implementation of contract.QueryBuilderFactory
type (
	MockQueryBuilderFactory struct {
		name string
	}
)

func (m *MockQueryBuilderFactory) NewQueryBuilder(model contract.Model, connection any) contract.QueryBuilder {
	return nil // Not needed for registry tests
}

func (m *MockQueryBuilderFactory) Name() string {
	return m.name
}

func TestGetQueryBuilderRegistry(t *testing.T) {
	registry := GetQueryBuilderRegistry()
	assert.NotNil(t, registry)
	assert.IsType(t, &queryBuilderRegistry{}, registry)
}

func TestQueryBuilderRegistry_Register(t *testing.T) {
	registry := &queryBuilderRegistry{
		factories: make(map[string]contract.QueryBuilderFactory),
	}

	factory := &MockQueryBuilderFactory{name: "test"}

	tests := []struct {
		name        string
		adapterName string
		factory     contract.QueryBuilderFactory
		expectPanic bool
		panicMsg    string
	}{
		{
			name:        "valid registration",
			adapterName: "test-adapter",
			factory:     factory,
			expectPanic: false,
		},
		{
			name:        "empty adapter name",
			adapterName: "",
			factory:     factory,
			expectPanic: true,
			panicMsg:    "adapter name cannot be empty",
		},
		{
			name:        "nil factory",
			adapterName: "test-adapter",
			factory:     nil,
			expectPanic: true,
			panicMsg:    "factory cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.PanicsWithValue(t, tt.panicMsg, func() {
					registry.Register(tt.adapterName, tt.factory)
				})
			} else {
				assert.NotPanics(t, func() {
					registry.Register(tt.adapterName, tt.factory)
				})
				// Verify registration
				retrieved, err := registry.Get(tt.adapterName)
				assert.NoError(t, err)
				assert.Equal(t, tt.factory, retrieved)
			}
		})
	}
}

func TestQueryBuilderRegistry_Get(t *testing.T) {
	registry := &queryBuilderRegistry{
		factories: make(map[string]contract.QueryBuilderFactory),
	}

	factory := &MockQueryBuilderFactory{name: "test"}
	registry.Register("test-adapter", factory)

	tests := []struct {
		name        string
		adapterName string
		expectError bool
		errorMsg    string
		expected    contract.QueryBuilderFactory
	}{
		{
			name:        "existing adapter",
			adapterName: "test-adapter",
			expectError: false,
			expected:    factory,
		},
		{
			name:        "non-existing adapter",
			adapterName: "non-existing",
			expectError: true,
			errorMsg:    "query builder factory not found for adapter: non-existing",
		},
		{
			name:        "empty adapter name",
			adapterName: "",
			expectError: true,
			errorMsg:    "query builder factory not found for adapter: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := registry.Get(tt.adapterName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestQueryBuilderRegistry_List(t *testing.T) {
	registry := &queryBuilderRegistry{
		factories: make(map[string]contract.QueryBuilderFactory),
	}

	// Test empty registry
	list := registry.List()
	assert.NotNil(t, list)
	assert.Len(t, list, 0)

	// Add factories
	factory1 := &MockQueryBuilderFactory{name: "factory1"}
	factory2 := &MockQueryBuilderFactory{name: "factory2"}
	factory3 := &MockQueryBuilderFactory{name: "factory3"}

	registry.Register("adapter1", factory1)
	registry.Register("adapter2", factory2)
	registry.Register("adapter3", factory3)

	// Test populated registry
	list = registry.List()
	assert.Len(t, list, 3)
	assert.Contains(t, list, "adapter1")
	assert.Contains(t, list, "adapter2")
	assert.Contains(t, list, "adapter3")
}

func TestRegisterQueryBuilderFactory(t *testing.T) {
	// Save original state
	originalFactories := make(map[string]contract.QueryBuilderFactory)
	for k, v := range globalQueryBuilderRegistry.factories {
		originalFactories[k] = v
	}

	// Clean up after test
	defer func() {
		globalQueryBuilderRegistry.factories = originalFactories
	}()

	factory := &MockQueryBuilderFactory{name: "global-test"}

	// Test registration
	assert.NotPanics(t, func() {
		RegisterQueryBuilderFactory("global-adapter", factory)
	})

	// Verify registration
	retrieved, err := globalQueryBuilderRegistry.Get("global-adapter")
	assert.NoError(t, err)
	assert.Equal(t, factory, retrieved)
}

func TestGetQueryBuilderFactory(t *testing.T) {
	// Save original state
	originalFactories := make(map[string]contract.QueryBuilderFactory)
	for k, v := range globalQueryBuilderRegistry.factories {
		originalFactories[k] = v
	}

	// Clean up after test
	defer func() {
		globalQueryBuilderRegistry.factories = originalFactories
	}()

	factory := &MockQueryBuilderFactory{name: "global-test"}
	globalQueryBuilderRegistry.Register("global-adapter", factory)

	tests := []struct {
		name        string
		adapterName string
		expectError bool
		expected    contract.QueryBuilderFactory
	}{
		{
			name:        "existing adapter",
			adapterName: "global-adapter",
			expectError: false,
			expected:    factory,
		},
		{
			name:        "non-existing adapter",
			adapterName: "non-existing",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetQueryBuilderFactory(tt.adapterName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestListQueryBuilderFactories(t *testing.T) {
	// Save original state
	originalFactories := make(map[string]contract.QueryBuilderFactory)
	for k, v := range globalQueryBuilderRegistry.factories {
		originalFactories[k] = v
	}

	// Clean up after test
	defer func() {
		globalQueryBuilderRegistry.factories = originalFactories
	}()

	// Clear registry for clean test
	globalQueryBuilderRegistry.factories = make(map[string]contract.QueryBuilderFactory)

	// Test empty registry
	list := ListQueryBuilderFactories()
	assert.NotNil(t, list)
	assert.Len(t, list, 0)

	// Add factories
	factory1 := &MockQueryBuilderFactory{name: "factory1"}
	factory2 := &MockQueryBuilderFactory{name: "factory2"}

	globalQueryBuilderRegistry.Register("adapter1", factory1)
	globalQueryBuilderRegistry.Register("adapter2", factory2)

	// Test populated registry
	list = ListQueryBuilderFactories()
	assert.Len(t, list, 2)
	assert.Contains(t, list, "adapter1")
	assert.Contains(t, list, "adapter2")
}

func TestQueryBuilderRegistry_ConcurrentAccess(t *testing.T) {
	registry := &queryBuilderRegistry{
		factories: make(map[string]contract.QueryBuilderFactory),
	}

	factory := &MockQueryBuilderFactory{name: "concurrent-test"}

	// Test concurrent registration and retrieval
	done := make(chan bool, 2)

	// Goroutine 1: Register
	go func() {
		defer func() { done <- true }()
		for range 100 {
			registry.Register("concurrent-adapter", factory)
		}
	}()

	// Goroutine 2: Get and List
	go func() {
		defer func() { done <- true }()
		for range 100 {
			registry.Get("concurrent-adapter")
			registry.List()
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	retrieved, err := registry.Get("concurrent-adapter")
	assert.NoError(t, err)
	assert.Equal(t, factory, retrieved)

	list := registry.List()
	assert.Contains(t, list, "concurrent-adapter")
}

func TestQueryBuilderRegistry_OverwriteRegistration(t *testing.T) {
	registry := &queryBuilderRegistry{
		factories: make(map[string]contract.QueryBuilderFactory),
	}

	factory1 := &MockQueryBuilderFactory{name: "factory1"}
	factory2 := &MockQueryBuilderFactory{name: "factory2"}

	// Register first factory
	registry.Register("test-adapter", factory1)
	retrieved, err := registry.Get("test-adapter")
	assert.NoError(t, err)
	assert.Equal(t, factory1, retrieved)

	// Overwrite with second factory
	registry.Register("test-adapter", factory2)
	retrieved, err = registry.Get("test-adapter")
	assert.NoError(t, err)
	assert.Equal(t, factory2, retrieved)
	assert.NotEqual(t, factory1, retrieved)
}
