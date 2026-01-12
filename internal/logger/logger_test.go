package logger_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/akfaiz/migris/internal/logger"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
)

func TestLogger_Get(t *testing.T) {
	instance1 := logger.Get()
	instance2 := logger.Get()

	assert.NotNil(t, instance1, "Get should not return nil")
	assert.Equal(t, instance1, instance2, "Get should return the same instance (singleton)")
}

func TestLogger_SetOutput(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}

	l.SetOutput(buf)

	l.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "INFO", "Output should contain INFO badge")
	assert.Contains(t, output, "test message", "Output should contain the message")
}

func TestLogger_Infof(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.Infof("test %s with %d args", "message", 2)

	output := buf.String()
	assert.Contains(t, output, "test message with 2 args", "Formatted message should be correct")
}

func TestLogger_DryRunLogging(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.DryRunStart(123)
	l.DryRunSQL("SELECT * FROM users")
	l.DryRunSummary(1, 1)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.GreaterOrEqual(t, len(lines), 3, "Should have multiple lines of output")
	assert.Contains(t, output, "DRY RUN", "Should contain DRY RUN badge")
	assert.Contains(t, output, "version 123", "Should contain version number")
	assert.Contains(
		t,
		output,
		"SELECT * FROM users",
		"Should contain SQL query",
	)
	assert.Contains(t, output, "SUMMARY", "Should contain summary")
}

func TestLogger_ConcurrentAccess(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	// Test that concurrent access doesn't cause data races
	done := make(chan bool, 2)

	go func() {
		for range 10 {
			l.Info("goroutine 1")
		}
		done <- true
	}()

	go func() {
		for range 10 {
			l.Info("goroutine 2")
		}
		done <- true
	}()

	<-done
	<-done

	output := buf.String()
	count1 := strings.Count(output, "goroutine 1")
	count2 := strings.Count(output, "goroutine 2")

	assert.Equal(t, 10, count1, "Should have all messages from goroutine 1")
	assert.Equal(t, 10, count2, "Should have all messages from goroutine 2")
}

func TestLogger_PrintResults(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	// Create mock migration results
	results := []*goose.MigrationResult{
		{
			Source: &goose.Source{
				Path: "001_create_users.sql",
			},
			Duration: 50 * time.Millisecond,
			Error:    nil,
		},
		{
			Source: &goose.Source{
				Path: "002_create_posts.sql",
			},
			Duration: 75 * time.Millisecond,
			Error:    nil,
		},
		{
			Source: &goose.Source{
				Path: "003_create_comments.sql",
			},
			Duration: 100 * time.Millisecond,
			Error:    errors.New("migration failed"),
		},
	}

	l.PrintResults(results)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 3 lines for 3 results
	assert.Len(t, lines, 3, "Should have output for each result")

	// Check that all migration files are printed
	assert.Contains(t, output, "001_create_users.sql", "Should contain first migration")
	assert.Contains(t, output, "002_create_posts.sql", "Should contain second migration")
	assert.Contains(t, output, "003_create_comments.sql", "Should contain third migration")

	// Check that DONE appears for successful migrations
	assert.Contains(t, output, "DONE", "Should contain DONE status")

	// Check that FAIL appears for failed migration
	assert.Contains(t, output, "FAIL", "Should contain FAIL status")
}

func TestLogger_PrintStatuses(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	// Create mock migration statuses
	statuses := []*goose.MigrationStatus{
		{
			Source: &goose.Source{
				Path: "001_create_users.sql",
			},
			State: goose.StateApplied,
		},
		{
			Source: &goose.Source{
				Path: "002_create_posts.sql",
			},
			State: goose.StateApplied,
		},
		{
			Source: &goose.Source{
				Path: "003_create_comments.sql",
			},
			State: goose.StatePending,
		},
	}

	l.PrintStatuses(statuses)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 3 lines for 3 statuses
	assert.Len(t, lines, 3, "Should have output for each status")

	// Check that all migration files are printed
	assert.Contains(t, output, "001_create_users.sql", "Should contain first migration")
	assert.Contains(t, output, "002_create_posts.sql", "Should contain second migration")
	assert.Contains(t, output, "003_create_comments.sql", "Should contain third migration")

	// Check that Applied appears for applied migrations
	assert.Contains(t, output, "Applied", "Should contain Applied status")

	// Check that Pending appears for pending migration
	assert.Contains(t, output, "Pending", "Should contain Pending status")
}

