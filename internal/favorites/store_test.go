package favorites_test

import (
	"testing"

	"github.com/rjbudzynski/radiogo/internal/favorites"
	"github.com/rjbudzynski/radiogo/internal/radio"
)

var (
	stationA = radio.Station{UUID: "uuid-a", Name: "Station A", URL: "http://a.example/stream"}
	stationB = radio.Station{UUID: "uuid-b", Name: "Station B", URL: "http://b.example/stream"}
	noUUID   = radio.Station{UUID: "", Name: "No UUID", URL: "http://nouuid.example/stream"}
)

func TestToggle_AddThenRemove(t *testing.T) {
	var favs []radio.Station

	favs = favorites.Toggle(favs, stationA)
	if len(favs) != 1 {
		t.Fatalf("want 1 favorite after add, got %d", len(favs))
	}

	favs = favorites.Toggle(favs, stationA)
	if len(favs) != 0 {
		t.Fatalf("want 0 favorites after remove, got %d", len(favs))
	}
}

func TestToggle_MultipleStations(t *testing.T) {
	var favs []radio.Station
	favs = favorites.Toggle(favs, stationA)
	favs = favorites.Toggle(favs, stationB)
	if len(favs) != 2 {
		t.Fatalf("want 2 favorites, got %d", len(favs))
	}

	// Removing A leaves B.
	favs = favorites.Toggle(favs, stationA)
	if len(favs) != 1 || favs[0].UUID != stationB.UUID {
		t.Fatalf("want only stationB remaining, got %v", favs)
	}
}

func TestToggle_FallbackToURL(t *testing.T) {
	var favs []radio.Station
	favs = favorites.Toggle(favs, noUUID)
	if len(favs) != 1 {
		t.Fatal("want 1 favorite after add by URL")
	}

	// Same URL, empty UUID — should be removed.
	dup := radio.Station{UUID: "", URL: noUUID.URL}
	favs = favorites.Toggle(favs, dup)
	if len(favs) != 0 {
		t.Fatalf("want 0 favorites after URL-match remove, got %d", len(favs))
	}
}

func TestContains(t *testing.T) {
	favs := []radio.Station{stationA}

	if !favorites.Contains(favs, stationA) {
		t.Error("Contains: expected true for stationA")
	}
	if favorites.Contains(favs, stationB) {
		t.Error("Contains: expected false for stationB")
	}
}

func TestContains_FallbackToURL(t *testing.T) {
	favs := []radio.Station{noUUID}
	match := radio.Station{UUID: "", URL: noUUID.URL}
	if !favorites.Contains(favs, match) {
		t.Error("Contains: expected URL-based match")
	}
}

func TestToggle_SymmetricURLMatch_StoredHasUUID(t *testing.T) {
	// Stored station has UUID, incoming has same URL but empty UUID.
	stored := radio.Station{UUID: "abc", Name: "Stored", URL: "http://example/stream"}
	incoming := radio.Station{UUID: "", Name: "Incoming", URL: "http://example/stream"}

	favs := []radio.Station{stored}
	favs = favorites.Toggle(favs, incoming)
	if len(favs) != 0 {
		t.Fatalf("want 0 favorites after toggle-remove (URL match against UUID entry), got %d", len(favs))
	}
}

func TestToggle_SymmetricURLMatch_IncomingHasUUID(t *testing.T) {
	// Stored station has empty UUID, incoming has UUID and same URL.
	// The old code already handled this case, but we test it explicitly
	// for confidence in the refactored sameStation helper.
	stored := radio.Station{UUID: "", Name: "Stored", URL: "http://example/stream"}
	incoming := radio.Station{UUID: "abc", Name: "Incoming", URL: "http://example/stream"}

	favs := []radio.Station{stored}
	favs = favorites.Toggle(favs, incoming)
	if len(favs) != 0 {
		t.Fatalf("want 0 favorites after toggle-remove (UUID match against URL entry), got %d", len(favs))
	}
}

func TestContains_SymmetricURLMatch_StoredHasUUID(t *testing.T) {
	stored := radio.Station{UUID: "abc", URL: "http://example/stream"}
	incoming := radio.Station{UUID: "", URL: "http://example/stream"}
	if !favorites.Contains([]radio.Station{stored}, incoming) {
		t.Error("Contains: expected match when stored has UUID and incoming has empty UUID")
	}
}

func TestContains_SymmetricURLMatch_IncomingHasUUID(t *testing.T) {
	stored := radio.Station{UUID: "", URL: "http://example/stream"}
	incoming := radio.Station{UUID: "abc", URL: "http://example/stream"}
	if !favorites.Contains([]radio.Station{stored}, incoming) {
		t.Error("Contains: expected match when stored has empty UUID and incoming has UUID")
	}
}

func TestSameStation_DifferentURLsNotMatched(t *testing.T) {
	a := radio.Station{UUID: "", URL: "http://a.example/stream"}
	b := radio.Station{UUID: "", URL: "http://b.example/stream"}
	// Both have empty UUIDs and different URLs — not a match.
	if favorites.Contains([]radio.Station{a}, b) {
		t.Error("Contains: should not match stations with different URLs when UUIDs are empty")
	}
}
