package logger

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/pressly/goose/v3"
	"golang.org/x/term"
)

// Constants.
const (
	DefaultTerminalWidth = 80
	DotChar              = "."
	BulletChar           = "   •"
)

var (
	grey       = color.New(color.FgHiBlack).SprintFunc()
	greenBold  = color.New(color.FgGreen, color.Bold).SprintFunc()
	yellowBold = color.New(color.FgYellow, color.Bold).SprintFunc()
	redBold    = color.New(color.FgRed, color.Bold).SprintFunc()

	// Badge colors.
	whiteBgBlue  = color.New(color.FgWhite, color.BgBlue).SprintFunc()
	whiteBgGreen = color.New(color.FgWhite, color.BgGreen).SprintFunc()
	whiteBgRed   = color.New(color.FgWhite, color.BgRed).SprintFunc()
)

// Logger represents a configurable logger instance.
type Logger struct {
	output io.Writer
	mu     sync.RWMutex
}

var (
	instance *Logger
	once     sync.Once
)

// Get returns the logger instance.
func Get() *Logger {
	once.Do(func() {
		instance = &Logger{
			output: os.Stdout,
		}
	})
	return instance
}

// SetOutput sets the output writer for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// Printf is a helper method that writes to the configured output.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.output, format, args...)
}

// Helper functions

// getTerminalWidth returns the terminal width with a fallback.
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return DefaultTerminalWidth
	}
	return width
}

// createDottedLine creates a line with dots to fill space between name and status.
func createDottedLine(name, durText, statusText string) string {
	width := getTerminalWidth()
	fillLen := width - len(name) - len(durText) - len(statusText) - 1
	if fillLen < 0 {
		fillLen = 0
	}
	return strings.Repeat(DotChar, fillLen)
}

// formatDuration formats duration in milliseconds.
func formatDuration(ms float64) string {
	return fmt.Sprintf(" %.2fms", ms)
}

// printBulletPoint prints a formatted bullet point with colored text.
func (l *Logger) printBulletPoint(label, value string, colorFunc func(...interface{}) string) {
	l.Printf("%s %s: %s\n", grey(BulletChar), label, colorFunc(value))
}

func (l *Logger) Info(msg string) {
	l.Infof("%s", msg)
}

func (l *Logger) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.Printf("%s %s\n", whiteBgBlue(" INFO "), msg)
}

func (l *Logger) PrintResults(results []*goose.MigrationResult) {
	for _, result := range results {
		l.PrintResult(result)
	}
}

func (l *Logger) PrintResult(result *goose.MigrationResult) {
	durText := formatDuration(result.Duration.Seconds() * 1000)
	statusText := " DONE"
	if result.Error != nil {
		statusText = " FAIL"
	}

	name := result.Source.Path
	dots := createDottedLine(name, durText, statusText)

	// Print with appropriate color
	l.Printf("%s %s%s", name, grey(dots), grey(durText))
	if result.Error != nil {
		l.Printf("%s\n", redBold(statusText))
	} else {
		l.Printf("%s\n", greenBold(statusText))
	}
}

func (l *Logger) PrintStatuses(statuses []*goose.MigrationStatus) {
	for _, status := range statuses {
		l.PrintStatus(status)
	}
}

func (l *Logger) PrintStatus(status *goose.MigrationStatus) {
	var statusText string
	if status.State == goose.StateApplied {
		statusText = " Applied"
	} else {
		statusText = " Pending"
	}

	name := status.Source.Path
	dots := createDottedLine(name, "", statusText)

	l.Printf("%s %s", name, grey(dots))
	if status.State == goose.StateApplied {
		l.Printf("%s\n", greenBold(statusText))
	} else {
		l.Printf("%s\n", yellowBold(statusText))
	}
}

// DryRun specific logger functions

func (l *Logger) DryRunStart(version int64) {
	l.Printf("%s Starting DRY RUN migration (UP) to version %d\n", whiteBgBlue(" DRY RUN "), version)
	l.Printf("%s Mode: DRY RUN - No actual database changes will be made\n\n", grey("📍"))
}

func (l *Logger) DryRunMigrationStart(source string, version int64) {
	l.Printf("%s %s (version %d)\n", yellowBold("PROCESSING"), source, version)
}

func (l *Logger) DryRunMigrationComplete(source string, duration float64) {
	durText := formatDuration(duration)
	statusText := " DRY RUN"
	dots := createDottedLine(source, durText, statusText)

	l.Printf("%s %s%s%s\n", source, grey(dots), grey(durText), greenBold(statusText))
}

func (l *Logger) DryRunSQL(query string, args ...any) {
	l.Printf("%s %s\n", whiteBgGreen(" SQL "), query)
	if len(args) > 0 {
		l.Printf("%s Arguments: %v\n", grey("   "), args)
	}
	l.Printf("\n")
}

func (l *Logger) DryRunSummary(totalMigrations, totalStatements int) {
	l.Printf("%s DRY RUN Summary:\n", whiteBgBlue(" SUMMARY "))
	l.printBulletPoint("Total migrations processed", strconv.Itoa(totalMigrations), greenBold)
	l.printBulletPoint("Total SQL statements generated", strconv.Itoa(totalStatements), greenBold)
	l.printBulletPoint("Mode", "DRY RUN (no changes applied to database)", yellowBold)
}

// DryRun DOWN specific logger functions

func (l *Logger) DryRunDownStart(version int64) {
	if version == 0 {
		l.Printf("%s Starting DRY RUN migration (RESET) - Rolling back all migrations\n", whiteBgRed(" DRY RUN "))
	} else {
		l.Printf("%s Starting DRY RUN migration (DOWN) to version %d\n", whiteBgRed(" DRY RUN "), version)
	}
	l.Printf("%s Mode: DRY RUN - No actual database changes will be made\n\n", grey("🔍"))
}

func (l *Logger) DryRunDownSummary(totalMigrations, totalStatements int, operation string) {
	l.Printf("%s DRY RUN %s Summary:\n", whiteBgRed(" SUMMARY "), operation)
	l.printBulletPoint("Total migrations processed", strconv.Itoa(totalMigrations), greenBold)
	l.printBulletPoint("Total SQL statements generated", strconv.Itoa(totalStatements), greenBold)
	l.printBulletPoint("Mode", "DRY RUN (no changes applied to database)", yellowBold)
}
