package favorites

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/rjbudzynski/radiogo/internal/config"
	"github.com/rjbudzynski/radiogo/internal/radio"
)

// Load reads saved favorites from disk. Returns empty slice on first run.
func Load() ([]radio.Station, error) {
	path := config.FavoritesPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var stations []radio.Station
	return stations, json.Unmarshal(data, &stations)
}

// Save writes the favorites list to disk atomically and syncs the compat M3U file.
func Save(stations []radio.Station) error {
	if err := config.EnsureDirs(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(stations, "", "  ")
	if err != nil {
		return err
	}

	// Write to a temp file in the same directory, then rename for atomicity.
	dest := config.FavoritesPath()
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return writeM3U(stations)
}

// Toggle adds the station if absent, removes it if present. Returns updated list.
func Toggle(stations []radio.Station, s radio.Station) []radio.Station {
	for i, f := range stations {
		if f.UUID == s.UUID || (f.UUID == "" && f.URL == s.URL) {
			return slices.Delete(stations, i, i+1)
		}
	}
	return append(stations, s)
}

// Contains reports whether s is in the favorites list.
func Contains(stations []radio.Station, s radio.Station) bool {
	for _, f := range stations {
		if f.UUID == s.UUID || (f.UUID == "" && f.URL == s.URL) {
			return true
		}
	}
	return false
}

// writeM3U writes a plain M3U file compatible with the legacy radiosh script,
// using an atomic tmp+rename to avoid partial writes.
func writeM3U(stations []radio.Station) error {
	dest := config.MPVPlaylistPath()
	tmp := dest + ".tmp"

	// Explicitly opened and closed so we can detect write errors before the rename.
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	for _, s := range stations {
		fmt.Fprintf(f, "%s | %s\n", s.Name, s.URL)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
