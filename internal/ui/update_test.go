package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rjbudzynski/radiogo/internal/radio"
)

func TestUpdate_PauseStateMsgSyncsPausedFlag(t *testing.T) {
	m := baseModel()

	updated, cmd := m.Update(pauseStateMsg{paused: true})
	if cmd != nil {
		t.Fatal("expected no command for pause state update")
	}

	got := updated.(Model)
	if !got.paused {
		t.Fatal("paused = false, want true")
	}

	updated, _ = got.Update(pauseStateMsg{paused: false})
	got = updated.(Model)
	if got.paused {
		t.Fatal("paused = true, want false")
	}
}

func TestHandleKey_SortsBrowseList(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseMode = browseModeTop
	m.browseSort = browseSortVotesDesc

	updated, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if cmd == nil {
		t.Fatal("expected command when cycling sort")
	}

	got := updated.(Model)
	if got.browseSort != browseSortBitrateDesc {
		t.Fatalf("browseSort = %v, want bitrate desc", got.browseSort)
	}
}

func TestHandleKey_BackspaceRestoresCategoryMenu(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.browseMode = browseModeTags
	m.browseHistory = []BrowseMode{browseModeCategories}
	m.browseCategories = []radio.Category{{Name: "Jazz"}}

	updated, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace, Alt: false})
	if cmd != nil {
		t.Fatal("expected no command on backspace navigation")
	}

	got := updated.(Model)
	if got.browseMode != browseModeCategories {
		t.Fatalf("browseMode = %v, want categories", got.browseMode)
	}
	if len(got.browseCategories) != 4 || got.browseCategories[3].Name != "Codecs" {
		t.Fatalf("browseCategories = %#v, want restored category menu", got.browseCategories)
	}
}

func TestHandleKey_BackspaceClearsNameSearchFirst(t *testing.T) {
	m := baseModel()
	m.activeTab = tabBrowse
	m.searchQuery = "jazz"
	m.browseMode = browseModeResults
	m.browseFilterType = "codec"
	m.browseFilterValue = "ogg"
	m.loading = false

	updated, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace, Alt: false})
	if cmd == nil {
		t.Fatal("expected command when clearing name search")
	}

	got := updated.(Model)
	if got.searchQuery != "" {
		t.Fatalf("searchQuery = %q, want empty", got.searchQuery)
	}
	if got.browseFilterType != "codec" || got.browseFilterValue != "ogg" {
		t.Fatalf("category filter changed unexpectedly: %#v", got)
	}
}

func TestStationsLoadedMsg_ResetsSelectionForSearchResults(t *testing.T) {
	m := baseModel()
	m.searchGen = 3
	m.searchQuery = "jazz"
	m.browseIndex = 2
	m.browseStations = stationsOf("old-a", "old-b", "old-c")

	updated, cmd := m.Update(stationsLoadedMsg{
		stations: stationsOf("new-a", "new-b"),
		gen:      3,
	})
	if cmd == nil {
		t.Fatal("expected persistence command after loading stations")
	}

	got := updated.(Model)
	if got.browseIndex != 0 {
		t.Fatalf("browseIndex = %d, want 0 for search results", got.browseIndex)
	}
}

func TestStationsLoadedMsg_ResetsSelectionForCategoryResults(t *testing.T) {
	m := baseModel()
	m.searchGen = 4
	m.browseMode = browseModeResults
	m.browseFilterType = "codec"
	m.browseFilterValue = "ogg"
	m.browseIndex = 2
	m.browseStations = stationsOf("old-a", "old-b", "old-c")

	updated, _ := m.Update(stationsLoadedMsg{
		stations: stationsOf("new-a", "new-b"),
		gen:      4,
	})

	got := updated.(Model)
	if got.browseIndex != 0 {
		t.Fatalf("browseIndex = %d, want 0 for category results", got.browseIndex)
	}
}
