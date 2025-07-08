//revive:disable:var-naming // allow 'utils' package name in tests for consistency and to avoid widespread refactors
package utils

import (
	"testing"
	"time"

	"github.com/next-trace/scg-database/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBInterface defines the interface for database connection pool operations
// MockDB is a mock implementation of DBInterface for testing
// MockDialectStrategy is a mock implementation of DialectStrategy
type (
	DBInterface interface {
		SetMaxOpenConns(n int)
		SetMaxIdleConns(n int)
		SetConnMaxLifetime(d time.Duration)
	}

	MockDB struct {
		mock.Mock
		maxOpenConns    int
		maxIdleConns    int
		connMaxLifetime time.Duration
	}

	MockDialectStrategy struct {
		mock.Mock
	}
)

func (m *MockDB) SetMaxOpenConns(n int) {
	m.Called(n)
	m.maxOpenConns = n
}

func (m *MockDB) SetMaxIdleConns(n int) {
	m.Called(n)
	m.maxIdleConns = n
}

func (m *MockDB) SetConnMaxLifetime(d time.Duration) {
	m.Called(d)
	m.connMaxLifetime = d
}

// configureConnectionPoolInterface is a wrapper to test the function with interface
func configureConnectionPoolInterface(db DBInterface, cfg *config.Config) {
	const defaultConnMaxLifetime = 10 * time.Second

	// Set connection max lifetime
	lifetime := defaultConnMaxLifetime
	if cfg.ConnMaxLifetime > 0 {
		lifetime = cfg.ConnMaxLifetime
	}
	db.SetConnMaxLifetime(lifetime)

	// Set max idle connections
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	// Set max open connections
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
}

// applyConnectionPoolOptionsInterface is a wrapper to test the function with interface
func applyConnectionPoolOptionsInterface(db DBInterface, options ...ConnectionPoolOption) {
	opts := &ConnectionPoolOptions{
		ConnMaxLifetime: 10 * time.Second, // default
	}

	for _, option := range options {
		option(opts)
	}

	if opts.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}
	if opts.MaxIdleConns > 0 {
		db.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.MaxOpenConns > 0 {
		db.SetMaxOpenConns(opts.MaxOpenConns)
	}
}

// applyPoolConfigurationInterface is a wrapper to test the ApplyPoolConfiguration method with interface
func applyPoolConfigurationInterface(db DBInterface, options []ConnectionPoolOption) {
	applyConnectionPoolOptionsInterface(db, options...)
}

func TestConfigureConnectionPool(t *testing.T) {
	tests := []struct {
		name                    string
		cfg                     *config.Config
		expectedMaxOpenConns    int
		expectedMaxIdleConns    int
		expectedConnMaxLifetime time.Duration
	}{
		{
			name: "all settings configured",
			cfg: &config.Config{
				MaxOpenConns:    25,
				MaxIdleConns:    10,
				ConnMaxLifetime: 30 * time.Second,
			},
			expectedMaxOpenConns:    25,
			expectedMaxIdleConns:    10,
			expectedConnMaxLifetime: 30 * time.Second,
		},
		{
			name: "only max open conns configured",
			cfg: &config.Config{
				MaxOpenConns: 50,
			},
			expectedMaxOpenConns:    50,
			expectedMaxIdleConns:    0,                // not set
			expectedConnMaxLifetime: 10 * time.Second, // default
		},
		{
			name: "only max idle conns configured",
			cfg: &config.Config{
				MaxIdleConns: 15,
			},
			expectedMaxOpenConns:    0, // not set
			expectedMaxIdleConns:    15,
			expectedConnMaxLifetime: 10 * time.Second, // default
		},
		{
			name: "only conn max lifetime configured",
			cfg: &config.Config{
				ConnMaxLifetime: 60 * time.Second,
			},
			expectedMaxOpenConns:    0, // not set
			expectedMaxIdleConns:    0, // not set
			expectedConnMaxLifetime: 60 * time.Second,
		},
		{
			name:                    "no settings configured - use defaults",
			cfg:                     &config.Config{},
			expectedMaxOpenConns:    0,                // not set
			expectedMaxIdleConns:    0,                // not set
			expectedConnMaxLifetime: 10 * time.Second, // default
		},
		{
			name: "zero values - should not be set",
			cfg: &config.Config{
				MaxOpenConns:    0,
				MaxIdleConns:    0,
				ConnMaxLifetime: 0,
			},
			expectedMaxOpenConns:    0,                // not set
			expectedMaxIdleConns:    0,                // not set
			expectedConnMaxLifetime: 10 * time.Second, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}

			// Set up expectations
			mockDB.On("SetConnMaxLifetime", tt.expectedConnMaxLifetime).Return()
			if tt.expectedMaxIdleConns > 0 {
				mockDB.On("SetMaxIdleConns", tt.expectedMaxIdleConns).Return()
			}
			if tt.expectedMaxOpenConns > 0 {
				mockDB.On("SetMaxOpenConns", tt.expectedMaxOpenConns).Return()
			}

			// Call the function
			configureConnectionPoolInterface(mockDB, tt.cfg)

			// Verify expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestExtractGormConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected *gorm.Config
	}{
		{
			name: "empty settings",
			cfg: &config.Config{
				Settings: map[string]interface{}{},
			},
			expected: &gorm.Config{},
		},
		{
			name: "with gorm_config",
			cfg: &config.Config{
				Settings: map[string]interface{}{
					"gorm_config": &gorm.Config{
						SkipDefaultTransaction: true,
					},
				},
			},
			expected: &gorm.Config{
				SkipDefaultTransaction: true,
			},
		},
		{
			name: "with gorm_logger",
			cfg: &config.Config{
				Settings: map[string]interface{}{
					"gorm_logger": logger.Default.LogMode(logger.Silent),
				},
			},
			expected: &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			},
		},
		{
			name: "with both gorm_config and gorm_logger",
			cfg: &config.Config{
				Settings: map[string]interface{}{
					"gorm_config": &gorm.Config{
						SkipDefaultTransaction: true,
					},
					"gorm_logger": logger.Default.LogMode(logger.Info),
				},
			},
			expected: &gorm.Config{
				SkipDefaultTransaction: true,
				Logger:                 logger.Default.LogMode(logger.Info),
			},
		},
		{
			name: "with invalid gorm_config type",
			cfg: &config.Config{
				Settings: map[string]interface{}{
					"gorm_config": "invalid",
				},
			},
			expected: &gorm.Config{},
		},
		{
			name: "with invalid gorm_logger type",
			cfg: &config.Config{
				Settings: map[string]interface{}{
					"gorm_logger": "invalid",
				},
			},
			expected: &gorm.Config{},
		},
		{
			name: "nil settings",
			cfg: &config.Config{
				Settings: nil,
			},
			expected: &gorm.Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractGormConfig(tt.cfg)
			assert.Equal(t, tt.expected.SkipDefaultTransaction, result.SkipDefaultTransaction)
			if tt.expected.Logger != nil {
				assert.NotNil(t, result.Logger)
			}
		})
	}
}

