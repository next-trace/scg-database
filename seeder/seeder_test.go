package seeder

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/require"
)

// Mock types for testing
type (
	// Mock connection for testing
	mockConnection struct{}

	// Mock sql.Result for testing
	mockResult struct{}

	// Mock seeders for testing
	successSeeder struct {
		name string
		ran  bool
	}

	errorSeeder struct {
		name string
		err  error
	}
)

func (m *mockResult) LastInsertId() (int64, error) { return 0, nil }
func (m *mockResult) RowsAffected() (int64, error) { return 0, nil }

func (m *mockConnection) GetConnection() any             { return nil }
func (m *mockConnection) Ping(ctx context.Context) error { return nil }
func (m *mockConnection) Close() error                   { return nil }
func (m *mockConnection) NewRepository(model contract.Model) (contract.Repository, error) {
	return nil, nil
}

func (m *mockConnection) Transaction(ctx context.Context, fn func(txConnection contract.Connection) error) error {
	return nil
}

func (m *mockConnection) Select(ctx context.Context, query string, bindings ...any) ([]map[string]any, error) {
	return nil, nil
}

func (m *mockConnection) Statement(ctx context.Context, query string, bindings ...any) (sql.Result, error) {
	return &mockResult{}, nil
}

func (s *successSeeder) Run(db contract.Connection) error {
	s.ran = true
	return nil
}

func (s *errorSeeder) Run(db contract.Connection) error {
	return s.err
}

func TestNew(t *testing.T) {
	db := &mockConnection{}
	runner := New(db)

	require.NotNil(t, runner)
}

func TestRunner_Run_Success(t *testing.T) {
	db := &mockConnection{}
	runner := New(db)

	seeder1 := &successSeeder{name: "seeder1"}
	seeder2 := &successSeeder{name: "seeder2"}

	err := runner.Run(seeder1, seeder2)

	require.NoError(t, err)
	require.True(t, seeder1.ran)
	require.True(t, seeder2.ran)
}

func TestRunner_Run_EmptySlice(t *testing.T) {
	db := &mockConnection{}
	runner := New(db)

	err := runner.Run()

	require.NoError(t, err)
}

func TestRunner_Run_SingleSeeder(t *testing.T) {
	db := &mockConnection{}
	runner := New(db)

	seeder := &successSeeder{name: "single"}

	err := runner.Run(seeder)

	require.NoError(t, err)
	require.True(t, seeder.ran)
}

func TestRunner_Run_Error(t *testing.T) {
	db := &mockConnection{}
	runner := New(db)

	expectedErr := errors.New("seeder failed")
	seeder1 := &successSeeder{name: "seeder1"}
	seeder2 := &errorSeeder{name: "seeder2", err: expectedErr}
	seeder3 := &successSeeder{name: "seeder3"}

	err := runner.Run(seeder1, seeder2, seeder3)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to run seeder")
	require.Contains(t, err.Error(), "*seeder.errorSeeder") // The error contains the type name
	require.True(t, seeder1.ran)                            // First seeder should have run
	require.False(t, seeder3.ran)                           // Third seeder should not have run due to error
}

func TestRunner_Run_ErrorWrapping(t *testing.T) {
	db := &mockConnection{}
	runner := New(db)

	originalErr := errors.New("database connection failed")
	seeder := &errorSeeder{name: "failing_seeder", err: originalErr}

	err := runner.Run(seeder)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to run seeder")
	require.Contains(t, err.Error(), "*seeder.errorSeeder")
	require.ErrorIs(t, err, originalErr)
}

func TestRunner_Run_MultipleErrors(t *testing.T) {
	db := &mockConnection{}
	runner := New(db)

	// Only the first error should be returned
	seeder1 := &errorSeeder{name: "seeder1", err: errors.New("first error")}
	seeder2 := &errorSeeder{name: "seeder2", err: errors.New("second error")}

	err := runner.Run(seeder1, seeder2)

	require.Error(t, err)
	require.Contains(t, err.Error(), "first error")
	require.NotContains(t, err.Error(), "second error")
}

func TestRunner_Run_NilSeeder(t *testing.T) {
	db := &mockConnection{}
	runner := New(db)

	// This should panic when trying to call Run on nil
	require.Panics(t, func() {
		runner.Run(nil)
	})
}
