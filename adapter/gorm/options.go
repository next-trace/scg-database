package gorm

import (
	"github.com/next-trace/scg-database/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// WithConfig is a GORM-specific option to provide a full gorm.Config object.
func WithConfig(gormCfg *gorm.Config) config.Option {
	return func(cfg *config.Config) {
		if cfg.Settings == nil {
			cfg.Settings = make(map[string]any)
		}
		cfg.Settings["gorm_config"] = gormCfg
	}
}

// WithLogger is a GORM-specific option to provide a custom GORM logger.
func WithLogger(l logger.Interface) config.Option {
	return func(cfg *config.Config) {
		if cfg.Settings == nil {
			cfg.Settings = make(map[string]any)
		}
		cfg.Settings["gorm_logger"] = l
	}
}
