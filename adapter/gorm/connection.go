package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"gorm.io/gorm"
)

type (
	connection struct {
		db     *gorm.DB
		config config.Config
	}
)

// Ensure the implementation satisfies the interface at compile time.
var (
	_ contract.Connection = (*connection)(nil)
)

func (c *connection) NewRepository(model contract.Model) (contract.Repository, error) {
	if model == nil || reflect.ValueOf(model).IsNil() {
		return nil, fmt.Errorf("model cannot be nil")
	}
	return newGormRepository(c.db, model), nil
}

func (c *connection) GetConnection() any {
	return c.db
}

func (c *connection) Ping(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (c *connection) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		// If we can't get the underlying SQL DB, it might already be closed
		// or there's a configuration issue. Return the error for proper handling.
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}
	return sqlDB.Close()
}

func (c *connection) Transaction(ctx context.Context, fn func(txConnection contract.Connection) error) error {
	return c.db.WithContext(ctx).Transaction(func(txGorm *gorm.DB) error {
		txConn := &connection{db: txGorm, config: c.config}
		return fn(txConn)
	})
}

// Select executes a raw read query, fulfilling the contract.
func (c *connection) Select(ctx context.Context, query string, bindings ...any) ([]map[string]any, error) {
	var results []map[string]any
	err := c.db.WithContext(ctx).Raw(query, bindings...).Scan(&results).Error
	return results, err
}

// Statement executes a raw write query, fulfilling the contract.
func (c *connection) Statement(ctx context.Context, query string, bindings ...any) (sql.Result, error) {
	sqlDB, err := c.db.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.ExecContext(ctx, query, bindings...)
}
