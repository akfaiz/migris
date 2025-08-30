package parser

import "regexp"

func ParseMigrationName(filename string) (tableName string, create bool) {
	// Regex patterns for common migration styles
	createPattern := regexp.MustCompile(`^create_(?P<table>[a-z0-9_]+?)(?:_table)?$`)
	addColPattern := regexp.MustCompile(`^add_(?P<columns>[a-z0-9_]+?)_to_(?P<table>[a-z0-9_]+?)(?:_table)?$`)
	removeColPattern := regexp.MustCompile(`^remove_(?P<columns>[a-z0-9_]+?)_from_(?P<table>[a-z0-9_]+?)(?:_table)?$`)

	switch {
	case createPattern.MatchString(filename):
		return createPattern.ReplaceAllString(filename, "${table}"), true
	case addColPattern.MatchString(filename):
		return addColPattern.ReplaceAllString(filename, "${table}"), false
	case removeColPattern.MatchString(filename):
		return removeColPattern.ReplaceAllString(filename, "${table}"), false
	default:
		return "", false // Unknown migration style
	}
}
