package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorAccent   = lipgloss.Color("86")  // cyan-green
	colorSubtle   = lipgloss.Color("240") // dark gray
	colorSelected = lipgloss.Color("212") // pink
	colorFav      = lipgloss.Color("220") // yellow
	colorPlaying  = lipgloss.Color("86")
	colorErr      = lipgloss.Color("196")
	colorMuted    = lipgloss.Color("245")

	// Pane borders
	stylePane = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSubtle)

	stylePaneFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent)

	// Tab bar
	styleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(colorSubtle).
				Padding(0, 1)

	// Station list items
	styleItemNormal = lipgloss.NewStyle().
			Padding(0, 1)

	styleItemSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSelected).
				Padding(0, 1)

	styleItemPlaying = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPlaying).
				Padding(0, 1)

	styleItemFav = lipgloss.NewStyle().
			Foreground(colorFav)

	// Info pane
	styleInfoLabel = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Width(10)

	styleInfoValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	styleNowPlaying = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPlaying)

	// Status bar
	styleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	styleStatusPlaying = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPlaying)

	styleStatusTrack = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleErr = lipgloss.NewStyle().
			Foreground(colorErr).
			Bold(true)

	styleSearchPrompt = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleSectionTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				MarginBottom(1)

	styleDivider = lipgloss.NewStyle().
			Foreground(colorSubtle)
)