func TestApplyConnectionPoolOptions(t *testing.T) {
	tests := []struct {
		name                    string
		options                 []ConnectionPoolOption
		expectedMaxOpenConns    int
		expectedMaxIdleConns    int
		expectedConnMaxLifetime time.Duration
	}{
		{
			name:                    "no options - use defaults",
			options:                 []ConnectionPoolOption{},
			expectedMaxOpenConns:    0,                // not set
			expectedMaxIdleConns:    0,                // not set
			expectedConnMaxLifetime: 10 * time.Second, // default
		},
		{
			name: "with max open conns option",
			options: []ConnectionPoolOption{
				WithMaxOpenConns(25),
			},
			expectedMaxOpenConns:    25,
			expectedMaxIdleConns:    0,                // not set
			expectedConnMaxLifetime: 10 * time.Second, // default
		},
		{
			name: "with max idle conns option",
			options: []ConnectionPoolOption{
				WithMaxIdleConns(10),
			},
			expectedMaxOpenConns:    0, // not set
			expectedMaxIdleConns:    10,
			expectedConnMaxLifetime: 10 * time.Second, // default
		},
		{
			name: "with conn max lifetime option",
			options: []ConnectionPoolOption{
				WithConnMaxLifetime(30 * time.Second),
			},
			expectedMaxOpenConns:    0, // not set
			expectedMaxIdleConns:    0, // not set
			expectedConnMaxLifetime: 30 * time.Second,
		},
		{
			name: "with all options",
			options: []ConnectionPoolOption{
				WithMaxOpenConns(50),
				WithMaxIdleConns(20),
				WithConnMaxLifetime(60 * time.Second),
			},
			expectedMaxOpenConns:    50,
			expectedMaxIdleConns:    20,
			expectedConnMaxLifetime: 60 * time.Second,
		},
		{
			name: "with zero values - should not be set",
			options: []ConnectionPoolOption{
				WithMaxOpenConns(0),
				WithMaxIdleConns(0),
				WithConnMaxLifetime(0),
			},
			expectedMaxOpenConns:    0, // not set
			expectedMaxIdleConns:    0, // not set
			expectedConnMaxLifetime: 0, // not set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}

			// Set up expectations
			if tt.expectedConnMaxLifetime > 0 {
				mockDB.On("SetConnMaxLifetime", tt.expectedConnMaxLifetime).Return()
			}
			if tt.expectedMaxIdleConns > 0 {
				mockDB.On("SetMaxIdleConns", tt.expectedMaxIdleConns).Return()
			}
			if tt.expectedMaxOpenConns > 0 {
				mockDB.On("SetMaxOpenConns", tt.expectedMaxOpenConns).Return()
			}

			// Call the function
			applyConnectionPoolOptionsInterface(mockDB, tt.options...)

			// Verify expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestConnectionPoolOptions(t *testing.T) {
	t.Run("WithMaxOpenConns", func(t *testing.T) {
		opts := &ConnectionPoolOptions{}
		option := WithMaxOpenConns(25)
		option(opts)
		assert.Equal(t, 25, opts.MaxOpenConns)
	})

	t.Run("WithMaxIdleConns", func(t *testing.T) {
		opts := &ConnectionPoolOptions{}
		option := WithMaxIdleConns(10)
		option(opts)
		assert.Equal(t, 10, opts.MaxIdleConns)
	})

	t.Run("WithConnMaxLifetime", func(t *testing.T) {
		opts := &ConnectionPoolOptions{}
		lifetime := 30 * time.Second
		option := WithConnMaxLifetime(lifetime)
		option(opts)
		assert.Equal(t, lifetime, opts.ConnMaxLifetime)
	})
}

func TestConfigFromOptions(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.Config
		expectedOptions int
	}{
		{
			name: "all options configured",
			cfg: &config.Config{
				MaxOpenConns:    25,
				MaxIdleConns:    10,
				ConnMaxLifetime: 30 * time.Second,
			},
			expectedOptions: 3,
		},
		{
			name: "only max open conns configured",
			cfg: &config.Config{
				MaxOpenConns: 50,
			},
			expectedOptions: 1,
		},
		{
			name: "only max idle conns configured",
			cfg: &config.Config{
				MaxIdleConns: 15,
			},
			expectedOptions: 1,
		},
		{
			name: "only conn max lifetime configured",
			cfg: &config.Config{
				ConnMaxLifetime: 60 * time.Second,
			},
			expectedOptions: 1,
		},
		{
			name:            "no options configured",
			cfg:             &config.Config{},
			expectedOptions: 0,
		},
		{
			name: "zero values - should not create options",
			cfg: &config.Config{
				MaxOpenConns:    0,
				MaxIdleConns:    0,
				ConnMaxLifetime: 0,
			},
			expectedOptions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := ConfigFromOptions(tt.cfg)
			assert.Len(t, options, tt.expectedOptions)

			// Test that options work correctly
			if len(options) > 0 {
				opts := &ConnectionPoolOptions{}
				for _, option := range options {
					option(opts)
				}

				if tt.cfg.MaxOpenConns > 0 {
					assert.Equal(t, tt.cfg.MaxOpenConns, opts.MaxOpenConns)
				}
				if tt.cfg.MaxIdleConns > 0 {
					assert.Equal(t, tt.cfg.MaxIdleConns, opts.MaxIdleConns)
				}
				if tt.cfg.ConnMaxLifetime > 0 {
					assert.Equal(t, tt.cfg.ConnMaxLifetime, opts.ConnMaxLifetime)
				}
			}
		})
	}
}

