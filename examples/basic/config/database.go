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

func getDatabaseConfig() Database {
	return Database{
		Host:     getEnv("DB_HOST"),
		Port:     getEnvInt("DB_PORT"),
		User:     getEnv("DB_USER"),
		Password: getEnv("DB_PASSWORD"),
		Database: getEnv("DB_NAME"),
		SSLMode:  getEnv("DB_SSLMODE"),
	}
}
