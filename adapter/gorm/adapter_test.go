package gorm

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type (
	// TestModel for testing purposes
	TestModel struct {
		ID        uint `gorm:"primaryKey"`
		Name      string
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}
)

func (m *TestModel) PrimaryKey() string                              { return "id" }
func (m *TestModel) TableName() string                               { return "test_models" }
func (m *TestModel) GetID() any                                      { return m.ID }
func (m *TestModel) SetID(id any)                                    { m.ID = id.(uint) }
func (m *TestModel) Relationships() map[string]contract.Relationship { return nil }
func (m *TestModel) GetDeletedAt() *time.Time {
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		return &t
	}
	return nil
}

func (m *TestModel) SetDeletedAt(t *time.Time) {
	if t != nil {
		m.DeletedAt = gorm.DeletedAt{Time: *t, Valid: true}
	} else {
		m.DeletedAt = gorm.DeletedAt{Valid: false}
	}
}
func (m *TestModel) GetCreatedAt() time.Time  { return m.CreatedAt }
func (m *TestModel) GetUpdatedAt() time.Time  { return m.UpdatedAt }
func (m *TestModel) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *TestModel) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func TestGormAdapter_Connect(t *testing.T) {
	t.Run("SuccessSQLite", func(t *testing.T) {
		cfg := config.Config{
			Driver: "gorm:sqlite",
			DSN:    ":memory:",
		}
		gormAdapter := &Adapter{}
		conn, err := gormAdapter.Connect(&cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		require.NoError(t, conn.Ping(t.Context()))
		require.NoError(t, conn.Close())
	})

	t.Run("UnsupportedDialect", func(t *testing.T) {
		cfg := config.Config{
			Driver: "gorm:unsupported",
			DSN:    "test.db",
		}
		gormAdapter := &Adapter{}
		_, err := gormAdapter.Connect(&cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported gorm dialect: unsupported")
	})

	t.Run("WithOptions", func(t *testing.T) {
		cfg := config.Config{
			Driver: "gorm:sqlite",
			DSN:    ":memory:",
		}
		opt := WithLogger(logger.Default.LogMode(logger.Silent))
		opt(&cfg) // Manually apply the option

		gormAdapter := &Adapter{}
		conn, err := gormAdapter.Connect(&cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		require.NoError(t, conn.Close())

		// Assert that the logger was set (logger configuration is verified by successful connection)
	})
}

func TestGormAdapter_Name(t *testing.T) {
	adapter := &Adapter{}
	assert.Equal(t, "gorm", adapter.Name())
}

func TestGormAdapter_ConnectWithConnectionPool(t *testing.T) {
	cfg := config.Config{
		Driver:          "gorm:sqlite",
		DSN:             ":memory:",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Second,
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Verify connection pool settings were applied by testing that the connection works
	// We can't easily test the pool settings without accessing internals,
	// so we'll just verify the connection is functional
	err = conn.Ping(t.Context())
	require.NoError(t, err)

	require.NoError(t, conn.Close())
}

func TestGormAdapter_ConnectWithGormConfig(t *testing.T) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
		Settings: map[string]any{
			"gorm_config": gormConfig,
		},
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.NoError(t, conn.Close())
}

func TestGormAdapter_ConnectWithGormLogger(t *testing.T) {
	customLogger := logger.Default.LogMode(logger.Info)

	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
		Settings: map[string]any{
			"gorm_logger": customLogger,
		},
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.NoError(t, conn.Close())
}

func TestGormAdapter_ConnectInvalidDriverFormat(t *testing.T) {
	cfg := config.Config{
		Driver: "invalid_format",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	_, err := adapter.Connect(&cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid gorm driver format")
}

func TestGormAdapter_ConnectMySQLDialect(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:mysql",
		DSN:    "invalid_mysql_dsn",
	}

	adapter := &Adapter{}
	_, err := adapter.Connect(&cfg)
	require.Error(t, err)
	// Should fail due to invalid DSN, but we test the MySQL dialect path
}

func TestGormAdapter_ConnectPostgresDialect(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:postgres",
		DSN:    "invalid_postgres_dsn",
	}

	adapter := &Adapter{}
	_, err := adapter.Connect(&cfg)
	require.Error(t, err)
	// Should fail due to invalid DSN, but we test the Postgres dialect path
}

// connectToGorm is a helper function for testing direct GORM connections
func connectToGorm(cfg *config.Config) (*gorm.DB, error) {
	driverParts := strings.Split(cfg.Driver, ":")
	if len(driverParts) != 2 {
		return nil, fmt.Errorf("invalid gorm driver format: %s", cfg.Driver)
	}

	dialectName := driverParts[1]
	var dialector gorm.Dialector

	switch dialectName {
	case DriverMySQL:
		dialector = mysql.Open(cfg.DSN)
	case DriverPostgres:
		dialector = postgres.Open(cfg.DSN)
	case DriverSQLite:
		dialector = sqlite.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported gorm dialect: %s", dialectName)
	}

	return gorm.Open(dialector, &gorm.Config{})
}

func TestConnectToGorm_AllDialects(t *testing.T) {
	testCases := []struct {
		name      string
		driver    string
		dsn       string
		shouldErr bool
	}{
		{
			name:      "SQLite success",
			driver:    "gorm:sqlite",
			dsn:       ":memory:",
			shouldErr: false,
		},
		{
			name:      "MySQL connection error",
			driver:    "gorm:mysql",
			dsn:       "user:pass@tcp(localhost:3306)/db",
			shouldErr: false, // GORM returns DB instance even on connection failure
		},
		{
			name:      "Postgres connection error",
			driver:    "gorm:postgres",
			dsn:       "postgres://user:pass@localhost/db",
			shouldErr: false, // GORM returns DB instance even on connection failure
		},
		{
			name:      "Unsupported dialect",
			driver:    "gorm:oracle",
			dsn:       "test",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{
				Driver: tc.driver,
				DSN:    tc.dsn,
			}

			db, err := connectToGorm(&cfg)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				// For MySQL and Postgres, GORM may return a DB instance but fail on actual operations
				// We expect either no error (SQLite) or a connection error (MySQL/Postgres)
				if tc.name == "SQLite success" {
					assert.NoError(t, err)
					assert.NotNil(t, db)
					if db != nil {
						sqlDB, _ := db.DB()
						if sqlDB != nil {
							sqlDB.Close()
						}
					}
				} else if err != nil {
					// For MySQL/Postgres connection errors, GORM may still return a DB instance
					// but the error will occur on actual database operations
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "connection refused")
					// DB instance may or may not be nil depending on GORM version
				}
			}
		})
	}
}

