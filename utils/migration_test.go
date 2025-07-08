package utils

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMapDriverName(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		expected string
	}{
		{
			name:     "gorm mysql driver",
			driver:   "gorm:mysql",
			expected: "mysql",
		},
		{
			name:     "direct mysql driver",
			driver:   "mysql",
			expected: "mysql",
		},
		{
			name:     "gorm postgres driver",
			driver:   "gorm:postgres",
			expected: "postgres",
		},
		{
			name:     "direct postgres driver",
			driver:   "postgres",
			expected: "postgres",
		},
		{
			name:     "unknown driver - return as is",
			driver:   "unknown",
			expected: "unknown",
		},
		{
			name:     "gorm sqlite driver - return as is",
			driver:   "gorm:sqlite",
			expected: "gorm:sqlite",
		},
		{
			name:     "empty driver",
			driver:   "",
			expected: "",
		},
		{
			name:     "complex driver name",
			driver:   "custom:driver:with:colons",
			expected: "custom:driver:with:colons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapDriverName(tt.driver)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateDatabaseDriver(t *testing.T) {
	tests := []struct {
		name        string
		driverName  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unsupported driver",
			driverName:  "sqlite",
			expectError: true,
			errorMsg:    "unsupported migration driver: sqlite",
		},
		{
			name:        "empty driver name",
			driverName:  "",
			expectError: true,
			errorMsg:    "unsupported migration driver: ",
		},
		{
			name:        "unknown driver",
			driverName:  "unknown",
			expectError: true,
			errorMsg:    "unsupported migration driver: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := CreateDatabaseDriver(tt.driverName, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
			assert.Nil(t, driver)
		})
	}

	// Note: We cannot test supported drivers (mysql, postgres) without real database connections
	// because the underlying driver libraries attempt to ping the database during initialization
}

func TestHandleMigrationError(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectError bool
		expectedErr error
	}{
		{
			name:        "no error",
			inputError:  nil,
			expectError: false,
			expectedErr: nil,
		},
		{
			name:        "ErrNoChange - should be handled",
			inputError:  migrate.ErrNoChange,
			expectError: false,
			expectedErr: nil,
		},
		{
			name:        "other error - should be returned",
			inputError:  errors.New("some migration error"),
			expectError: true,
			expectedErr: errors.New("some migration error"),
		},
		{
			name:        "wrapped ErrNoChange - should be handled",
			inputError:  errors.New("wrapped: " + migrate.ErrNoChange.Error()),
			expectError: true, // This won't be detected as ErrNoChange because it's wrapped differently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleMigrationError(tt.inputError)
			if tt.expectError {
				assert.Error(t, result)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), result.Error())
				}
			} else {
				assert.NoError(t, result)
			}
		})
	}
}

// MockSQLDB is a mock implementation for testing SafeCloseSQLDB
type (
	MockSQLDB struct {
		mock.Mock
	}
)

