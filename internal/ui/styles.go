package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors - Vibrant Neon Palette
	colorAccent   = lipgloss.Color("51")  // electric cyan
	colorSubtle   = lipgloss.Color("147") // soft purple
	colorSelected = lipgloss.Color("201") // hot pink
	colorFav      = lipgloss.Color("226") // electric yellow
	colorPlaying  = lipgloss.Color("82")  // lime green
	colorErr      = lipgloss.Color("196") // bright red
	colorMuted    = lipgloss.Color("245") // gray

	// Pane borders
	stylePane = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSubtle).
			Background(lipgloss.Color("236"))

	stylePaneFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Background(lipgloss.Color("236"))

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
			Foreground(lipgloss.Color("231")).
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
			Foreground(colorFav).
			Bold(true)

	// Info pane
	styleInfoLabel = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Width(10)

	styleInfoValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("231"))

	styleNowPlaying = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPlaying).
			Underline(true)

	// Status bar - deep purple with neon accents
	styleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("55")).
			Foreground(lipgloss.Color("231")).
			Padding(0, 1)

	styleStatusPlaying = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPlaying)

	styleStatusTrack = lipgloss.NewStyle().
				Foreground(colorAccent)

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
				Foreground(colorSelected).
				MarginBottom(1)

	styleDivider = lipgloss.NewStyle().
			Foreground(colorSubtle)
)
