package schema_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcmariadb "github.com/testcontainers/testcontainers-go/modules/mariadb"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testDBName     = "db_test"
	testDBUser     = "root"
	testDBPassword = "password"
)

func startMySQLTestDB(ctx context.Context) (testcontainers.Container, *sql.DB, error) {
	container, err := tcmysql.Run(ctx, "mysql:8.0.36",
		tcmysql.WithDatabase(testDBName),
		tcmysql.WithUsername(testDBUser),
		tcmysql.WithPassword(testDBPassword),
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

	if err := waitForDBPing(ctx, db, 30*time.Second); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	return container, db, nil
}

func startMariaDBTestDB(ctx context.Context) (testcontainers.Container, *sql.DB, error) {
	container, err := tcmariadb.Run(ctx, "mariadb:11.0.3",
		tcmariadb.WithDatabase(testDBName),
		tcmariadb.WithUsername(testDBUser),
		tcmariadb.WithPassword(testDBPassword),
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

	if err := waitForDBPing(ctx, db, 30*time.Second); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	return container, db, nil
}

func startPostgresTestDB(ctx context.Context) (testcontainers.Container, *sql.DB, error) {
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testDBUser),
		tcpostgres.WithPassword(testDBPassword),
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

	if err := waitForDBPing(ctx, db, 30*time.Second); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	return container, db, nil
}

func waitForDBPing(ctx context.Context, db *sql.DB, timeout time.Duration) error {
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
