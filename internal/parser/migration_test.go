package parser_test

import (
	"testing"

	"github.com/afkdevs/go-schema/internal/parser"
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

		// Unknown patterns
		{
			name:       "unknown migration pattern",
			filename:   "update_users_table",
			wantTable:  "",
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
			if gotTable != tt.wantTable {
				t.Errorf("ParseMigrationName() gotTable = %v, want %v", gotTable, tt.wantTable)
			}
			if gotCreate != tt.wantCreate {
				t.Errorf("ParseMigrationName() gotCreate = %v, want %v", gotCreate, tt.wantCreate)
			}
		})
	}
}
