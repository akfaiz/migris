package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/akfaiz/migris/examples/migriscobra/migrations" // Import migrations directory
	"github.com/akfaiz/migris/extra/migriscobra"
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

	cmd := migriscobra.NewCLI(migriscobra.Config{
		DB:            db,
		Dialect:       "pgx",
		MigrationsDir: migrationDir,
	})
	err = cmd.Execute()
}