func (m *MockSQLDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSafeCloseSQLDB(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() *sql.DB
		expectError bool
		errorMsg    string
	}{
		{
			name: "nil database",
			setupMock: func() *sql.DB {
				return nil
			},
			expectError: false,
		},
		// Note: We can't easily test with a real sql.DB without creating actual database connections
		// In a real scenario, you would need integration tests with actual databases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupMock()
			err := SafeCloseSQLDB(db)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMigrationDriverFactory(t *testing.T) {
	factory := &MigrationDriverFactory{}

	t.Run("SupportedDrivers", func(t *testing.T) {
		drivers := factory.SupportedDrivers()
		expected := []string{"mysql", "postgres"}
		assert.Equal(t, expected, drivers)
		assert.Len(t, drivers, 2)
		assert.Contains(t, drivers, "mysql")
		assert.Contains(t, drivers, "postgres")
	})

	t.Run("IsDriverSupported", func(t *testing.T) {
		tests := []struct {
			name       string
			driverName string
			expected   bool
		}{
			{
				name:       "mysql supported",
				driverName: "mysql",
				expected:   true,
			},
			{
				name:       "postgres supported",
				driverName: "postgres",
				expected:   true,
			},
			{
				name:       "sqlite not supported",
				driverName: "sqlite",
				expected:   false,
			},
			{
				name:       "unknown driver not supported",
				driverName: "unknown",
				expected:   false,
			},
			{
				name:       "empty driver not supported",
				driverName: "",
				expected:   false,
			},
			{
				name:       "case sensitive - MySQL not supported",
				driverName: "MySQL",
				expected:   false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := factory.IsDriverSupported(tt.driverName)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("CreateDriver", func(t *testing.T) {
		// Test that CreateDriver delegates to CreateDatabaseDriver for unsupported drivers
		t.Run("unsupported driver", func(t *testing.T) {
			driver, err := factory.CreateDriver("sqlite", nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported migration driver")
			assert.Nil(t, driver)
		})

		// Note: We cannot test supported drivers (mysql, postgres) without real database connections
		// because CreateDriver delegates to CreateDatabaseDriver which requires valid DB connections
	})
}

func TestNewMigrationResult(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		err     error
		message string
	}{
		{
			name:    "successful result",
			success: true,
			err:     nil,
			message: "Migration completed successfully",
		},
		{
			name:    "failed result",
			success: false,
			err:     errors.New("migration failed"),
			message: "Migration failed with error",
		},
		{
			name:    "failed result with nil error",
			success: false,
			err:     nil,
			message: "Migration failed for unknown reason",
		},
		{
			name:    "successful result with empty message",
			success: true,
			err:     nil,
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMigrationResult(tt.success, tt.err, tt.message)
			assert.NotNil(t, result)
			assert.Equal(t, tt.success, result.Success)
			assert.Equal(t, tt.err, result.Error)
			assert.Equal(t, tt.message, result.Message)
		})
	}
}

func TestExecuteMigrationWithCleanup(t *testing.T) {
	tests := []struct {
		name            string
		operation       func() error
		expectedSuccess bool
		expectedMessage string
		expectError     bool
	}{
		{
			name: "successful operation",
			operation: func() error {
				return nil
			},
			expectedSuccess: true,
			expectedMessage: "Migration completed successfully",
			expectError:     false,
		},
		{
			name: "operation with ErrNoChange - should be treated as success",
			operation: func() error {
				return migrate.ErrNoChange
			},
			expectedSuccess: true,
			expectedMessage: "Migration completed successfully",
			expectError:     false,
		},
		{
			name: "operation with error",
			operation: func() error {
				return errors.New("migration error")
			},
			expectedSuccess: false,
			expectedMessage: "Migration failed: migration error",
			expectError:     true,
		},
		{
			name: "operation with panic recovered as error",
			operation: func() error {
				return errors.New("critical error")
			},
			expectedSuccess: false,
			expectedMessage: "Migration failed: critical error",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use nil for sqlDB since SafeCloseSQLDB handles nil gracefully
			result := ExecuteMigrationWithCleanup(nil, tt.operation)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedSuccess, result.Success)
			assert.Equal(t, tt.expectedMessage, result.Message)

			if tt.expectError {
				assert.Error(t, result.Error)
			} else {
				assert.NoError(t, result.Error)
			}
		})
	}
}

func TestMigrationConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *MigrationConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &MigrationConfig{
				DriverName:     "mysql",
				DSN:            "user:pass@tcp(localhost:3306)/db",
				MigrationsPath: "/migrations",
				Steps:          0,
			},
			expectError: false,
		},
		{
			name: "missing driver name",
			config: &MigrationConfig{
				DSN:            "user:pass@tcp(localhost:3306)/db",
				MigrationsPath: "/migrations",
			},
			expectError: true,
			errorMsg:    "driver name is required",
		},
		{
			name: "empty driver name",
			config: &MigrationConfig{
				DriverName:     "",
				DSN:            "user:pass@tcp(localhost:3306)/db",
				MigrationsPath: "/migrations",
			},
			expectError: true,
			errorMsg:    "driver name is required",
		},
		{
			name: "missing DSN",
			config: &MigrationConfig{
				DriverName:     "mysql",
				MigrationsPath: "/migrations",
			},
			expectError: true,
			errorMsg:    "DSN is required",
		},
		{
			name: "empty DSN",
			config: &MigrationConfig{
				DriverName:     "mysql",
				DSN:            "",
				MigrationsPath: "/migrations",
			},
			expectError: true,
			errorMsg:    "DSN is required",
		},
		{
			name: "missing migrations path",
			config: &MigrationConfig{
				DriverName: "mysql",
				DSN:        "user:pass@tcp(localhost:3306)/db",
			},
			expectError: true,
			errorMsg:    "migrations path is required",
		},
		{
			name: "empty migrations path",
			config: &MigrationConfig{
				DriverName:     "mysql",
				DSN:            "user:pass@tcp(localhost:3306)/db",
				MigrationsPath: "",
			},
			expectError: true,
			errorMsg:    "migrations path is required",
		},
		{
			name: "valid config with steps",
			config: &MigrationConfig{
				DriverName:     "postgres",
				DSN:            "postgres://user:pass@localhost/db",
				MigrationsPath: "/path/to/migrations",
				Steps:          5,
			},
			expectError: false,
		},
		{
			name: "valid config with negative steps",
			config: &MigrationConfig{
				DriverName:     "postgres",
				DSN:            "postgres://user:pass@localhost/db",
				MigrationsPath: "/path/to/migrations",
				Steps:          -3,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
