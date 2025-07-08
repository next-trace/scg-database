package contract

import (
	"time"
)

type (
	Model interface {
		PrimaryKey() string
		TableName() string
		GetID() any
		SetID(any)
		Relationships() map[string]Relationship
	}

	SoftDelete interface {
		GetDeletedAt() *time.Time
		SetDeletedAt(*time.Time)
	}

	Timestamps interface {
		GetCreatedAt() time.Time
		GetUpdatedAt() time.Time
		SetCreatedAt(time.Time)
		SetUpdatedAt(time.Time)
	}
)

// Relationship implementations
type (
	// HasOneRelationship represents a one-to-one relationship
	HasOneRelationship struct {
		related    Model
		foreignKey string
		ownerKey   string
	}

	// HasManyRelationship represents a one-to-many relationship
	HasManyRelationship struct {
		related    Model
		foreignKey string
		ownerKey   string
	}

	// BelongsToRelationship represents a many-to-one relationship
	BelongsToRelationship struct {
		related    Model
		foreignKey string
		ownerKey   string
	}

	// BelongsToManyRelationship represents a many-to-many relationship
	BelongsToManyRelationship struct {
		related   Model
		joinTable string
	}
)

func NewHasOne(related Model, foreignKey, ownerKey string) *HasOneRelationship {
	return &HasOneRelationship{
		related:    related,
		foreignKey: foreignKey,
		ownerKey:   ownerKey,
	}
}

func (r *HasOneRelationship) Type() RelationshipType      { return HasOne }
func (r *HasOneRelationship) RelatedModel() Model         { return r.related }
func (r *HasOneRelationship) ForeignKey() string          { return r.foreignKey }
func (r *HasOneRelationship) OwnerKey() string            { return r.ownerKey }
func (r *HasOneRelationship) ManyToManyJoinTable() string { return "" }

func NewHasMany(related Model, foreignKey, ownerKey string) *HasManyRelationship {
	return &HasManyRelationship{
		related:    related,
		foreignKey: foreignKey,
		ownerKey:   ownerKey,
	}
}

func (r *HasManyRelationship) Type() RelationshipType      { return HasMany }
func (r *HasManyRelationship) RelatedModel() Model         { return r.related }
func (r *HasManyRelationship) ForeignKey() string          { return r.foreignKey }
func (r *HasManyRelationship) OwnerKey() string            { return r.ownerKey }
func (r *HasManyRelationship) ManyToManyJoinTable() string { return "" }

func NewBelongsTo(related Model, foreignKey, ownerKey string) *BelongsToRelationship {
	return &BelongsToRelationship{
		related:    related,
		foreignKey: foreignKey,
		ownerKey:   ownerKey,
	}
}

func (r *BelongsToRelationship) Type() RelationshipType      { return BelongsTo }
func (r *BelongsToRelationship) RelatedModel() Model         { return r.related }
func (r *BelongsToRelationship) ForeignKey() string          { return r.foreignKey }
func (r *BelongsToRelationship) OwnerKey() string            { return r.ownerKey }
func (r *BelongsToRelationship) ManyToManyJoinTable() string { return "" }

func NewBelongsToMany(related Model, joinTable string) *BelongsToManyRelationship {
	return &BelongsToManyRelationship{
		related:   related,
		joinTable: joinTable,
	}
}

func (r *BelongsToManyRelationship) Type() RelationshipType      { return BelongsToMany }
func (r *BelongsToManyRelationship) RelatedModel() Model         { return r.related }
func (r *BelongsToManyRelationship) ForeignKey() string          { return "" }
func (r *BelongsToManyRelationship) OwnerKey() string            { return "" }
func (r *BelongsToManyRelationship) ManyToManyJoinTable() string { return r.joinTable }
