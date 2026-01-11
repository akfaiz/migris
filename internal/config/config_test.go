package config_test

import (
	"testing"

	"github.com/akfaiz/migris/internal/config"
	"github.com/akfaiz/migris/internal/dialect"
	"github.com/stretchr/testify/assert"
)

func TestSetGetDialect(t *testing.T) {
	// Test setting and getting dialect
	config.SetDialect(dialect.Postgres)
	result := config.GetDialect()
	assert.Equal(t, dialect.Postgres, result)

	// Test setting MySQL dialect
	config.SetDialect(dialect.MySQL)
	result = config.GetDialect()
	assert.Equal(t, dialect.MySQL, result)
}
