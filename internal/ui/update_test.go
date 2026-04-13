package ui

import "testing"

func TestUpdate_PauseStateMsgSyncsPausedFlag(t *testing.T) {
	m := baseModel()

	updated, cmd := m.Update(pauseStateMsg{paused: true})
	if cmd != nil {
		t.Fatal("expected no command for pause state update")
	}

	got := updated.(Model)
	if !got.paused {
		t.Fatal("paused = false, want true")
	}

	updated, _ = got.Update(pauseStateMsg{paused: false})
	got = updated.(Model)
	if got.paused {
		t.Fatal("paused = true, want false")
	}
}