func TestLogger_DryRunMigrationStart(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	source := "001_create_users.sql"
	version := int64(123)

	l.DryRunMigrationStart(source, version)

	output := buf.String()

	assert.Contains(t, output, "PROCESSING", "Should contain PROCESSING text")
	assert.Contains(t, output, source, "Should contain the source filename")
	assert.Contains(t, output, "version 123", "Should contain the version number")
	assert.Contains(t, output, "(version 123)", "Should have proper version format")
}

func TestLogger_DryRunMigrationComplete(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	source := "001_create_users.sql"
	duration := 50.5 // milliseconds

	l.DryRunMigrationComplete(source, duration)

	output := buf.String()

	assert.Contains(t, output, source, "Should contain the source filename")
	assert.Contains(t, output, "DRY RUN", "Should contain DRY RUN status")
	assert.Contains(t, output, "50.50ms", "Should contain formatted duration")
}

func TestLogger_DryRunSQL(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.DryRunSQL("SELECT * FROM users")

	output := buf.String()
	assert.Contains(t, output, "SQL", "Should contain SQL badge")
	assert.Contains(
		t,
		output,
		"SELECT * FROM users",
		"Should contain the query",
	)
}

func TestLogger_DryRunSQL_WithArgs(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.DryRunSQL("INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com")

	output := buf.String()
	assert.Contains(t, output, "SQL", "Should contain SQL badge")
	assert.Contains(t, output, "INSERT INTO users (name, email) VALUES (?, ?)", "Should contain the query")
	assert.Contains(t, output, "Arguments:", "Should contain Arguments label")
	assert.Contains(t, output, "John", "Should contain first argument")
	assert.Contains(t, output, "john@example.com", "Should contain second argument")
}

func TestLogger_DryRunSummary(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.DryRunSummary(5, 12)

	output := buf.String()
	assert.Contains(t, output, "SUMMARY", "Should contain SUMMARY badge")
	assert.Contains(t, output, "DRY RUN Summary", "Should contain DRY RUN Summary header")
	assert.Contains(t, output, "Total migrations processed", "Should contain migrations label")
	assert.Contains(t, output, "5", "Should contain total migrations count")
	assert.Contains(t, output, "Total SQL statements generated", "Should contain statements label")
	assert.Contains(t, output, "12", "Should contain total statements count")
	assert.Contains(t, output, "Mode", "Should contain Mode label")
	assert.Contains(t, output, "DRY RUN (no changes applied to database)", "Should contain mode description")
}

func TestLogger_DryRunDownSummary(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.DryRunDownSummary(3, 8, "ROLLBACK")

	output := buf.String()
	assert.Contains(t, output, "SUMMARY", "Should contain SUMMARY badge")
	assert.Contains(t, output, "DRY RUN ROLLBACK Summary", "Should contain operation in header")
	assert.Contains(t, output, "Total migrations processed", "Should contain migrations label")
	assert.Contains(t, output, "3", "Should contain total migrations count")
	assert.Contains(t, output, "Total SQL statements generated", "Should contain statements label")
	assert.Contains(t, output, "8", "Should contain total statements count")
	assert.Contains(t, output, "Mode", "Should contain Mode label")
}

func TestLogger_DryRunDownStart(t *testing.T) {
	l := logger.Get()
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.DryRunDownStart(456)

	output := buf.String()
	assert.Contains(t, output, "DRY RUN", "Should contain DRY RUN badge")
	assert.Contains(t, output, "DOWN", "Should contain DOWN keyword")
	assert.Contains(t, output, "version 456", "Should contain the version number")
	assert.Contains(t, output, "Mode: DRY RUN", "Should contain mode information")
	assert.Contains(t, output, "No actual database changes will be made", "Should contain warning message")
}
