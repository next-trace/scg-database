package user

import (
	"github.com/next-trace/scg-database/contract"
)

type (
	User struct {
		ID    uint `gorm:"primaryKey" json:"id"`
		Name  string
		Email string `gorm:"uniqueIndex"`
	}
)

// TableName returns the database table name for the User model.
func (m *User) TableName() string {
	// Return the table name for this model
	// Example: return "users"
	return "users"
}

// Relationships defines the relationships for the User model.
func (m *User) Relationships() map[string]contract.Relationship {
	return map[string]contract.Relationship{
		// Example relationships:
		// "Profile": contract.NewHasOne(&Profile{}, "user_id", "id"),
		// "Orders": contract.NewHasMany(&Order{}, "user_id", "id"),
		// "Roles": contract.NewBelongsToMany(&Role{}, "user_roles"),
	}
}

// Model interface implementation
func (m *User) PrimaryKey() string { return "id" }
func (m *User) GetID() any         { return m.ID }
func (m *User) SetID(id any) {
	if idVal, ok := id.(uint); ok {
		m.ID = idVal
	}
}
