package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rjbudzynski/radiogo/internal/appstate"
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
type pauseStateMsg struct{ paused bool }
type playerStoppedMsg struct{}
type favSaveErrMsg struct{ err error }
type stateSavedMsg struct{}
type stateSaveErrMsg struct{ err error }

// Model is the root Bubble Tea model.
type Model struct {
	// Layout
	width  int
	height int

	// Tabs / navigation
	activeTab   int
	browseIndex int
	favIndex    int
	searching   bool
	searchInput textinput.Model
	searchQuery string

	// Data
	browseStations    []radio.Station
	favorites         []radio.Station
	searchGen         int // incremented on each search/load; stale responses are discarded
	restoredSelection appstate.StationRef

	// Playback
	player     *radio.Player
	nowPlaying *radio.Station
	trackTitle string
	volume     int
	paused     bool

	// UI state
	loading   bool
	browseErr error // non-fatal: shown as inline banner
	saveErr   error // non-fatal: shown as inline banner
	stateErr  error // non-fatal: shown as inline banner
	spinner   spinner.Model
}

// New constructs a fresh Model.
func New(favs []radio.Station, restored *appstate.State) Model {
	si := textinput.New()
	si.Placeholder = "station name…"
	si.CharLimit = 80

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorAccent)

	m := Model{
		activeTab:   tabBrowse,
		favorites:   favs,
		player:      &radio.Player{},
		volume:      80,
		searchInput: si,
		spinner:     sp,
		loading:     true,
		searchGen:   1, // matches the gen passed by Init
	}

	if restored == nil {
		return m
	}

	m.volume = clamp(restored.Volume, 0, 100)
	m.activeTab = normalizeTab(restored.ActiveTab)
	m.searchQuery = restored.SearchQuery
	m.restoredSelection = restored.SelectedStation

	if idx := stationIndex(favs, restored.SelectedStation); idx >= 0 {
		m.favIndex = idx
	}

	return m
}

// Init starts the spinner and loads the initial browse list.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadBrowseStations(m.searchQuery, m.searchGen),
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

func loadBrowseStations(query string, gen int) tea.Cmd {
	if query != "" {
		return searchStations(query, gen)
	}
	return loadTopStations(gen)
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

func (m *Model) stationForTab(tab int) *radio.Station {
	switch tab {
	case tabFavorites:
		return stationAt(m.favorites, m.favIndex)
	case tabBrowse:
		if len(m.browseStations) == 0 {
			return stationAt(m.favorites, m.browseIndex)
		}
		return stationAt(m.browseStations, m.browseIndex)
	default:
		if s := stationAt(m.browseStations, m.browseIndex); s != nil {
			return s
		}
		return stationAt(m.favorites, m.favIndex)
	}
}

func (m *Model) selectedStationRef() appstate.StationRef {
	s := m.stationForTab(m.activeTab)
	if s == nil {
		return appstate.StationRef{}
	}
	return appstate.StationRef{
		UUID: s.UUID,
		URL:  s.URL,
	}
}

func (m *Model) stateSnapshot() appstate.State {
	return appstate.State{
		Volume:          m.volume,
		ActiveTab:       m.activeTab,
		SearchQuery:     m.searchQuery,
		SelectedStation: m.selectedStationRef(),
	}
}

func (m *Model) persistStateCmd() tea.Cmd {
	state := m.stateSnapshot()
	return func() tea.Msg {
		if err := appstate.Save(state); err != nil {
			return stateSaveErrMsg{err: err}
		}
		return stateSavedMsg{}
	}
}

func (m *Model) restoreBrowseSelection() {
	if ref := m.restoredSelection; !ref.IsZero() {
		if idx := stationIndex(m.browseStations, ref); idx >= 0 {
			m.browseIndex = idx
		}
	}
	m.restoredSelection = appstate.StationRef{}
	m.browseIndex = clampIndex(m.browseIndex, len(m.browseStations))
}

func (m *Model) clampSelection() {
	m.browseIndex = clampIndex(m.browseIndex, len(m.activeBrowseList()))
	m.favIndex = clampIndex(m.favIndex, len(m.favorites))
}

func (m *Model) activeBrowseList() []radio.Station {
	if len(m.browseStations) == 0 {
		return m.favorites
	}
	return m.browseStations
}

func stationAt(list []radio.Station, idx int) *radio.Station {
	if idx < 0 || idx >= len(list) {
		return nil
	}
	s := list[idx]
	return &s
}

func stationIndex(list []radio.Station, ref appstate.StationRef) int {
	for i, station := range list {
		if ref.UUID != "" && station.UUID == ref.UUID {
			return i
		}
		if ref.URL != "" && station.URL == ref.URL {
			return i
		}
	}
	return -1
}

func clampIndex(idx, n int) int {
	if n <= 0 {
		return 0
	}
	if idx < 0 {
		return 0
	}
	if idx >= n {
		return n - 1
	}
	return idx
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func normalizeTab(tab int) int {
	if tab < tabBrowse || tab > tabHelp {
		return tabBrowse
	}
	return tab
}