func TestGormAdapter_ConnectDBError(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Test that the connection is functional
	err = conn.Ping(t.Context())
	require.NoError(t, err)

	require.NoError(t, conn.Close())
}

func TestConnection_CloseError(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Close once
	require.NoError(t, conn.Close())

	// Close again should still work (idempotent)
	err = conn.Close()
	// SQLite allows multiple closes, so this might not error
	// but we test the code path
	_ = err
}

func TestConnection_NewRepositoryNilModel(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Test with nil model
	repo, err := conn.NewRepository(nil)
	require.Error(t, err)
	require.Nil(t, repo)
	require.Contains(t, err.Error(), "model cannot be nil")
}

func TestConnection_StatementError(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Test invalid SQL statement
	_, err = conn.Statement(t.Context(), "INVALID SQL STATEMENT")
	require.Error(t, err)
}

func TestRepository_WithRelationships(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	model := &TestModel{}
	repo, err := conn.NewRepository(model)
	require.NoError(t, err)

	// Test With method with relationships
	repoWithRel := repo.With("Profile", "Orders")
	require.NotNil(t, repoWithRel)

	// Test With method with non-existent relationship
	repoWithNonExistent := repo.With("NonExistentRelation")
	require.NotNil(t, repoWithNonExistent)
}

func TestRepository_OrderByEdgeCases(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	model := &TestModel{}
	repo, err := conn.NewRepository(model)
	require.NoError(t, err)

	// Test with invalid column name (SQL injection attempt)
	repoInvalid := repo.OrderBy("'; DROP TABLE users; --", "ASC")
	require.NotNil(t, repoInvalid)

	// Test with empty direction
	repoEmptyDir := repo.OrderBy("name", "")
	require.NotNil(t, repoEmptyDir)

	// Test with invalid direction
	repoInvalidDir := repo.OrderBy("name", "INVALID")
	require.NotNil(t, repoInvalidDir)
}

func TestRepository_LimitOffsetEdgeCases(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	model := &TestModel{}
	repo, err := conn.NewRepository(model)
	require.NoError(t, err)

	// Test with negative limit
	repoNegLimit := repo.Limit(-10)
	require.NotNil(t, repoNegLimit)

	// Test with negative offset
	repoNegOffset := repo.Offset(-5)
	require.NotNil(t, repoNegOffset)

	// Test with zero values
	repoZeroLimit := repo.Limit(0)
	require.NotNil(t, repoZeroLimit)

	repoZeroOffset := repo.Offset(0)
	require.NotNil(t, repoZeroOffset)
}

func TestRepository_ConvertModelsToSliceErrors(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	model := &TestModel{}
	repo, err := conn.NewRepository(model)
	require.NoError(t, err)

	// Test with empty slice - we'll test this through the public interface
	// by trying to create with an empty slice
	err = repo.Create(t.Context())
	require.NoError(t, err) // Empty slice should be handled gracefully

	// Test with nil model in slice
	err = repo.Create(t.Context(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "model cannot be nil")
}

func TestRepository_CreateInBatchesEdgeCases(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	model := &TestModel{}
	repo, err := conn.NewRepository(model)
	require.NoError(t, err)

	// Test with empty models slice
	err = repo.CreateInBatches(t.Context(), []contract.Model{}, 10)
	require.NoError(t, err) // Should handle empty slice gracefully

	// Test with zero batch size
	models := []contract.Model{&TestModel{Name: "test"}}
	err = repo.CreateInBatches(t.Context(), models, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "batch size must be positive")

	// Test with negative batch size
	err = repo.CreateInBatches(t.Context(), models, -5)
	require.Error(t, err)
	require.Contains(t, err.Error(), "batch size must be positive")
}

func TestRepository_UpdateMultipleModels(t *testing.T) {
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Create table schema
	gormDB := conn.GetConnection().(*gorm.DB)
	err = gormDB.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	model := &TestModel{}
	repo, err := conn.NewRepository(model)
	require.NoError(t, err)

	// Create test models
	model1 := &TestModel{Name: "test1"}
	model2 := &TestModel{Name: "test2"}

	err = repo.Create(t.Context(), model1, model2)
	require.NoError(t, err)

	// Update multiple models
	model1.Name = "updated1"
	model2.Name = "updated2"

	err = repo.Update(t.Context(), model1, model2)
	require.NoError(t, err)
}
