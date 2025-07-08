package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRelationshipTypes tests the relationship type constants
func TestRelationshipTypes(t *testing.T) {
	assert.Equal(t, RelationshipType("HasOne"), HasOne)
	assert.Equal(t, RelationshipType("HasMany"), HasMany)
	assert.Equal(t, RelationshipType("BelongsTo"), BelongsTo)
	assert.Equal(t, RelationshipType("BelongsToMany"), BelongsToMany)
	assert.Equal(t, RelationshipType("Many2Many"), Many2Many)
}

// TestHasOneRelationship tests the HasOneRelationship implementation
func TestHasOneRelationship(t *testing.T) {
	relatedModel := NewBaseModel()
	rel := NewHasOne(relatedModel, "user_id", "id")

	assert.NotNil(t, rel)
	assert.Equal(t, HasOne, rel.Type())
	assert.Equal(t, relatedModel, rel.RelatedModel())
	assert.Equal(t, "user_id", rel.ForeignKey())
	assert.Equal(t, "id", rel.OwnerKey())
	assert.Equal(t, "", rel.ManyToManyJoinTable())
}

// TestHasManyRelationship tests the HasManyRelationship implementation
func TestHasManyRelationship(t *testing.T) {
	relatedModel := NewBaseModel()
	rel := NewHasMany(relatedModel, "user_id", "id")

	assert.NotNil(t, rel)
	assert.Equal(t, HasMany, rel.Type())
	assert.Equal(t, relatedModel, rel.RelatedModel())
	assert.Equal(t, "user_id", rel.ForeignKey())
	assert.Equal(t, "id", rel.OwnerKey())
	assert.Equal(t, "", rel.ManyToManyJoinTable())
}

// TestBelongsToRelationship tests the BelongsToRelationship implementation
func TestBelongsToRelationship(t *testing.T) {
	relatedModel := NewBaseModel()
	rel := NewBelongsTo(relatedModel, "user_id", "id")

	assert.NotNil(t, rel)
	assert.Equal(t, BelongsTo, rel.Type())
	assert.Equal(t, relatedModel, rel.RelatedModel())
	assert.Equal(t, "user_id", rel.ForeignKey())
	assert.Equal(t, "id", rel.OwnerKey())
	assert.Equal(t, "", rel.ManyToManyJoinTable())
}

// TestBelongsToManyRelationship tests the BelongsToManyRelationship implementation
func TestBelongsToManyRelationship(t *testing.T) {
	relatedModel := NewBaseModel()
	rel := NewBelongsToMany(relatedModel, "user_roles")

	assert.NotNil(t, rel)
	assert.Equal(t, BelongsToMany, rel.Type())
	assert.Equal(t, relatedModel, rel.RelatedModel())
	assert.Equal(t, "", rel.ForeignKey())
	assert.Equal(t, "", rel.OwnerKey())
	assert.Equal(t, "user_roles", rel.ManyToManyJoinTable())
}

// TestRelationshipInterface tests that all relationship types implement the interface
func TestRelationshipInterface(t *testing.T) {
	relatedModel := NewBaseModel()

	relationships := []Relationship{
		NewHasOne(relatedModel, "foreign_key", "owner_key"),
		NewHasMany(relatedModel, "foreign_key", "owner_key"),
		NewBelongsTo(relatedModel, "foreign_key", "owner_key"),
		NewBelongsToMany(relatedModel, "join_table"),
	}

	for _, rel := range relationships {
		assert.NotNil(t, rel.Type())
		assert.NotNil(t, rel.RelatedModel())
		// ForeignKey and OwnerKey can be empty for some relationship types
		assert.NotNil(t, rel.ForeignKey())
		assert.NotNil(t, rel.OwnerKey())
		assert.NotNil(t, rel.ManyToManyJoinTable())
	}
}

// TestRelationshipTypeString tests string representation of relationship types
func TestRelationshipTypeString(t *testing.T) {
	assert.Equal(t, "HasOne", string(HasOne))
	assert.Equal(t, "HasMany", string(HasMany))
	assert.Equal(t, "BelongsTo", string(BelongsTo))
	assert.Equal(t, "BelongsToMany", string(BelongsToMany))
	assert.Equal(t, "Many2Many", string(Many2Many))
}

// TestRelationshipEdgeCases tests edge cases for relationships
func TestRelationshipEdgeCases(t *testing.T) {
	t.Run("HasOne with empty keys", func(t *testing.T) {
		relatedModel := NewBaseModel()
		rel := NewHasOne(relatedModel, "", "")

		assert.Equal(t, HasOne, rel.Type())
		assert.Equal(t, relatedModel, rel.RelatedModel())
		assert.Equal(t, "", rel.ForeignKey())
		assert.Equal(t, "", rel.OwnerKey())
	})

	t.Run("BelongsToMany with empty join table", func(t *testing.T) {
		relatedModel := NewBaseModel()
		rel := NewBelongsToMany(relatedModel, "")

		assert.Equal(t, BelongsToMany, rel.Type())
		assert.Equal(t, relatedModel, rel.RelatedModel())
		assert.Equal(t, "", rel.ManyToManyJoinTable())
	})
}
