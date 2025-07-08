package contract

type (
	RelationshipType string

	Relationship interface {
		Type() RelationshipType
		RelatedModel() Model
		ForeignKey() string
		OwnerKey() string
		ManyToManyJoinTable() string
	}
)

const (
	HasOne        RelationshipType = "HasOne"
	HasMany       RelationshipType = "HasMany"
	BelongsTo     RelationshipType = "BelongsTo"
	BelongsToMany RelationshipType = "BelongsToMany"
	Many2Many     RelationshipType = "Many2Many"
)
