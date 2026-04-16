package radio

import (
	"net/url"
	"testing"
)

func TestBuildSearchPath(t *testing.T) {
	path := buildSearchPath(SearchOptions{
		Name:       "lofi jazz",
		Country:    "Japan",
		Codec:      "mp3",
		Order:      "bitrate",
		Reverse:    true,
		Limit:      25,
		BitrateMin: 128,
	})

	u, err := url.Parse(path)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := u.Path; got != "/stations/search" {
		t.Fatalf("path = %q, want /stations/search", got)
	}

	q := u.Query()
	if got := q.Get("name"); got != "lofi jazz" {
		t.Fatalf("name = %q, want lofi jazz", got)
	}
	if got := q.Get("country"); got != "Japan" {
		t.Fatalf("country = %q, want Japan", got)
	}
	if got := q.Get("codec"); got != "mp3" {
		t.Fatalf("codec = %q, want mp3", got)
	}
	if got := q.Get("order"); got != "bitrate" {
		t.Fatalf("order = %q, want bitrate", got)
	}
	if got := q.Get("reverse"); got != "true" {
		t.Fatalf("reverse = %q, want true", got)
	}
	if got := q.Get("limit"); got != "25" {
		t.Fatalf("limit = %q, want 25", got)
	}
	if got := q.Get("bitrateMin"); got != "128" {
		t.Fatalf("bitrateMin = %q, want 128", got)
	}
	if got := q.Get("hidebroken"); got != "true" {
		t.Fatalf("hidebroken = %q, want true", got)
	}
}
