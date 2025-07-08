package db

import (
	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
)

// WithAdapter allows injection of a custom DBAdapter at runtime, bypassing the registry.
// This is the core of the Open/Closed Principle for this package.
func WithAdapter(adapter contract.DBAdapter) config.Option {
	return func(cfg *config.Config) {
		cfg.Adapter = adapter
	}
}
