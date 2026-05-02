package migris_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akfaiz/migris"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		migName     string
		wantErr     bool
		checkExists bool
	}{
		{
			name:        "create migration with simple name",
			migName:     "add_users_table",
			wantErr:     false,
			checkExists: true,
		},
		{
			name:        "create migration with create prefix",
			migName:     "create_posts_table",
			wantErr:     false,
			checkExists: true,
		},
		{
			name:        "create migration with update prefix",
			migName:     "update_users_table",
			wantErr:     false,
			checkExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			defer os.RemoveAll(tmpDir)

			// Run Create
			err := migris.Create(tmpDir, tt.migName)
			if tt.wantErr {
				require.Error(t, err, "Create() should return error")
				return
			}
			require.NoError(t, err, "Create() should not return error")

			// Check if migration file was created
			if tt.checkExists {
				entries, err := os.ReadDir(tmpDir)
				require.NoError(t, err, "failed to read temp dir")
				assert.NotEmpty(t, entries, "expected migration file to be created")

				// Check if file contains expected content
				filename := filepath.Join(tmpDir, entries[0].Name())
				content, err := os.ReadFile(filename)
				require.NoError(t, err, "failed to read migration file")

				contentStr := string(content)
				assert.Contains(
					t,
					contentStr,
					"package migrations",
					"migration file should contain 'package migrations'",
				)
				assert.Contains(t, contentStr, "schema.Context", "migration file should contain 'schema.Context'")
			}
		})
	}
}

func TestCreate_InvalidDirectory(t *testing.T) {
	err := migris.Create("/nonexistent/invalid/path", "test_migration")
	assert.Error(t, err, "Create() should return error for invalid directory")
}
