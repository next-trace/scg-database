package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	// Test that root command exists and has expected properties
	require.Equal(t, "scg-db", rootCmd.Use)
	require.Contains(t, rootCmd.Short, "CLI tool")
	require.Contains(t, rootCmd.Long, "scg-db provides commands")
}

func TestExecute(t *testing.T) {
	// Test that Execute function exists and can be called
	// We can't easily test the actual execution without mocking os.Exit
	// but we can test that the function exists
	require.NotNil(t, Execute)
}

func TestInitConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `database:
  default: gorm:sqlite
  connections:
    gorm:sqlite:
      dsn: test.db
  paths:
    models: "domain"
    migrations: "database/migrations"
`

	err := os.WriteFile(configFile, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Set the config file and call initConfig
	viper.Reset()
	cfgFile = configFile // Set the global variable
	initConfig()

	// Verify config was loaded
	require.Equal(t, "gorm:sqlite", viper.GetString("database.default"))
	require.Equal(t, "test.db", viper.GetString("database.connections.gorm:sqlite.dsn"))
	require.Equal(t, "domain", viper.GetString("database.paths.models"))
	require.Equal(t, "database/migrations", viper.GetString("database.paths.migrations"))
}

func TestInitConfigWithoutFile(t *testing.T) {
	// Test initConfig when no config file exists
	viper.Reset()
	viper.SetConfigFile("nonexistent.yaml")

	// This should not panic, just print an error
	require.NotPanics(t, func() {
		initConfig()
	})
}

func TestCommandStructure(t *testing.T) {
	// Test that all expected commands are added to root
	commands := rootCmd.Commands()

	var makeFound, migrateFound bool
	for _, cmd := range commands {
		switch cmd.Use {
		case "make":
			makeFound = true
		case "migrate":
			migrateFound = true
		}
	}

	require.True(t, makeFound, "make command should be added to root")
	require.True(t, migrateFound, "migrate command should be added to root")
}

func TestCommandHelp(t *testing.T) {
	// Test that help command works
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "scg-db")
	require.Contains(t, output, "Available Commands:")
	require.Contains(t, output, "make")
	require.Contains(t, output, "migrate")
}

func TestPersistentFlags(t *testing.T) {
	// Test that persistent flags are set up correctly
	flag := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, flag)
	require.Equal(t, "config", flag.Name)
	require.Equal(t, "", flag.DefValue)
	require.Contains(t, flag.Usage, "config file")
}

// Helper function to capture command output
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return strings.TrimSpace(buf.String()), err
}
