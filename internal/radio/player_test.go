package radio

import "testing"

func TestDispatchEvent_MetaUpdate(t *testing.T) {
	var got MetaMsg
	dispatchEvent(
		map[string]any{
			"event": "property-change",
			"name":  "metadata/by-key/icy-title",
			"data":  "Artist - Track",
		},
		func(msg MetaMsg) { got = msg },
		nil,
	)

	if got.Title != "Artist - Track" {
		t.Fatalf("title = %q, want Artist - Track", got.Title)
	}
}

func TestDispatchEvent_PauseUpdate(t *testing.T) {
	var got PauseStateMsg
	dispatchEvent(
		map[string]any{
			"event": "property-change",
			"name":  "pause",
			"data":  true,
		},
		nil,
		func(msg PauseStateMsg) { got = msg },
	)

	if !got.Paused {
		t.Fatal("paused = false, want true")
	}
}

func TestDispatchEvent_IgnoresUnexpectedPayload(t *testing.T) {
	called := false
	dispatchEvent(
		map[string]any{
			"event": "property-change",
			"name":  "pause",
			"data":  "true",
		},
		nil,
		func(PauseStateMsg) { called = true },
	)

	if called {
		t.Fatal("pause callback was called for invalid payload")
	}
}
