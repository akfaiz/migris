package schema

import (
	"errors"

	"github.com/afkdevs/go-schema/internal/config"
	"github.com/afkdevs/go-schema/internal/dialect"
)

// SetDialect sets the migrator dialect
func SetDialect(d string) error {
	dialectValue := dialect.FromString(d)
	if dialectValue == dialect.Unknown {
		return errors.New("unsupported dialect: " + d)
	}
	cfg := config.Get()
	cfg.Dialect = dialectValue
	config.Set(cfg)
	return nil
}

// SetVerbose enables or disables verbose mode
func SetVerbose(enabled bool) {
	cfg := config.Get()
	cfg.Verbose = enabled
	config.Set(cfg)
}
