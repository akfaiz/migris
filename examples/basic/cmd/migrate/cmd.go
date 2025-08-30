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
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "dry-run",
						Aliases:  []string{"d"},
						Usage:    "Preview the migration without applying changes",
						Required: false,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					return Up(c.Bool("dry-run"))
				},
			},
			{
				Name:  "reset",
				Usage: "Reset all migrations",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "dry-run",
						Aliases:  []string{"d"},
						Usage:    "Preview the migration without applying changes",
						Required: false,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					return Reset(c.Bool("dry-run"))
				},
			},
		},
	}
	return cmd
}
