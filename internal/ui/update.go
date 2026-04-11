package ui

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rjbudzynski/radiogo/internal/favorites"
	"github.com/rjbudzynski/radiogo/internal/radio"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case stationsLoadedMsg:
		if msg.gen != m.searchGen {
			return m, nil // discard stale response
		}
		m.loading = false
		m.browseErr = nil
		m.browseStations = msg.stations
		if m.browseIndex >= len(m.browseStations) {
			m.browseIndex = 0
		}
		return m, nil

	case stationsErrMsg:
		m.loading = false
		m.browseErr = msg.err
		// browseStations stays empty; activeList() will fall back to favorites
		return m, nil

	case metaUpdateMsg:
		m.trackTitle = msg.title
		return m, nil

	case playerStoppedMsg:
		m.nowPlaying = nil
		m.trackTitle = ""
		m.paused = false
		return m, nil

	case favSaveErrMsg:
		m.saveErr = msg.err
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Propagate spinner ticks.
	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Search input mode — route most keys to the textinput.
	if m.searching {
		switch {
		case isKey(msg, keys.Escape):
			m.searching = false
			m.searchInput.Blur()
			return m, nil

		case isKey(msg, keys.Enter):
			m.searching = false
			m.searchInput.Blur()
			q := m.searchInput.Value()
			if q == "" {
				return m, nil
			}
			m.searchQuery = q
			m.loading = true
			m.browseIndex = 0
			m.searchGen++
			return m, tea.Batch(m.spinner.Tick, searchStations(q, m.searchGen))

		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	switch {
	case isKey(msg, keys.Quit):
		m.player.Stop()
		return m, tea.Quit

	case isKey(msg, keys.Tab):
		m.activeTab = (m.activeTab + 1) % 3
		return m, nil

	case isKey(msg, keys.Up):
		if m.activeTab == tabHelp {
			return m, nil
		}
		idx := m.activeIndex() - 1
		if idx < 0 {
			idx = 0
		}
		m.setActiveIndex(idx)
		return m, nil

	case isKey(msg, keys.Down):
		if m.activeTab == tabHelp {
			return m, nil
		}
		list := m.activeList()
		idx := m.activeIndex() + 1
		if idx >= len(list) {
			idx = len(list) - 1
		}
		m.setActiveIndex(idx)
		return m, nil

	case isKey(msg, keys.Enter):
		if m.activeTab == tabHelp {
			return m, nil
		}
		return m, m.playSelected()

	case isKey(msg, keys.Fav):
		if m.activeTab == tabHelp {
			return m, nil
		}
		if s := m.selectedStation(); s != nil {
			m.favorites = favorites.Toggle(m.favorites, *s)
			favs := m.favorites
			return m, func() tea.Msg {
				if err := favorites.Save(favs); err != nil {
					return favSaveErrMsg{err}
				}
				return nil
			}
		}
		return m, nil

	case isKey(msg, keys.Search):
		if m.activeTab == tabBrowse {
			m.searching = true
			m.searchInput.SetValue("")
			m.searchInput.Focus()
		}
		return m, nil

	case isKey(msg, keys.Pause):
		if m.player.IsRunning() {
			m.player.Pause()
			m.paused = !m.paused
		}
		return m, nil

	case isKey(msg, keys.VolUp):
		if m.volume < 100 {
			m.volume += 5
			m.player.SetVolume(m.volume)
		}
		return m, nil

	case isKey(msg, keys.VolDown):
		if m.volume > 0 {
			m.volume -= 5
			m.player.SetVolume(m.volume)
		}
		return m, nil
	}

	return m, nil
}

// playSelected starts playback of the currently highlighted station.
func (m *Model) playSelected() tea.Cmd {
	s := m.selectedStation()
	if s == nil {
		return nil
	}
	station := *s

	prog := currentProgram // set in main.go via SetProgram
	m.nowPlaying = &station
	m.trackTitle = ""
	m.paused = false

	err := m.player.Play(
		station,
		func(mm radio.MetaMsg) {
			if prog != nil {
				prog.Send(metaUpdateMsg{title: mm.Title})
			}
		},
		func(_ radio.PlayerStoppedMsg) {
			if prog != nil {
				prog.Send(playerStoppedMsg{})
			}
		},
	)
	if err != nil {
		if isExecNotFound(err) {
			m.browseErr = fmt.Errorf("mpv is required for playback but was not found in PATH")
		} else {
			m.browseErr = err
		}
		m.nowPlaying = nil
		return nil
	}
	m.player.SetVolume(m.volume)
	return nil
}

// isKey is a small helper to check a key binding.
func isKey(msg tea.KeyMsg, b key.Binding) bool {
	return key.Matches(msg, b)
}

// isExecNotFound reports whether err is an "executable not found" error from exec.Command.
func isExecNotFound(err error) bool {
	var execErr *exec.Error
	return errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound
}

// currentProgram is set by main so callbacks can send messages.
var currentProgram *tea.Program

// SetProgram injects the running tea.Program reference for async callbacks.
func SetProgram(p *tea.Program) { currentProgram = p }
