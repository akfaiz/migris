package parser_test

import (
	"testing"

	"github.com/akfaiz/migris/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestParseMigrationName(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		wantTable  string
		wantCreate bool
	}{
		// Create table patterns
		{
			name:       "create table with table suffix",
			filename:   "create_users_table",
			wantTable:  "users",
			wantCreate: true,
		},
		{
			name:       "create table without table suffix",
			filename:   "create_posts",
			wantTable:  "posts",
			wantCreate: true,
		},
		{
			name:       "create table with underscores",
			filename:   "create_user_profiles_table",
			wantTable:  "user_profiles",
			wantCreate: true,
		},
		{
			name:       "create table with numbers",
			filename:   "create_table_v2",
			wantTable:  "table_v2",
			wantCreate: true,
		},

		// Add column patterns
		{
			name:       "add column with table suffix",
			filename:   "add_email_to_users_table",
			wantTable:  "users",
			wantCreate: false,
		},
		{
			name:       "add column without table suffix",
			filename:   "add_name_to_posts",
			wantTable:  "posts",
			wantCreate: false,
		},
		{
			name:       "add multiple columns",
			filename:   "add_first_name_last_name_to_users_table",
			wantTable:  "users",
			wantCreate: false,
		},
		{
			name:       "add column with underscores",
			filename:   "add_created_at_to_user_profiles",
			wantTable:  "user_profiles",
			wantCreate: false,
		},

		// Remove column patterns
		{
			name:       "remove column with table suffix",
			filename:   "remove_email_from_users_table",
			wantTable:  "users",
			wantCreate: false,
		},
		{
			name:       "remove column without table suffix",
			filename:   "remove_name_from_posts",
			wantTable:  "posts",
			wantCreate: false,
		},
		{
			name:       "remove multiple columns",
			filename:   "remove_old_field_from_users_table",
			wantTable:  "users",
			wantCreate: false,
		},
		{
			name:       "remove column with underscores",
			filename:   "remove_updated_at_from_user_profiles",
			wantTable:  "user_profiles",
			wantCreate: false,
		},
		{
			name:       "update table",
			filename:   "update_users_table",
			wantTable:  "users",
			wantCreate: false,
		},
		{
			name:       "empty filename",
			filename:   "",
			wantTable:  "",
			wantCreate: false,
		},
		{
			name:       "invalid format",
			filename:   "invalid_migration_name",
			wantTable:  "",
			wantCreate: false,
		},
		{
			name:       "uppercase letters",
			filename:   "create_Users_table",
			wantTable:  "",
			wantCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTable, gotCreate := parser.ParseMigrationName(tt.filename)
			assert.Equal(t, tt.wantTable, gotTable, "for filename: %s", tt.filename)
			assert.Equal(t, tt.wantCreate, gotCreate, "for filename: %s", tt.filename)
		})
	}
}
