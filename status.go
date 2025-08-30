package migris

import (
	"context"

	"github.com/akfaiz/migris/internal/logger"
)

// Status returns the status of the migrations.
func (m *Migrate) Status() error {
	ctx := context.Background()
	return m.StatusContext(ctx)
}

// StatusContext returns the status of the migrations.
func (m *Migrate) StatusContext(ctx context.Context) error {
	provider, err := m.newProvider()
	if err != nil {
		return err
	}
	migrations, err := provider.Status(ctx)
	if err != nil {
		return err
	}
	logger.PrintStatuses(migrations)
	return nil
}
