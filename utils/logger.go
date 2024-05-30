package utils

import (
	"log"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var do_log = true

var (
	StartTime time.Time
	LastTime  time.Time
)

func SetupLog() {
	LOG_FILE := "/tmp/nvim-gui.log"
	// open log file
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Panic(err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func Log(text string, args ...interface{}) {
	if do_log {
		message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
		log.Println(message, args)
	}
}

func LogTime(text string) {
	if do_log {
		duration := time.Since(StartTime)
		LastTime = time.Now()
		message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
		log.Println(message+" took ", duration)
	}
}

func LogTimeSinceLast(text string) {
	if do_log {
		duration := time.Since(LastTime)
		message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
		log.Printf("\noperation '%s' took %v", message, duration)
		LastTime = time.Now()
	}
}
