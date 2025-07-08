package db

import (
	"fmt"
	"sync"

	"github.com/next-trace/scg-database/contract"
)

var (
	adaptersMu sync.RWMutex
	adapters   = make(map[string]contract.DBAdapter)
)

// RegisterAdapter adds one or more database adapters to the global registry.
func RegisterAdapter(adapter contract.DBAdapter, names ...string) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	if adapter == nil {
		panic("cannot register a nil database adapter")
	}
	if len(names) == 0 {
		panic("cannot register adapter without at least one name")
	}

	for _, name := range names {
		if name == "" {
			panic("cannot register adapter with an empty name")
		}
		adapters[name] = adapter
	}
}

// GetAdapter retrieves a registered database adapter by its name.
func GetAdapter(name string) (contract.DBAdapter, error) {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	adapter, ok := adapters[name]
	if !ok {
		return nil, fmt.Errorf("unknown database adapter %q (is the adapter package imported?)", name)
	}
	return adapter, nil
}
