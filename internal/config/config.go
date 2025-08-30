package config

import (
	"sync/atomic"

	"github.com/afkdevs/migris/internal/dialect"
)

type Config struct {
	Dialect   dialect.Dialect
	TableName string
	Verbose   bool
}

var config = atomic.Pointer[Config]{}

func init() {
	config.Store(&Config{
		Dialect:   dialect.Unknown,
		TableName: "migris_db_version",
		Verbose:   true,
	})
}

func Set(cfg *Config) {
	config.Store(cfg)
}

func Get() *Config {
	return config.Load()
}
