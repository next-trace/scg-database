package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg := New()

	// Test default values
	require.Equal(t, 10, cfg.MaxIdleConns)
	require.Equal(t, 100, cfg.MaxOpenConns)
	require.Equal(t, time.Hour, cfg.ConnMaxLifetime)
	require.Equal(t, "domain", cfg.ModelsPath)
	require.NotNil(t, cfg.Settings)
	require.Empty(t, cfg.Settings)

	// Test empty required fields
	require.Empty(t, cfg.Driver)
	require.Empty(t, cfg.DSN)
	require.Empty(t, cfg.MigrationsPath)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Driver: "gorm:sqlite",
				DSN:    "test.db",
			},
			wantErr: false,
		},
		{
			name: "missing driver",
			config: &Config{
				DSN: "test.db",
			},
			wantErr: true,
			errMsg:  "database driver is required",
		},
		{
			name: "missing DSN",
			config: &Config{
				Driver: "gorm:sqlite",
			},
			wantErr: true,
			errMsg:  "database dsn is required",
		},
		{
			name: "empty driver",
			config: &Config{
				Driver: "",
				DSN:    "test.db",
			},
			wantErr: true,
			errMsg:  "database driver is required",
		},
		{
			name: "empty DSN",
			config: &Config{
				Driver: "gorm:sqlite",
				DSN:    "",
			},
			wantErr: true,
			errMsg:  "database dsn is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_FunctionalOptions(t *testing.T) {
	cfg := New()

	// Test that we can apply functional options
	option1 := func(c *Config) {
		c.Driver = "test-driver"
	}

	option2 := func(c *Config) {
		c.DSN = "test-dsn"
		c.MaxIdleConns = 5
	}

	option1(cfg)
	option2(cfg)

	require.Equal(t, "test-driver", cfg.Driver)
	require.Equal(t, "test-dsn", cfg.DSN)
	require.Equal(t, 5, cfg.MaxIdleConns)
	require.Equal(t, 100, cfg.MaxOpenConns) // Should remain default
}

func TestConfig_Settings(t *testing.T) {
	cfg := New()

	// Test that Settings map is initialized and can be used
	cfg.Settings["custom_key"] = "custom_value"
	cfg.Settings["timeout"] = 30

	require.Equal(t, "custom_value", cfg.Settings["custom_key"])
	require.Equal(t, 30, cfg.Settings["timeout"])
	require.Len(t, cfg.Settings, 2)
}

func TestConfig_AllFields(t *testing.T) {
	cfg := &Config{
		Driver:          "gorm:mysql",
		DSN:             "user:pass@tcp(localhost:3306)/dbname",
		MigrationsPath:  "file://migrations",
		ModelsPath:      "models",
		MaxIdleConns:    5,
		MaxOpenConns:    50,
		ConnMaxLifetime: 30 * time.Minute,
		Adapter:         "custom-adapter",
		Settings:        map[string]any{"key": "value"},
	}

	require.NoError(t, cfg.Validate())
	require.Equal(t, "gorm:mysql", cfg.Driver)
	require.Equal(t, "user:pass@tcp(localhost:3306)/dbname", cfg.DSN)
	require.Equal(t, "file://migrations", cfg.MigrationsPath)
	require.Equal(t, "models", cfg.ModelsPath)
	require.Equal(t, 5, cfg.MaxIdleConns)
	require.Equal(t, 50, cfg.MaxOpenConns)
	require.Equal(t, 30*time.Minute, cfg.ConnMaxLifetime)
	require.Equal(t, "custom-adapter", cfg.Adapter)
	require.Equal(t, "value", cfg.Settings["key"])
}
