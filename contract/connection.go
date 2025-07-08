package contract

import (
	"context"
	"database/sql"
)

type (
	Connection interface {
		GetConnection() any
		Ping(context.Context) error
		Close() error
		NewRepository(Model) (Repository, error)
		Transaction(context.Context, func(Connection) error) error
		Select(context.Context, string, ...any) ([]map[string]any, error)
		Statement(context.Context, string, ...any) (sql.Result, error)
	}
)
