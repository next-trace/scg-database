package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/next-trace/scg-database/example/domain/user"

	gormadapter "github.com/next-trace/scg-database/adapter/gorm"
	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-database/db"
	"github.com/next-trace/scg-database/migration"
)

// runMigrations handles the migration system demo
func runMigrations(cfg *config.Config) {
	fmt.Println("\n📦 Migration System Demo...")
	if strings.Contains(cfg.Driver, "sqlite") {
		fmt.Println("⚠ SQLite migrations not yet supported by migrator")
		fmt.Println("✔ Skipping migrations for SQLite demo")
		fmt.Println("  (Database tables will be auto-created by GORM)")
		return
	}

	fmt.Println("Running migrations...")
	migrator, err := migration.NewMigrator(cfg)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}
	if err := migrator.Up(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	if sourceErr, dbErr := migrator.Close(); sourceErr != nil || dbErr != nil {
		if sourceErr != nil {
			log.Printf("warning: failed to close migrator source: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("warning: failed to close migrator database: %v", dbErr)
		}
	}
	fmt.Println("✔ Migrations completed successfully")
}

// setupDatabase handles database connection and table creation
func setupDatabase(cfg *config.Config) (contract.Connection, contract.Repository) {
	fmt.Println("\n🔌 Connecting to Database...")
	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	fmt.Println("✔ Database connection established")

	userRepo, err := conn.NewRepository(&user.User{})
	if err != nil {
		log.Fatalf("failed to create user repository: %v", err)
	}

	ctx := context.Background()

	// Create table for SQLite (since we skip migrations)
	if strings.Contains(cfg.Driver, "sqlite") {
		fmt.Println("\n🔧 Creating tables for SQLite...")
		createTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			email TEXT UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)`
		_, err = conn.Statement(ctx, createTableSQL)
		if err != nil {
			log.Fatalf("failed to create users table: %v", err)
		}
		fmt.Println("✔ Tables created successfully")
	}

	return conn, userRepo
}

// demonstrateCreateOperations shows CREATE and batch operations
func demonstrateCreateOperations(ctx context.Context, userRepo contract.Repository) {
	fmt.Println("\n🆕 Creating Users...")
	users := []*user.User{
		{Name: "Alice Johnson", Email: "alice@example.com"},
		{Name: "Bob Smith", Email: "bob@example.com"},
		{Name: "Charlie Brown", Email: "charlie@example.com"},
		{Name: "Diana Prince", Email: "diana@example.com"},
	}

	for _, u := range users {
		err := userRepo.Create(ctx, u)
		if err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Fatalf("failed to create user %s: %v", u.Name, err)
		}
		if err == nil {
			fmt.Printf("  ✔ Created user: %s (ID: %d)\n", u.Name, u.ID)
		} else {
			fmt.Printf("  ⚠ User %s already exists, skipping...\n", u.Name)
		}
	}

	// Batch Create Demo
	fmt.Println("\n📦 Batch Operations Demo...")
	batchUsers := []contract.Model{
		&user.User{Name: "Eve Wilson", Email: "eve@example.com"},
		&user.User{Name: "Frank Miller", Email: "frank@example.com"},
	}

	err := userRepo.CreateInBatches(ctx, batchUsers, 2)
	if err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed") {
		log.Printf("batch create warning: %v", err)
	} else {
		fmt.Println("  ✔ Batch users created successfully")
	}
}

// demonstrateReadOperations shows READ operations with query building
func demonstrateReadOperations(ctx context.Context, userRepo contract.Repository) {
	fmt.Println("\n🔍 Query Building Demo...")

	// Find all users
	allUsers, err := userRepo.Get(ctx)
	if err != nil {
		log.Fatalf("failed to get all users: %v", err)
	}
	fmt.Printf("  ✔ Found %d total users\n", len(allUsers))

	// Query with conditions
	filteredUsers, err := userRepo.Where("name LIKE ?", "%Alice%").
		OrderBy("created_at", "DESC").
		Limit(5).
		Get(ctx)
	if err != nil {
		log.Fatalf("failed to query users: %v", err)
	}
	fmt.Printf("  ✔ Found %d users matching 'Alice'\n", len(filteredUsers))
}

// demonstrateUpdateDeleteOperations shows UPDATE and DELETE operations
func demonstrateUpdateDeleteOperations(ctx context.Context, userRepo contract.Repository) {
	// Find first user for update demo
	firstUser, err := userRepo.OrderBy("id", "ASC").First(ctx)
	if err != nil {
		log.Fatalf("failed to find first user: %v", err)
	}
	if firstUserTyped, ok := firstUser.(*user.User); ok {
		fmt.Printf("  ✔ First user: %s (ID: %d)\n", firstUserTyped.Name, firstUserTyped.ID)

		// UPDATE Operations
		fmt.Println("\n✏️ Update Operations Demo...")
		originalName := firstUserTyped.Name
		firstUserTyped.Name += " (Updated)"

		err = userRepo.Update(ctx, firstUserTyped)
		if err != nil {
			log.Fatalf("failed to update user: %v", err)
		}
		fmt.Printf("  ✔ Updated user name from '%s' to '%s'\n", originalName, firstUserTyped.Name)

		// Verify update
		updatedUser, err := userRepo.Find(ctx, firstUserTyped.ID)
		if err != nil {
			log.Fatalf("failed to find updated user: %v", err)
		}
		if updatedUserTyped, ok := updatedUser.(*user.User); ok {
			fmt.Printf("  ✔ Verified update: %s\n", updatedUserTyped.Name)
		}
	}

	// DELETE Operations (Soft Delete)
	fmt.Println("\n🗑️ Soft Delete Demo...")
	lastUser, err := userRepo.OrderBy("id", "DESC").First(ctx)
	if err != nil {
		log.Fatalf("failed to find last user: %v", err)
	}
	if lastUserTyped, ok := lastUser.(*user.User); ok {
		err = userRepo.Delete(ctx, lastUserTyped)
		if err != nil {
			log.Fatalf("failed to soft delete user: %v", err)
		}
		fmt.Printf("  ✔ Soft deleted user: %s (ID: %d)\n", lastUserTyped.Name, lastUserTyped.ID)

		// Verify soft delete - user should not appear in normal queries
		remainingUsers, err := userRepo.Get(ctx)
		if err != nil {
			log.Fatalf("failed to get remaining users: %v", err)
		}
		fmt.Printf("  ✔ Remaining active users: %d\n", len(remainingUsers))
	}
}

// demonstrateAdvancedRepositoryOperations shows advanced repository features
func demonstrateAdvancedRepositoryOperations(ctx context.Context, userRepo contract.Repository) {
	fmt.Println("\n🔧 Advanced Repository Operations Demo...")

	// FirstOrCreate Demo
	fmt.Println("\n🔍 FirstOrCreate Demo...")
	condition := &user.User{Email: "unique@example.com"}
	createData := &user.User{Name: "Unique User", Email: "unique@example.com"}

	foundOrCreated, err := userRepo.FirstOrCreate(ctx, condition, createData)
	if err != nil {
		log.Printf("FirstOrCreate warning: %v", err)
	} else if foundUser, ok := foundOrCreated.(*user.User); ok {
		fmt.Printf("  ✔ FirstOrCreate result: %s (ID: %d)\n", foundUser.Name, foundUser.ID)
	}

	// UpdateOrCreate Demo
	fmt.Println("\n🔄 UpdateOrCreate Demo...")
	condition2 := &user.User{Email: "updateorcreate@example.com"}
	updateData := map[string]interface{}{"name": "Updated Or Created User"}

	updatedOrCreated, err := userRepo.UpdateOrCreate(ctx, condition2, updateData)
	if err != nil {
		log.Printf("UpdateOrCreate warning: %v", err)
	} else if updatedUser, ok := updatedOrCreated.(*user.User); ok {
		fmt.Printf("  ✔ UpdateOrCreate result: %s (ID: %d)\n", updatedUser.Name, updatedUser.ID)
	}

	// Pluck Demo
	fmt.Println("\n📋 Pluck Demo...")
	var emails []string
	err = userRepo.Pluck(ctx, "email", &emails)
	if err != nil {
		log.Printf("Pluck warning: %v", err)
	} else {
		fmt.Printf("  ✔ Plucked %d email addresses\n", len(emails))
		if len(emails) > 0 {
			fmt.Printf("  ✔ First email: %s\n", emails[0])
		}
	}

	// Unscoped Demo (show soft-deleted records)
	fmt.Println("\n👻 Unscoped Query Demo...")
	unscopedUsers, err := userRepo.Unscoped().Get(ctx)
	if err != nil {
		log.Printf("Unscoped query warning: %v", err)
	} else {
		fmt.Printf("  ✔ Unscoped query found %d users (including soft-deleted)\n", len(unscopedUsers))
	}

	// ForceDelete Demo (hard delete)
	fmt.Println("\n💥 ForceDelete Demo...")
	// Find a soft-deleted user to force delete
	softDeletedUsers, err := userRepo.Unscoped().Where("deleted_at IS NOT NULL").Get(ctx)
	switch {
	case err != nil:
		log.Printf("Finding soft-deleted users warning: %v", err)
	case len(softDeletedUsers) > 0:
		if softDeletedUser, ok := softDeletedUsers[0].(*user.User); ok {
			err = userRepo.ForceDelete(ctx, softDeletedUser)
			if err != nil {
				log.Printf("ForceDelete warning: %v", err)
			} else {
				fmt.Printf("  ✔ Force deleted user: %s (ID: %d)\n", softDeletedUser.Name, softDeletedUser.ID)
			}
		}
	default:
		fmt.Println("  ⚠ No soft-deleted users found to force delete")
	}
}

// demonstrateTransactionDemo shows transaction support
func demonstrateTransactionDemo(ctx context.Context, conn contract.Connection) {
	fmt.Println("\n💳 Transaction Demo...")
	err := conn.Transaction(ctx, func(txConn contract.Connection) error {
		txUserRepo, err := txConn.NewRepository(&user.User{})
		if err != nil {
			return fmt.Errorf("failed to create tx repository: %w", err)
		}

		txUser := &user.User{Name: "Transaction User", Email: "tx@example.com"}
		err = txUserRepo.Create(ctx, txUser)
		if err != nil {
			return fmt.Errorf("failed to create user in transaction: %w", err)
		}

		fmt.Printf("  ✔ Created user in transaction: %s (ID: %d)\n", txUser.Name, txUser.ID)
		return nil
	})
	if err != nil {
		log.Fatalf("transaction failed: %v", err)
	}
	fmt.Println("  ✔ Transaction completed successfully")
}

// demonstrateRawSQLDemo shows raw SQL execution
func demonstrateRawSQLDemo(ctx context.Context, conn contract.Connection) {
	fmt.Println("\n🔧 Raw SQL Demo...")
	results, err := conn.Select(ctx, "SELECT COUNT(*) as user_count FROM users WHERE deleted_at IS NULL")
	if err != nil {
		log.Fatalf("failed to execute raw query: %v", err)
	}
	fmt.Printf("  ✔ Raw query executed, results: %v\n", results)
}

// demonstrateSeedingDemo shows database seeding
func demonstrateSeedingDemo(ctx context.Context, userRepo contract.Repository) {
	fmt.Println("\n🌱 Seeding Demo...")
	seedUsers := []*user.User{
		{Name: "Seed User 1", Email: "seed1@example.com"},
		{Name: "Seed User 2", Email: "seed2@example.com"},
	}

	fmt.Println("  Creating seed users...")
	for _, seedUser := range seedUsers {
		err := userRepo.Create(ctx, seedUser)
		switch {
		case err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed"):
			log.Printf("seeding warning for %s: %v", seedUser.Name, err)
		case err == nil:
			fmt.Printf("  ✔ Seeded user: %s\n", seedUser.Name)
		default:
			fmt.Printf("  ⚠ Seed user %s already exists\n", seedUser.Name)
		}
	}
	fmt.Println("  ✔ Seeding completed successfully")
}

// showFinalStatistics displays final statistics and summary
func showFinalStatistics(ctx context.Context, userRepo contract.Repository, cfg *config.Config) {
	fmt.Println("\n📊 Final Statistics...")
	allFinalUsers, err := userRepo.Get(ctx)
	if err != nil {
		log.Fatalf("failed to get final user count: %v", err)
	}
	fmt.Printf("  ✔ Total active users in database: %d\n", len(allFinalUsers))

	fmt.Println("\n🧹 Cleanup Demo (commented out for preservation)...")
	fmt.Println("  // userRepo.Where(\"email LIKE ?\", \"%@example.com\").Delete(ctx)")
	fmt.Println("  // This would soft-delete all example users")

	fmt.Println("\n🎉 SCG-Database Example Completed Successfully!")
	fmt.Println("===============================================")
	fmt.Println("\nFeatures Demonstrated:")
	fmt.Println("  ✔ Database connection and configuration")
	fmt.Println("  ✔ Migration system")
	fmt.Println("  ✔ Repository pattern with CRUD operations")
	fmt.Println("  ✔ Query building (Where, OrderBy, Limit)")
	fmt.Println("  ✔ Batch operations")
	fmt.Println("  ✔ Advanced repository operations:")
	fmt.Println("    • FirstOrCreate (find or create)")
	fmt.Println("    • UpdateOrCreate (upsert operations)")
	fmt.Println("    • Pluck (extract column values)")
	fmt.Println("    • Unscoped queries (include soft-deleted)")
	fmt.Println("    • ForceDelete (hard delete)")
	fmt.Println("  ✔ Transaction support")
	fmt.Println("  ✔ Soft delete functionality")
	fmt.Println("  ✔ Raw SQL execution")
	fmt.Println("  ✔ Database seeding")
	fmt.Println("  ✔ Error handling and graceful degradation")
	fmt.Printf("  ✔ Database file: %s\n", cfg.DSN)
}

func main() {
	fmt.Println("🚀 SCG-Database Comprehensive Example")
	fmt.Println("=====================================")

	// Register the GORM adapter
	gormadapter.Register()

	// 1. Configuration
	cfg := config.Config{
		Driver:         "gorm:sqlite",
		DSN:            "scg_example.db",
		MigrationsPath: "file://database/migrations",
	}

	// 2. Run Migrations
	runMigrations(&cfg)

	// 3. Connect to Database and setup
	conn, userRepo := setupDatabase(&cfg)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("warning: failed to close database connection: %v", err)
		}
	}()

	ctx := context.Background()

	// 5. Demonstrate CRUD Operations
	fmt.Println("\n📝 CRUD Operations Demo")
	fmt.Println("-----------------------")

	// CREATE Operations
	demonstrateCreateOperations(ctx, userRepo)

	// READ Operations with Query Building
	demonstrateReadOperations(ctx, userRepo)

	// UPDATE and DELETE Operations
	demonstrateUpdateDeleteOperations(ctx, userRepo)

	// 6. Advanced Repository Operations
	demonstrateAdvancedRepositoryOperations(ctx, userRepo)

	// 7. Additional Demos
	demonstrateTransactionDemo(ctx, conn)
	demonstrateRawSQLDemo(ctx, conn)
	demonstrateSeedingDemo(ctx, userRepo)
	showFinalStatistics(ctx, userRepo, &cfg)
}
