package ui

import (
	"fmt"
	"testing"

	"github.com/rjbudzynski/radiogo/internal/appstate"
	"github.com/rjbudzynski/radiogo/internal/radio"
)

func baseModel() Model {
	m := New(nil, nil)
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

// --- listHeaderLines ---

func TestListHeaderLines_Browse(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseErr = nil
	if got := m.listHeaderLines(); got != 1 {
		t.Errorf("want 1 (search line only), got %d", got)
	}
}

func TestListHeaderLines_BrowseWithError(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseErr = fmt.Errorf("oops")
	m.browseStations = stationsOf("a") // stations present → not a fallback
	if got := m.listHeaderLines(); got != 2 {
		t.Errorf("want 2 (error + search line), got %d", got)
	}
}

func TestListHeaderLines_Favorites(t *testing.T) {
	m := baseModel()
	m.activeTab = tabFavorites
	if got := m.listHeaderLines(); got != 0 {
		t.Errorf("want 0 for favorites tab, got %d", got)
	}
}

// --- listHitTest ---

func TestListHitTest_BasicClick(t *testing.T) {
	m := baseModel()
	m.height = 40
	m.loading = false
	m.activeTab = tabBrowse
	m.browseStations = stationsOf("a", "b", "c", "d", "e")
	m.browseIndex = 0

	// listStartY = tabBar(1) + top border(1) + headerLines(1) = 3
	if got := m.listHitTest(3); got != 0 {
		t.Errorf("Y=3 should hit station 0, got %d", got)
	}
	if got := m.listHitTest(4); got != 1 {
		t.Errorf("Y=4 should hit station 1, got %d", got)
	}
}

func TestListHitTest_OutOfBounds(t *testing.T) {
	m := baseModel()
	m.height = 40
	m.loading = false
	m.activeTab = tabBrowse
	m.browseStations = stationsOf("a")
	m.browseIndex = 0

	if got := m.listHitTest(0); got != -1 {
		t.Errorf("Y=0 (tab bar) should miss, got %d", got)
	}
	if got := m.listHitTest(1); got != -1 {
		t.Errorf("Y=1 (pane border) should miss, got %d", got)
	}
}

func TestListHitTest_HelpTab(t *testing.T) {
	m := baseModel()
	m.loading = false
	m.activeTab = tabHelp
	if got := m.listHitTest(3); got != -1 {
		t.Errorf("help tab should always return -1, got %d", got)
	}
}

func TestListHitTest_FavoritesTab(t *testing.T) {
	m := baseModel()
	m.height = 40
	m.loading = false
	m.activeTab = tabFavorites
	m.favorites = stationsOf("fav0", "fav1")
	m.favIndex = 0

	// No header lines for favorites: listStartY = 2
	if got := m.listHitTest(2); got != 0 {
		t.Errorf("Y=2 on favorites should hit fav0, got %d", got)
	}
	if got := m.listHitTest(3); got != 1 {
		t.Errorf("Y=3 on favorites should hit fav1, got %d", got)
	}
}

func TestNew_AppliesRestoredState(t *testing.T) {
	restored := &appstate.State{
		Volume:      35,
		ActiveTab:   tabFavorites,
		SearchQuery: "jazz",
		BrowseSort:  "bitrate_desc",
		SelectedStation: appstate.StationRef{
			UUID: "fav2",
			URL:  "http://fav2/stream",
		},
	}

	m := New(stationsOf("fav1", "fav2"), restored)

	if m.volume != 35 {
		t.Fatalf("volume = %d, want 35", m.volume)
	}
	if m.activeTab != tabFavorites {
		t.Fatalf("activeTab = %d, want favorites", m.activeTab)
	}
	if m.searchQuery != "jazz" {
		t.Fatalf("searchQuery = %q, want jazz", m.searchQuery)
	}
	if m.browseSort != browseSortBitrateDesc {
		t.Fatalf("browseSort = %v, want bitrate desc", m.browseSort)
	}
	if m.favIndex != 1 {
		t.Fatalf("favIndex = %d, want 1", m.favIndex)
	}
}

func TestBrowseSortCycle(t *testing.T) {
	if got := cycleBrowseSort(browseSortVotesDesc); got != browseSortBitrateDesc {
		t.Fatalf("cycle from votes = %v, want bitrate", got)
	}
	if got := cycleBrowseSort(browseSortBitrateDesc); got != browseSortNameAsc {
		t.Fatalf("cycle from bitrate = %v, want name", got)
	}
	if got := cycleBrowseSort(browseSortNameAsc); got != browseSortVotesDesc {
		t.Fatalf("cycle from name = %v, want votes", got)
	}
}

func TestCurrentFilterType_Codecs(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseMode = browseModeCodecs
	if got := m.currentFilterType(); got != "codec" {
		t.Fatalf("currentFilterType() = %q, want codec", got)
	}
}

func TestCategoryMenuIncludesCodecs(t *testing.T) {
	menu := categoryMenuCategories()
	if len(menu) != 4 {
		t.Fatalf("menu length = %d, want 4", len(menu))
	}
	if menu[3].Name != "Codecs" {
		t.Fatalf("menu[3] = %q, want Codecs", menu[3].Name)
	}
}

func TestBrowseSearchOptions_CombinesNameAndCategory(t *testing.T) {
	m := baseModel()
	m.browseSort = browseSortBitrateDesc
	m.browseFilterType = "codec"
	m.browseFilterValue = "ogg"

	opts := m.browseSearchOptions("jazz")
	if opts.Name != "jazz" {
		t.Fatalf("Name = %q, want jazz", opts.Name)
	}
	if opts.Codec != "ogg" {
		t.Fatalf("Codec = %q, want ogg", opts.Codec)
	}
	if opts.Order != "bitrate" || !opts.Reverse {
		t.Fatalf("Order/Reverse = %q/%v, want bitrate/true", opts.Order, opts.Reverse)
	}
}

func TestRestoreBrowseSelection(t *testing.T) {
	m := baseModel()
	m.restoredSelection = appstate.StationRef{
		UUID: "b",
		URL:  "http://b/stream",
	}
	m.browseStations = stationsOf("a", "b", "c")

	m.restoreBrowseSelection()

	if m.browseIndex != 1 {
		t.Fatalf("browseIndex = %d, want 1", m.browseIndex)
	}
	if !m.restoredSelection.IsZero() {
		t.Fatalf("restoredSelection = %#v, want cleared after restore", m.restoredSelection)
	}
}
