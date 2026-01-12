package dialect_test

import (
	"testing"

	"github.com/akfaiz/migris/internal/dialect"
	"github.com/stretchr/testify/assert"
)

func TestDialectString(t *testing.T) {
	tests := []struct {
		dialect  dialect.Dialect
		expected string
	}{
		{dialect.Postgres, "postgres"},
		{dialect.MySQL, "mysql"},
		{dialect.SQLite3, "sqlite3"},
		{dialect.Unknown, ""},
	}

	for _, test := range tests {
		result := test.dialect.String()
		assert.Equal(t, test.expected, result)
	}
}

func TestGooseDialect(t *testing.T) {
	tests := []struct {
		dialect  dialect.Dialect
		expected string
	}{
		{dialect.Postgres, "postgres"},
		{dialect.MySQL, "mysql"},
		{dialect.SQLite3, "sqlite3"},
		{dialect.Unknown, ""},
	}

	for _, test := range tests {
		result := test.dialect.GooseDialect()
		assert.Equal(t, test.expected, string(result))
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected dialect.Dialect
	}{
		{"postgres", dialect.Postgres},
		{"pgx", dialect.Postgres},
		{"mysql", dialect.MySQL},
		{"mariadb", dialect.MySQL},
		{"sqlite3", dialect.SQLite3},
		{"sqlite", dialect.SQLite3},
		{"unknown", dialect.Unknown}, // default
	}

	for _, test := range tests {
		result := dialect.FromString(test.input)
		assert.Equal(t, test.expected, result)
	}
}
