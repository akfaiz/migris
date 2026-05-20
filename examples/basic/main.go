package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/akfaiz/migris"
	_ "github.com/akfaiz/migris/examples/basic/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const migrationDir = "migrations"

func loadDatabaseURL() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set in the environment")
	}
	return databaseURL
}

func newMigrator(db *sql.DB) *migris.Migrate {
	m, err := migris.New("pgx",
		migris.WithDB(db),
		migris.WithMigrationDir(migrationDir),
	)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	return m
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	databaseURL := loadDatabaseURL()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	subcmd := os.Args[1]
	args := os.Args[2:]

	switch subcmd {
	case "create":
		fs := flag.NewFlagSet("create", flag.ExitOnError)
		name := fs.String("name", "", "Name of the migration (required)")
		fs.StringVar(name, "n", "", "Name of the migration (required)")
		_ = fs.Parse(args)
		if *name == "" {
			fmt.Fprintln(os.Stderr, "error: --name is required")
			fs.Usage()
			os.Exit(1)
		}
		if err := newMigrator(db).Create(*name); err != nil {
			log.Fatalf("create failed: %v", err)
		}

	case "up":
		fs := flag.NewFlagSet("up", flag.ExitOnError)
		dryRun := fs.Bool("dry-run", false, "Simulate without applying changes")
		fs.BoolVar(dryRun, "d", false, "Simulate without applying changes")
		_ = fs.Parse(args)
		if err := newMigrator(db).UpContext(ctx, runOpts(*dryRun)...); err != nil {
			log.Fatalf("up failed: %v", err)
		}

	case "down":
		fs := flag.NewFlagSet("down", flag.ExitOnError)
		dryRun := fs.Bool("dry-run", false, "Simulate without applying changes")
		fs.BoolVar(dryRun, "d", false, "Simulate without applying changes")
		_ = fs.Parse(args)
		if err := newMigrator(db).DownContext(ctx, runOpts(*dryRun)...); err != nil {
			log.Fatalf("down failed: %v", err)
		}

	case "reset":
		fs := flag.NewFlagSet("reset", flag.ExitOnError)
		dryRun := fs.Bool("dry-run", false, "Simulate without applying changes")
		fs.BoolVar(dryRun, "d", false, "Simulate without applying changes")
		_ = fs.Parse(args)
		if err := newMigrator(db).ResetContext(ctx, runOpts(*dryRun)...); err != nil {
			log.Fatalf("reset failed: %v", err)
		}

	case "status":
		if err := newMigrator(db).StatusContext(ctx); err != nil {
			log.Fatalf("status failed: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", subcmd)
		printUsage()
		os.Exit(1)
	}
}

func runOpts(dryRun bool) []migris.Option {
	if dryRun {
		return []migris.Option{migris.WithDryRun(true)}
	}
	return nil
}

func printUsage() {
	fmt.Println("Usage: migrate <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create  -n <name>          Create a new migration file")
	fmt.Println("  up      [-d|--dry-run]      Apply all pending migrations")
	fmt.Println("  down    [-d|--dry-run]      Rollback the last migration")
	fmt.Println("  reset   [-d|--dry-run]      Rollback all migrations")
	fmt.Println("  status                      Show migration status")
}
