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

func TestTruncate_NoEllipsisWhenFits(t *testing.T) {
	got := truncate("abc", 3)
	if strings.Contains(got, "…") {
		t.Errorf("expected no ellipsis for exact-fit, got %q", got)
	}
}
