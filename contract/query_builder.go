package contract

import (
	"context"
)

type (
	// QueryBuilder provides a fluent interface for building database queries
	// Each adapter should implement this interface to provide database-specific query building
	QueryBuilder interface {
		// Query building methods
		Select(...string) QueryBuilder
		Where(string, ...any) QueryBuilder
		WhereIn(string, []any) QueryBuilder
		WhereNotIn(string, []any) QueryBuilder
		WhereNull(string) QueryBuilder
		WhereNotNull(string) QueryBuilder
		WhereBetween(string, any, any) QueryBuilder
		OrWhere(string, ...any) QueryBuilder

		// Join methods
		Join(string, string) QueryBuilder
		LeftJoin(string, string) QueryBuilder
		RightJoin(string, string) QueryBuilder
		InnerJoin(string, string) QueryBuilder

		// Ordering and grouping
		OrderBy(string, string) QueryBuilder
		GroupBy(...string) QueryBuilder
		Having(string, ...any) QueryBuilder

		// Limiting and pagination
		Limit(int) QueryBuilder
		Offset(int) QueryBuilder

		// Relationships
		With(...string) QueryBuilder
		WithCount(...string) QueryBuilder

		// Scopes and advanced features
		Scoped() QueryBuilder
		Unscoped() QueryBuilder

		// Execution methods
		Find(context.Context, any) error
		First(context.Context, any) error
		Get(context.Context, any) error
		Count(context.Context) (int64, error)
		Exists(context.Context) (bool, error)

		// Mutation methods
		Create(context.Context, any) error
		Update(context.Context, any) error
		Delete(context.Context) error

		// Raw query methods
		Raw(string, ...any) QueryBuilder
		Exec(context.Context, string, ...any) error

		// Utility methods
		ToSQL() (string, []any, error)
		Clone() QueryBuilder
		Reset() QueryBuilder
	}

	// QueryBuilderFactory creates QueryBuilder instances for specific models and connections
	QueryBuilderFactory interface {
		NewQueryBuilder(Model, any) QueryBuilder
		Name() string
	}

	// QueryBuilderRegistry manages query builder factories for different adapters
	QueryBuilderRegistry interface {
		Register(string, QueryBuilderFactory)
		Get(string) (QueryBuilderFactory, error)
		List() []string
	}
)
