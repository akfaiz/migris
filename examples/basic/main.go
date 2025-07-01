package main

import (
	"context"
	"log"
	"os"

	"github.com/ahmadfaizk/schema/examples/basic/cmd/migrate"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "schema-example",
		Usage: "A simple schema example",
		Commands: []*cli.Command{
			{
				Name:  "migrate",
				Usage: "Run database migrations",
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
							return migrate.Create(c.String("name"))
						},
					},
					{
						Name:  "up",
						Usage: "Run all pending migrations",
						Action: func(ctx context.Context, c *cli.Command) error {
							return migrate.Up()
						},
					},
					{
						Name:  "down",
						Usage: "Reset the database by rolling back all migrations",
						Action: func(ctx context.Context, c *cli.Command) error {
							return migrate.Down()
						},
					},
				},
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
