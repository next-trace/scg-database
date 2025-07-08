package db

import (
	"context"

	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
)

// Connect is the single, unified function for creating a database connection.
// It accepts a populated Config struct and optional settings.
func Connect(cfg *config.Config, opts ...config.Option) (contract.Connection, error) {
	// Apply functional options, like WithAdapter or WithLogger
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, NewConfigValidationError(err)
	}

	// Determine the adapter, either from injection or the registry
	var adapter contract.DBAdapter
	if cfg.Adapter != nil {
		var ok bool
		adapter, ok = cfg.Adapter.(contract.DBAdapter)
		if !ok {
			return nil, NewInvalidAdapterError()
		}
	} else {
		var err error
		adapter, err = GetAdapter(cfg.Driver)
		if err != nil {
			return nil, NewAdapterLookupError(err)
		}
	}

	// Use the adapter to establish the connection
	conn, err := adapter.Connect(cfg)
	if err != nil {
		return nil, NewAdapterConnectError(err)
	}

	// Ping to ensure the connection is live
	if err := conn.Ping(context.Background()); err != nil {
		_ = conn.Close()
		return nil, NewConnectionPingError(err)
	}

	return conn, nil
}
