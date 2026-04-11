package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rjbudzynski/radiogo/internal/config"
	"github.com/rjbudzynski/radiogo/internal/favorites"
	"github.com/rjbudzynski/radiogo/internal/ui"
)

func main() {
	if err := config.EnsureDirs(); err != nil {
		fmt.Fprintf(os.Stderr, "radiogo: config error: %v\n", err)
		os.Exit(1)
	}

	favs, err := favorites.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "radiogo: could not load favorites: %v\n", err)
	}

	model := ui.New(favs)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	ui.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "radiogo: %v\n", err)
		os.Exit(1)
	}
}
