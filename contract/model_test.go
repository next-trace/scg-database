package contract

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBaseModel tests the BaseModel implementation
func TestBaseModel(t *testing.T) {
	t.Run("NewBaseModel", func(t *testing.T) {
		model := NewBaseModel()
		assert.NotNil(t, model)
		assert.Nil(t, model.ID)
	})

	t.Run("Model interface methods", func(t *testing.T) {
		model := NewBaseModel()

		// Test PrimaryKey
		assert.Equal(t, "id", model.PrimaryKey())

		// Test TableName (should be empty for base model)
		assert.Equal(t, "", model.TableName())

		// Test ID methods
		assert.Nil(t, model.GetID())
		model.SetID(123)
		assert.Equal(t, 123, model.GetID())

		// Test Relationships (should be nil for base model)
		assert.Nil(t, model.Relationships())
	})
}

// TestTimestampedModel tests the TimestampedModel implementation
func TestTimestampedModel(t *testing.T) {
	t.Run("NewTimestampedModel", func(t *testing.T) {
		model := NewTimestampedModel()
		assert.NotNil(t, model)
		assert.Nil(t, model.ID)
		assert.True(t, model.CreatedAt.IsZero())
		assert.True(t, model.UpdatedAt.IsZero())
	})

	t.Run("Timestamps interface methods", func(t *testing.T) {
		model := NewTimestampedModel()

		// Test initial state
		assert.True(t, model.GetCreatedAt().IsZero())
		assert.True(t, model.GetUpdatedAt().IsZero())

		// Test setting timestamps
		createdAt := time.Now()
		updatedAt := time.Now().Add(time.Hour)

		model.SetCreatedAt(createdAt)
		model.SetUpdatedAt(updatedAt)

		assert.Equal(t, createdAt, model.GetCreatedAt())
		assert.Equal(t, updatedAt, model.GetUpdatedAt())
	})
}

// TestSoftDeletableModel tests the SoftDeletableModel implementation
func TestSoftDeletableModel(t *testing.T) {
	t.Run("NewSoftDeletableModel", func(t *testing.T) {
		model := NewSoftDeletableModel()
		assert.NotNil(t, model)
		assert.Nil(t, model.ID)
		assert.True(t, model.CreatedAt.IsZero())
		assert.True(t, model.UpdatedAt.IsZero())
		assert.Nil(t, model.DeletedAt)
	})

	t.Run("SoftDelete interface methods", func(t *testing.T) {
		model := NewSoftDeletableModel()

		// Test initial state
		assert.Nil(t, model.GetDeletedAt())

		// Test setting deleted at
		now := time.Now()
		model.SetDeletedAt(&now)
		assert.NotNil(t, model.GetDeletedAt())
		assert.Equal(t, now, *model.GetDeletedAt())

		// Test setting to nil
		model.SetDeletedAt(nil)
		assert.Nil(t, model.GetDeletedAt())
	})

	t.Run("Timestamps interface methods", func(t *testing.T) {
		model := NewSoftDeletableModel()

		// Test initial state
		assert.True(t, model.GetCreatedAt().IsZero())
		assert.True(t, model.GetUpdatedAt().IsZero())

		// Test setting timestamps
		createdAt := time.Now()
		updatedAt := time.Now().Add(time.Hour)

		model.SetCreatedAt(createdAt)
		model.SetUpdatedAt(updatedAt)

		assert.Equal(t, createdAt, model.GetCreatedAt())
		assert.Equal(t, updatedAt, model.GetUpdatedAt())
	})
}

// TestModelInterfaces tests that each model type implements the correct interfaces
func TestModelInterfaces(t *testing.T) {
	t.Run("BaseModel interfaces", func(t *testing.T) {
		model := NewBaseModel()
		// Test that BaseModel implements Model interface
		var _ Model = model
	})

	t.Run("TimestampedModel interfaces", func(t *testing.T) {
		model := NewTimestampedModel()
		// Test that TimestampedModel implements Model and Timestamps interfaces
		var _ Model = model
		var _ Timestamps = model
	})

	t.Run("SoftDeletableModel interfaces", func(t *testing.T) {
		model := NewSoftDeletableModel()
		// Test that SoftDeletableModel implements Model, Timestamps, and SoftDelete interfaces
		var _ Model = model
		var _ Timestamps = model
		var _ SoftDelete = model
	})
}

// TestBaseModelWithDifferentIDTypes tests BaseModel with different ID types
func TestBaseModelWithDifferentIDTypes(t *testing.T) {
	model := NewBaseModel()

	// Test with int
	model.SetID(42)
	assert.Equal(t, 42, model.GetID())

	// Test with string
	model.SetID("uuid-123")
	assert.Equal(t, "uuid-123", model.GetID())

	// Test with int64
	model.SetID(int64(9223372036854775807))
	assert.Equal(t, int64(9223372036854775807), model.GetID())
}
