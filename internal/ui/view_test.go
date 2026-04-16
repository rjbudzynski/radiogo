package ui

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate with ellipsis", "hello world", 8, "hello w…"},
		{"max=1", "hello", 1, "…"},
		{"max=0", "hello", 0, "…"},
		{"empty string", "", 5, ""},
		{"unicode runes", "héllo", 4, "hél…"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := truncate(tc.s, tc.max)
			if got != tc.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tc.s, tc.max, got, tc.want)
			}
		})
	}
}

func TestTabAtX(t *testing.T) {
	tests := []struct {
		x    int
		want int
	}{
		{0, 0},   // start of Browse
		{6, 0},   // middle of Browse
		{11, 0},  // end of Browse
		{12, -1}, // separator
		{13, 1},  // start of Favorites
		{20, 1},  // middle of Favorites
		{27, 1},  // end of Favorites
		{28, -1}, // separator
		{29, 2},  // start of Help
		{34, 2},  // middle of Help
		{38, 2},  // end of Help
		{39, -1}, // past all tabs
	}
	for _, tc := range tests {
		got := tabAtX(tc.x)
		if got != tc.want {
			t.Errorf("tabAtX(%d) = %d, want %d", tc.x, got, tc.want)
		}
	}
}

func TestTruncate_NoEllipsisWhenFits(t *testing.T) {
	got := truncate("abc", 3)
	if strings.Contains(got, "…") {
		t.Errorf("expected no ellipsis for exact-fit, got %q", got)
	}
}

func TestWrapText(t *testing.T) {
	got := wrapText("alpha beta gamma", 10)
	want := []string{"alpha beta", "gamma"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRenderHelpParagraph_IndentsWrappedLines(t *testing.T) {
	got := renderHelpParagraph("alpha beta gamma delta", 12, 2)
	if !strings.Contains(got, "\n  gamma") {
		t.Fatalf("wrapped paragraph missing indented continuation line: %q", got)
	}
}
