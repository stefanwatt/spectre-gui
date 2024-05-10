package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func Log(text string) {
	message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
	fmt.Println(message)
}
