package migriscli

import (
	"context"
	"database/sql"

	"github.com/akfaiz/migris"
	"github.com/urfave/cli/v3"
)

// Config holds the configuration for the migris CLI commands.
type Config struct {
	DB            *sql.DB // Database connection
	Dialect       string  // Database dialect (e.g., "pgx", "mysql", etc.)
	MigrationsDir string  // Directory where migration files are stored
}

// NewCommand creates a new CLI command for migris with subcommands.
func NewCommand(cfg Config) *cli.Command {
	cmd := &cli.Command{
		Name:  "migrate",
		Usage: "Database migration CLI tool",
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a new migration file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Name of the migration",
						Required: true,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					return migris.Create(cfg.MigrationsDir, c.String("name"))
				},
			},
			{
				Name:  "up",
				Usage: "Apply all up migrations",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Simulate the migration without applying changes",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					migrator, err := createMigrator(c, cfg.DB, cfg)
					if err != nil {
						return err
					}
					return migrator.UpContext(ctx)
				},
			},
			{
				Name:  "up-to",
				Usage: "Apply migrations up to a specific version",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Simulate the migration without applying changes",
					},
					&cli.Int64Flag{
						Name:     "version",
						Aliases:  []string{"v"},
						Usage:    "Target version to migrate up to",
						Required: true,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					migrator, err := createMigrator(c, cfg.DB, cfg)
					if err != nil {
						return err
					}
					return migrator.UpToContext(ctx, c.Int64("version"))
				},
			},
			{
				Name:  "down",
				Usage: "Rollback the last migration",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Simulate the migration without applying changes",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					migrator, err := createMigrator(c, cfg.DB, cfg)
					if err != nil {
						return err
					}
					return migrator.DownContext(ctx)
				},
			},
			{
				Name:  "down-to",
				Usage: "Rollback migrations down to a specific version",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Simulate the migration without applying changes",
					},
					&cli.Int64Flag{
						Name:     "version",
						Aliases:  []string{"v"},
						Usage:    "Target version to migrate down to",
						Required: true,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					migrator, err := createMigrator(c, cfg.DB, cfg)
					if err != nil {
						return err
					}
					return migrator.DownToContext(ctx, c.Int64("version"))
				},
			},
			{
				Name:  "reset",
				Usage: "Rollback all migrations",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Simulate the migration without applying changes",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					migrator, err := createMigrator(c, cfg.DB, cfg)
					if err != nil {
						return err
					}
					return migrator.ResetContext(ctx)
				},
			},
			{
				Name:  "status",
				Usage: "Show the status of migrations",
				Action: func(ctx context.Context, c *cli.Command) error {
					migrator, err := createMigrator(c, cfg.DB, cfg)
					if err != nil {
						return err
					}
					return migrator.StatusContext(ctx)
				},
			},
		},
	}

	return cmd
}

func createMigrator(c *cli.Command, db *sql.DB, cfg Config) (*migris.Migrate, error) {
	options := []migris.Option{
		migris.WithDB(db),
		migris.WithMigrationDir(cfg.MigrationsDir),
	}

	if c.Bool("dry-run") {
		options = append(options, migris.WithDryRun(true))
	}

	migrator, err := migris.New(cfg.Dialect, options...)
	if err != nil {
		return nil, err
	}

	return migrator, nil
}
