package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName      = "radiogo"
	APIBase      = "https://all.api.radio-browser.info/json"
	APIUserAgent = "radiogo/1.0 (github.com/rjbudzynski/radiogo)"
	DefaultLimit = 40
)

// MPVSocketPath returns a PID-scoped socket path on Unix or a named pipe on Windows.
func MPVSocketPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\radiogo-mpv`
	}
	return fmt.Sprintf("/tmp/radiogo-%d.sock", os.Getpid())
}

func ConfigDir() string {
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, AppName)
		}
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", AppName)
}

func FavoritesPath() string {
	return filepath.Join(ConfigDir(), "favorites.json")
}

func StatePath() string {
	return filepath.Join(ConfigDir(), "state.json")
}

// MPVPlaylistPath returns the compat M3U path used by the legacy radiosh script.
func MPVPlaylistPath() string {
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "mpv", "radio.m3u")
		}
	}
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
