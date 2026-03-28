package utils

import (
	"os"
	"strings"
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

	level := log.InfoLevel
	switch strings.ToLower(strings.TrimSpace(os.Getenv("NVIM_GUI_LOG_LEVEL"))) {
	case "debug":
		level = log.DebugLevel
	case "warn", "warning":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	case "fatal":
		level = log.FatalLevel
	case "info", "":
		level = log.InfoLevel
	}

	Logger = log.NewWithOptions(f, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "nvim-gui",
		Level:           level,
	})
	log.SetDefault(Logger)
	log.Info("logger initialized", "level", level.String(), "path", "/tmp/nvim-gui.log")
}
