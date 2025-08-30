package config

import (
	"sync/atomic"

	"github.com/afkdevs/migris/internal/dialect"
)

type Config struct {
	Dialect dialect.Dialect
}

var config = atomic.Pointer[Config]{}

func init() {
	config.Store(&Config{
		Dialect: dialect.Unknown,
	})
}

func SetDialect(dialect dialect.Dialect) {
	cfg := config.Load()
	cfg.Dialect = dialect
	config.Store(cfg)
}

func GetDialect() dialect.Dialect {
	return config.Load().Dialect
}
