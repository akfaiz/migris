package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/akfaiz/migris"
	_ "github.com/akfaiz/migris/examples/basic/migrations" // Import migrations directory
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

const migrationDir = "migrations"

func loadDatabaseURL() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set in the environment")
	}
	return databaseURL
}

func main() {
	databaseURL := loadDatabaseURL()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrator, err := migris.New("pgx", migris.WithDB(db), migris.WithMigrationDir(migrationDir))
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}

	cmd := &cli.Command{
		Name:  "migrate",
		Usage: "Migration tool",
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
					return migrator.Create(c.String("name"))
				},
			},
			{
				Name:  "up",
				Usage: "Run all pending migrations",
				Action: func(ctx context.Context, c *cli.Command) error {
					return migrator.UpContext(ctx)
				},
			},
			{
				Name:  "reset",
				Usage: "Rollback all migrations",
				Action: func(ctx context.Context, c *cli.Command) error {
					return migrator.ResetContext(ctx)
				},
			},
			{
				Name:  "down",
				Usage: "Rollback the last migration",
				Action: func(ctx context.Context, c *cli.Command) error {
					return migrator.DownContext(ctx)
				},
			},
			{
				Name:  "status",
				Usage: "Show the status of migrations",
				Action: func(ctx context.Context, c *cli.Command) error {
					return migrator.StatusContext(ctx)
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
