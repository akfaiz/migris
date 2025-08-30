package migrate

import (
	"context"

	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
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
					return Create(c.String("name"))
				},
			},
			{
				Name:  "up",
				Usage: "Run all pending migrations",
				Action: func(ctx context.Context, c *cli.Command) error {
					return Up()
				},
			},
			{
				Name:  "reset",
				Usage: "Reset all migrations",
				Action: func(ctx context.Context, c *cli.Command) error {
					return Reset()
				},
			},
			{
				Name:  "down",
				Usage: "Revert the last migration",
				Action: func(ctx context.Context, c *cli.Command) error {
					return Down()
				},
			},
			{
				Name:  "status",
				Usage: "Show the status of migrations",
				Action: func(ctx context.Context, c *cli.Command) error {
					return Status()
				},
			},
		},
	}
	return cmd
}
