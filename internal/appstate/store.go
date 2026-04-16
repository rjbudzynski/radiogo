package appstate

import (
	"encoding/json"
	"os"

	"github.com/rjbudzynski/radiogo/internal/config"
)

type StationRef struct {
	UUID string `json:"uuid,omitempty"`
	URL  string `json:"url,omitempty"`
}

func (r StationRef) IsZero() bool {
	return r.UUID == "" && r.URL == ""
}

type State struct {
	Volume          int        `json:"volume"`
	ActiveTab       int        `json:"active_tab,omitempty"`
	SearchQuery     string     `json:"search_query,omitempty"`
	BrowseSort      string     `json:"browse_sort,omitempty"`
	SelectedStation StationRef `json:"selected_station,omitempty"`
}

// Load reads the saved UI state. It returns nil on first run.
func Load() (*State, error) {
	data, err := os.ReadFile(config.StatePath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// Save writes the saved UI state to disk atomically.
func Save(state State) error {
	if err := config.EnsureDirs(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	dest := config.StatePath()
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
