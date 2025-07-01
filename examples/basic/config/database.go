package config

import "fmt"

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

func (db Database) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Database, db.SSLMode)
}

func GetDatabase() Database {
	return Database{
		Host:     "localhost",
		Port:     5432,
		User:     "root",
		Password: "password",
		Database: "schema_example",
		SSLMode:  "disable",
	}
}
