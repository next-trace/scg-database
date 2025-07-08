package contract

import (
	"github.com/next-trace/scg-database/config"
)

type (
	DBAdapter interface {
		// Connect's only job is to create our rich Connection object from a config struct.
		// Options should be applied *before* this is called.
		Connect(*config.Config) (Connection, error)
		Name() string
	}
)
