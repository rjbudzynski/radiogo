package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/woodz-dot/radiogo/internal/config"
	"github.com/woodz-dot/radiogo/internal/radio"
)

const (
	tabBrowse    = 0
	tabFavorites = 1
	tabHelp      = 2
)

// Async messages sent by background goroutines into the Bubble Tea loop.

type stationsLoadedMsg struct{ stations []radio.Station }
type stationsErrMsg struct{ err error }
type metaUpdateMsg struct{ title string }
type playerStoppedMsg struct{}

// Model is the root Bubble Tea model.
type Model struct {
	// Layout
	width  int
	height int

	// Tabs / navigation
	activeTab    int
	browseIndex  int
	favIndex     int
	searching    bool
	searchInput  textinput.Model
	searchQuery  string

	// Data
	browseStations []radio.Station
	favorites      []radio.Station

	// Playback
	player      *radio.Player
	nowPlaying  *radio.Station
	trackTitle  string
	volume      int
	paused      bool

	// UI state
	loading   bool
	browseErr error // non-fatal: shown as inline banner
	spinner   spinner.Model
}

// New constructs a fresh Model.
func New(favs []radio.Station) Model {
	si := textinput.New()
	si.Placeholder = "station name…"
	si.CharLimit = 80

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorAccent)

	return Model{
		activeTab:    tabBrowse,
		favorites:    favs,
		player:       &radio.Player{},
		volume:       80,
		searchInput:  si,
		spinner:      sp,
		loading:      true,
	}
}

// Init starts the spinner and loads the default top-stations list.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadTopStations(),
	)
}

// loadTopStations is a Bubble Tea command that fetches top stations in the background.
func loadTopStations() tea.Cmd {
	return func() tea.Msg {
		stations, err := radio.TopStations(config.DefaultLimit)
		if err != nil {
			return stationsErrMsg{err}
		}
		return stationsLoadedMsg{stations}
	}
}

// searchStations is a Bubble Tea command for a name search.
func searchStations(query string) tea.Cmd {
	return func() tea.Msg {
		stations, err := radio.SearchStations(query, config.DefaultLimit)
		if err != nil {
			return stationsErrMsg{err}
		}
		return stationsLoadedMsg{stations}
	}
}

// activeList returns the station list for the current tab.
// For the Help tab (or Browse fallback when empty), returns favorites.
func (m *Model) activeList() []radio.Station {
	switch m.activeTab {
	case tabFavorites:
		return m.favorites
	case tabHelp:
		return nil
	default: // tabBrowse
		if len(m.browseStations) == 0 {
			return m.favorites // fallback when directory unreachable
		}
		return m.browseStations
	}
}

// browseIsFallback reports whether the Browse tab is showing favorites as a fallback.
func (m *Model) browseIsFallback() bool {
	return m.activeTab == tabBrowse && len(m.browseStations) == 0 && len(m.favorites) > 0
}

// activeIndex returns/sets the cursor for the current tab.
func (m *Model) activeIndex() int {
	if m.activeTab == tabFavorites {
		return m.favIndex
	}
	return m.browseIndex
}

func (m *Model) setActiveIndex(i int) {
	if m.activeTab == tabFavorites {
		m.favIndex = i
	} else {
		m.browseIndex = i
	}
}

func (m *Model) selectedStation() *radio.Station {
	list := m.activeList()
	idx := m.activeIndex()
	if idx < 0 || idx >= len(list) {
		return nil
	}
	s := list[idx]
	return &s
}
