package appstate_test

import (
	"path/filepath"
	"testing"

	"github.com/rjbudzynski/radiogo/internal/appstate"
	"github.com/rjbudzynski/radiogo/internal/config"
)

func TestLoad_FirstRunReturnsNil(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	state, err := appstate.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state != nil {
		t.Fatalf("Load() = %#v, want nil on first run", state)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	want := appstate.State{
		Volume:      35,
		ActiveTab:   1,
		SearchQuery: "jazz",
		SelectedStation: appstate.StationRef{
			UUID: "station-uuid",
			URL:  "https://example.com/stream",
		},
	}

	if err := appstate.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := appstate.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got == nil {
		t.Fatal("Load() = nil, want saved state")
	}

	if *got != want {
		t.Fatalf("Load() = %#v, want %#v", *got, want)
	}

	if path := config.StatePath(); filepath.Base(path) != "state.json" {
		t.Fatalf("StatePath() = %q, want state.json suffix", path)
	}
}
