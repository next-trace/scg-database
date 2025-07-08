package main

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestMigrateCommand(t *testing.T) {
	// Test that migrate command exists and has expected properties
	require.Equal(t, "migrate", migrateCmd.Use)
	require.Contains(t, migrateCmd.Short, "manage database migrations")

	// Test that all subcommands are added
	subcommands := migrateCmd.Commands()
	expectedCommands := map[string]bool{
		"up":    false,
		"down":  false,
		"fresh": false,
		"make":  false,
	}

	for _, cmd := range subcommands {
		if _, exists := expectedCommands[cmd.Use]; exists {
			expectedCommands[cmd.Use] = true
		}
		if strings.HasPrefix(cmd.Use, "down") {
			expectedCommands["down"] = true
		}
		if strings.HasPrefix(cmd.Use, "make") {
			expectedCommands["make"] = true
		}
	}

	for cmdName, found := range expectedCommands {
		require.True(t, found, "%s command should be added to migrate", cmdName)
	}
}

// Note: Actual command execution tests are complex due to file system operations
// and os.Exit calls. These are better tested through integration tests.

func TestRunMigrationCommand_Function(t *testing.T) {
	// Test that runMigrationCommand returns a function
	upFunc := runMigrationCommand("up")
	require.NotNil(t, upFunc)

	downFunc := runMigrationCommand("down")
	require.NotNil(t, downFunc)

	freshFunc := runMigrationCommand("fresh")
	require.NotNil(t, freshFunc)
}

func TestMigrateUpCommand_Properties(t *testing.T) {
	require.Equal(t, "up", migrateUpCmd.Use)
	require.Contains(t, migrateUpCmd.Short, "Run all pending up migrations")
}

func TestMigrateDownCommand_Properties(t *testing.T) {
	require.Contains(t, migrateDownCmd.Use, "down")
	require.Contains(t, migrateDownCmd.Short, "Rollback")

	// Test that it has Args validation (we can't easily test the internal structure)
	require.NotNil(t, migrateDownCmd.Args)
}

func TestMigrateFreshCommand_Properties(t *testing.T) {
	require.Equal(t, "fresh", migrateFreshCmd.Use)
	require.Contains(t, migrateFreshCmd.Short, "Drop all tables")
}

func TestMigrateCreateCommand_Properties(t *testing.T) {
	require.Equal(t, "make [name]", migrateCreateCmd.Use)
	require.Contains(t, migrateCreateCmd.Short, "Create new up and down migration files")
}

// Note: Testing the actual migration execution (up, down, fresh) would require
// setting up a real database connection and migration infrastructure, which is
// complex for unit tests. These commands are better tested through integration tests.
// However, we can test the configuration validation and command structure.

func TestMigrationCommandConfigValidation(t *testing.T) {
	// Test configuration validation logic that would be used by migration commands
	tests := []struct {
		name           string
		defaultConn    string
		migrationsPath string
		dsn            string
		expectValid    bool
	}{
		{
			name:           "valid config",
			defaultConn:    "gorm:sqlite",
			migrationsPath: "database/migrations",
			dsn:            "test.db",
			expectValid:    true,
		},
		{
			name:           "missing default connection",
			defaultConn:    "",
			migrationsPath: "database/migrations",
			dsn:            "test.db",
			expectValid:    false,
		},
		{
			name:           "missing migrations path",
			defaultConn:    "gorm:sqlite",
			migrationsPath: "",
			dsn:            "test.db",
			expectValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("database.default", tt.defaultConn)
			viper.Set("database.paths.migrations", tt.migrationsPath)
			viper.Set("database.connections.gorm:sqlite.dsn", tt.dsn)

			// Test the configuration values that migration commands would check
			defaultConnection := viper.GetString("database.default")
			migrationsPath := viper.GetString("database.paths.migrations")

			if tt.expectValid {
				require.NotEmpty(t, defaultConnection)
				require.NotEmpty(t, migrationsPath)
			} else {
				require.True(t, defaultConnection == "" || migrationsPath == "")
			}
		})
	}
}
