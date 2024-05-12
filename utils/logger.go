package utils

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var do_log = false

var StartTime time.Time

func Log(text string, args ...interface{}) {
	if do_log {
		message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
		fmt.Println(message, args)
	}
}

func LogTime(text string) {
	duration := time.Since(StartTime)
	message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
	fmt.Println(message+" : ", duration)
}
