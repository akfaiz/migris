package parser

import "regexp"

// Migration pattern types
type MigrationType int

const (
	MigrationTypeUnknown MigrationType = iota
	MigrationTypeCreate
	MigrationTypeUpdate
)

// Migration patterns compiled once for performance
var (
	// Create table patterns: create_users, create_users_table
	createTablePattern = regexp.MustCompile(`^create_(?P<table>[a-z0-9_]+?)(?:_table)?$`)

	// Add column patterns: add_email_to_users, add_columns_to_users_table
	addColumnPattern = regexp.MustCompile(`^add_(?P<columns>[a-z0-9_]+?)_to_(?P<table>[a-z0-9_]+?)(?:_table)?$`)

	// Remove column patterns: remove_email_from_users, drop_column_from_users_table
	removeColumnPattern = regexp.MustCompile(`^(?:remove|drop)_(?P<columns>[a-z0-9_]+?)_(?:from|in)_(?P<table>[a-z0-9_]+?)(?:_table)?$`)

	// Update/modify table patterns: update_users, modify_users_table, alter_users
	updateTablePattern = regexp.MustCompile(`^(?:update|modify|alter)_(?P<table>[a-z0-9_]+?)(?:_table)?$`)

	// Index patterns: add_index_to_users, create_index_on_users
	indexPattern = regexp.MustCompile(`^(?:add|create)_(?:index|idx)_(?:to|on)_(?P<table>[a-z0-9_]+?)(?:_table)?$`)

	// Foreign key patterns: add_foreign_key_to_users
	foreignKeyPattern = regexp.MustCompile(`^add_(?:foreign_key|fk)_to_(?P<table>[a-z0-9_]+?)(?:_table)?$`)

	// Drop table patterns: drop_users, drop_users_table
	dropTablePattern = regexp.MustCompile(`^drop_(?P<table>[a-z0-9_]+?)(?:_table)?$`)
)

// MigrationInfo contains parsed migration information
type MigrationInfo struct {
	TableName string
	Type      MigrationType
	IsCreate  bool // Backwards compatibility
}

// ParseMigrationName analyzes a migration filename and extracts table information
func ParseMigrationName(filename string) (tableName string, create bool) {
	info := ParseMigrationInfo(filename)
	return info.TableName, info.IsCreate
}

// ParseMigrationInfo provides detailed migration information
func ParseMigrationInfo(filename string) MigrationInfo {
	patterns := []struct {
		pattern  *regexp.Regexp
		isCreate bool
		migType  MigrationType
	}{
		{createTablePattern, true, MigrationTypeCreate},
		{addColumnPattern, false, MigrationTypeUpdate},
		{removeColumnPattern, false, MigrationTypeUpdate},
		{updateTablePattern, false, MigrationTypeUpdate},
		{indexPattern, false, MigrationTypeUpdate},
		{foreignKeyPattern, false, MigrationTypeUpdate},
		{dropTablePattern, false, MigrationTypeUpdate},
	}

	for _, p := range patterns {
		if matches := p.pattern.FindStringSubmatch(filename); matches != nil {
			tableName := extractTableName(p.pattern, matches)
			return MigrationInfo{
				TableName: tableName,
				Type:      p.migType,
				IsCreate:  p.isCreate,
			}
		}
	}

	return MigrationInfo{
		TableName: "",
		Type:      MigrationTypeUnknown,
		IsCreate:  false,
	}
}

// extractTableName extracts the table name from regex matches
func extractTableName(pattern *regexp.Regexp, matches []string) string {
	names := pattern.SubexpNames()
	for i, name := range names {
		if name == "table" && i < len(matches) {
			return matches[i]
		}
	}
	return ""
}
