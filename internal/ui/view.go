package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/rjbudzynski/radiogo/internal/favorites"
)

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	totalHeight := m.height
	statusBarHeight := 2 // status line + help line
	tabBarHeight := 1
	innerHeight := totalHeight - statusBarHeight - tabBarHeight - 2 // border top+bottom

	// Pane widths (left ~40%, right ~60%).
	leftWidth := m.width * 40 / 100
	rightWidth := m.width - leftWidth - 1 // -1 for separator gap

	tabBar := m.renderTabBar(m.width)
	leftPane := m.renderListPane(leftWidth, innerHeight)
	rightPane := m.renderInfoPane(rightWidth, innerHeight)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	statusBar := m.renderStatusBar(m.width)
	helpBar := m.renderHelpBar(m.width)

	return lipgloss.JoinVertical(lipgloss.Left,
		tabBar,
		mainRow,
		statusBar,
		helpBar,
	)
}

// tabAtX returns the tab index (0=Browse, 1=Favorites, 2=Help) for a given
// X coordinate in the tab bar row, or -1 if x doesn't hit any tab.
func tabAtX(x int) int {
	labels := []string{"Browse", "Favorites", "Help"}
	cursor := 0
	for i, l := range labels {
		w := lipgloss.Width(styleTabInactive.Render("[ " + l + " ]"))
		if x >= cursor && x < cursor+w {
			return i
		}
		cursor += w + 1 // +1 for the " " separator between tabs
	}
	return -1
}


func (m Model) renderTabBar(width int) string {
	tabs := []string{"Browse", "Favorites", "Help"}
	var parts []string
	for i, t := range tabs {
		if i == m.activeTab {
			parts = append(parts, styleTabActive.Render("[ "+t+" ]"))
		} else {
			parts = append(parts, styleTabInactive.Render("[ "+t+" ]"))
		}
	}
	bar := strings.Join(parts, " ")
	appName := styleTabInactive.Render("radiogo")
	gap := width - lipgloss.Width(bar) - lipgloss.Width(appName)
	if gap > 0 {
		bar += strings.Repeat(" ", gap) + appName
	}
	return bar
}

// renderListPane renders the station list (or help content) on the left.
func (m Model) renderListPane(width, height int) string {
	innerW := width - 2 // border

	// Help tab — static content.
	if m.activeTab == tabHelp {
		return stylePane.Width(width).Height(height).Render(renderHelpContent(innerW))
	}

	var sb strings.Builder

	// Loading spinner.
	if m.loading {
		sb.WriteString("\n  " + m.spinner.View() + " Loading stations…")
		return stylePane.Width(width).Height(height).Render(sb.String())
	}

	// Inline error banner — shown at top but does not hide the list.
	if m.browseErr != nil && m.activeTab == tabBrowse {
		errText := truncate("⚠  "+m.browseErr.Error(), innerW-2)
		sb.WriteString(styleErr.Render(errText) + "\n")
		if m.browseIsFallback() {
			sb.WriteString(styleHelp.Render("  directory unavailable — showing favorites") + "\n")
		}
	}

	// Search prompt (browse tab only).
	searchLine := ""
	if m.activeTab == tabBrowse {
		if m.searching {
			searchLine = styleSearchPrompt.Render("/") + " " + m.searchInput.View()
		} else if m.searchQuery != "" {
			searchLine = styleHelp.Render("search: " + m.searchQuery + "  (/ to re-search)")
		} else if !m.browseIsFallback() {
			searchLine = styleHelp.Render("top stations  (/ to search)")
		}
	}

	list := m.activeList()
	if len(list) == 0 {
		msg := "No stations."
		if m.activeTab == tabFavorites {
			msg = "No favorites yet.  Press f on any station to add one."
		}
		if searchLine != "" {
			sb.WriteString(searchLine + "\n")
		}
		content := sb.String() + "\n  " + styleHelp.Render(msg)
		return stylePaneFocused.Width(width).Height(height).Render(content)
	}

	// Write the search/context line before the station list.
	if searchLine != "" {
		sb.WriteString(searchLine + "\n")
	}

	idx := m.activeIndex()
	// Count already-written header lines to size the scroll window.
	headerLines := strings.Count(sb.String(), "\n")
	innerH := height - headerLines - 2 // borders
	if innerH < 1 {
		innerH = 1
	}

	// Simple scrolling window centred on cursor.
	start := 0
	if idx > innerH-1 {
		start = idx - innerH + 1
	}
	end := start + innerH
	if end > len(list) {
		end = len(list)
	}

	for i := start; i < end; i++ {
		s := list[i]
		label := truncate(s.Name, innerW-3)
		isFav := favorites.Contains(m.favorites, s)
		isPlaying := m.nowPlaying != nil && (m.nowPlaying.UUID == s.UUID || m.nowPlaying.URL == s.URL)

		var line string
		switch {
		case i == idx && isPlaying:
			line = styleItemSelected.Render(styleItemPlaying.Render("▶ ")+label) + favMarker(isFav)
		case i == idx:
			line = styleItemSelected.Render("  "+label) + favMarker(isFav)
		case isPlaying:
			line = styleItemPlaying.Render("▶ " + label)
		default:
			line = styleItemNormal.Render("  " + label)
			if isFav {
				line += styleItemFav.Render(" ★")
			}
		}
		sb.WriteString(line + "\n")
	}

	paneStyle := stylePane
	if !m.searching {
		paneStyle = stylePaneFocused
	}
	return paneStyle.Width(width).Height(height).Render(sb.String())
}

