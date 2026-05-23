package migriscobra //nolint:testpackage // Tests cover unexported command builders and option helpers.

import (
	"bytes"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCLI(t *testing.T) {
	cmd := NewCLI(Config{MigrationsDir: "migrations"})

	if cmd.Use != "migrate" {
		t.Fatalf("expected command use migrate, got %q", cmd.Use)
	}
	if cmd.Short != "Database migration CLI tool" {
		t.Fatalf("unexpected command short help: %q", cmd.Short)
	}
	if cmd.Long != "A powerful database migration tool powered by migris" {
		t.Fatalf("unexpected command long help: %q", cmd.Long)
	}
	wantCommands := []string{"create", "up", "up-to", "down", "down-to", "reset", "status"}
	gotCommands := cobraCommandNames(cmd.Commands())
	slices.Sort(wantCommands)
	slices.Sort(gotCommands)
	if !slices.Equal(gotCommands, wantCommands) {
		t.Fatalf("expected commands %v, got %v", wantCommands, gotCommands)
	}

	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("expected to find create command: %v", err)
	}
	if createCmd == nil {
		t.Fatal("expected create command")
	}
	nameFlag := createCmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatal("expected create name flag")
	}
	if nameFlag.Shorthand != "n" ||
		!slices.Contains(nameFlag.Annotations["cobra_annotation_bash_completion_one_required_flag"], "true") {
		t.Fatalf("unexpected create name flag: %#v", nameFlag)
	}

	upToCmd, _, err := cmd.Find([]string{"up-to"})
	if err != nil {
		t.Fatalf("expected to find up-to command: %v", err)
	}
	if upToCmd == nil {
		t.Fatal("expected up-to command")
	}
	versionFlag := upToCmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Fatal("expected up-to version flag")
	}
	if versionFlag.Shorthand != "v" ||
		!slices.Contains(versionFlag.Annotations["cobra_annotation_bash_completion_one_required_flag"], "true") {
		t.Fatalf("unexpected up-to version flag: %#v", versionFlag)
	}
}

func TestCreateCommandCreatesMigrationFile(t *testing.T) {
	dir := t.TempDir()
	cmd := NewCLI(Config{MigrationsDir: dir})
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{"create", "--name", "create_users_table"})

	err := cmd.Execute()
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
			args := []string{name}
			if name == "up-to" || name == "down-to" {
				args = append(args, "--version", "20250101000000")
			}

			cmd := NewCLI(Config{Dialect: "unknown", MigrationsDir: t.TempDir()})
			cmd.SetOut(bytes.NewBuffer(nil))
			cmd.SetErr(bytes.NewBuffer(nil))
			cmd.SetArgs(args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected migrator config error")
			}
			if !strings.Contains(err.Error(), "unknown database dialect") {
				t.Fatalf("expected unknown dialect error, got %v", err)
			}
		})
	}
}

func TestRequiredFlags(t *testing.T) {
	for _, args := range [][]string{{"create"}, {"up-to"}, {"down-to"}} {
		t.Run(args[0], func(t *testing.T) {
			cmd := NewCLI(Config{MigrationsDir: t.TempDir()})
			cmd.SetOut(bytes.NewBuffer(nil))
			cmd.SetErr(bytes.NewBuffer(nil))
			cmd.SetArgs(args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected required flag error")
			}
			if !strings.Contains(err.Error(), "required flag") {
				t.Fatalf("expected required flag error, got %v", err)
			}
		})
	}
}

func TestRunOpts(t *testing.T) {
	cmd := createUpCommand(Config{})
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("expected to set dry-run flag: %v", err)
	}

	opts := runOpts(cmd, Config{AllowMissing: true})
	if len(opts) != 2 {
		t.Fatalf("expected 2 run options, got %d", len(opts))
	}

	cmd = createUpCommand(Config{})
	opts = runOpts(cmd, Config{})
	if len(opts) != 0 {
		t.Fatalf("expected no run options, got %d", len(opts))
	}
}

func cobraCommandNames(commands []*cobra.Command) []string {
	names := make([]string, 0, len(commands))
	for _, cmd := range commands {
		if cmd.Use == "help [command]" {
			continue
		}
		names = append(names, cmd.Use)
	}
	return names
}
