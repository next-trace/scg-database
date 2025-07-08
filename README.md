# SCG-Database: A Contract-Driven Database Toolkit for Go

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build](https://github.com/next-trace/scg-database/actions/workflows/ci.yml/badge.svg)](https://github.com/next-trace/scg-database/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/next-trace/scg-database/branch/main/graph/badge.svg)](https://codecov.io/gh/next-trace/scg-database)

SCG-Database is a modern, contract-driven database toolkit for Go applications. It provides a clean, extensible architecture that promotes separation of concerns and makes your applications database-agnostic through a powerful adapter pattern.

## Overview & Philosophy

- Single responsibility: this library does not log. Your app owns logging. We return rich, wrapped errors with context so you can decide what and how to log.
- Open/Closed by design: core is closed for modification but open for extension via adapters. Add new databases/ORMs by implementing small interfaces.
- Minimal surface area: one Config, one Connect entrypoint, one Connection with Repository + QueryBuilder.
- Safe defaults: connection ping on startup, typed errors for common failure modes.

## 🚀 Features

- **Contract-Based Architecture**: Clean interfaces that decouple your business logic from database implementations
- **Multiple Database Support**: Built-in GORM adapter with support for MySQL, PostgreSQL, SQLite
- **Repository Pattern**: Rich repository interface with query building, CRUD operations, and relationships
- **Migration System**: Integrated database migration management with up/down migrations
- **CLI Tools**: Command-line interface for generating models and managing migrations
- **Seeding Support**: Database seeding functionality for development and testing
- **Soft Deletes**: Built-in soft delete support with timestamp management
- **Transaction Support**: Full transaction support with context-aware operations
- **Extensible**: Easy to add custom adapters for other databases or ORMs

## 📦 Installation

```bash
go get github.com/next-trace/scg-database
```

## 🏗️ Architecture

Contract-first, adapter-based design embracing the Open/Closed Principle (OCP):

- Contracts (interfaces) define behavior:
  - Connection, Repository, QueryBuilder, Model, Migrator, Seeder.
- Providers/Adapters implement contracts for specific technologies (e.g., GORM for MySQL/Postgres/SQLite).
- Registry resolves adapter by name at runtime: db.RegisterAdapter(...) and db.GetAdapter(...).
- Single entrypoint: db.Connect(cfg) returns a contract.Connection independent of the underlying database.

Flow:
1. You build a config.Config and call db.Connect(cfg).
2. db.Connect applies options, validates cfg, looks up the adapter from the registry (or uses WithAdapter for injection), and connects.
3. The adapter returns a rich contract.Connection that exposes repositories, transactions, and raw queries.

Extensibility:
- Implement contract.DBAdapter with Connect(*config.Config) and Name().
- Register it with db.RegisterAdapter(adapter, "your:driver").
- Now db.Connect can create connections for your driver without changing core code.

## 🚀 Quick Start

### 1. Basic Usage

```go
package main

import (
    "context"
    "log"

    gormadapter "github.com/next-trace/scg-database/adapter/gorm"
    "github.com/next-trace/scg-database/config"
    "github.com/next-trace/scg-database/db"
)

func main() {
    // Register GORM adapter once at startup
    gormadapter.Register()

    // Configure database connection
    cfg := config.New()
    cfg.Driver = "gorm:sqlite"
    cfg.DSN = "app.db"

    // Connect to database
    conn, err := db.Connect(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Use the connection...
}
```

### 2. Define Models

```go
package user

import (
    "github.com/next-trace/scg-database/contract"
)

type User struct {
    contract.BaseModel
    Name  string
    Email string
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) Relationships() map[string]contract.Relationship {
    return map[string]contract.Relationship{
        // Example relationships:
        // "Profile": contract.NewHasOne(&Profile{}, "user_id", "id"),
        // "Orders": contract.NewHasMany(&Order{}, "user_id", "id"),
        // "Roles": contract.NewBelongsToMany(&Role{}, "user_roles"),
    }
}
```

### 3. Repository Operations

```go
// Create repository
userRepo, err := conn.NewRepository(&user.User{})
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()

// Create user
newUser := &user.User{
    Name:  "John Doe",
    Email: "john@example.com",
}
err = userRepo.Create(ctx, newUser)

// Find user
foundUser, err := userRepo.Find(ctx, newUser.ID)

// Query with conditions
users, err := userRepo.Where("name LIKE ?", "John%").
    OrderBy("created_at", "DESC").
    Limit(10).
    Get(ctx)

// Update user
newUser.Name = "John Smith"
err = userRepo.Update(ctx, newUser)

// Soft delete
err = userRepo.Delete(ctx, newUser)
```

## 🛠️ CLI Tools

The package includes a powerful CLI tool for code generation and migration management.

### Generate Models

```bash
go run ./cmd/scg-db make model User
```

This creates a new model file with the proper structure and contracts.

### Migration Management

```bash
# Create a new migration
go run ./cmd/scg-db migrate make create_users_table

# Run pending migrations
go run ./cmd/scg-db migrate up

# Rollback migrations
go run ./cmd/scg-db migrate down

# Fresh migration (drop all tables and re-run)
go run ./cmd/scg-db migrate fresh
```

### Configuration

Create a `config.yaml` file:

```yaml
database:
  default: gorm:sqlite
  connections:
    gorm:sqlite:
      dsn: app.db
    gorm:mysql:
      dsn: "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
    gorm:postgres:
      dsn: "host=localhost user=username password=password dbname=database port=5432 sslmode=disable"
  paths:
    models: "domain"
    migrations: "database/migrations"
```

## 📚 Advanced Usage

### Observability and Tracing (GORM Plugins)

If you use the GORM adapter directly, you can attach observability plugins
(such as OpenTelemetry tracing) to the underlying *gorm.DB instance.

The adapter exposes a helper to create a raw GORM connection and register
plugins:

```go
package main

import (
    "log"

    gormadapter "github.com/next-trace/scg-database/adapter/gorm"
    "github.com/next-trace/scg-database/config"
    oteltracing "gorm.io/plugin/opentelemetry/tracing"
)

func main() {
    // Build config (Postgres shown as example)
    cfg := &config.Config{
        Driver: gormadapter.GormDriverPostgres, // "gorm:postgres"
        DSN:    "postgres://user:pass@localhost:5432/app?sslmode=disable",
    }

    // Create *gorm.DB and register tracing plugin
    gdb, err := gormadapter.New(cfg, oteltracing.New())
    if err != nil {
        log.Fatal(err)
    }

    // Use gdb as a regular *gorm.DB with tracing enabled
    _ = gdb
}
```

Notes:
- This keeps the core library free of tracing code while enabling you to plug
  in any GORM plugin you need.
- If you prefer to remain fully decoupled from GORM in your microservice,
  continue using db.Connect(cfg) which returns a contract.Connection. In that
  mode, GORM plugins are not injected by default.

### Transactions

```go
err := conn.Transaction(ctx, func(txConn contract.Connection) error {
    userRepo, _ := txConn.NewRepository(&user.User{})
    
    // All operations within this function are part of the transaction
    err := userRepo.Create(ctx, &user.User{Name: "Alice"})
    if err != nil {
        return err // This will rollback the transaction
    }
    
    // More operations...
    return nil // This will commit the transaction
})
```

### Relationships and Eager Loading

```go
// Load user with related data
users, err := userRepo.With("Profile", "Orders").Get(ctx)
```

### Batch Operations

```go
users := []contract.Model{
    &user.User{Name: "Alice", Email: "alice@example.com"},
    &user.User{Name: "Bob", Email: "bob@example.com"},
}

// Create in batches
err := userRepo.CreateInBatches(ctx, users, 100)
```

### Custom Queries

```go
// Raw SQL queries
results, err := conn.Select(ctx, "SELECT * FROM users WHERE created_at > ?", time.Now().AddDate(0, -1, 0))

// Execute statements
result, err := conn.Statement(ctx, "UPDATE users SET status = ? WHERE last_login < ?", "inactive", cutoffDate)
```

## 🧪 Testing

The library includes comprehensive test coverage and powerful testing utilities designed specifically for microservices that need to test against real databases running in Docker containers.

### Basic Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### DatabaseTestSuite for Microservices

The `testing` package provides a comprehensive testing framework for microservices that need to test against real databases:

```go
package user_test

import (
    "testing"
    "time"
    
    "github.com/next-trace/scg-database/testing"
    "github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
    *testing.DatabaseTestSuite
    userService *UserService
}

func (suite *UserServiceTestSuite) SetupSuite() {
    // Configure test database (works with Docker containers)
    cfg := testing.DatabaseTestConfig{
        Driver:          "gorm:postgres", // or mysql, sqlite
        DSN:             "postgres://user:pass@localhost:5432/testdb",
        MigrationsPath:  "/migrations",
        CleanupStrategy: testing.CleanupTruncate,
        Timeout:         30 * time.Second,
    }
    
    suite.DatabaseTestSuite = testing.NewDatabaseTestSuite(&cfg)
    suite.DatabaseTestSuite.SetupSuite()
    
    // Initialize your service with the test database connection
    suite.userService = NewUserService(suite.Connection)
}

func (suite *UserServiceTestSuite) TestCreateUser() {
    // Test your service methods
    user, err := suite.userService.CreateUser("John Doe", "john@example.com")
    suite.NoError(err)
    suite.NotNil(user)
    
    // Use built-in assertion helpers
    suite.AssertRecordExists(&User{}, user.ID)
}

func (suite *UserServiceTestSuite) TestUserRepository() {
    // Create repository for direct testing
    userRepo := suite.CreateRepository(&User{})
    
    // Seed test data
    testUser := &User{Name: "Test User", Email: "test@example.com"}
    suite.SeedData(testUser)
    
    // Test repository operations
    found, err := userRepo.Find(context.Background(), testUser.ID)
    suite.NoError(err)
    suite.Equal("Test User", found.(*User).Name)
}

func (suite *UserServiceTestSuite) TestTransactions() {
    // Test transaction handling
    err := suite.ExecuteInTransaction(func(conn contract.Connection) error {
        userRepo, _ := conn.NewRepository(&User{})
        return userRepo.Create(context.Background(), &User{Name: "TX User"})
    })
    suite.NoError(err)
}

func (suite *UserServiceTestSuite) TearDownTest() {
    // Clean up after each test
    suite.TruncateTable("users")
}

func TestUserServiceSuite(t *testing.T) {
    suite.Run(t, new(UserServiceTestSuite))
}
```

### Testing Features

#### Cleanup Strategies

Choose how to clean up test data between tests:

```go
// Truncate tables after each test (fastest)
CleanupStrategy: testing.CleanupTruncate

// Wrap each test in a transaction and rollback (isolated)
CleanupStrategy: testing.CleanupTransaction

// Drop and recreate database (thorough but slow)
CleanupStrategy: testing.CleanupRecreate

// No cleanup (useful for debugging)
CleanupStrategy: testing.CleanupNone
```

#### Database Readiness Checking

The test suite automatically waits for your Docker database to be ready:

```bash
# Set environment variables for database timeout
export TEST_DB_TIMEOUT=60s
export TEST_MIGRATIONS_PATH=/path/to/migrations
```

#### Built-in Assertions

```go
// Assert record exists
suite.AssertRecordExists(&User{}, userID)

// Assert record doesn't exist
suite.AssertRecordNotExists(&User{}, userID)

// Assert table is empty
suite.AssertTableEmpty("users")

// Seed test data
suite.SeedData(&User{Name: "Test"}, &User{Name: "Test2"})

// Truncate specific tables
suite.TruncateTable("users")

// Get raw SQL connection for advanced operations
rawDB := suite.GetRawConnection()
```

#### Docker Integration Example

Use with Docker Compose for integration testing:

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
    ports:
      - "5432:5432"
    
  test:
    build: .
    depends_on:
      - postgres
    environment:
      TEST_DB_DSN: "postgres://testuser:testpass@postgres:5432/testdb"
    command: go test ./...
```

### Testing Best Practices

1. **Use Real Databases**: Test against the same database type you use in production
2. **Isolate Tests**: Each test should be independent and clean up after itself
3. **Seed Consistently**: Use the `SeedData` method for consistent test data
4. **Test Transactions**: Verify your transaction handling works correctly
5. **Performance Testing**: Use the testing suite for performance benchmarks

## 🔧 Extending the Library

### Custom Adapters

You can create custom adapters by implementing the `contract.DBAdapter` interface:

```go
type MyCustomAdapter struct{}

func (a *MyCustomAdapter) Name() string {
    return "mycustom"
}

func (a *MyCustomAdapter) Connect(cfg *config.Config) (contract.Connection, error) {
    // Implement your connection logic
    return nil, nil
}

// Register your adapter
// Note: RegisterAdapter takes the adapter first, then one or more names.
db.RegisterAdapter(&MyCustomAdapter{}, "mycustom")
```

## 📁 Project Structure

```
scg-database/
├── adapter/gorm/          # GORM database adapter
├── cmd/scg-db/           # CLI application
├── config/               # Configuration management
├── contract/             # Interface definitions
├── db/                   # Core database functionality
├── example/              # Usage examples
├── migration/            # Migration system
├── seeder/              # Database seeding
└── testing/             # Testing utilities
```

## 🎯 Example Application

Run the complete example to see the library in action:

```bash
cd example
chmod +x example.sh
./example.sh
```

This script demonstrates:
- Model generation via CLI
- Migration creation and execution
- Repository operations
- Error handling

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📦 Migrations (CLI and API)

- CLI (recommended for apps):

```bash
# Create a new migration
go run ./cmd/scg-db migrate make create_users_table

# Apply all pending migrations
go run ./cmd/scg-db migrate up

# Rollback last N steps
go run ./cmd/scg-db migrate down --steps 1

# Fresh (drop all and re-run)
go run ./cmd/scg-db migrate fresh
```

- API (useful in tests/tools):

```go
cfg := config.New()
cfg.Driver = "gorm:postgres"
cfg.DSN = "postgres://user:pass@localhost:5432/app?sslmode=disable"
cfg.MigrationsPath = "file://database/migrations"

m, err := migration.NewMigrator(cfg)
if err != nil { /* handle */ }
defer m.Close()

if err := m.Up(); err != nil { /* handle */ }
```

## 🧰 Error Handling (no logging inside)

All public APIs return rich errors. Use errors.Is / errors.As with db error sentinels:

```go
import (
    "errors"
    "github.com/next-trace/scg-database/db"
)

if err != nil {
    if errors.Is(err, db.ErrRecordNotFound) {
        // 404 or ignore
    }
    // log with your logger of choice, e.g., zap/logrus/slog
}
```

## 🧪 Integration Testing with Postgres (Docker)

- docker-compose example:

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U testuser"]
      interval: 2s
      timeout: 2s
      retries: 20
```

- Test suite setup (minimal):

```go
cfg := testing.DatabaseTestConfig{
    Driver:         "gorm:postgres",
    DSN:            "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
    MigrationsPath: "file://./database/migrations",
}
suite := testing.NewDatabaseTestSuite(&cfg)
suite.SetupSuite()
defer suite.TearDownSuite()
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/) (`MAJOR.MINOR.PATCH`).

- **MAJOR**: Breaking API changes
- **MINOR**: New features (backward-compatible)
- **PATCH**: Bug fixes and improvements (backward-compatible)

Consumers should always pin to a specific tag (e.g. `v1.2.3`) to avoid accidental breaking changes.


## 🔐 Security

- Please report vulnerabilities via GitHub Security Advisories or open a private issue.
- Do not disclose security issues publicly until a fix is available.

## 🗒️ Changelog

- Changes are tracked in GitHub Releases.
- A structured CHANGELOG.md may be added following Keep a Changelog.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [GORM](https://gorm.io/) for the default database adapter
- Uses [golang-migrate](https://github.com/golang-migrate/migrate) for migration management
- CLI powered by [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper)
