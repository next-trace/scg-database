package contract

import (
	"context"
)

type (
	Repository interface {
		With(...string) Repository
		Where(any, ...any) Repository
		Unscoped() Repository
		Limit(int) Repository
		Offset(int) Repository
		OrderBy(string, string) Repository

		Find(context.Context, any) (Model, error)
		FindOrFail(context.Context, any) (Model, error)
		First(context.Context) (Model, error)
		FirstOrFail(context.Context) (Model, error)
		Get(context.Context) ([]Model, error)
		Pluck(context.Context, string, any) error

		Create(context.Context, ...Model) error
		CreateInBatches(context.Context, []Model, int) error
		Update(context.Context, ...Model) error
		Delete(context.Context, ...Model) error
		ForceDelete(context.Context, ...Model) error

		FirstOrCreate(context.Context, Model, ...Model) (Model, error)
		UpdateOrCreate(context.Context, Model, any) (Model, error)

		// QueryBuilder provides access to the fluent query builder interface
		QueryBuilder() QueryBuilder
	}
)
