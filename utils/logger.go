package utils

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

var (
	StartTime time.Time
	LastTime  time.Time
)

func SetupLog() {
	f, err := os.OpenFile("/tmp/nvim-gui.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	Logger = log.NewWithOptions(f, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "nvim-gui",
	})
	log.SetDefault(Logger)
}
