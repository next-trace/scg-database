// Package main provides the SCG database CLI tool with commands for managing
// database migrations including up, down, fresh, and create operations.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	// Import GORM adapter to register it.
	_ "github.com/next-trace/scg-database/adapter/gorm"
	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/migration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// validatePath ensures the path is safe and doesn't contain directory traversal attempts
func validatePath(basePath, targetPath string) error {
	// Clean the paths to resolve any .. or . components
	cleanBase := filepath.Clean(basePath)
	cleanTarget := filepath.Clean(targetPath)

	// Get absolute paths
	absBase, err := filepath.Abs(cleanBase)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for base: %w", err)
	}

	absTarget, err := filepath.Abs(cleanTarget)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for target: %w", err)
	}

	// Check if target is within the base directory
	relPath, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// If the relative path starts with "..", it's trying to escape the base directory
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("path traversal attempt detected: %s", targetPath)
	}

	return nil
}

// secureCreateFile creates a file safely after validating the path
func secureCreateFile(basePath, targetPath string) (*os.File, error) {
	// Validate the path first
	if err := validatePath(basePath, targetPath); err != nil {
		return nil, err
	}

	// Clean the target path to remove any remaining path traversal attempts
	cleanPath := filepath.Clean(targetPath)

	// Create the file using the cleaned path
	return os.Create(cleanPath)
}

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Commands to manage database migrations",
	}

	migrateUpCmd = &cobra.Command{
		Use:   "up",
		Short: "Run all pending up migrations",
		Run:   runMigrationCommand("up"),
	}

	migrateDownCmd = &cobra.Command{
		Use:   "down [steps]",
		Short: "Rollback the last migration, or N steps",
		Args:  cobra.MaximumNArgs(1),
		Run:   runMigrationCommand("down"),
	}

	migrateFreshCmd = &cobra.Command{
		Use:   "fresh",
		Short: "Drop all tables and re-run all migrations",
		Run:   runMigrationCommand("fresh"),
	}

	migrateCreateCmd = &cobra.Command{
		Use:   "make [name]",
		Short: "Create new up and down migration files",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd
			name := args[0]
			// Viper returns the raw path, e.g., "database/migrations"
			migrationsPath := viper.GetString("database.paths.migrations")
			if migrationsPath == "" {
				fmt.Println("Error: 'database.paths.migrations' path not set in config.")
				os.Exit(1)
			}

			// Ensure the directory exists
			if err := os.MkdirAll(migrationsPath, 0o750); err != nil {
				fmt.Printf("Error: could not create migrations directory: %v\n", err)
				os.Exit(1)
			}

			timestamp := time.Now().UTC().Format("20060102150405")
			baseFileName := fmt.Sprintf("%s_%s", timestamp, name)

			upFile := filepath.Join(migrationsPath, baseFileName+".up.sql")
			downFile := filepath.Join(migrationsPath, baseFileName+".down.sql")

			// Create migration files securely
			if _, err := secureCreateFile(migrationsPath, upFile); err != nil {
				fmt.Printf("Error: could not create up migration file: %v\n", err)
				os.Exit(1)
			}

			if _, err := secureCreateFile(migrationsPath, downFile); err != nil {
				fmt.Printf("Error: could not create down migration file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Created migration: %s\n", upFile)
			fmt.Printf("Created migration: %s\n", downFile)
		},
	}
)

func init() {
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateFreshCmd, migrateCreateCmd)
}

func runMigrationCommand(direction string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		_ = cmd
		defaultConnection := viper.GetString("database.default")
		if defaultConnection == "" {
			fmt.Println("Error: 'database.default' connection not set in config.")
			os.Exit(1)
		}

		migrationsPath := viper.GetString("database.paths.migrations")
		if migrationsPath == "" {
			fmt.Println("Error: 'database.paths.migrations' path not set in config.")
			os.Exit(1)
		}

		// The path for golang-migrate must be a source URL, like "file://path/to/migrations"
		if !strings.HasPrefix(migrationsPath, "file://") {
			absPath, err := filepath.Abs(migrationsPath)
			if err != nil {
				fmt.Printf("Error: could not get absolute path for migrations: %v\n", err)
				os.Exit(1)
			}
			migrationsPath = "file://" + absPath
		}

		cfg := config.Config{
			Driver:         defaultConnection,
			DSN:            viper.GetString(fmt.Sprintf("database.connections.%s.dsn", defaultConnection)),
			MigrationsPath: migrationsPath,
		}

		migrator, err := migration.NewMigrator(&cfg)
		if err != nil {
			fmt.Printf("Error creating migrator: %v\n", err)
			os.Exit(1)
		}
		defer migrator.Close()

		var opErr error
		switch direction {
		case "up":
			fmt.Println("Running up migrations...")
			opErr = migrator.Up()
		case "down":
			steps := 1
			if len(args) > 0 {
				var errAtoi error
				steps, errAtoi = strconv.Atoi(args[0])
				if errAtoi != nil || steps <= 0 {
					fmt.Println("Error: [steps] must be a positive integer.")
					os.Exit(1)
				}
			}
			fmt.Printf("Rolling back %d migration(s)...\n", steps)
			opErr = migrator.Down(steps)
		case "fresh":
			fmt.Println("Dropping all tables and re-running migrations...")
			opErr = migrator.Fresh()
		}

		if opErr != nil {
			fmt.Printf("Migration command failed: %v\n", opErr)
			os.Exit(1)
		}
		fmt.Println("Migration command completed successfully.")
	}
}
