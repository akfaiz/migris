package migris

import (
	"errors"

	"github.com/afkdevs/migris/internal/config"
	"github.com/afkdevs/migris/internal/dialect"
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

// SetTableName sets the table name for the migrator
func SetTableName(name string) {
	cfg := config.Get()
	cfg.TableName = name
	config.Set(cfg)
}
