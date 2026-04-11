package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rjbudzynski/radiogo/internal/config"
	"github.com/rjbudzynski/radiogo/internal/radio"
)

const (
	tabBrowse    = 0
	tabFavorites = 1
	tabHelp      = 2
)

// Async messages sent by background goroutines into the Bubble Tea loop.

type stationsLoadedMsg struct {
	stations []radio.Station
	gen      int
}
type stationsErrMsg struct{ err error }
type metaUpdateMsg struct{ title string }
type playerStoppedMsg struct{}
type favSaveErrMsg struct{ err error }

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
	searchGen      int // incremented on each search/load; stale responses are discarded

	// Playback
	player      *radio.Player
	nowPlaying  *radio.Station
	trackTitle  string
	volume      int
	paused      bool

	// UI state
	loading   bool
	browseErr error // non-fatal: shown as inline banner
	saveErr   error // non-fatal: shown as inline banner
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
		searchGen:    1, // matches the gen passed by Init
	}
}

// Init starts the spinner and loads the default top-stations list.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadTopStations(m.searchGen),
	)
}

// loadTopStations is a Bubble Tea command that fetches top stations in the background.
func loadTopStations(gen int) tea.Cmd {
	return func() tea.Msg {
		stations, err := radio.TopStations(config.DefaultLimit)
		if err != nil {
			return stationsErrMsg{err}
		}
		return stationsLoadedMsg{stations, gen}
	}
}

// searchStations is a Bubble Tea command for a name search.
func searchStations(query string, gen int) tea.Cmd {
	return func() tea.Msg {
		stations, err := radio.SearchStations(query, config.DefaultLimit)
		if err != nil {
			return stationsErrMsg{err}
		}
		return stationsLoadedMsg{stations, gen}
	}
}

// listHeaderLines returns the number of lines rendered above the station rows
// inside the list pane (error banner + search/context line).
func (m *Model) listHeaderLines() int {
	n := 0
	if m.browseErr != nil && m.activeTab == tabBrowse {
		n++ // error line
		if m.browseIsFallback() {
			n++ // "directory unavailable — showing favorites" notice
		}
	}
	if m.activeTab == tabBrowse {
		n++ // search/context line (always present on Browse tab)
	}
	return n
}

// listHitTest maps a terminal Y coordinate to a station index in activeList,
// or returns -1 if the coordinate is outside the visible station rows.
// It mirrors the scroll-window calculation in renderListPane.
func (m *Model) listHitTest(y int) int {
	if m.activeTab == tabHelp || m.loading {
		return -1
	}
	list := m.activeList()
	if len(list) == 0 {
		return -1
	}

	headerLines := m.listHeaderLines()
	// innerH mirrors renderListPane: height arg is (m.height - statusBarHeight - tabBarHeight - 2)
	innerH := (m.height - 5) - headerLines - 2
	if innerH < 1 {
		return -1
	}

	// Reproduce the scroll-window start offset used at render time.
	idx := m.activeIndex()
	start := 0
	if idx > innerH-1 {
		start = idx - innerH + 1
	}

	// List rows start at: tabBar(1) + top pane border(1) + headerLines
	listStartY := 2 + headerLines
	row := y - listStartY
	if row < 0 || row >= innerH {
		return -1
	}
	station := start + row
	if station < 0 || station >= len(list) {
		return -1
	}
	return station
}


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