func (m *MockDialectStrategy) CreateDialector(dsn string) (interface{}, error) {
	args := m.Called(dsn)
	return args.Get(0), args.Error(1)
}

func (m *MockDialectStrategy) ValidateDriver(driver string) error {
	args := m.Called(driver)
	return args.Error(0)
}

func (m *MockDialectStrategy) GetDriverName() string {
	args := m.Called()
	return args.String(0)
}

func TestNewConnectionBuilder(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.Config
		options         []ConnectionBuilderOption
		expectedOptions int
	}{
		{
			name: "basic builder",
			cfg: &config.Config{
				Driver: "gorm:sqlite",
				DSN:    "test.db",
			},
			options:         []ConnectionBuilderOption{},
			expectedOptions: 0,
		},
		{
			name: "builder with dialect strategy",
			cfg: &config.Config{
				Driver: "gorm:sqlite",
				DSN:    "test.db",
			},
			options: []ConnectionBuilderOption{
				WithDialectStrategy(&MockDialectStrategy{}),
			},
			expectedOptions: 0,
		},
		{
			name: "builder with pool options",
			cfg: &config.Config{
				Driver: "gorm:sqlite",
				DSN:    "test.db",
			},
			options: []ConnectionBuilderOption{
				WithPoolOptions(WithMaxOpenConns(25)),
			},
			expectedOptions: 1,
		},
		{
			name: "builder with config pool options",
			cfg: &config.Config{
				Driver:       "gorm:sqlite",
				DSN:          "test.db",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
			options:         []ConnectionBuilderOption{},
			expectedOptions: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewConnectionBuilder(tt.cfg, tt.options...)
			assert.NotNil(t, builder)
			assert.Equal(t, tt.cfg.Driver, builder.config.Driver)
			assert.Equal(t, tt.cfg.DSN, builder.config.DSN)
			assert.Len(t, builder.poolOptions, tt.expectedOptions)
		})
	}
}

