package migriscli //nolint:testpackage // Tests cover unexported command builders and option helpers.

import (
	"bytes"
	"context"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestNewCLI(t *testing.T) {
	cmd := NewCLI(Config{MigrationsDir: "migrations"})

	if cmd.Name != "migrate" {
		t.Fatalf("expected command name migrate, got %q", cmd.Name)
	}
	if cmd.Usage != "Database migration CLI tool" {
		t.Fatalf("unexpected command usage: %q", cmd.Usage)
	}
	if len(cmd.Commands) != 7 {
		t.Fatalf("expected 7 commands, got %d", len(cmd.Commands))
	}
	wantCommands := []string{"create", "up", "up-to", "down", "down-to", "reset", "status"}
	if !slices.Equal(commandNames(cmd.Commands), wantCommands) {
		t.Fatalf("expected commands %v, got %v", wantCommands, commandNames(cmd.Commands))
	}

	createCmd := cmd.Command("create")
	if createCmd == nil {
		t.Fatal("expected create command")
	}
	if len(createCmd.Flags) != 1 {
		t.Fatalf("expected 1 create flag, got %d", len(createCmd.Flags))
	}
	nameFlag, ok := createCmd.Flags[0].(*cli.StringFlag)
	if !ok {
		t.Fatalf("expected create flag to be *cli.StringFlag, got %T", createCmd.Flags[0])
	}
	if nameFlag.Name != "name" || !slices.Equal(nameFlag.Aliases, []string{"n"}) || !nameFlag.Required {
		t.Fatalf("unexpected create name flag: %#v", nameFlag)
	}

	upToCmd := cmd.Command("up-to")
	if upToCmd == nil {
		t.Fatal("expected up-to command")
	}
	if len(upToCmd.Flags) != 2 {
		t.Fatalf("expected 2 up-to flags, got %d", len(upToCmd.Flags))
	}
	versionFlag, ok := upToCmd.Flags[1].(*cli.Int64Flag)
	if !ok {
		t.Fatalf("expected version flag to be *cli.Int64Flag, got %T", upToCmd.Flags[1])
	}
	if versionFlag.Name != "version" || !slices.Equal(versionFlag.Aliases, []string{"v"}) || !versionFlag.Required {
		t.Fatalf("unexpected up-to version flag: %#v", versionFlag)
	}
}

func TestCreateCommandCreatesMigrationFile(t *testing.T) {
	dir := t.TempDir()
	cmd := NewCLI(Config{MigrationsDir: dir})
	cmd.Writer = bytes.NewBuffer(nil)
	cmd.ErrWriter = bytes.NewBuffer(nil)

	err := cmd.Run(context.Background(), []string{"migrate", "create", "--name", "create_users_table"})
	if err != nil {
		t.Fatalf("expected create command to succeed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("expected to read migration dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 migration file, got %d", len(entries))
	}
	if !strings.Contains(entries[0].Name(), "_create_users_table.go") {
		t.Fatalf("expected generated create users migration, got %q", entries[0].Name())
	}
}

func TestMigrationCommandsReturnMigratorConfigErrors(t *testing.T) {
	for _, name := range []string{"up", "up-to", "down", "down-to", "reset", "status"} {
		t.Run(name, func(t *testing.T) {
			args := []string{"migrate", name}
			if name == "up-to" || name == "down-to" {
				args = append(args, "--version", "20250101000000")
			}

			cmd := NewCLI(Config{Dialect: "unknown", MigrationsDir: t.TempDir()})
			cmd.Writer = bytes.NewBuffer(nil)
			cmd.ErrWriter = bytes.NewBuffer(nil)

			err := cmd.Run(context.Background(), args)
			if err == nil {
				t.Fatal("expected migrator config error")
			}
			if !strings.Contains(err.Error(), "unknown database dialect") {
				t.Fatalf("expected unknown dialect error, got %v", err)
			}
		})
	}
}

func TestRunOpts(t *testing.T) {
	var got []any
	cmd := runOptsTestCommand(&got, Config{AllowMissing: true})

	err := cmd.Run(context.Background(), []string{"up", "--dry-run"})
	if err != nil {
		t.Fatalf("expected dry-run command to succeed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 run options, got %d", len(got))
	}

	got = nil
	cmd = runOptsTestCommand(&got, Config{})
	err = cmd.Run(context.Background(), []string{"up"})
	if err != nil {
		t.Fatalf("expected command to succeed: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no run options, got %d", len(got))
	}
}

func runOptsTestCommand(got *[]any, cfg Config) *cli.Command {
	return &cli.Command{
		Name:  "up",
		Flags: []cli.Flag{dryRunFlag()},
		Action: func(_ context.Context, c *cli.Command) error {
			for _, opt := range runOpts(c, cfg) {
				*got = append(*got, opt)
			}
			return nil
		},
	}
}

func commandNames(commands []*cli.Command) []string {
	names := make([]string, len(commands))
	for i, cmd := range commands {
		names[i] = cmd.Name
	}
	return names
}
