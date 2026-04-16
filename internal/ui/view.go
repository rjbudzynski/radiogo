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
		} else {
			// Breadcrumbs / Mode context
			switch m.browseMode {
			case browseModeTop:
				if m.searchQuery != "" {
					searchLine = styleHelp.Render("search: " + m.searchQuery + "  (/ to edit, Backspace to clear)")
				} else {
					searchLine = styleHelp.Render("top stations  (/ search, c categories)")
				}
			case browseModeCategories:
				searchLine = styleHelp.Render("Browse > Categories")
			case browseModeTags:
				searchLine = styleHelp.Render("Browse > Tags")
			case browseModeCountries:
				searchLine = styleHelp.Render("Browse > Countries")
			case browseModeLanguages:
				searchLine = styleHelp.Render("Browse > Languages")
			case browseModeCodecs:
				searchLine = styleHelp.Render("Browse > Codecs")
			case browseModeResults:
				base := fmt.Sprintf("Browse > %s > %s", prettyFilterType(m.browseFilterType), m.browseFilterValue)
				if m.searchQuery != "" {
					searchLine = styleHelp.Render(base + "  · search: " + m.searchQuery)
				} else {
					searchLine = styleHelp.Render(base)
				}
			}
		}
	}

	if m.activeTab == tabBrowse && !m.isCategoryMode() {
		sortLine := styleHelp.Render("sort: " + m.browseSort.String())
		if searchLine != "" {
			searchLine += "  " + sortLine
		} else {
			searchLine = sortLine
		}
	}

	if m.isCategoryMode() {
		return m.renderCategoryList(width, height, innerW, searchLine)
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

func (m Model) renderCategoryList(width, height, innerW int, searchLine string) string {
	var sb strings.Builder
	if searchLine != "" {
		sb.WriteString(searchLine + "\n")
	}

	idx := m.browseIndex
	headerLines := strings.Count(sb.String(), "\n")
	innerH := height - headerLines - 2
	if innerH < 1 {
		innerH = 1
	}

	start := 0
	if idx > innerH-1 {
		start = idx - innerH + 1
	}
	end := start + innerH
	if end > len(m.browseCategories) {
		end = len(m.browseCategories)
	}

	for i := start; i < end; i++ {
		c := m.browseCategories[i]
		label := truncate(c.Name, innerW-10)
		if c.Count > 0 {
			label = fmt.Sprintf("%- *s %d", innerW-10, label, c.Count)
		}

		line := styleItemNormal.Render("  " + label)
		if i == idx {
			line = styleItemSelected.Render("  " + label)
		}
		sb.WriteString(line + "\n")
	}

	return stylePaneFocused.Width(width).Height(height).Render(sb.String())
}

func favMarker(isFav bool) string {
	if isFav {
		return styleItemFav.Render(" ★")
	}
	return ""
}

// renderHelpContent returns the text shown in the list pane on the Help tab.
func renderHelpContent(width int) string {
	// Clamp to a sensible minimum so sub-expressions don't go negative.
	if width < 20 {
		width = 20
	}
	var sb strings.Builder

	section := func(title string) {
		sb.WriteString("\n" + styleSectionTitle.Render(title) + "\n")
	}
	keyW := 12
	row := func(key, desc string) {
		sb.WriteString(renderHelpRow(key, desc, keyW, width))
	}

	sb.WriteString(styleSectionTitle.Render("radiogo — keyboard shortcuts") + "\n")

	section("Navigation")
	row("↑ / k", "move up")
	row("↓ / j", "move down")
	row("Tab", "cycle tabs: Browse → Favorites → Help")
	row("Backspace", "back (browse categories)")
	row("Esc", "cancel search edit")

	section("Playback")
	row("Enter", "play selected station")
	row("p", "pause / resume")
	row("+ / =", "volume up  (+5%)")
	row("-", "volume down  (−5%)")

	section("Stations")
	row("/", "search by name in the current browse context")
	row("c", "browse categories: tags, countries, languages, codecs")
	row("s", "cycle sort: votes, bitrate, name")
	row("f", "toggle favorite on selected")

	section("General")
	row("q / Ctrl+C", "quit")

	divider := strings.Repeat("─", width)
	sb.WriteString("\n" + styleDivider.Render(divider) + "\n")

	notes := "Stations sourced from radio-browser.info\n" +
		"Favorites: ~/.config/radiogo/favorites.json\n" +
		"If the directory is unreachable, favorites are shown in Browse as a fallback.\n" +
		"Backspace clears the name search first, then backs out of categories.\n" +
		"Browse sort preference is saved automatically."
	sb.WriteString(styleHelp.Render(renderHelpParagraph(notes, width, 2)) + "\n")

	return sb.String()
}

// renderInfoPane renders the right-hand station detail + now playing panel.
func (m Model) renderInfoPane(width, height int) string {
	if m.activeTab == tabHelp {
		return stylePane.Width(width).Height(height).Render(
			"\n" + styleHelp.Render("  Press Tab to return to Browse or Favorites."),
		)
	}

	if m.isCategoryMode() {
		return m.renderCategoryInfo(width, height)
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

func (m Model) renderCategoryInfo(width, height int) string {
	idx := m.browseIndex
	if idx < 0 || idx >= len(m.browseCategories) {
		return stylePane.Width(width).Height(height).Render("\n  Select a category.")
	}
	cat := m.browseCategories[idx]

	var sb strings.Builder
	sb.WriteString(styleSectionTitle.Render(cat.Name) + "\n\n")
	if cat.Count > 0 {
		sb.WriteString(infoRow("Stations", fmt.Sprintf("%d", cat.Count)))
	}

	switch m.browseMode {
	case browseModeCategories:
		sb.WriteString("\n  Press Enter to browse by " + cat.Name + ".")
	default:
		sb.WriteString("\n  Press Enter to see stations for this " + m.currentFilterType() + ".")
	}

	sb.WriteString("\n\n  Press Backspace to go back.")

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
	if m.stateErr != nil {
		msg := truncate("⚠  could not save state: "+m.stateErr.Error(), width-2)
		content := styleErr.Render(msg)
		pad := width - lipgloss.Width(content)
		if pad > 0 {
			content += strings.Repeat(" ", pad)
		}
		return styleStatusBar.Width(width).Render(content)
	}
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
		hints = append(hints, keys.Category.Help().Key+" cat")
		if !m.isCategoryMode() {
			hints = append(hints, keys.Sort.Help().Key+" sort")
		}
		if len(m.browseHistory) > 0 {
			hints = append(hints, keys.Back.Help().Key+" back")
		}
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

func prettyFilterType(s string) string {
	switch s {
	case "tag":
		return "Tags"
	case "country":
		return "Countries"
	case "language":
		return "Languages"
	case "codec":
		return "Codecs"
	default:
		return s
	}
}

func renderHelpRow(key, desc string, keyW, width int) string {
	descW := width - keyW
	if descW < 1 {
		descW = 1
	}

	lines := wrapText(desc, descW)
	if len(lines) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styleInfoLabel.Width(keyW).Render(key))
	sb.WriteString(styleInfoValue.Width(descW).Render(lines[0]))
	sb.WriteString("\n")

	pad := strings.Repeat(" ", keyW)
	for _, line := range lines[1:] {
		sb.WriteString(pad)
		sb.WriteString(styleInfoValue.Width(descW).Render(line))
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderHelpParagraph(text string, width, indent int) string {
	if width < 1 {
		width = 1
	}
	if indent < 0 {
		indent = 0
	}

	var sb strings.Builder
	paragraphs := strings.Split(text, "\n")
	for i, para := range paragraphs {
		if i > 0 {
			sb.WriteString("\n")
		}
		lines := wrapText(para, width-indent)
		if len(lines) == 0 {
			sb.WriteString(strings.Repeat(" ", indent))
			continue
		}
		for j, line := range lines {
			if j > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(strings.Repeat(" ", indent))
			sb.WriteString(line)
		}
	}
	return sb.String()
}

func wrapText(text string, width int) []string {
	if width < 1 {
		width = 1
	}

	var lines []string
	for _, para := range strings.Split(text, "\n") {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		current := words[0]
		currentLen := len([]rune(current))
		for currentLen > width {
			runes := []rune(current)
			lines = append(lines, string(runes[:width]))
			current = string(runes[width:])
			currentLen = len([]rune(current))
		}
		for _, word := range words[1:] {
			wordLen := len([]rune(word))
			if current == "" {
				current = word
				currentLen = wordLen
				for currentLen > width {
					runes := []rune(current)
					lines = append(lines, string(runes[:width]))
					current = string(runes[width:])
					currentLen = len([]rune(current))
				}
				continue
			}
			if currentLen+1+wordLen <= width {
				current += " " + word
				currentLen += 1 + wordLen
				continue
			}
			lines = append(lines, current)
			current = word
			currentLen = wordLen
			for currentLen > width {
				runes := []rune(current)
				lines = append(lines, string(runes[:width]))
				current = string(runes[width:])
				currentLen = len([]rune(current))
			}
		}
		if current != "" {
			lines = append(lines, current)
		}
	}

	return lines
}
