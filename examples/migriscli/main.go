package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/akfaiz/migris/examples/migriscli/migrations" // Import migrations directory
	"github.com/akfaiz/migris/extra/migriscli"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
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

	cmd := migriscli.NewCLI(migriscli.Config{
		DB:            db,
		Dialect:       "pgx",
		MigrationsDir: migrationDir,
	})
	err = cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
