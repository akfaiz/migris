package cmd

import (
	"context"

	"github.com/akfaiz/migris/examples/basic/cmd/migrate"
	"github.com/urfave/cli/v3"
)

var cmd = &cli.Command{
	Name:  "schema-example",
	Usage: "A simple schema example",
	Commands: []*cli.Command{
		migrate.Command(),
	},
}

func Execute(args []string) error {
	return cmd.Run(context.Background(), args)
}
