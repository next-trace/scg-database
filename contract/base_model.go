package contract

import (
	"time"
)

type (
	// BaseModel provides a minimal, adapter-agnostic base model implementation
	// that satisfies only the core Model interface without forcing timestamps or soft deletes.
	//
	// Note: Fields are public to allow database adapters to access them,
	// but should be accessed through the interface methods for consistency.
	BaseModel struct {
		ID any `json:"id"`
	}

	// TimestampedModel extends BaseModel with timestamp functionality
	// Use this when you need created_at and updated_at fields
	TimestampedModel struct {
		BaseModel
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	// SoftDeletableModel extends TimestampedModel with soft delete functionality
	// Use this when you need soft delete capability
	SoftDeletableModel struct {
		TimestampedModel
		DeletedAt *time.Time `json:"deleted_at,omitempty"`
	}
)

// NewBaseModel creates a new BaseModel instance
func NewBaseModel() *BaseModel {
	return &BaseModel{}
}

// NewTimestampedModel creates a new TimestampedModel instance
func NewTimestampedModel() *TimestampedModel {
	return &TimestampedModel{
		BaseModel: BaseModel{},
	}
}

// NewSoftDeletableModel creates a new SoftDeletableModel instance
func NewSoftDeletableModel() *SoftDeletableModel {
	return &SoftDeletableModel{
		TimestampedModel: TimestampedModel{
			BaseModel: BaseModel{},
		},
	}
}

// BaseModel - Model interface implementation
func (b *BaseModel) PrimaryKey() string                     { return "id" }
func (b *BaseModel) TableName() string                      { return "" } // Override in concrete models
func (b *BaseModel) GetID() any                             { return b.ID }
func (b *BaseModel) SetID(id any)                           { b.ID = id }
func (b *BaseModel) Relationships() map[string]Relationship { return nil } // Override in concrete models

// TimestampedModel - Timestamps interface implementation
func (t *TimestampedModel) GetCreatedAt() time.Time   { return t.CreatedAt }
func (t *TimestampedModel) GetUpdatedAt() time.Time   { return t.UpdatedAt }
func (t *TimestampedModel) SetCreatedAt(tm time.Time) { t.CreatedAt = tm }
func (t *TimestampedModel) SetUpdatedAt(tm time.Time) { t.UpdatedAt = tm }

// SoftDeletableModel - SoftDelete interface implementation
func (s *SoftDeletableModel) GetDeletedAt() *time.Time  { return s.DeletedAt }
func (s *SoftDeletableModel) SetDeletedAt(t *time.Time) { s.DeletedAt = t }
