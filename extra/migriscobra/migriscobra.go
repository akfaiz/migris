package migriscobra

import (
	"context"
	"database/sql"

	"github.com/akfaiz/migris"
	"github.com/spf13/cobra"
)

// Config holds the configuration for the migris CLI commands.
type Config struct {
	DB            *sql.DB // Database connection
	Dialect       string  // Database dialect (e.g., "pgx", "mysql", etc.)
	MigrationsDir string  // Directory where migration files are stored
}

// NewCLI creates a new CLI interface for migris with subcommands using Cobra.
func NewCLI(cfg Config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration CLI tool",
		Long:  "A powerful database migration tool powered by migris",
	}

	// Add subcommands
	rootCmd.AddCommand(
		createCreateCommand(cfg),
		createUpCommand(cfg),
		createUpToCommand(cfg),
		createDownCommand(cfg),
		createDownToCommand(cfg),
		createResetCommand(cfg),
		createStatusCommand(cfg),
	)

	return rootCmd
}

func createCreateCommand(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new migration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return cmd.Help()
			}
			return migris.Create(cfg.MigrationsDir, name)
		},
	}
	cmd.Flags().StringP("name", "n", "", "Name of the migration (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func createUpCommand(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply all up migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrator, err := createMigrator(cmd, cfg)
			if err != nil {
				return err
			}
			return migrator.UpContext(context.Background())
		},
	}
	cmd.Flags().Bool("dry-run", false, "Simulate the migration without applying changes")
	return cmd
}

func createUpToCommand(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up-to",
		Short: "Apply migrations up to a specific version",
		RunE: func(cmd *cobra.Command, args []string) error {
			version, _ := cmd.Flags().GetInt64("version")
			migrator, err := createMigrator(cmd, cfg)
			if err != nil {
				return err
			}
			return migrator.UpToContext(context.Background(), version)
		},
	}
	cmd.Flags().Bool("dry-run", false, "Simulate the migration without applying changes")
	cmd.Flags().Int64P("version", "v", 0, "Target version to migrate up to (required)")
	cmd.MarkFlagRequired("version")
	return cmd
}

func createDownCommand(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrator, err := createMigrator(cmd, cfg)
			if err != nil {
				return err
			}
			return migrator.DownContext(context.Background())
		},
	}
	cmd.Flags().Bool("dry-run", false, "Simulate the migration without applying changes")
	return cmd
}

func createDownToCommand(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down-to",
		Short: "Rollback migrations down to a specific version",
		RunE: func(cmd *cobra.Command, args []string) error {
			version, _ := cmd.Flags().GetInt64("version")
			migrator, err := createMigrator(cmd, cfg)
			if err != nil {
				return err
			}
			return migrator.DownToContext(context.Background(), version)
		},
	}
	cmd.Flags().Bool("dry-run", false, "Simulate the migration without applying changes")
	cmd.Flags().Int64P("version", "v", 0, "Target version to migrate down to (required)")
	cmd.MarkFlagRequired("version")
	return cmd
}

func createResetCommand(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Rollback all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrator, err := createMigrator(cmd, cfg)
			if err != nil {
				return err
			}
			return migrator.ResetContext(context.Background())
		},
	}
	cmd.Flags().Bool("dry-run", false, "Simulate the migration without applying changes")
	return cmd
}

func createStatusCommand(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrator, err := createMigrator(cmd, cfg)
			if err != nil {
				return err
			}
			return migrator.StatusContext(context.Background())
		},
	}
	return cmd
}

func createMigrator(cmd *cobra.Command, cfg Config) (*migris.Migrate, error) {
	options := []migris.Option{
		migris.WithDB(cfg.DB),
		migris.WithMigrationDir(cfg.MigrationsDir),
	}

	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		options = append(options, migris.WithDryRun(true))
	}

	migrator, err := migris.New(cfg.Dialect, options...)
	if err != nil {
		return nil, err
	}

	return migrator, nil
}
