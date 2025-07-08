package gorm

import (
	"fmt"
	"testing"
	"time"

	"github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-database/db"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type (
	// Test Model
	testModel struct {
		ID        uint `gorm:"primaryKey"`
		Name      string
		Email     string `gorm:"uniqueIndex"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}
)

func (m *testModel) PrimaryKey() string                              { return "id" }
func (m *testModel) TableName() string                               { return "test_models" }
func (m *testModel) GetID() any                                      { return m.ID }
func (m *testModel) SetID(id any)                                    { m.ID = id.(uint) }
func (m *testModel) Relationships() map[string]contract.Relationship { return nil }
func (m *testModel) GetDeletedAt() *time.Time {
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		return &t
	}
	return nil
}

func (m *testModel) SetDeletedAt(t *time.Time) {
	if t != nil {
		m.DeletedAt = gorm.DeletedAt{Time: *t, Valid: true}
	} else {
		m.DeletedAt = gorm.DeletedAt{Valid: false}
	}
}
func (m *testModel) GetCreatedAt() time.Time  { return m.CreatedAt }
func (m *testModel) GetUpdatedAt() time.Time  { return m.UpdatedAt }
func (m *testModel) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *testModel) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

// Test Helper to create an isolated DB for each test
func setupTest(t *testing.T) contract.Repository {
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&testModel{}))

	conn := &connection{db: gormDB}
	repo, err := conn.NewRepository(&testModel{})
	require.NoError(t, err)

	t.Cleanup(func() { conn.Close() })
	return repo
}

// --- Test Cases ---
func TestRepository_CreateInBatches(t *testing.T) {
	repo := setupTest(t)
	models := make([]contract.Model, 0, 5)
	for i := range 5 {
		models = append(models, &testModel{Name: fmt.Sprintf("User %d", i), Email: fmt.Sprintf("user%d@example.com", i)})
	}
	require.NoError(t, repo.CreateInBatches(t.Context(), models, 2))

	all, err := repo.Get(t.Context())
	require.NoError(t, err)
	require.Len(t, all, 5)
}

func TestRepository_Create(t *testing.T) {
	repo := setupTest(t)
	user := &testModel{Name: "Alice", Email: "alice@example.com"}
	require.NoError(t, repo.Create(t.Context(), user))
	require.NotZero(t, user.ID)
}

func TestRepository_FirstOrFail(t *testing.T) {
	repo := setupTest(t)
	_, err := repo.FirstOrFail(t.Context())
	require.ErrorIs(t, err, db.ErrRecordNotFound)

	require.NoError(t, repo.Create(t.Context(), &testModel{Name: "Zeke"}))
	found, err := repo.FirstOrFail(t.Context())
	require.NoError(t, err)
	require.Equal(t, "Zeke", found.(*testModel).Name)
}

// --- Query Builder Tests ---
func TestRepository_Where(t *testing.T) {
	repo := setupTest(t)
	user1 := &testModel{Name: "Alice", Email: "alice@example.com"}
	user2 := &testModel{Name: "Bob", Email: "bob@example.com"}
	require.NoError(t, repo.Create(t.Context(), user1, user2))

	found, err := repo.Where("name = ?", "Alice").First(t.Context())
	require.NoError(t, err)
	require.Equal(t, "Alice", found.(*testModel).Name)
}

func TestRepository_OrderBy(t *testing.T) {
	t.Run("ASC order", func(t *testing.T) {
		repo := setupTest(t)
		user1 := &testModel{Name: "Bob", Email: "bob1@example.com"}
		user2 := &testModel{Name: "Alice", Email: "alice1@example.com"}
		require.NoError(t, repo.Create(t.Context(), user1, user2))

		found, err := repo.OrderBy("name", "ASC").First(t.Context())
		require.NoError(t, err)
		require.Equal(t, "Alice", found.(*testModel).Name)
	})

	t.Run("DESC order", func(t *testing.T) {
		repo := setupTest(t)
		user1 := &testModel{Name: "Bob", Email: "bob2@example.com"}
		user2 := &testModel{Name: "Alice", Email: "alice2@example.com"}
		require.NoError(t, repo.Create(t.Context(), user1, user2))

		found, err := repo.OrderBy("name", "DESC").First(t.Context())
		require.NoError(t, err)
		require.Equal(t, "Bob", found.(*testModel).Name)
	})

	t.Run("invalid direction defaults to ASC", func(t *testing.T) {
		repo := setupTest(t)
		user1 := &testModel{Name: "Bob", Email: "bob3@example.com"}
		user2 := &testModel{Name: "Alice", Email: "alice3@example.com"}
		require.NoError(t, repo.Create(t.Context(), user1, user2))

		found, err := repo.OrderBy("name", "invalid").First(t.Context())
		require.NoError(t, err)
		require.Equal(t, "Alice", found.(*testModel).Name)
	})
}

func TestRepository_LimitOffset(t *testing.T) {
	repo := setupTest(t)
	for i := range 5 {
		user := &testModel{Name: fmt.Sprintf("User%d", i), Email: fmt.Sprintf("user%d@example.com", i)}
		require.NoError(t, repo.Create(t.Context(), user))
	}

	results, err := repo.Limit(2).Offset(1).Get(t.Context())
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestRepository_Unscoped(t *testing.T) {
	repo := setupTest(t)
	user := &testModel{Name: "ToDelete", Email: "delete@example.com"}
	require.NoError(t, repo.Create(t.Context(), user))
	require.NoError(t, repo.Delete(t.Context(), user))

	// Should not find soft deleted record (returns nil, not error)
	found, err := repo.Where("name = ?", "ToDelete").First(t.Context())
	require.NoError(t, err)
	require.Nil(t, found)

	// Should find with Unscoped
	found, err = repo.Unscoped().Where("name = ?", "ToDelete").First(t.Context())
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "ToDelete", found.(*testModel).Name)
}

// --- Read Operation Tests ---
func TestRepository_Find(t *testing.T) {
	repo := setupTest(t)
	user := &testModel{Name: "Alice", Email: "alice@example.com"}
	require.NoError(t, repo.Create(t.Context(), user))

	found, err := repo.Find(t.Context(), user.ID)
	require.NoError(t, err)
	require.Equal(t, "Alice", found.(*testModel).Name)

	// Test not found returns nil, nil
	notFound, err := repo.Find(t.Context(), 999)
	require.NoError(t, err)
	require.Nil(t, notFound)
}

func TestRepository_FindOrFail_NotFound(t *testing.T) {
	repo := setupTest(t)
	_, err := repo.FindOrFail(t.Context(), 999)
	require.ErrorIs(t, err, db.ErrRecordNotFound)
}

func TestRepository_Get(t *testing.T) {
	repo := setupTest(t)
	user1 := &testModel{Name: "Alice", Email: "alice@example.com"}
	user2 := &testModel{Name: "Bob", Email: "bob@example.com"}
	require.NoError(t, repo.Create(t.Context(), user1, user2))

	all, err := repo.Get(t.Context())
	require.NoError(t, err)
	require.Len(t, all, 2)
}

func TestRepository_Pluck(t *testing.T) {
	repo := setupTest(t)
	user1 := &testModel{Name: "Alice", Email: "alice@example.com"}
	user2 := &testModel{Name: "Bob", Email: "bob@example.com"}
	require.NoError(t, repo.Create(t.Context(), user1, user2))

	var names []string
	err := repo.Pluck(t.Context(), "name", &names)
	require.NoError(t, err)
	require.Len(t, names, 2)
	require.Contains(t, names, "Alice")
	require.Contains(t, names, "Bob")
}

// --- Write Operation Tests ---
func TestRepository_Create_NilModel(t *testing.T) {
	repo := setupTest(t)
	err := repo.Create(t.Context(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "model cannot be nil")
}

func TestRepository_Create_EmptySlice(t *testing.T) {
	repo := setupTest(t)
	err := repo.Create(t.Context())
	require.NoError(t, err) // Empty slice should be no-op
}

func TestRepository_CreateInBatches_InvalidBatchSize(t *testing.T) {
	repo := setupTest(t)
	models := []contract.Model{&testModel{Name: "Test", Email: "test@example.com"}}
	err := repo.CreateInBatches(t.Context(), models, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "batch size must be positive")
}

func TestRepository_Update(t *testing.T) {
	repo := setupTest(t)
	user := &testModel{Name: "Alice", Email: "alice@example.com"}
	require.NoError(t, repo.Create(t.Context(), user))
	require.NotZero(t, user.ID) // Ensure ID is set

	// Create a new instance with the same ID to avoid GORM state issues
	userToUpdate := &testModel{
		ID:    user.ID,
		Name:  "Alice Updated",
		Email: user.Email,
	}
	require.NoError(t, repo.Update(t.Context(), userToUpdate))

	found, err := repo.Find(t.Context(), user.ID)
	require.NoError(t, err)
	require.Equal(t, "Alice Updated", found.(*testModel).Name)
}

func TestRepository_Update_NilModel(t *testing.T) {
	repo := setupTest(t)
	err := repo.Update(t.Context(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "model cannot be nil")
}

func TestRepository_Delete(t *testing.T) {
	repo := setupTest(t)
	user := &testModel{Name: "ToDelete", Email: "delete@example.com"}
	require.NoError(t, repo.Create(t.Context(), user))

	require.NoError(t, repo.Delete(t.Context(), user))

	// Should not find soft deleted record
	found, err := repo.Find(t.Context(), user.ID)
	require.NoError(t, err)
	require.Nil(t, found)
}

func TestRepository_Delete_NilModel(t *testing.T) {
	repo := setupTest(t)
	err := repo.Delete(t.Context(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "model cannot be nil")
}

func TestRepository_ForceDelete(t *testing.T) {
	repo := setupTest(t)
	user := &testModel{Name: "ToForceDelete", Email: "forcedelete@example.com"}
	require.NoError(t, repo.Create(t.Context(), user))

	require.NoError(t, repo.ForceDelete(t.Context(), user))

	// Should not find even with Unscoped
	found, err := repo.Unscoped().Find(t.Context(), user.ID)
	require.NoError(t, err)
	require.Nil(t, found)
}

func TestRepository_ForceDelete_NilModel(t *testing.T) {
	repo := setupTest(t)
	err := repo.ForceDelete(t.Context(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "model cannot be nil")
}

// --- Upsert Operation Tests ---
func TestRepository_FirstOrCreate(t *testing.T) {
	repo := setupTest(t)
	condition := &testModel{Email: "test@example.com"}
	create := &testModel{Name: "Test User", Email: "test@example.com"}

	// First call should create
	result, err := repo.FirstOrCreate(t.Context(), condition, create)
	require.NoError(t, err)
	require.Equal(t, "Test User", result.(*testModel).Name)

	// Second call should find existing
	result2, err := repo.FirstOrCreate(t.Context(), condition, create)
	require.NoError(t, err)
	require.Equal(t, result.(*testModel).ID, result2.(*testModel).ID)
}

func TestRepository_UpdateOrCreate(t *testing.T) {
	repo := setupTest(t)
	condition := &testModel{Email: "update@example.com"}
	values := map[string]interface{}{"name": "Updated User"}

	// First call should create
	result, err := repo.UpdateOrCreate(t.Context(), condition, values)
	require.NoError(t, err)
	require.Equal(t, "Updated User", result.(*testModel).Name)

	// Second call should update existing
	values["name"] = "Updated Again"
	result2, err := repo.UpdateOrCreate(t.Context(), condition, values)
	require.NoError(t, err)
	require.Equal(t, result.(*testModel).ID, result2.(*testModel).ID)
	require.Equal(t, "Updated Again", result2.(*testModel).Name)
}
