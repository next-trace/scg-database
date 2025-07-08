package gorm

import (
	"context"
	"fmt"
	"strings"

	"github.com/next-trace/scg-database/contract"
	"gorm.io/gorm"
)

// SQL direction constants
const (
	OrderDirectionASC  = "ASC"
	OrderDirectionDESC = "DESC"
)

type (
	// gormQueryBuilder implements the contract.QueryBuilder interface for GORM
	gormQueryBuilder struct {
		db    *gorm.DB
		model contract.Model
	}
)

// Ensure gormQueryBuilder implements contract.QueryBuilder
var (
	_ contract.QueryBuilder = (*gormQueryBuilder)(nil)
)

// newGormQueryBuilder creates a new GORM query builder instance
func newGormQueryBuilder(model contract.Model, connection any) *gormQueryBuilder {
	gormDB, ok := connection.(*gorm.DB)
	if !ok {
		panic("connection must be a *gorm.DB instance")
	}

	return &gormQueryBuilder{
		db:    gormDB.Model(model),
		model: model,
	}
}

// Query building methods

func (q *gormQueryBuilder) Select(columns ...string) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Select(columns),
		model: q.model,
	}
}

func (q *gormQueryBuilder) Where(condition string, args ...any) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Where(condition, args...),
		model: q.model,
	}
}

func (q *gormQueryBuilder) WhereIn(column string, values []any) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Where(fmt.Sprintf("%s IN ?", column), values),
		model: q.model,
	}
}

func (q *gormQueryBuilder) WhereNotIn(column string, values []any) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Where(fmt.Sprintf("%s NOT IN ?", column), values),
		model: q.model,
	}
}

func (q *gormQueryBuilder) WhereNull(column string) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Where(fmt.Sprintf("%s IS NULL", column)),
		model: q.model,
	}
}

func (q *gormQueryBuilder) WhereNotNull(column string) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Where(fmt.Sprintf("%s IS NOT NULL", column)),
		model: q.model,
	}
}

func (q *gormQueryBuilder) WhereBetween(column string, start, end any) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", column), start, end),
		model: q.model,
	}
}

func (q *gormQueryBuilder) OrWhere(condition string, args ...any) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Or(condition, args...),
		model: q.model,
	}
}

// Join methods

func (q *gormQueryBuilder) Join(table, condition string) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Joins(fmt.Sprintf("JOIN %s ON %s", table, condition)),
		model: q.model,
	}
}

func (q *gormQueryBuilder) LeftJoin(table, condition string) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", table, condition)),
		model: q.model,
	}
}

func (q *gormQueryBuilder) RightJoin(table, condition string) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Joins(fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition)),
		model: q.model,
	}
}

func (q *gormQueryBuilder) InnerJoin(table, condition string) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Joins(fmt.Sprintf("INNER JOIN %s ON %s", table, condition)),
		model: q.model,
	}
}

// Ordering and grouping

func (q *gormQueryBuilder) OrderBy(column, direction string) contract.QueryBuilder {
	// Validate direction to prevent SQL injection
	direction = strings.ToUpper(strings.TrimSpace(direction))
	if direction != OrderDirectionASC && direction != OrderDirectionDESC {
		direction = OrderDirectionASC
	}

	return &gormQueryBuilder{
		db:    q.db.Order(fmt.Sprintf("%s %s", column, direction)),
		model: q.model,
	}
}

func (q *gormQueryBuilder) GroupBy(columns ...string) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Group(strings.Join(columns, ", ")),
		model: q.model,
	}
}

func (q *gormQueryBuilder) Having(condition string, args ...any) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Having(condition, args...),
		model: q.model,
	}
}

// Limiting and pagination

func (q *gormQueryBuilder) Limit(limit int) contract.QueryBuilder {
	if limit < 0 {
		return q
	}
	return &gormQueryBuilder{
		db:    q.db.Limit(limit),
		model: q.model,
	}
}

func (q *gormQueryBuilder) Offset(offset int) contract.QueryBuilder {
	if offset < 0 {
		return q
	}
	return &gormQueryBuilder{
		db:    q.db.Offset(offset),
		model: q.model,
	}
}

// Relationships

func (q *gormQueryBuilder) With(relations ...string) contract.QueryBuilder {
	db := q.db
	for _, relation := range relations {
		db = db.Preload(relation)
	}
	return &gormQueryBuilder{
		db:    db,
		model: q.model,
	}
}

func (q *gormQueryBuilder) WithCount(...string) contract.QueryBuilder {
	// GORM doesn't have a direct WithCount equivalent, but we can simulate it
	// This is a simplified implementation
	return q
}

// Scopes and advanced features

func (q *gormQueryBuilder) Scoped() contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Scopes(),
		model: q.model,
	}
}

func (q *gormQueryBuilder) Unscoped() contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Unscoped(),
		model: q.model,
	}
}

// Execution methods

func (q *gormQueryBuilder) Find(ctx context.Context, dest any) error {
	return q.db.WithContext(ctx).Find(dest).Error
}

func (q *gormQueryBuilder) First(ctx context.Context, dest any) error {
	return q.db.WithContext(ctx).First(dest).Error
}

func (q *gormQueryBuilder) Get(ctx context.Context, dest any) error {
	return q.db.WithContext(ctx).Find(dest).Error
}

func (q *gormQueryBuilder) Count(ctx context.Context) (int64, error) {
	var count int64
	err := q.db.WithContext(ctx).Count(&count).Error
	return count, err
}

func (q *gormQueryBuilder) Exists(ctx context.Context) (bool, error) {
	count, err := q.Count(ctx)
	return count > 0, err
}

// Mutation methods

func (q *gormQueryBuilder) Create(ctx context.Context, value any) error {
	return q.db.WithContext(ctx).Create(value).Error
}

func (q *gormQueryBuilder) Update(ctx context.Context, values any) error {
	return q.db.WithContext(ctx).Updates(values).Error
}

func (q *gormQueryBuilder) Delete(ctx context.Context) error {
	return q.db.WithContext(ctx).Delete(q.model).Error
}

// Raw query methods

func (q *gormQueryBuilder) Raw(sql string, args ...any) contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Raw(sql, args...),
		model: q.model,
	}
}

func (q *gormQueryBuilder) Exec(ctx context.Context, sql string, args ...any) error {
	return q.db.WithContext(ctx).Exec(sql, args...).Error
}

// Utility methods

func (q *gormQueryBuilder) ToSQL() (sql string, args []any, err error) {
	// GORM doesn't provide a direct way to get SQL without executing
	// This is a simplified implementation
	return "", nil, fmt.Errorf("ToSQL not fully implemented for GORM adapter")
}

func (q *gormQueryBuilder) Clone() contract.QueryBuilder {
	return &gormQueryBuilder{
		db:    q.db.Session(&gorm.Session{}),
		model: q.model,
	}
}

func (q *gormQueryBuilder) Reset() contract.QueryBuilder {
	return newGormQueryBuilder(q.model, q.db.Session(&gorm.Session{NewDB: true}))
}

type (
	// GormQueryBuilderFactory implements contract.QueryBuilderFactory for GORM
	GormQueryBuilderFactory struct{}
)

// Ensure GormQueryBuilderFactory implements contract.QueryBuilderFactory
var (
	_ contract.QueryBuilderFactory = (*GormQueryBuilderFactory)(nil)
)

func (f *GormQueryBuilderFactory) NewQueryBuilder(model contract.Model, connection any) contract.QueryBuilder {
	return newGormQueryBuilder(model, connection)
}

func (f *GormQueryBuilderFactory) Name() string {
	return "gorm"
}
