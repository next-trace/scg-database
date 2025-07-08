package gorm

import (
	"testing"

	"github.com/next-trace/scg-database/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestQueryBuilder tests the GORM query builder implementation
func TestQueryBuilder(t *testing.T) {
	// Register the GORM adapter and query builder factory
	Register()

	// Setup test database
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Create table
	gormDB := conn.GetConnection().(*gorm.DB)
	err = gormDB.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Create repository
	model := &TestModel{}
	repo, err := conn.NewRepository(model)
	require.NoError(t, err)

	// Get query builder
	qb := repo.QueryBuilder()
	require.NotNil(t, qb)

	t.Run("Basic query building", func(t *testing.T) {
		// Test chaining methods
		qb2 := qb.Where("name = ?", "test").
			OrderBy("id", "ASC").
			Limit(10)

		assert.NotNil(t, qb2)
	})

	t.Run("Count functionality", func(t *testing.T) {
		// Create some test data
		testModel := &TestModel{Name: "test"}
		err := repo.Create(t.Context(), testModel)
		require.NoError(t, err)

		// Test count
		count, err := qb.Count(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Exists functionality", func(t *testing.T) {
		exists, err := qb.Where("name = ?", "test").Exists(t.Context())
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = qb.Where("name = ?", "nonexistent").Exists(t.Context())
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Clone functionality", func(t *testing.T) {
		qb2 := qb.Where("name = ?", "test")
		qb3 := qb2.Clone()

		assert.NotNil(t, qb3)
		// Both should be independent
		qb4 := qb2.Where("id > ?", 0)
		qb5 := qb3.Where("id < ?", 100)

		assert.NotNil(t, qb4)
		assert.NotNil(t, qb5)
	})

	t.Run("Reset functionality", func(t *testing.T) {
		qb2 := qb.Where("name = ?", "test").OrderBy("id", "DESC")
		qb3 := qb2.Reset()

		assert.NotNil(t, qb3)
	})
}

func TestQueryBuilderFactory(t *testing.T) {
	factory := &GormQueryBuilderFactory{}

	t.Run("Factory name", func(t *testing.T) {
		assert.Equal(t, "gorm", factory.Name())
	})

	t.Run("Create query builder", func(t *testing.T) {
		// Setup test database
		cfg := config.Config{
			Driver: "gorm:sqlite",
			DSN:    ":memory:",
		}

		adapter := &Adapter{}
		conn, err := adapter.Connect(&cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()

		gormDB := conn.GetConnection().(*gorm.DB)
		model := &TestModel{}

		qb := factory.NewQueryBuilder(model, gormDB)
		assert.NotNil(t, qb)
	})

	t.Run("Invalid connection type", func(t *testing.T) {
		model := &TestModel{}

		assert.Panics(t, func() {
			factory.NewQueryBuilder(model, "invalid_connection")
		})
	})
}

func TestQueryBuilderMethods(t *testing.T) {
	// Setup test database
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	gormDB := conn.GetConnection().(*gorm.DB)
	err = gormDB.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	model := &TestModel{}
	qb := newGormQueryBuilder(model, gormDB)

	t.Run("Select method", func(t *testing.T) {
		qb2 := qb.Select("name", "id")
		assert.NotNil(t, qb2)
	})

	t.Run("Where methods", func(t *testing.T) {
		qb2 := qb.Where("name = ?", "test")
		assert.NotNil(t, qb2)

		qb3 := qb.WhereIn("id", []any{1, 2, 3})
		assert.NotNil(t, qb3)

		qb4 := qb.WhereNotIn("id", []any{4, 5, 6})
		assert.NotNil(t, qb4)

		qb5 := qb.WhereNull("deleted_at")
		assert.NotNil(t, qb5)

		qb6 := qb.WhereNotNull("created_at")
		assert.NotNil(t, qb6)

		qb7 := qb.WhereBetween("id", 1, 10)
		assert.NotNil(t, qb7)

		qb8 := qb.OrWhere("name = ?", "other")
		assert.NotNil(t, qb8)
	})

	t.Run("Join methods", func(t *testing.T) {
		qb2 := qb.Join("profiles", "profiles.user_id = users.id")
		assert.NotNil(t, qb2)

		qb3 := qb.LeftJoin("orders", "orders.user_id = users.id")
		assert.NotNil(t, qb3)

		qb4 := qb.RightJoin("roles", "roles.id = users.role_id")
		assert.NotNil(t, qb4)

		qb5 := qb.InnerJoin("companies", "companies.id = users.company_id")
		assert.NotNil(t, qb5)
	})

	t.Run("Ordering and grouping", func(t *testing.T) {
		qb2 := qb.OrderBy("name", "ASC")
		assert.NotNil(t, qb2)

		qb3 := qb.OrderBy("name", "invalid_direction") // Should default to ASC
		assert.NotNil(t, qb3)

		qb4 := qb.GroupBy("status", "type")
		assert.NotNil(t, qb4)

		qb5 := qb.Having("COUNT(*) > ?", 5)
		assert.NotNil(t, qb5)
	})

	t.Run("Limiting and pagination", func(t *testing.T) {
		qb2 := qb.Limit(10)
		assert.NotNil(t, qb2)

		qb3 := qb.Limit(-5) // Should return same builder
		assert.Equal(t, qb, qb3)

		qb4 := qb.Offset(20)
		assert.NotNil(t, qb4)

		qb5 := qb.Offset(-10) // Should return same builder
		assert.Equal(t, qb, qb5)
	})

	t.Run("Relationships", func(t *testing.T) {
		qb2 := qb.With("Profile", "Orders")
		assert.NotNil(t, qb2)

		qb3 := qb.WithCount("Orders") // Simplified implementation
		assert.NotNil(t, qb3)
	})

	t.Run("Scopes", func(t *testing.T) {
		qb2 := qb.Scoped()
		assert.NotNil(t, qb2)

		qb3 := qb.Unscoped()
		assert.NotNil(t, qb3)
	})

	t.Run("Raw queries", func(t *testing.T) {
		qb2 := qb.Raw("SELECT * FROM test_models WHERE name = ?", "test")
		assert.NotNil(t, qb2)
	})

	t.Run("ToSQL method", func(t *testing.T) {
		sql, args, err := qb.ToSQL()
		assert.Error(t, err) // Not fully implemented
		assert.Empty(t, sql)
		assert.Nil(t, args)
	})
}

func TestQueryBuilderExecution(t *testing.T) {
	// Setup test database
	cfg := config.Config{
		Driver: "gorm:sqlite",
		DSN:    ":memory:",
	}

	adapter := &Adapter{}
	conn, err := adapter.Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	gormDB := conn.GetConnection().(*gorm.DB)
	err = gormDB.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	model := &TestModel{}
	qb := newGormQueryBuilder(model, gormDB)

	ctx := t.Context()

	t.Run("Create and find", func(t *testing.T) {
		testModel := &TestModel{Name: "query_test"}
		err := qb.Create(ctx, testModel)
		assert.NoError(t, err)

		var result TestModel
		err = qb.Where("name = ?", "query_test").First(ctx, &result)
		assert.NoError(t, err)
		assert.Equal(t, "query_test", result.Name)
	})

	t.Run("Update", func(t *testing.T) {
		// Create fresh query builder for update
		updateQB := newGormQueryBuilder(model, gormDB)
		err := updateQB.Where("name = ?", "query_test").Update(ctx, map[string]any{"name": "updated_test"})
		assert.NoError(t, err)

		// Create fresh query builder for finding updated record
		findQB := newGormQueryBuilder(model, gormDB)
		var result TestModel
		err = findQB.Where("name = ?", "updated_test").First(ctx, &result)
		assert.NoError(t, err)
		assert.Equal(t, "updated_test", result.Name)
	})

	t.Run("Get multiple", func(t *testing.T) {
		// Create another record using fresh query builder
		createQB := newGormQueryBuilder(model, gormDB)
		testModel2 := &TestModel{Name: "query_test2"}
		err := createQB.Create(ctx, testModel2)
		assert.NoError(t, err)

		// Get all records using fresh query builder
		getQB := newGormQueryBuilder(model, gormDB)
		var results []TestModel
		err = getQB.Get(ctx, &results)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)
	})

	t.Run("Delete", func(t *testing.T) {
		// Create fresh query builder for delete
		deleteQB := newGormQueryBuilder(model, gormDB)
		err := deleteQB.Where("name = ?", "updated_test").Delete(ctx)
		assert.NoError(t, err)

		// Create fresh query builder for count check
		countQB := newGormQueryBuilder(model, gormDB)
		count, err := countQB.Where("name = ?", "updated_test").Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Exec raw SQL", func(t *testing.T) {
		// Create fresh query builder for raw SQL execution
		execQB := newGormQueryBuilder(model, gormDB)
		err := execQB.Exec(ctx, "UPDATE test_models SET name = ? WHERE name = ?", "exec_test", "query_test2")
		assert.NoError(t, err)

		// Create fresh query builder for count check
		countQB := newGormQueryBuilder(model, gormDB)
		count, err := countQB.Where("name = ?", "exec_test").Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}
