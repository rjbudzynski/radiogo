package config

import (
	"os"
	"path/filepath"
)

const (
	AppName       = "radiogo"
	MPVSocketPath = "/tmp/radiogo.sock"
	APIBase       = "https://de1.api.radio-browser.info/json"
	APIUserAgent  = "radiogo/1.0 (github.com/woodz-dot/radiogo)"
	DefaultLimit  = 40
)

func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", AppName)
}

func FavoritesPath() string {
	return filepath.Join(ConfigDir(), "favorites.json")
}

// MPVPlaylistPath returns the compat M3U path used by the legacy radiosh script.
func MPVPlaylistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mpv", "radio.m3u")
}

func EnsureDirs() error {
	dirs := []string{
		ConfigDir(),
		filepath.Dir(MPVPlaylistPath()),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}
