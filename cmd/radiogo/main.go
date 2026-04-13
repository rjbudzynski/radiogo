package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rjbudzynski/radiogo/internal/appstate"
	"github.com/rjbudzynski/radiogo/internal/config"
	"github.com/rjbudzynski/radiogo/internal/favorites"
	"github.com/rjbudzynski/radiogo/internal/ui"
)

func main() {
	showHelp, err := parseArgs(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		os.Exit(2)
	}
	if showHelp {
		return
	}

	if err := config.EnsureDirs(); err != nil {
		fmt.Fprintf(os.Stderr, "radiogo: config error: %v\n", err)
		os.Exit(1)
	}

	favs, err := favorites.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "radiogo: could not load favorites: %v\n", err)
	}

	state, err := appstate.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "radiogo: could not load saved state: %v\n", err)
	}

	model := ui.New(favs, state)

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

func parseArgs(args []string, stdout, stderr io.Writer) (bool, error) {
	fs := flag.NewFlagSet("radiogo", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	help := fs.Bool("help", false, "show help")
	fs.BoolVar(help, "h", false, "show help")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			printUsage(stdout)
			return true, nil
		}
		fmt.Fprintf(stderr, "radiogo: %v\n\n", err)
		printUsage(stderr)
		return false, err
	}

	if *help {
		printUsage(stdout)
		return true, nil
	}

	return false, nil
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, usageText())
}

func usageText() string {
	name := filepath.Base(os.Args[0])
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "radiogo"
	}

	return fmt.Sprintf(`%s - terminal radio player

Usage:
  %s [--help]

Options:
  -h, --help   Show this help message
`, name, name)
}
