package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/require"
)

// --- Mocks for Testing ---
type (
	fakeAdapter struct {
		connectErr error
		conn       contract.Connection
	}

	// fakeConn is a mock connection that allows us to track method calls.
	fakeConn struct {
		pingErr     error
		closeErr    error
		closeCalled bool // track if Close() was called
	}
)

func (f *fakeAdapter) Connect(cfg *config.Config) (contract.Connection, error) {
	return f.conn, f.connectErr
}

func (f *fakeAdapter) Name() string { return "fake" }

func (f *fakeConn) Ping(ctx context.Context) error { return f.pingErr }
func (f *fakeConn) Close() error {
	f.closeCalled = true
	return f.closeErr
}
func (f *fakeConn) GetConnection() any                                              { return nil }
func (f *fakeConn) NewRepository(model contract.Model) (contract.Repository, error) { return nil, nil }
func (f *fakeConn) Transaction(ctx context.Context, fn func(contract.Connection) error) error {
	return nil
}

func (f *fakeConn) Select(ctx context.Context, query string, bindings ...any) ([]map[string]any, error) {
	return nil, nil
}

func (f *fakeConn) Statement(ctx context.Context, query string, bindings ...any) (sql.Result, error) {
	return nil, nil
}

// --- Test Cases ---
func TestConnect_WithInjectedAdapter(t *testing.T) {
	cfg := config.Config{Driver: "any", DSN: "any"}
	adapter := &fakeAdapter{conn: &fakeConn{}}
	cfg.Adapter = adapter // Inject the adapter directly

	conn, err := Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
}

func TestConnect_WithRegisteredAdapter(t *testing.T) {
	// Register a fake adapter for this test
	adapter := &fakeAdapter{conn: &fakeConn{}}
	RegisterAdapter(adapter, "registered-fake")

	cfg := config.Config{Driver: "registered-fake", DSN: "any"}
	conn, err := Connect(&cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
}

func TestConnect_AdapterLookupError(t *testing.T) {
	cfg := config.Config{Driver: "nonexistent-adapter", DSN: "any"}
	_, err := Connect(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unknown database adapter "nonexistent-adapter"`)
}

func TestConnect_ValidationFail(t *testing.T) {
	cfg := config.Config{Driver: "fake", DSN: ""} // DSN is empty
	_, err := Connect(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database dsn is required")
}

func TestConnect_AdapterConnectFail(t *testing.T) {
	cfg := config.Config{Driver: "fake", DSN: "any"}
	connectError := errors.New("adapter failed")
	adapter := &fakeAdapter{connectErr: connectError}
	cfg.Adapter = adapter

	_, err := Connect(&cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, connectError)
}

func TestConnect_PingFail_CallsClose(t *testing.T) {
	cfg := config.Config{Driver: "fake", DSN: "any"}
	pingError := errors.New("ping failed")
	mockConn := &fakeConn{pingErr: pingError}
	adapter := &fakeAdapter{conn: mockConn}
	cfg.Adapter = adapter

	_, err := Connect(&cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, pingError, "The wrapping error should contain the ping error")
	require.True(t, mockConn.closeCalled, "Close() should be called on ping failure")
}
