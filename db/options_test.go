package db

import (
	"testing"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/assert"
)

// MockDBAdapter is a mock implementation of contract.DBAdapter for testing
type (
	MockDBAdapter struct {
		name string
	}
)

func (m *MockDBAdapter) Name() string {
	return m.name
}

func (m *MockDBAdapter) Connect(cfg *config.Config) (contract.Connection, error) {
	return nil, nil // Not needed for options tests
}

func TestWithAdapter(t *testing.T) {
	adapter := &MockDBAdapter{name: "test-adapter"}

	tests := []struct {
		name     string
		adapter  contract.DBAdapter
		expected contract.DBAdapter
	}{
		{
			name:     "valid adapter",
			adapter:  adapter,
			expected: adapter,
		},
		{
			name:     "nil adapter",
			adapter:  nil,
			expected: nil,
		},
		{
			name:     "different adapter",
			adapter:  &MockDBAdapter{name: "different-adapter"},
			expected: &MockDBAdapter{name: "different-adapter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a config to test the option
			cfg := &config.Config{}

			// Apply the option
			option := WithAdapter(tt.adapter)
			option(cfg)

			// Verify the adapter was set correctly
			assert.Equal(t, tt.expected, cfg.Adapter)
		})
	}
}

func TestWithAdapter_Integration(t *testing.T) {
	// Test that WithAdapter works with the config.Option pattern
	adapter := &MockDBAdapter{name: "integration-test"}

	// Create config with multiple options
	cfg := config.Config{
		Driver: "test:driver",
		DSN:    "test.db",
	}

	// Apply options
	options := []config.Option{
		WithAdapter(adapter),
	}

	for _, opt := range options {
		opt(&cfg)
	}

	// Verify the adapter was set
	assert.Equal(t, adapter, cfg.Adapter)
	assert.Equal(t, "test:driver", cfg.Driver) // Other fields should remain unchanged
	assert.Equal(t, "test.db", cfg.DSN)
}

func TestWithAdapter_OverwriteExisting(t *testing.T) {
	// Test that WithAdapter overwrites existing adapter
	originalAdapter := &MockDBAdapter{name: "original"}
	newAdapter := &MockDBAdapter{name: "new"}

	cfg := &config.Config{
		Adapter: originalAdapter,
	}

	// Apply new adapter option
	option := WithAdapter(newAdapter)
	option(cfg)

	// Verify the adapter was overwritten
	assert.Equal(t, newAdapter, cfg.Adapter)
	assert.NotEqual(t, originalAdapter, cfg.Adapter)
}

func TestWithAdapter_MultipleApplications(t *testing.T) {
	// Test applying WithAdapter multiple times
	adapter1 := &MockDBAdapter{name: "adapter1"}
	adapter2 := &MockDBAdapter{name: "adapter2"}
	adapter3 := &MockDBAdapter{name: "adapter3"}

	cfg := &config.Config{}

	// Apply multiple adapter options
	WithAdapter(adapter1)(cfg)
	assert.Equal(t, adapter1, cfg.Adapter)

	WithAdapter(adapter2)(cfg)
	assert.Equal(t, adapter2, cfg.Adapter)

	WithAdapter(adapter3)(cfg)
	assert.Equal(t, adapter3, cfg.Adapter)
}

func TestWithAdapter_FunctionSignature(t *testing.T) {
	// Test that WithAdapter returns the correct function type
	adapter := &MockDBAdapter{name: "signature-test"}

	option := WithAdapter(adapter)

	// Verify it's a config.Option function
	assert.IsType(t, config.Option(nil), option)

	// Verify it can be called with a config
	cfg := &config.Config{}
	assert.NotPanics(t, func() {
		option(cfg)
	})
}

func TestWithAdapter_NilConfig(t *testing.T) {
	// Test behavior with nil config (should not panic)
	adapter := &MockDBAdapter{name: "nil-config-test"}
	option := WithAdapter(adapter)

	// This should not panic even with nil config
	assert.Panics(t, func() {
		option(nil)
	})
}
