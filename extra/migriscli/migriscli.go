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
	AllowMissing  bool    // Allow out-of-order migrations
}

// NewCLI creates a new CLI interface for migris with subcommands.
func NewCLI(cfg Config) *cli.Command {
	return &cli.Command{
		Name:     "migrate",
		Usage:    "Database migration CLI tool",
		Commands: buildCommands(cfg),
	}
}

func buildCommands(cfg Config) []*cli.Command {
	return []*cli.Command{
		newCreateCommand(cfg),
		newUpCommand(cfg),
		newUpToCommand(cfg),
		newDownCommand(cfg),
		newDownToCommand(cfg),
		newResetCommand(cfg),
		newStatusCommand(cfg),
	}
}

func newCreateCommand(cfg Config) *cli.Command {
	return &cli.Command{
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
		Action: func(_ context.Context, c *cli.Command) error {
			return migris.Create(cfg.MigrationsDir, c.String("name"))
		},
	}
}

func newUpCommand(cfg Config) *cli.Command {
	return &cli.Command{
		Name:  "up",
		Usage: "Apply all up migrations",
		Flags: []cli.Flag{dryRunFlag()},
		Action: func(ctx context.Context, c *cli.Command) error {
			migrator, err := createMigrator(cfg)
			if err != nil {
				return err
			}
			return migrator.UpContext(ctx, runOpts(c, cfg)...)
		},
	}
}

func newUpToCommand(cfg Config) *cli.Command {
	return &cli.Command{
		Name:  "up-to",
		Usage: "Apply migrations up to a specific version",
		Flags: []cli.Flag{dryRunFlag(), versionFlag("Target version to migrate up to")},
		Action: func(ctx context.Context, c *cli.Command) error {
			migrator, err := createMigrator(cfg)
			if err != nil {
				return err
			}
			return migrator.UpToContext(ctx, c.Int64("version"), runOpts(c, cfg)...)
		},
	}
}

func newDownCommand(cfg Config) *cli.Command {
	return &cli.Command{
		Name:  "down",
		Usage: "Rollback the last migration",
		Flags: []cli.Flag{dryRunFlag()},
		Action: func(ctx context.Context, c *cli.Command) error {
			migrator, err := createMigrator(cfg)
			if err != nil {
				return err
			}
			return migrator.DownContext(ctx, runOpts(c, cfg)...)
		},
	}
}

func newDownToCommand(cfg Config) *cli.Command {
	return &cli.Command{
		Name:  "down-to",
		Usage: "Rollback migrations down to a specific version",
		Flags: []cli.Flag{dryRunFlag(), versionFlag("Target version to migrate down to")},
		Action: func(ctx context.Context, c *cli.Command) error {
			migrator, err := createMigrator(cfg)
			if err != nil {
				return err
			}
			return migrator.DownToContext(ctx, c.Int64("version"), runOpts(c, cfg)...)
		},
	}
}

func newResetCommand(cfg Config) *cli.Command {
	return &cli.Command{
		Name:  "reset",
		Usage: "Rollback all migrations",
		Flags: []cli.Flag{dryRunFlag()},
		Action: func(ctx context.Context, c *cli.Command) error {
			migrator, err := createMigrator(cfg)
			if err != nil {
				return err
			}
			return migrator.ResetContext(ctx, runOpts(c, cfg)...)
		},
	}
}

func newStatusCommand(cfg Config) *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Show the status of migrations",
		Action: func(ctx context.Context, _ *cli.Command) error {
			migrator, err := createMigrator(cfg)
			if err != nil {
				return err
			}
			return migrator.StatusContext(ctx)
		},
	}
}

func dryRunFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  "dry-run",
		Usage: "Simulate the migration without applying changes",
	}
}

func versionFlag(usage string) *cli.Int64Flag {
	return &cli.Int64Flag{
		Name:     "version",
		Aliases:  []string{"v"},
		Usage:    usage,
		Required: true,
	}
}

func createMigrator(cfg Config) (*migris.Migrate, error) {
	return migris.New(cfg.Dialect,
		migris.WithDB(cfg.DB),
		migris.WithMigrationDir(cfg.MigrationsDir),
	)
}

func runOpts(c *cli.Command, cfg Config) []migris.Option {
	var opts []migris.Option
	if c.Bool("dry-run") {
		opts = append(opts, migris.WithDryRun(true))
	}
	if cfg.AllowMissing {
		opts = append(opts, migris.WithAllowMissing(true))
	}
	return opts
}
