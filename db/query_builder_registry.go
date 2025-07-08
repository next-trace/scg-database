package db

import (
	"fmt"
	"sync"

	"github.com/next-trace/scg-database/contract"
)

// queryBuilderRegistry implements contract.QueryBuilderRegistry
type (
	queryBuilderRegistry struct {
		factories map[string]contract.QueryBuilderFactory
		mu        sync.RWMutex
	}
)

var (
	// Ensure queryBuilderRegistry implements contract.QueryBuilderRegistry
	_ contract.QueryBuilderRegistry = (*queryBuilderRegistry)(nil)

	// Global registry instance
	globalQueryBuilderRegistry = &queryBuilderRegistry{
		factories: make(map[string]contract.QueryBuilderFactory),
	}
)

// GetQueryBuilderRegistry returns the global query builder registry instance
func GetQueryBuilderRegistry() contract.QueryBuilderRegistry {
	return globalQueryBuilderRegistry
}

// Register registers a query builder factory for a specific adapter
func (r *queryBuilderRegistry) Register(adapterName string, factory contract.QueryBuilderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if adapterName == "" {
		panic("adapter name cannot be empty")
	}
	if factory == nil {
		panic("factory cannot be nil")
	}

	r.factories[adapterName] = factory
}

// Get retrieves a query builder factory for a specific adapter
func (r *queryBuilderRegistry) Get(adapterName string) (contract.QueryBuilderFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[adapterName]
	if !exists {
		return nil, fmt.Errorf("query builder factory not found for adapter: %s", adapterName)
	}

	return factory, nil
}

// List returns a list of all registered adapter names
func (r *queryBuilderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}

	return names
}

// RegisterQueryBuilderFactory is a convenience function to register a query builder factory
func RegisterQueryBuilderFactory(adapterName string, factory contract.QueryBuilderFactory) {
	globalQueryBuilderRegistry.Register(adapterName, factory)
}

// GetQueryBuilderFactory is a convenience function to get a query builder factory
func GetQueryBuilderFactory(adapterName string) (contract.QueryBuilderFactory, error) {
	return globalQueryBuilderRegistry.Get(adapterName)
}

// ListQueryBuilderFactories is a convenience function to list all registered factories
func ListQueryBuilderFactories() []string {
	return globalQueryBuilderRegistry.List()
}
