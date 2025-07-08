package utils_test

import (
	"testing"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfigForMigration(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Driver:         "gorm:sqlite",
				DSN:            "test.db",
				MigrationsPath: "migrations",
			},
			expectError: false,
		},
		{
			name: "empty migrations path",
			cfg: &config.Config{
				Driver: "gorm:sqlite",
				DSN:    "test.db",
			},
			expectError: true,
			errorMsg:    "migrations path is required",
		},
		{
			name: "whitespace migrations path",
			cfg: &config.Config{
				Driver:         "gorm:sqlite",
				DSN:            "test.db",
				MigrationsPath: "   ",
			},
			expectError: true,
			errorMsg:    "migrations path is required",
		},
		{
			name: "empty DSN",
			cfg: &config.Config{
				Driver:         "gorm:sqlite",
				MigrationsPath: "migrations",
			},
			expectError: true,
			errorMsg:    "database dsn is required",
		},
		{
			name: "whitespace DSN",
			cfg: &config.Config{
				Driver:         "gorm:sqlite",
				DSN:            "   ",
				MigrationsPath: "migrations",
			},
			expectError: true,
			errorMsg:    "database dsn is required",
		},
		{
			name: "empty driver",
			cfg: &config.Config{
				DSN:            "test.db",
				MigrationsPath: "migrations",
			},
			expectError: true,
			errorMsg:    "database driver is required",
		},
		{
			name: "whitespace driver",
			cfg: &config.Config{
				Driver:         "   ",
				DSN:            "test.db",
				MigrationsPath: "migrations",
			},
			expectError: true,
			errorMsg:    "database driver is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidateConfigForMigration(tt.cfg)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNonNegativeInt(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		fieldName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "positive value",
			value:       10,
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "zero value",
			value:       0,
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "negative value",
			value:       -1,
			fieldName:   "test_field",
			expectError: true,
			errorMsg:    "test_field cannot be negative",
		},
		{
			name:        "large negative value",
			value:       -100,
			fieldName:   "batch_size",
			expectError: true,
			errorMsg:    "batch_size cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidateNonNegativeInt(tt.value, tt.fieldName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		fieldName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "positive value",
			value:       10,
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "zero value",
			value:       0,
			fieldName:   "test_field",
			expectError: true,
			errorMsg:    "test_field must be positive",
		},
		{
			name:        "negative value",
			value:       -1,
			fieldName:   "test_field",
			expectError: true,
			errorMsg:    "test_field must be positive",
		},
		{
			name:        "large positive value",
			value:       1000,
			fieldName:   "limit",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidatePositiveInt(tt.value, tt.fieldName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateColumnName(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected bool
	}{
		{
			name:     "valid simple column",
			column:   "name",
			expected: true,
		},
		{
			name:     "valid column with underscore",
			column:   "user_name",
			expected: true,
		},
		{
			name:     "valid column with numbers",
			column:   "field123",
			expected: true,
		},
		{
			name:     "valid column starting with underscore",
			column:   "_private",
			expected: true,
		},
		{
			name:     "valid qualified column",
			column:   "users.name",
			expected: true,
		},
		{
			name:     "valid complex qualified column",
			column:   "schema.table.column",
			expected: true,
		},
		{
			name:     "invalid column starting with number",
			column:   "123field",
			expected: false,
		},
		{
			name:     "invalid column with special characters",
			column:   "user-name",
			expected: false,
		},
		{
			name:     "invalid column with spaces",
			column:   "user name",
			expected: false,
		},
		{
			name:     "invalid empty column",
			column:   "",
			expected: false,
		},
		{
			name:     "invalid column with SQL injection attempt",
			column:   "name'; DROP TABLE users; --",
			expected: false,
		},
		{
			name:     "invalid qualified column with bad part",
			column:   "users.123name",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidateColumnName(tt.column)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateOrderDirection(t *testing.T) {
	tests := []struct {
		name             string
		direction        string
		defaultDirection string
		expected         string
	}{
		{
			name:             "valid ASC",
			direction:        "ASC",
			defaultDirection: "DESC",
			expected:         "ASC",
		},
		{
			name:             "valid DESC",
			direction:        "DESC",
			defaultDirection: "ASC",
			expected:         "DESC",
		},
		{
			name:             "lowercase asc",
			direction:        "asc",
			defaultDirection: "DESC",
			expected:         "ASC",
		},
		{
			name:             "lowercase desc",
			direction:        "desc",
			defaultDirection: "ASC",
			expected:         "DESC",
		},
		{
			name:             "mixed case ASC",
			direction:        "Asc",
			defaultDirection: "DESC",
			expected:         "ASC",
		},
		{
			name:             "with whitespace",
			direction:        "  DESC  ",
			defaultDirection: "ASC",
			expected:         "DESC",
		},
		{
			name:             "invalid direction",
			direction:        "INVALID",
			defaultDirection: "ASC",
			expected:         "ASC",
		},
		{
			name:             "empty direction",
			direction:        "",
			defaultDirection: "DESC",
			expected:         "DESC",
		},
		{
			name:             "whitespace only direction",
			direction:        "   ",
			defaultDirection: "ASC",
			expected:         "ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidateOrderDirection(tt.direction, tt.defaultDirection)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateModelsSlice(t *testing.T) {
	tests := []struct {
		name        string
		models      []interface{}
		operation   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid models slice",
			models:      []interface{}{"model1", "model2"},
			operation:   "create",
			expectError: false,
		},
		{
			name:        "single valid model",
			models:      []interface{}{"model1"},
			operation:   "update",
			expectError: false,
		},
		{
			name:        "empty models slice",
			models:      []interface{}{},
			operation:   "delete",
			expectError: true,
			errorMsg:    "models slice cannot be empty for delete operation",
		},
		{
			name:        "nil models slice",
			models:      nil,
			operation:   "create",
			expectError: true,
			errorMsg:    "models slice cannot be empty for create operation",
		},
		{
			name:        "models slice with nil value at start",
			models:      []interface{}{nil, "model2"},
			operation:   "update",
			expectError: true,
			errorMsg:    "model at index 0 cannot be nil for update operation",
		},
		{
			name:        "models slice with nil value in middle",
			models:      []interface{}{"model1", nil, "model3"},
			operation:   "create",
			expectError: true,
			errorMsg:    "model at index 1 cannot be nil for create operation",
		},
		{
			name:        "models slice with nil value at end",
			models:      []interface{}{"model1", "model2", nil},
			operation:   "delete",
			expectError: true,
			errorMsg:    "model at index 2 cannot be nil for delete operation",
		},
		{
			name:        "models slice with all nil values",
			models:      []interface{}{nil, nil, nil},
			operation:   "batch_create",
			expectError: true,
			errorMsg:    "model at index 0 cannot be nil for batch_create operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidateModelsSlice(tt.models, tt.operation)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDriverFormat(t *testing.T) {
	tests := []struct {
		name        string
		driver      string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid gorm sqlite driver",
			driver:      "gorm:sqlite",
			expected:    "sqlite",
			expectError: false,
		},
		{
			name:        "valid gorm mysql driver",
			driver:      "gorm:mysql",
			expected:    "mysql",
			expectError: false,
		},
		{
			name:        "valid gorm postgres driver",
			driver:      "gorm:postgres",
			expected:    "postgres",
			expectError: false,
		},
		{
			name:        "invalid driver format - no colon",
			driver:      "gorm_sqlite",
			expected:    "",
			expectError: true,
			errorMsg:    "invalid gorm driver format: gorm_sqlite (expected 'gorm:dialect')",
		},
		{
			name:        "invalid driver format - too many parts",
			driver:      "gorm:sqlite:extra",
			expected:    "",
			expectError: true,
			errorMsg:    "invalid gorm driver format: gorm:sqlite:extra (expected 'gorm:dialect')",
		},
		{
			name:        "invalid driver format - empty",
			driver:      "",
			expected:    "",
			expectError: true,
			errorMsg:    "invalid gorm driver format:  (expected 'gorm:dialect')",
		},
		{
			name:        "invalid driver format - only colon",
			driver:      ":",
			expected:    "",
			expectError: false,
		},
		{
			name:        "invalid driver format - missing dialect",
			driver:      "gorm:",
			expected:    "",
			expectError: false,
		},
		{
			name:        "invalid driver format - missing prefix",
			driver:      ":sqlite",
			expected:    "sqlite",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ValidateDriverFormat(tt.driver)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Equal(t, tt.expected, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
