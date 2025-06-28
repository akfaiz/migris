package main

import (
	"database/sql"
	"fmt"
	"os"
	"slices"

	"github.com/ahmadfaizk/schema"
	_ "github.com/ahmadfaizk/schema/examples/basic/migrations"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pressly/goose/v3"
)

func main() {
	validArgs := []string{"up", "down", "create"}
	if len(os.Args) < 2 || !slices.Contains(validArgs, os.Args[1]) {
		fmt.Println("Usage: go run main.go [up|down|create <migration_name>]")
		return
	}
	if os.Args[1] == "create" && len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go create <migration_name>")
		return
	}

	goose.SetDialect("postgres")
	schema.SetDialect("postgres")
	// schema.SetDebug(true)

	cfg := dbConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		Database: "schema_example",
		SSLMode:  "disable",
	}
	db, err := newDatabase(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	switch os.Args[1] {
	case "up":
		if err := goose.Up(db, "migrations"); err != nil {
			panic(fmt.Errorf("failed to run migrations: %w", err))
		}
	case "down":
		if err := goose.Reset(db, "migrations"); err != nil {
			panic(fmt.Errorf("failed to reset migrations: %w", err))
		}
	case "create":
		migrationName := os.Args[2]
		if err := goose.Create(db, "migrations", migrationName, "go"); err != nil {
			panic(fmt.Errorf("failed to create migration %s: %w", migrationName, err))
		}
	}
}

type dbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

func (db *dbConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Database, db.SSLMode)
}

func newDatabase(cfg dbConfig) (*sql.DB, error) {
	dsn := cfg.DSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
