package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver for testing
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for testing
	"github.com/testcontainers/testcontainers-go"
	tcmariadb "github.com/testcontainers/testcontainers-go/modules/mariadb"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	TestDBName     = "db_test"
	TestDBUser     = "root"
	TestDBPassword = "password"
)

func StartMySQLTestDB(ctx context.Context) (testcontainers.Container, *sql.DB, error) {
	container, err := tcmysql.Run(ctx, "mysql:8.0.36",
		tcmysql.WithDatabase(TestDBName),
		tcmysql.WithUsername(TestDBUser),
		tcmysql.WithPassword(TestDBPassword),
	)
	if err != nil {
		return nil, nil, err
	}

	connectionString, err := container.ConnectionString(ctx, "parseTime=true", "loc=Local")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	err = WaitForDBPing(ctx, db, 30*time.Second)
	if err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	return container, db, nil
}

func StartMariaDBTestDB(ctx context.Context) (testcontainers.Container, *sql.DB, error) {
	container, err := tcmariadb.Run(ctx, "mariadb:11.0.3",
		tcmariadb.WithDatabase(TestDBName),
		tcmariadb.WithUsername(TestDBUser),
		tcmariadb.WithPassword(TestDBPassword),
	)
	if err != nil {
		return nil, nil, err
	}

	connectionString, err := container.ConnectionString(ctx, "parseTime=true", "loc=Local")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	err = WaitForDBPing(ctx, db, 30*time.Second)
	if err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	return container, db, nil
}

func StartPostgresTestDB(ctx context.Context) (testcontainers.Container, *sql.DB, error) {
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase(TestDBName),
		tcpostgres.WithUsername(TestDBUser),
		tcpostgres.WithPassword(TestDBPassword),
	)
	if err != nil {
		return nil, nil, err
	}

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	err = WaitForDBPing(ctx, db, 30*time.Second)
	if err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	return container, db, nil
}

func WaitForDBPing(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("database did not become ready within %s", timeout)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
