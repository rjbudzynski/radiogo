package ui

import (
	"testing"

	"github.com/rjbudzynski/radiogo/internal/radio"
)

func baseModel() Model {
	m := New(nil)
	m.width = 120
	m.height = 40
	return m
}

func stationsOf(names ...string) []radio.Station {
	out := make([]radio.Station, len(names))
	for i, n := range names {
		out[i] = radio.Station{UUID: n, Name: n, URL: "http://" + n + "/stream"}
	}
	return out
}

// --- activeList ---

func TestActiveList_Browse(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseStations = stationsOf("a", "b")
	list := m.activeList()
	if len(list) != 2 {
		t.Fatalf("want 2 browse stations, got %d", len(list))
	}
}

func TestActiveList_BrowseFallback(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseStations = nil
	m.favorites = stationsOf("fav1")
	list := m.activeList()
	if len(list) != 1 || list[0].UUID != "fav1" {
		t.Fatalf("expected favorites fallback, got %v", list)
	}
}

func TestActiveList_Favorites(t *testing.T) {
	m := baseModel()
	m.activeTab = tabFavorites
	m.favorites = stationsOf("fav1", "fav2")
	list := m.activeList()
	if len(list) != 2 {
		t.Fatalf("want 2 favorites, got %d", len(list))
	}
}

func TestActiveList_Help(t *testing.T) {
	m := baseModel()
	m.activeTab = tabHelp
	list := m.activeList()
	if list != nil {
		t.Fatalf("want nil for help tab, got %v", list)
	}
}

// --- browseIsFallback ---

func TestBrowseIsFallback_True(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseStations = nil
	m.favorites = stationsOf("x")
	if !m.browseIsFallback() {
		t.Error("expected browseIsFallback() = true")
	}
}

func TestBrowseIsFallback_FalseWhenStationsExist(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseStations = stationsOf("a")
	if m.browseIsFallback() {
		t.Error("expected browseIsFallback() = false when browse stations exist")
	}
}

func TestBrowseIsFallback_FalseOnFavTab(t *testing.T) {
	m := baseModel()
	m.activeTab = tabFavorites
	m.favorites = stationsOf("f")
	if m.browseIsFallback() {
		t.Error("expected browseIsFallback() = false on favorites tab")
	}
}

// --- selectedStation ---

func TestSelectedStation(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseStations = stationsOf("a", "b", "c")
	m.browseIndex = 1
	s := m.selectedStation()
	if s == nil || s.UUID != "b" {
		t.Fatalf("want station 'b', got %v", s)
	}
}

func TestSelectedStation_OutOfBounds(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseStations = stationsOf("a")
	m.browseIndex = 5
	if s := m.selectedStation(); s != nil {
		t.Fatalf("want nil for out-of-bounds index, got %v", s)
	}
}

func TestSelectedStation_EmptyList(t *testing.T) {
	m := baseModel()
	m.activeTab = tabFavorites
	m.favorites = nil
	if s := m.selectedStation(); s != nil {
		t.Fatalf("want nil for empty list, got %v", s)
	}
}
