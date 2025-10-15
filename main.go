package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nkzk/xtree/internal/ui"
)

func main() {
	if _, err := tea.NewProgram(ui.NewModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("failed to start: %v", err)
		os.Exit(1)
	}
}
