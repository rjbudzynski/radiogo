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
