package utils

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var do_log = false

var (
	StartTime time.Time
	LastTime  time.Time
)

func Log(text string, args ...interface{}) {
	if do_log {
		message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
		fmt.Println(message, args)
	}
}

func LogTime(text string) {
	duration := time.Since(StartTime)
	LastTime = time.Now()
	message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
	fmt.Println(message+" took ", duration)
}

func LogTimeSinceLast(text string) {
	duration := time.Since(LastTime)
	message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
	fmt.Printf("\noperation '%s' took %v", message, duration)
	LastTime = time.Now()
}
