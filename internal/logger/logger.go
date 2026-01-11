package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
func printBulletPoint(label, value string, colorFunc func(...interface{}) string) {
	fmt.Printf("%s %s: %s\n", grey(BulletChar), label, colorFunc(value))
}

func Info(msg string) {
	Infof("%s", msg)
}

func Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", whiteBgBlue(" INFO "), msg)
}

func PrintResults(results []*goose.MigrationResult) {
	for _, result := range results {
		PrintResult(result)
	}
}

func PrintResult(result *goose.MigrationResult) {
	durText := formatDuration(result.Duration.Seconds() * 1000)
	statusText := " DONE"
	if result.Error != nil {
		statusText = " FAIL"
	}

	name := result.Source.Path
	dots := createDottedLine(name, durText, statusText)

	// Print with appropriate color
	fmt.Printf("%s %s%s", name, grey(dots), grey(durText))
	if result.Error != nil {
		fmt.Printf("%s\n", redBold(statusText))
	} else {
		fmt.Printf("%s\n", greenBold(statusText))
	}
}

func PrintStatuses(statuses []*goose.MigrationStatus) {
	for _, status := range statuses {
		PrintStatus(status)
	}
}

func PrintStatus(status *goose.MigrationStatus) {
	var statusText string
	if status.State == goose.StateApplied {
		statusText = " Applied"
	} else {
		statusText = " Pending"
	}

	name := status.Source.Path
	dots := createDottedLine(name, "", statusText)

	fmt.Printf("%s %s", name, grey(dots))
	if status.State == goose.StateApplied {
		fmt.Printf("%s\n", greenBold(statusText))
	} else {
		fmt.Printf("%s\n", yellowBold(statusText))
	}
}

// DryRun specific logger functions

func DryRunStart(version int64) {
	fmt.Printf("%s Starting DRY RUN migration (UP) to version %d\n", whiteBgBlue(" DRY RUN "), version)
	fmt.Printf("%s Mode: DRY RUN - No actual database changes will be made\n\n", grey("📍"))
}

func DryRunMigrationStart(source string, version int64) {
	fmt.Printf("%s %s (version %d)\n", yellowBold("PROCESSING"), source, version)
}

func DryRunMigrationComplete(source string, duration float64) {
	durText := formatDuration(duration)
	statusText := " DRY RUN"
	dots := createDottedLine(source, durText, statusText)

	fmt.Printf("%s %s%s%s\n", source, grey(dots), grey(durText), greenBold(statusText))
}

func DryRunSQL(query string, args ...any) {
	fmt.Printf("%s %s\n", whiteBgGreen(" SQL "), query)
	if len(args) > 0 {
		fmt.Printf("%s Arguments: %v\n", grey("   "), args)
	}
	fmt.Println()
}

func DryRunSummary(totalMigrations, totalStatements int, duration float64) {
	fmt.Printf("%s DRY RUN Summary:\n", whiteBgBlue(" SUMMARY "))
	printBulletPoint("Total migrations processed", strconv.Itoa(totalMigrations), greenBold)
	printBulletPoint("Total SQL statements generated", strconv.Itoa(totalStatements), greenBold)
	printBulletPoint("Total execution time", fmt.Sprintf("%.2fms", duration), greenBold)
	printBulletPoint("Mode", "DRY RUN (no changes applied to database)", yellowBold)
}

// DryRun DOWN specific logger functions

func DryRunDownStart(version int64) {
	if version == 0 {
		fmt.Printf("%s Starting DRY RUN migration (RESET) - Rolling back all migrations\n", whiteBgRed(" DRY RUN "))
	} else {
		fmt.Printf("%s Starting DRY RUN migration (DOWN) to version %d\n", whiteBgRed(" DRY RUN "), version)
	}
	fmt.Printf("%s Mode: DRY RUN - No actual database changes will be made\n\n", grey("🔍"))
}

func DryRunDownSummary(totalMigrations, totalStatements int, duration float64, operation string) {
	fmt.Printf("%s DRY RUN %s Summary:\n", whiteBgRed(" SUMMARY "), operation)
	printBulletPoint("Total migrations processed", strconv.Itoa(totalMigrations), greenBold)
	printBulletPoint("Total SQL statements generated", strconv.Itoa(totalStatements), greenBold)
	printBulletPoint("Total execution time", fmt.Sprintf("%.2fms", duration), greenBold)
	printBulletPoint("Mode", "DRY RUN (no changes applied to database)", yellowBold)
}
