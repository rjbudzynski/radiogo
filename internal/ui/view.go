package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/woodz-dot/radiogo/internal/favorites"
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

// renderTabBar renders the [ Browse ] [ Favorites ] tabs.
func (m Model) renderTabBar(width int) string {
	tabs := []string{"Browse", "Favorites"}
	var parts []string
	for i, t := range tabs {
		if i == m.activeTab {
			parts = append(parts, styleTabActive.Render("[ "+t+" ]"))
		} else {
			parts = append(parts, styleTabInactive.Render("[ "+t+" ]"))
		}
	}
	bar := strings.Join(parts, " ")
	// right-align app name
	appName := styleTabInactive.Render("radiogo")
	gap := width - lipgloss.Width(bar) - lipgloss.Width(appName)
	if gap > 0 {
		bar += strings.Repeat(" ", gap) + appName
	}
	return bar
}

// renderListPane renders the station list on the left.
func (m Model) renderListPane(width, height int) string {
	innerW := width - 2 // border
	innerH := height

	var sb strings.Builder

	// Loading / error states.
	if m.loading {
		sb.WriteString("\n  " + m.spinner.View() + " Loading stations…")
		return stylePane.Width(width).Height(height).Render(sb.String())
	}
	if m.err != nil {
		return stylePane.Width(width).Height(height).Render(
			"\n  " + styleErr.Render("Error: "+m.err.Error()),
		)
	}

	// Search prompt (browse tab only).
	searchLine := ""
	if m.activeTab == tabBrowse {
		if m.searching {
			searchLine = styleSearchPrompt.Render("/") + " " + m.searchInput.View()
		} else if m.searchQuery != "" {
			searchLine = styleHelp.Render("search: " + m.searchQuery + "  (/ to re-search)")
		} else {
			searchLine = styleHelp.Render("top stations  (/ to search)")
		}
	}

	list := m.activeList()
	idx := m.activeIndex()
	maxRows := innerH - 2 // leave room for search line + border
	if searchLine != "" {
		maxRows--
	}

	// Simple scrolling window centred on cursor.
	start := 0
	if idx > maxRows-1 {
		start = idx - maxRows + 1
	}
	end := start + maxRows
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
			prefix := styleItemPlaying.Render("▶ ")
			favMark := favMarker(isFav, innerW)
			line = styleItemSelected.Render(prefix+label) + favMark
		case i == idx:
			prefix := "  "
			favMark := favMarker(isFav, innerW)
			line = styleItemSelected.Render(prefix+label) + favMark
		case isPlaying:
			prefix := styleItemPlaying.Render("▶ ")
			line = styleItemPlaying.Render(prefix + label)
		default:
			prefix := "  "
			line = styleItemNormal.Render(prefix + label)
			if isFav {
				line += styleItemFav.Render(" ★")
			}
		}
		sb.WriteString(line + "\n")
	}

	content := sb.String()
	if searchLine != "" {
		content = searchLine + "\n" + content
	}

	paneStyle := stylePane
	if !m.searching {
		paneStyle = stylePaneFocused
	}
	return paneStyle.Width(width).Height(height).Render(content)
}

func favMarker(isFav bool, _ int) string {
	if isFav {
		return styleItemFav.Render(" ★")
	}
	return ""
}

// renderInfoPane renders the right-hand station detail + now playing panel.
func (m Model) renderInfoPane(width, height int) string {
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
