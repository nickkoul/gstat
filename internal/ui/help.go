package ui

import "fmt"

const (
	normalHelpPanelLines = 5
	searchHelpPanelLines = 4
)

// HelpPanelLineCount returns the number of rendered content lines in the help panel.
func HelpPanelLineCount(searchMode bool) int {
	if searchMode {
		return searchHelpPanelLines
	}
	return normalHelpPanelLines
}

// RenderHelpPanel renders the expanded on-screen hotkey help.
func RenderHelpPanel(width int, searchMode bool, roundMode string, favoritesOnly bool) string {
	s := DefaultStyles()

	var lines []string
	if searchMode {
		lines = []string{
			s.HelpTitle.Render(" Hotkeys"),
			fmt.Sprintf("  %s %s  %s %s", s.StatusKey.Render("type"), s.StatusDim.Render("filter players"), s.StatusKey.Render("backspace"), s.StatusDim.Render("delete")),
			fmt.Sprintf("  %s %s  %s %s", s.StatusKey.Render("enter"), s.StatusDim.Render("keep filter"), s.StatusKey.Render("esc"), s.StatusDim.Render("clear + exit")),
			fmt.Sprintf("  %s %s", s.StatusKey.Render("^c"), s.StatusDim.Render("quit")),
		}
	} else {
		favoriteViewLabel := "favorites only"
		if favoritesOnly {
			favoriteViewLabel = "all players"
		}
		lines = []string{
			s.HelpTitle.Render(" Hotkeys"),
			fmt.Sprintf("  %s %s  %s %s  %s %s", s.StatusKey.Render("j/k"), s.StatusDim.Render("move"), s.StatusKey.Render("^d/^u"), s.StatusDim.Render("half page"), s.StatusKey.Render("g/G"), s.StatusDim.Render("top/bottom")),
			fmt.Sprintf("  %s %s  %s %s  %s %s", s.StatusKey.Render("f"), s.StatusDim.Render("favorite"), s.StatusKey.Render("F"), s.StatusDim.Render(favoriteViewLabel), s.StatusKey.Render("/"), s.StatusDim.Render("search")),
			fmt.Sprintf("  %s %s  %s %s", s.StatusKey.Render("t"), s.StatusDim.Render(fmt.Sprintf("rounds (%s)", roundMode)), s.StatusKey.Render("?"), s.StatusDim.Render("toggle help")),
			fmt.Sprintf("  %s %s", s.StatusKey.Render("q"), s.StatusDim.Render("quit")),
		}
	}

	return s.HelpPanel.Width(width).Render(joinLines(lines))
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	joined := lines[0]
	for _, line := range lines[1:] {
		joined += "\n" + line
	}
	return joined
}
