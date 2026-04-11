# radiogo

A terminal radio player for Linux and macOS. Browse thousands of stations from [radio-browser.info](https://radio-browser.info), search by name, save favorites, and control playback — all from the keyboard.

## Requirements

- **Go 1.22+** (to build)
- **[mpv](https://mpv.io/)** (runtime — must be on `$PATH`)

## Install

```sh
go install github.com/rjbudzynski/radiogo/cmd/radiogo@latest
```

Or build from source:

```sh
git clone https://github.com/rjbudzynski/radiogo
cd radiogo
go build -o radiogo ./cmd/radiogo
```

## Usage

```sh
radiogo
```

The app opens in your terminal with three tabs — **Browse**, **Favorites**, and **Help** — navigated by `Tab`.

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Tab` | Cycle tabs (Browse → Favorites → Help) |
| `Enter` | Play selected station |
| `p` | Pause / resume |
| `+` / `=` | Volume up (+5%) |
| `-` | Volume down (−5%) |
| `/` | Search by name *(Browse tab)* |
| `Esc` | Cancel search |
| `f` | Toggle favorite on selected station |
| `q` / `Ctrl+C` | Quit |

### Browse tab

Opens with the 40 top-voted stations from the Radio Browser directory. Press `/` to search by name. If the directory is unreachable, your saved favorites are shown as a fallback.

### Favorites tab

Stations you've marked with `f`. Favorites are stored locally and survive restarts.

## Configuration

radiogo follows the [XDG Base Directory spec](https://specifications.freedesktop.org/basedir-spec/latest/):

| File | Purpose |
|------|---------|
| `$XDG_CONFIG_HOME/radiogo/favorites.json` | Saved favorites |
| `~/.config/mpv/radio.m3u` | Compat M3U playlist (written on every favorites save) |

`$XDG_CONFIG_HOME` defaults to `~/.config` when unset.

## License

See [LICENSE](LICENSE) if present; otherwise ask the author.
