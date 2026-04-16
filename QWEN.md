# radiogo - Terminal Radio Player

## Project Overview

**radiogo** is a terminal-based radio player for Linux and macOS that allows users to browse thousands of radio stations from [radio-browser.info](https://radio-browser.info), search by name, save favorites, and control playback entirely from the keyboard.

### Technology Stack

- **Language:** Go 1.26.2+
- **UI Framework:** Bubble Tea (TUI framework by Charmbracelet)
- **Styling:** Lipgloss for terminal styling
- **Player:** mpv (must be installed separately and on `$PATH`)

### Architecture

The project follows a clean layered architecture:

```
cmd/radiogo/          # Entry point (main.go, argument parsing)
internal/
  ├── appstate/       # UI state persistence (volume, selection, active tab)
  ├── config/         # Configuration paths (XDG spec compliance)
  ├── favorites/      # Favorites management (save/load, toggle, M3U export)
  ├── radio/          # Radio API (Browser) and Player (mpv IPC)
  └── ui/             # Bubble Tea model, views, updates, key bindings
```

### Key Components

| Component | Responsibility |
|-----------|----------------|
| `radio.Player` | Manages mpv subprocess via JSON IPC socket; controls playback, volume, pause/resume |
| `radio.Browser` | Fetches stations from radio-browser.info API; supports top stations and name search |
| `favorites.Store` | Saves/loads favorites to `favorites.json`; syncs with legacy `radio.m3u` for mpv compatibility |
| `appstate.Store` | Persists UI state (volume, active tab, search query, selection) between runs |
| `ui.Model` | Bubble Tea model representing the full TUI state (3 tabs: Browse, Favorites, Help) |

## Building and Running

### Requirements

- Go 1.26.2+ (go version in `go.mod`)
- mpv must be installed and on `$PATH`

### Installation

```sh
# Install latest release
go install github.com/rjbudzynski/radiogo/cmd/radiogo@latest

# Build from source
git clone https://github.com/rjbudzynski/radiogo
cd radiogo
go build -o radiogo ./cmd/radiogo
```

### Running

```sh
./radiogo        # Launch the TUI
./radiogo --help # Show usage information
```

### Testing

```sh
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/radio/...
go test ./internal/appstate/...
go test ./internal/favorites/...
go test ./internal/ui/...
```

## Usage

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up in list |
| `↓` / `j` | Move down in list |
| `Tab` | Cycle tabs (Browse → Favorites → Help) |
| `Enter` | Play selected station |
| `p` | Pause / resume playback |
| `+` / `=` | Volume up (+5%) |
| `-` | Volume down (−5%) |
| `/` | Enter search mode (Browse tab only) |
| `Esc` | Cancel search mode |
| `f` | Toggle favorite on selected station |
| `q` / `Ctrl+C` | Quit |

### Tabs

1. **Browse** - Displays top 40 stations from radio-browser.info. Press `/` to search by station name. Falls back to favorites if API is unreachable.

2. **Favorites** - Stations marked with `f` are saved here. Survives restarts.

3. **Help** - Displays keyboard shortcuts information.

## Configuration

The app follows the [XDG Base Directory spec](https://specifications.freedesktop.org/basedir-spec/latest/):

| File | Purpose |
|------|---------|
| `$XDG_CONFIG_HOME/radiogo/favorites.json` | Saved favorites (default: `~/.config/radiogo/favorites.json`) |
| `$XDG_CONFIG_HOME/radiogo/state.json` | Saved UI state (volume, selection, active tab, search query) |
| `~/.config/mpv/radio.m3u` | Compat M3U playlist (written on every favorites save) |

### mpv Socket

Each instance creates a PID-scoped Unix socket at `/tmp/radiogo-<pid>.sock` to allow multiple concurrent instances.

## Development Conventions

### Code Style

- Uses standard Go formatting (`gofmt`)
- Bubble Tea pattern: `Model` with `Init()`, `Update()`, `View()` methods
-Messages passed through Bubble Tea's msg system (`tea.Cmd` for async operations)

### Testing Practices

- Tests located alongside source in `_test.go` files
- Tests cover argument parsing, state persistence, and UI logic
- Uses Go's standard `testing` package

### API Integration

- Radio Browser API base: `https://all.api.radio-browser.info/json`
- User-Agent: `radiogo/1.0 (github.com/rjbudzynski/radiogo)`
- Default limit: 40 stations per request

### Async Pattern

Background operations (API calls, playback events) use Bubble Tea commands that return messages:

```go
tea.Batch(
    m.spinner.Tick,
    loadBrowseStations(m.searchQuery, m.searchGen),
)
```

Callback pattern for player events (metadata updates, pause state, player stop) uses function parameters passed to `Player.Play()`.

### State Management

- **Transient UI state**: Held in `ui.Model`
- **Persistent state**: `appstate.State` saved to `state.json` on quit and significant state changes
- **Favorites**: Saved to `favorites.json` on toggle; also writes `radio.m3u` for mpv compatibility

## Troubleshooting

- If mpv is not installed: `brew install mpv` (macOS) or `sudo apt install mpv` (Linux)
- If API is unreachable, favorites are shown as fallback in the Browse tab
- Check `/tmp/radiogo-*.sock` files are being cleaned up if multiple instances were killed unexpectedly

## Project Structure Details

### `internal/config/config.go`
 Defines app constants, paths, and directory creation logic.

### `internal/radio/player.go`
 Manages mpv subprocess, IPC communication, and property observation (metadata, pause state, volume).

### `internal/radio/browser.go`
 API client for radio-browser.info; provides `TopStations()` and `SearchStations()` functions.

### `internal/favorites/store.go`
 Favorites persistence with atomic write (temp file + rename) pattern and M3U export.

### `internal/appstate/store.go`
 UI state persistence (volume, tab selection, search query, selected station).

### `internal/ui/`
 Bubble Tea implementation with:
 - `model.go` - Data structures and helper functions
 - `update.go` - Message handlers and state transitions
 - `view.go` - TUI rendering
 - `keys.go` - Key bindings
 - `styles.go` - Lipgloss styling
