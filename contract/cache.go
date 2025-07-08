package contract

import (
	"time"
)

type (
	Cache interface {
		Get(string) (any, bool)
		Set(string, any, time.Duration) error
		Delete(string) error
		Flush() error
	}
)