func favMarker(isFav bool) string {
	if isFav {
		return styleItemFav.Render(" ★")
	}
	return ""
}

// renderHelpContent returns the text shown in the list pane on the Help tab.
func renderHelpContent(width int) string {
	_ = width
	var sb strings.Builder

	section := func(title string) {
		sb.WriteString("\n" + styleSectionTitle.Render(title) + "\n")
	}
	row := func(key, desc string) {
		sb.WriteString(styleInfoLabel.Width(12).Render(key) + styleInfoValue.Render(desc) + "\n")
	}

	sb.WriteString(styleSectionTitle.Render("radiogo — keyboard shortcuts") + "\n")

	section("Navigation")
	row("↑ / k", "move up")
	row("↓ / j", "move down")
	row("Tab", "cycle tabs: Browse → Favorites → Help")
	row("Esc", "cancel search")

	section("Playback")
	row("Enter", "play selected station")
	row("p", "pause / resume")
	row("+ / =", "volume up  (+5%)")
	row("-", "volume down  (−5%)")

	section("Stations")
	row("/", "search by name  (Browse tab)")
	row("f", "toggle favorite on selected")

	section("General")
	row("q / Ctrl+C", "quit")

	sb.WriteString("\n" + styleDivider.Render("─────────────────────────────────") + "\n")
	sb.WriteString(styleHelp.Render("Stations sourced from radio-browser.info\n"))
	sb.WriteString(styleHelp.Render("Favorites: ~/.config/radiogo/favorites.json\n"))
	sb.WriteString(styleHelp.Render("If the directory is unreachable, favorites\n"))
	sb.WriteString(styleHelp.Render("are shown in Browse as a fallback.\n"))

	return sb.String()
}

