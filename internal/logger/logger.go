package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pressly/goose/v3"
	"golang.org/x/term"
)

var (
	grey       = color.New(color.FgHiBlack).SprintFunc()
	greenBold  = color.New(color.FgGreen, color.Bold).SprintFunc()
	yellowBold = color.New(color.FgYellow, color.Bold).SprintFunc()
	redBold    = color.New(color.FgRed, color.Bold).SprintFunc()
)

func Info(msg string) {
	Infof("%s", msg)
}

func Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)

	whiteBgBlue := color.New(color.FgWhite, color.BgBlue).SprintFunc()
	fmt.Printf("%s %s\n", whiteBgBlue(" INFO "), msg)
}

func PrintResults(results []*goose.MigrationResult) {
	for _, result := range results {
		PrintResult(result)
	}
}

func PrintResult(result *goose.MigrationResult) {
	// Get terminal width
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80 // fallback
	}

	durText := fmt.Sprintf(" %.2fms", result.Duration.Seconds()*1000)
	statusText := " DONE"
	if result.Error != nil {
		statusText = " FAIL"
	}

	name := result.Source.Path

	// Calculate how many dots to add
	fillLen := width - len(name) - len(durText) - len(statusText) - 1
	if fillLen < 0 {
		fillLen = 0
	}
	dots := strings.Repeat(".", fillLen)

	// Print
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
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80 // fallback
	}

	var statusText string
	if status.State == goose.StateApplied {
		statusText = " Applied"
	} else {
		statusText = " Pending"
	}

	name := status.Source.Path

	fillLen := width - len(name) - len(statusText) - 1
	if fillLen < 0 {
		fillLen = 0
	}
	dots := strings.Repeat(".", fillLen)

	fmt.Printf("%s %s", name, grey(dots))
	if status.State == goose.StateApplied {
		fmt.Printf("%s\n", greenBold(statusText))
	} else {
		fmt.Printf("%s\n", yellowBold(statusText))
	}
}