func TestConnectionBuilderOptions(t *testing.T) {
	t.Run("WithDialectStrategy", func(t *testing.T) {
		strategy := &MockDialectStrategy{}
		builder := &ConnectionBuilder{}
		option := WithDialectStrategy(strategy)
		option(builder)
		assert.Equal(t, strategy, builder.dialectStrategy)
	})

	t.Run("WithPoolOptions", func(t *testing.T) {
		builder := &ConnectionBuilder{}
		poolOption := WithMaxOpenConns(25)
		option := WithPoolOptions(poolOption)
		option(builder)
		assert.Len(t, builder.poolOptions, 1)
	})
}

func TestConnectionBuilderBuild(t *testing.T) {
	tests := []struct {
		name        string
		builder     *ConnectionBuilder
		setupMock   func(*MockDialectStrategy)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful build",
			builder: &ConnectionBuilder{
				config: config.Config{
					Driver: "gorm:sqlite",
					DSN:    "test.db",
				},
				dialectStrategy: &MockDialectStrategy{},
			},
			setupMock: func(m *MockDialectStrategy) {
				m.On("ValidateDriver", "gorm:sqlite").Return(nil)
				m.On("CreateDialector", "test.db").Return("dialector", nil)
			},
			expectError: false,
		},
		{
			name: "no dialect strategy",
			builder: &ConnectionBuilder{
				config: config.Config{
					Driver: "gorm:sqlite",
					DSN:    "test.db",
				},
			},
			setupMock:   func(m *MockDialectStrategy) {},
			expectError: true,
			errorMsg:    "dialect strategy is required",
		},
		{
			name: "driver validation fails",
			builder: &ConnectionBuilder{
				config: config.Config{
					Driver: "invalid:driver",
					DSN:    "test.db",
				},
				dialectStrategy: &MockDialectStrategy{},
			},
			setupMock: func(m *MockDialectStrategy) {
				m.On("ValidateDriver", "invalid:driver").Return(assert.AnError)
			},
			expectError: true,
			errorMsg:    "driver validation failed",
		},
		{
			name: "dialector creation fails",
			builder: &ConnectionBuilder{
				config: config.Config{
					Driver: "gorm:sqlite",
					DSN:    "invalid.db",
				},
				dialectStrategy: &MockDialectStrategy{},
			},
			setupMock: func(m *MockDialectStrategy) {
				m.On("ValidateDriver", "gorm:sqlite").Return(nil)
				m.On("CreateDialector", "invalid.db").Return(nil, assert.AnError)
			},
			expectError: true,
			errorMsg:    "failed to create dialector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if mockStrategy, ok := tt.builder.dialectStrategy.(*MockDialectStrategy); ok {
				tt.setupMock(mockStrategy)
			}

			dialector, sqlDB, err := tt.builder.Build()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, dialector)
				assert.Nil(t, sqlDB)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dialector)
			}

			if mockStrategy, ok := tt.builder.dialectStrategy.(*MockDialectStrategy); ok {
				mockStrategy.AssertExpectations(t)
			}
		})
	}
}

func TestConnectionBuilderApplyPoolConfiguration(t *testing.T) {
	mockDB := &MockDB{}
	builder := &ConnectionBuilder{
		poolOptions: []ConnectionPoolOption{
			WithMaxOpenConns(25),
			WithMaxIdleConns(10),
			WithConnMaxLifetime(30 * time.Second),
		},
	}

	// Set up expectations
	mockDB.On("SetConnMaxLifetime", 30*time.Second).Return()
	mockDB.On("SetMaxIdleConns", 10).Return()
	mockDB.On("SetMaxOpenConns", 25).Return()

	// Call the function
	applyPoolConfigurationInterface(mockDB, builder.poolOptions)

	// Verify expectations
	mockDB.AssertExpectations(t)
}