// renderInfoPane renders the right-hand station detail + now playing panel.
func (m Model) renderInfoPane(width, height int) string {
	if m.activeTab == tabHelp {
		return stylePane.Width(width).Height(height).Render(
			"\n" + styleHelp.Render("  Press Tab to return to Browse or Favorites."),
		)
	}

	var sb strings.Builder

	sel := m.selectedStation()
	if sel == nil {
		return stylePane.Width(width).Height(height).Render("\n  No station selected.")
	}

	sb.WriteString(styleSectionTitle.Render(sel.Name) + "\n")
	sb.WriteString(infoRow("Country", sel.Country))
	sb.WriteString(infoRow("Language", sel.Language))
	if sel.Tags != "" {
		tags := truncate(sel.Tags, width-14)
		sb.WriteString(infoRow("Tags", tags))
	}
	if sel.Codec != "" {
		codec := sel.Codec
		if sel.Bitrate > 0 {
			codec += fmt.Sprintf("  %d kbps", sel.Bitrate)
		}
		sb.WriteString(infoRow("Codec", codec))
	}
	if sel.Votes > 0 {
		sb.WriteString(infoRow("Votes", fmt.Sprintf("%d", sel.Votes)))
	}
	if sel.Homepage != "" {
		sb.WriteString(infoRow("Home", truncate(sel.Homepage, width-14)))
	}
	if sel.URL != "" {
		sb.WriteString(infoRow("Stream", truncate(sel.URL, width-14)))
	}

	isFav := favorites.Contains(m.favorites, *sel)
	favStr := "no  (press f to add)"
	if isFav {
		favStr = styleItemFav.Render("★ yes  (press f to remove)")
	}
	sb.WriteString(infoRow("Favorite", favStr))

	// Now playing section.
	sb.WriteString("\n" + styleDivider.Render(strings.Repeat("─", width-4)) + "\n")

	if m.nowPlaying != nil {
		isThisStation := m.nowPlaying.UUID == sel.UUID || m.nowPlaying.URL == sel.URL
		if isThisStation {
			status := "▶ Playing"
			if m.paused {
				status = "⏸ Paused"
			}
			sb.WriteString(styleNowPlaying.Render(status) + "\n")
			if m.trackTitle != "" {
				sb.WriteString(styleNowPlaying.Render("♪ "+m.trackTitle) + "\n")
			} else {
				sb.WriteString(styleHelp.Render("  (no track metadata)") + "\n")
			}
		} else {
			sb.WriteString(styleHelp.Render("Now playing: "+truncate(m.nowPlaying.Name, width-16)) + "\n")
			if m.trackTitle != "" {
				sb.WriteString(styleHelp.Render("♪ "+m.trackTitle) + "\n")
			}
		}
	} else {
		sb.WriteString(styleHelp.Render("Not playing") + "\n")
	}

	return stylePane.Width(width).Height(height).Render(sb.String())
}

func infoRow(label, value string) string {
	if value == "" {
		return ""
	}
	return styleInfoLabel.Render(label+":") + " " + styleInfoValue.Render(value) + "\n"
}

// renderStatusBar renders the bottom now-playing status line.
func (m Model) renderStatusBar(width int) string {
	// Prioritise error messages over playback status.
	if m.saveErr != nil {
		msg := truncate("⚠  could not save favorites: "+m.saveErr.Error(), width-2)
		content := styleErr.Render(msg)
		pad := width - lipgloss.Width(content)
		if pad > 0 {
			content += strings.Repeat(" ", pad)
		}
		return styleStatusBar.Width(width).Render(content)
	}

	var parts []string

	if m.nowPlaying != nil {
		icon := "▶"
		if m.paused {
			icon = "⏸"
		}
		parts = append(parts, styleStatusPlaying.Render(icon+" "+m.nowPlaying.Name))
		if m.trackTitle != "" {
			parts = append(parts, styleStatusTrack.Render("  ♪ "+m.trackTitle))
		}
		parts = append(parts, styleHelp.Render(fmt.Sprintf("  vol:%d%%", m.volume)))
	} else {
		parts = append(parts, styleHelp.Render("No station playing"))
	}

	content := strings.Join(parts, "")
	pad := width - lipgloss.Width(content)
	if pad > 0 {
		content += strings.Repeat(" ", pad)
	}
	return styleStatusBar.Width(width).Render(content)
}

// renderHelpBar renders the one-line key-hint footer.
func (m Model) renderHelpBar(_ int) string {
	if m.activeTab == tabHelp {
		return styleHelp.Render("  Tab  switch tab    q  quit")
	}
	hints := []string{
		keys.Up.Help().Key + " " + keys.Down.Help().Key + " navigate",
		keys.Enter.Help().Key + " play",
		keys.Fav.Help().Key + " fav",
	}
	if m.activeTab == tabBrowse {
		hints = append(hints, keys.Search.Help().Key+" search")
	}
	hints = append(hints,
		keys.Tab.Help().Key+" tab",
		keys.Pause.Help().Key+" pause",
		keys.VolUp.Help().Key+keys.VolDown.Help().Key+" vol",
		keys.Quit.Help().Key+" quit",
	)
	return styleHelp.Render("  " + strings.Join(hints, "  "))
}

// truncate shortens s to max runes, appending "…" if cut.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}
