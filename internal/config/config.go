package config

import (
	"sync/atomic"

	"github.com/afkdevs/go-schema/internal/dialect"
)

type Config struct {
	Dialect dialect.Dialect
	Verbose bool
}

var config = atomic.Pointer[Config]{}

func init() {
	config.Store(&Config{
		Dialect: dialect.Unknown,
		Verbose: true,
	})
}

func Set(newConfig *Config) {
	config.Store(newConfig)
}

func Get() *Config {
	return config.Load()
}
