# radiogo — Copilot Instructions

## What this is

A terminal radio player written in Go. It streams audio via an `mpv` subprocess and discovers stations through the [Radio Browser API](https://de1.api.radio-browser.info). The UI is a [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI with three tabs: Browse, Favorites, Help.

## Build & run

```sh
go build ./cmd/radiogo      # build binary
go run   ./cmd/radiogo      # run directly
```

No test files exist yet. There is no linter config — `//nolint:errcheck` is used inline where fire-and-forget errors are intentional.

## Architecture

```
cmd/radiogo/main.go          entry point — wires config, favorites, ui.Model, tea.Program
internal/config/config.go    constants & XDG paths (config dir, socket, API base URL)
internal/radio/browser.go    Radio Browser API client; defines Station struct
internal/radio/player.go     manages a single mpv subprocess via JSON IPC socket
internal/favorites/store.go  load/save favorites (JSON) + writes compat M3U for mpv
internal/ui/
  model.go                   root Bubble Tea Model and constructor
  update.go                  Update() + handleKey(); all state transitions live here
  view.go                    View() — renders tabs, list, info pane, status bar
  styles.go                  all lipgloss style variables
  keys.go                    key.Binding definitions
```

### Bubble Tea message flow

Background work (API calls, mpv events) communicates with the Bubble Tea loop via message types defined at the top of `model.go`:

| Message type        | Source                              |
|---------------------|-------------------------------------|
| `stationsLoadedMsg` | `loadTopStations` / `searchStations` commands |
| `stationsErrMsg`    | same commands on error              |
| `metaUpdateMsg`     | mpv icy-title property observer     |
| `playerStoppedMsg`  | mpv process exit                    |

The `Player.Play` callback goroutines can't call `tea.Cmd` directly; they use `currentProgram.Send(...)` (set via `ui.SetProgram`).

### Player

`radio.Player` wraps a single `mpv --input-ipc-server` subprocess. All methods are goroutine-safe via `sync.Mutex`. Control (pause, volume, stop) is sent as JSON over the Unix socket at `config.MPVSocketPath` (`/tmp/radiogo.sock`). The icy-title metadata is read by observing the `metadata/by-key/icy-title` property.

## Key conventions

- **Station identity**: equality is checked as `UUID == UUID`, with a fallback to `URL == URL` when UUID is empty (see `favorites.Toggle` / `favorites.Contains`).
- **Async saves**: `favorites.Save` is called with `go favorites.Save(...)` and `//nolint:errcheck` — save errors are silently dropped.
- **Browse fallback**: when `browseStations` is empty (API unreachable), `activeList()` returns `favorites` so the Browse tab still shows something useful.
- **Styles**: all colors use lipgloss 256-color codes (e.g., `lipgloss.Color("86")`), not hex/RGB. Color constants are defined in `styles.go`.
- **Key bindings**: all bindings live in the `keys` package-level var in `keys.go`; use `isKey(msg, keys.X)` to match them in `update.go`.
- **Config paths**: always go through `config.*Path()` helpers, which respect `$XDG_CONFIG_HOME`.
