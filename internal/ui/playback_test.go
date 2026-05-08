package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rjbudzynski/radiogo/internal/radio"
)

// playerSpy records playback commands so tests can verify what the model sends.
type playerSpy struct {
	playStation    *radio.Station
	playVolume     int
	setVolumeCalls []int
	pauseCalled    bool
	stopCalled     bool
	running        bool
}

func (s *playerSpy) Play(station radio.Station, volume int, onMeta radio.MetaCallback, onPause radio.PauseCallback, onStop radio.StopCallback) error {
	s.playStation = &station
	s.playVolume = volume
	s.running = true
	return nil
}

func (s *playerSpy) Pause()                               { s.pauseCalled = true }
func (s *playerSpy) SetVolume(vol int)                     { s.setVolumeCalls = append(s.setVolumeCalls, vol) }
func (s *playerSpy) Stop()                                 { s.stopCalled = true }
func (s *playerSpy) IsRunning() bool                       { return s.running }

// spyModel creates a Model with a playerSpy and a single station selected in Browse.
func spyModel() (Model, *playerSpy) {
	m := baseModel()
	spy := &playerSpy{}
	m.player = spy
	m.browseStations = stationsOf("test-station")
	m.browseIndex = 0
	return m, spy
}

func TestPlaySelected_PassesVolumeToPlay(t *testing.T) {
	m, spy := spyModel()
	m.volume = 42

	m.playSelected()

	if spy.playStation == nil {
		t.Fatal("Play was not called")
	}
	if spy.playStation.Name != "test-station" {
		t.Fatalf("Play station = %q, want test-station", spy.playStation.Name)
	}
	if spy.playVolume != 42 {
		t.Fatalf("Play volume = %d, want 42", spy.playVolume)
	}
}

func TestPlaySelected_DefaultsToModelVolume(t *testing.T) {
	m, spy := spyModel()
	m.volume = 80 // default

	m.playSelected()

	if spy.playVolume != 80 {
		t.Fatalf("Play volume = %d, want 80 (model default)", spy.playVolume)
	}
}

func TestPlaySelected_UpdatesNowPlaying(t *testing.T) {
	m, spy := spyModel()
	m.volume = 50

	m.playSelected()
	_ = spy // unused but keeps the spy in scope

	if m.nowPlaying == nil {
		t.Fatal("nowPlaying is nil after playSelected")
	}
	if m.nowPlaying.Name != "test-station" {
		t.Fatalf("nowPlaying = %q, want test-station", m.nowPlaying.Name)
	}
	if m.trackTitle != "" {
		t.Fatalf("trackTitle = %q, want empty after station switch", m.trackTitle)
	}
	if m.paused {
		t.Fatal("paused should be false after playSelected")
	}
}

// updatedModel is a helper that calls m.handleKey and returns the resulting Model.
func updatedModel(m Model, msg tea.KeyMsg) Model {
	updated, _ := m.handleKey(msg)
	return updated.(Model)
}

func TestVolUp_IncrementsVolumeAndCallsSetVolume(t *testing.T) {
	m, spy := spyModel()
	m.volume = 75

	got := updatedModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})

	if got.volume != 80 {
		t.Fatalf("model volume = %d, want 80", got.volume)
	}
	if len(spy.setVolumeCalls) != 1 {
		t.Fatalf("SetVolume called %d times, want 1", len(spy.setVolumeCalls))
	}
	if spy.setVolumeCalls[0] != 80 {
		t.Fatalf("SetVolume(80) not sent: got %d", spy.setVolumeCalls[0])
	}
}

func TestVolUp_CapsAt100(t *testing.T) {
	m, spy := spyModel()
	m.volume = 98

	got := updatedModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if got.volume != 100 {
		t.Fatalf("model volume = %d, want 100", got.volume)
	}

	// + again at 99 — should land on 100
	got = updatedModel(got, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if got.volume != 100 {
		t.Fatalf("model volume = %d, want 100 (cap)", got.volume)
	}

	// + at 100 — no-op
	got = updatedModel(got, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if got.volume != 100 {
		t.Fatalf("model volume = %d, want 100 (still capped)", got.volume)
	}
	if len(spy.setVolumeCalls) != 1 {
		t.Fatalf("SetVolume called %d times, want 1 (only the 98→100 step)", len(spy.setVolumeCalls))
	}
	if spy.setVolumeCalls[0] != 100 {
		t.Fatalf("SetVolume(100) not sent: got %d", spy.setVolumeCalls[0])
	}
}

func TestVolDown_DecrementsVolumeAndCallsSetVolume(t *testing.T) {
	m, spy := spyModel()
	m.volume = 75

	got := updatedModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})

	if got.volume != 70 {
		t.Fatalf("model volume = %d, want 70", got.volume)
	}
	if len(spy.setVolumeCalls) != 1 {
		t.Fatalf("SetVolume called %d times, want 1", len(spy.setVolumeCalls))
	}
	if spy.setVolumeCalls[0] != 70 {
		t.Fatalf("SetVolume(70) not sent: got %d", spy.setVolumeCalls[0])
	}
}

func TestVolDown_FloorsAt0(t *testing.T) {
	m, spy := spyModel()
	m.volume = 3

	got := updatedModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if got.volume != 0 {
		t.Fatalf("model volume = %d, want 0", got.volume)
	}

	// - at 0 — no-op
	got = updatedModel(got, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if got.volume != 0 {
		t.Fatalf("model volume = %d, want 0 (still floored)", got.volume)
	}
	if len(spy.setVolumeCalls) != 1 {
		t.Fatalf("SetVolume called %d times, want 1 (only for actual decrements)", len(spy.setVolumeCalls))
	}
}

func TestVolUp_NoCallWhenNotPlaying(t *testing.T) {
	m, spy := spyModel()
	m.volume = 50
	m.player = spy // no station playing, but player is ready

	got := updatedModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})

	// VolUp works regardless of playback state (the key handler doesn't check IsRunning)
	if got.volume != 55 {
		t.Fatalf("model volume = %d, want 55", got.volume)
	}
	if len(spy.setVolumeCalls) != 1 {
		t.Fatalf("SetVolume called %d times, want 1", len(spy.setVolumeCalls))
	}
}

func TestPause_CallsPlayerPause(t *testing.T) {
	m, spy := spyModel()
	spy.running = true

	updatedModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})

	if !spy.pauseCalled {
		t.Fatal("Pause() was not called")
	}
}

func TestPause_NoCallWhenNotRunning(t *testing.T) {
	m, spy := spyModel()
	spy.running = false

	updatedModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})

	if spy.pauseCalled {
		t.Fatal("Pause() was called even though player is not running")
	}
}
