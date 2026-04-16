package radio

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/rjbudzynski/radiogo/internal/config"
)

// MetaMsg carries an icy-title update from the player goroutine.
type MetaMsg struct{ Title string }

// PauseStateMsg carries a pause-state update from the player goroutine.
type PauseStateMsg struct{ Paused bool }

// PlayerStoppedMsg signals mpv has exited.
type PlayerStoppedMsg struct{ Err error }

// Player manages a single mpv subprocess via its JSON IPC socket.
type Player struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	conn       net.Conn
	reqID      int
	stopped    bool
	generation int // incremented on each Play; guards against stale goroutine callbacks
}

// MetaCallback is called from a background goroutine when icy-title changes.
type MetaCallback func(MetaMsg)

// PauseCallback is called from a background goroutine when pause state changes.
type PauseCallback func(PauseStateMsg)

// StopCallback is called when mpv exits.
type StopCallback func(PlayerStoppedMsg)

// Play stops any current stream and starts playing url.
func (p *Player) Play(station Station, onMeta MetaCallback, onPause PauseCallback, onStop StopCallback) error {
	p.Stop()

	p.mu.Lock()
	p.stopped = false
	p.generation++
	gen := p.generation
	p.mu.Unlock()

	socketPath := config.MPVSocketPath()

	// Remove stale socket.
	_ = os.Remove(socketPath)

	cmd := exec.Command("mpv",
		"--no-video",
		"--really-quiet",
		"--idle=yes",
		"--input-ipc-server="+socketPath,
		station.URL,
	)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("mpv start: %w", err)
	}

	p.mu.Lock()
	p.cmd = cmd
	p.mu.Unlock()

	// Wait for socket to appear (up to 3 seconds).
	go func() {
		var conn net.Conn
		for i := 0; i < 30; i++ {
			time.Sleep(100 * time.Millisecond)
			var c net.Conn
			var err error
			if runtime.GOOS == "windows" {
				c, err = net.Dial("tcp", socketPath)
			} else {
				c, err = net.Dial("unix", socketPath)
			}
			if err == nil {
				conn = c
				break
			}
		}

		p.mu.Lock()
		p.conn = conn
		p.mu.Unlock()

		if conn != nil {
			go p.readEvents(conn, onMeta, onPause)
			go p.observeProperties(conn)
		}

		// Wait for mpv to exit.
		err := cmd.Wait()
		p.mu.Lock()
		stale := p.stopped || p.generation != gen
		p.mu.Unlock()
		if !stale && onStop != nil {
			onStop(PlayerStoppedMsg{Err: err})
		}
	}()

	return nil
}

// observeProperties registers the properties the UI mirrors via mpv IPC.
func (p *Player) observeProperties(conn net.Conn) {
	p.sendCommand(conn, "observe_property", 1, "metadata/by-key/icy-title")
	p.sendCommand(conn, "observe_property", 2, "pause")
}

// sendCommand sends a JSON IPC command on conn.
// Returns false if the write failed (broken socket).
func (p *Player) sendCommand(conn net.Conn, cmd string, args ...any) bool {
	p.mu.Lock()
	p.reqID++
	id := p.reqID
	p.mu.Unlock()

	req := map[string]any{"command": append([]any{cmd}, args...), "request_id": id}
	data, _ := json.Marshal(req)
	data = append(data, '\n')
	_, err := conn.Write(data)
	if err != nil {
		p.mu.Lock()
		if p.conn == conn {
			p.conn = nil
		}
		p.mu.Unlock()
		conn.Close()
		return false
	}
	return true
}

// readEvents reads mpv JSON events from the socket and fires callbacks.
func (p *Player) readEvents(conn net.Conn, onMeta MetaCallback, onPause PauseCallback) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var ev map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}
		dispatchEvent(ev, onMeta, onPause)
	}
}

func dispatchEvent(ev map[string]any, onMeta MetaCallback, onPause PauseCallback) {
	if ev["event"] != "property-change" {
		return
	}

	name, ok := ev["name"].(string)
	if !ok {
		return
	}

	switch name {
	case "metadata/by-key/icy-title":
		title, ok := ev["data"].(string)
		if ok && onMeta != nil {
			onMeta(MetaMsg{Title: title})
		}
	case "pause":
		paused, ok := ev["data"].(bool)
		if ok && onPause != nil {
			onPause(PauseStateMsg{Paused: paused})
		}
	}
}

// Pause toggles pause via mpv IPC.
func (p *Player) Pause() {
	p.mu.Lock()
	conn := p.conn
	p.mu.Unlock()
	if conn != nil {
		p.sendCommand(conn, "cycle", "pause")
	}
}

// SetVolume sets mpv volume (0–100).
func (p *Player) SetVolume(vol int) {
	p.mu.Lock()
	conn := p.conn
	p.mu.Unlock()
	if conn != nil {
		p.sendCommand(conn, "set_property", "volume", vol)
	}
}

// Stop kills the current mpv process.
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopped = true
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		p.cmd = nil
	}
}

// IsRunning reports whether mpv is currently active.
func (p *Player) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cmd != nil && !p.stopped
}
